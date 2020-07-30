/*
Copyright 2019 The KubeCarrier Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package integration

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"

	catalogv1alpha1 "k8c.io/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "k8c.io/kubecarrier/pkg/apis/core/v1alpha1"
	apiserverv1 "k8c.io/kubecarrier/pkg/apiserver/api/v1"
	"k8c.io/kubecarrier/pkg/testutil"

	kubermatictestutil "github.com/kubermatic/utils/pkg/testutil"
)

const (
	localAPIServerPort = 9443
	// Htpasswd
	username = "user1"
	password = "mickey5"
)

func newAPIServer(f *testutil.Framework) func(t *testing.T) {
	return func(t *testing.T) {
		t.Log("testing how API Server works")
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		managementClient, err := f.ManagementClient(t)
		require.NoError(t, err, "creating management client")
		t.Cleanup(managementClient.CleanUpFunc(ctx))

		ns := &corev1.Namespace{}
		ns.Name = "kubecarrier-system"

		pfCmd := exec.CommandContext(ctx,
			"kubectl",
			"--kubeconfig", f.Config().ManagementExternalKubeconfigPath,
			"--namespace", ns.GetName(),
			"port-forward",
			// well known service name since it's assumed only one API server shall be deployed
			"service/kubecarrier-api-server-manager",
			fmt.Sprintf("%d:https", localAPIServerPort),
		)
		pfCmd.Stdout = &kubermatictestutil.TestingLogWriter{T: t}
		pfCmd.Stderr = &kubermatictestutil.TestingLogWriter{T: t}
		require.NoError(t, pfCmd.Start())

		token := fetchUserToken(ctx, t, managementClient, f.Config().ManagementExternalKubeconfigPath)
		t.Log("token", token)
		servingTLSSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "apiserver-tls-cert",
				Namespace: ns.GetName(),
			},
		}

		require.NoError(t, managementClient.WaitUntil(ctx, servingTLSSecret, func() (done bool, err error) {
			data, ok := servingTLSSecret.Data[corev1.TLSCertKey]
			return ok && len(data) > 0, nil
		}))

		certPool := x509.NewCertPool()
		require.True(t, certPool.AppendCertsFromPEM(servingTLSSecret.Data["ca.crt"]))
		require.True(t, certPool.AppendCertsFromPEM(servingTLSSecret.Data[corev1.TLSCertKey]))

		conn, err := newGRPCConn(ctx, certPool, gRPCWithAuthToken{token: token})
		require.NoError(t, err)
		testRunningAPIServer(t, ctx, conn)
		t.Run("auth-modes", func(t *testing.T) {
			authModes(ctx, certPool, managementClient, f)(t)
		})

		docClient := apiserverv1.NewDocClient(conn)
		_, err = docClient.Swagger(ctx, &apiserverv1.DocStaticRequest{Path: "/"})
		if assert.NoError(t, err, "docs gRPC") {
			t.Log("Docs")
		}

		// Create an account to test authorization
		accountName := "kubecarrier-admin"
		account := testutil.NewTenantAccount(accountName, rbacv1.Subject{
			Kind:     rbacv1.UserKind,
			APIGroup: "rbac.authorization.k8s.io",
			Name:     "admin@kubecarrier.io",
		})
		require.NoError(t, managementClient.Create(ctx, account))
		require.NoError(t, kubermatictestutil.WaitUntilReady(ctx, managementClient, account), "account not ready")

		for name, testFn := range map[string]func(ctx context.Context, conn *grpc.ClientConn, account *catalogv1alpha1.Account, f *testutil.Framework) func(t *testing.T){
			"account-service":  accountService,
			"offering-service": offeringService,
			"region-service":   regionService,
			"provider-service": providerService,
			"instance-service": instanceService,
		} {
			name := name
			testFn := testFn

			t.Run(name, func(t *testing.T) {
				t.Parallel()
				testFn(ctx, conn, account, f)(t)
			})
		}
	}
}

func testRunningAPIServer(t *testing.T, ctx context.Context, conn *grpc.ClientConn) {
	client := apiserverv1.NewKubeCarrierClient(conn)
	versionCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	t.Cleanup(cancel)
	require.NoError(t, wait.PollUntil(time.Second, func() (done bool, err error) {
		version, err := client.Version(versionCtx, &apiserverv1.VersionRequest{})
		if err == nil {
			assert.NotEmpty(t, version.Version)
			assert.NotEmpty(t, version.Branch)
			assert.NotEmpty(t, version.BuildDate)
			assert.NotEmpty(t, version.GoVersion)
			t.Log("got response for gRPC server version")
			return true, nil
		}
		if grpcStatus, ok := status.FromError(err); ok {
			if grpcStatus.Code() == codes.Unavailable {
				t.Log("gRPC server temporary unavailable, retrying")
				return false, nil
			}
			t.Logf("gRPC server errored out, retrying : %d %v %v",
				grpcStatus.Code(),
				grpcStatus.Message(),
				grpcStatus.Err(),
			)
			return false, nil
		}
		return false, err
	}, versionCtx.Done()), "client version gRPC call")
}

func newGRPCConn(ctx context.Context, certPool *x509.CertPool, creds credentials.PerRPCCredentials) (*grpc.ClientConn, error) {
	return grpc.DialContext(
		ctx,
		fmt.Sprintf("localhost:%d", localAPIServerPort),
		grpc.WithTransportCredentials(
			credentials.NewClientTLSFromCert(certPool, ""),
		),
		grpc.WithPerRPCCredentials(creds),
	)
}

func authModes(ctx context.Context, certPool *x509.CertPool, managementClient *kubermatictestutil.RecordingClient, f *testutil.Framework) func(t *testing.T) {
	return func(t *testing.T) {
		fetchUserInfo := func(creds credentials.PerRPCCredentials) *apiserverv1.UserInfo {
			t.Helper()
			conn, err := newGRPCConn(ctx, certPool, creds)
			require.NoError(t, err)
			client := apiserverv1.NewKubeCarrierClient(conn)
			userInfo, err := client.WhoAmI(ctx, &empty.Empty{})
			require.NoError(t, err)
			t.Log("userInfo")
			kubermatictestutil.LogObject(t, userInfo)
			return userInfo
		}

		t.Run("OIDC", func(t *testing.T) {
			t.Parallel()
			token := fetchUserToken(ctx, t, managementClient, f.Config().ManagementExternalKubeconfigPath)
			userInfo := fetchUserInfo(gRPCWithAuthToken{token: token})
			assert.Equal(t, "admin@kubecarrier.io", userInfo.User)
		})

		t.Run("Token", func(t *testing.T) {
			t.Parallel()
			sa := &corev1.ServiceAccount{}
			require.NoError(t, managementClient.Get(ctx, types.NamespacedName{Namespace: "default", Name: "default"}, sa))
			secret := &corev1.Secret{}
			require.NoError(t, managementClient.Get(ctx, types.NamespacedName{Namespace: "default", Name: sa.Secrets[0].Name}, secret))
			token := string(secret.Data["token"])
			userInfo := fetchUserInfo(gRPCWithAuthToken{token: token})
			assert.Equal(t, "system:serviceaccount:default:default", userInfo.User)
		})

		t.Run("htpasswd", func(t *testing.T) {
			t.Parallel()
			userInfo := fetchUserInfo(gRPCBasicAuthToken{username: username, password: password})
			assert.Equal(t, "user1", userInfo.User)
		})
	}
}

func fetchUserToken(ctx context.Context, t *testing.T, managementClient *kubermatictestutil.RecordingClient, kubeconfig string) string {
	const localDexServerPort = 10443
	pfDex := exec.CommandContext(ctx,
		"kubectl",
		"--kubeconfig", kubeconfig,
		"--namespace", "kubecarrier-system",
		"port-forward",
		"service/dex",
		fmt.Sprintf("%d:https", localDexServerPort),
	)

	pfDex.Stdout = &kubermatictestutil.TestingLogWriter{T: t}
	pfDex.Stderr = &kubermatictestutil.TestingLogWriter{T: t}

	require.NoError(t, pfDex.Start())
	certPool := x509.NewCertPool()
	dexTLSSecret := &corev1.Secret{}
	require.NoError(t, managementClient.Get(ctx, types.NamespacedName{Name: "dex-web-server", Namespace: "kubecarrier-system"}, dexTLSSecret))
	require.True(t, certPool.AppendCertsFromPEM(dexTLSSecret.Data["ca.crt"]))
	require.True(t, certPool.AppendCertsFromPEM(dexTLSSecret.Data[corev1.TLSCertKey]))
	token, err := kubermatictestutil.DexFakeClientCredentialsGrant(
		ctx,
		kubermatictestutil.NewLogger(t),
		&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs:    certPool,
					ServerName: "dex.kubecarrier-system.svc",
				},
			},
			Timeout: 5 * time.Second,
		},
		fmt.Sprintf("https://localhost:%d/auth", localDexServerPort),
		"admin@example.com",
		"password",
	)
	require.NoError(t, err, "getting token from internal dex instance")
	return token
}

type gRPCWithAuthToken struct {
	token string
}

var _ credentials.PerRPCCredentials = gRPCWithAuthToken{}

func (w gRPCWithAuthToken) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"Authorization": "Bearer " + w.token,
	}, nil
}

func (w gRPCWithAuthToken) RequireTransportSecurity() bool {
	return true
}

type gRPCBasicAuthToken struct {
	username string
	password string
}

var _ credentials.PerRPCCredentials = gRPCBasicAuthToken{}

func (w gRPCBasicAuthToken) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	token := base64.StdEncoding.EncodeToString([]byte(w.username + ":" + w.password))
	return map[string]string{
		"Authorization": "Basic " + token,
	}, nil
}

func (w gRPCBasicAuthToken) RequireTransportSecurity() bool {
	return true
}

func instanceService(ctx context.Context, conn *grpc.ClientConn, tenantAccount *catalogv1alpha1.Account, f *testutil.Framework) func(t *testing.T) {
	return func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		serviceClient, err := f.ServiceClient(t)
		require.NoError(t, err, "creating service client")
		t.Cleanup(serviceClient.CleanUpFunc(ctx))

		managementClient, err := f.ManagementClient(t)
		require.NoError(t, err, "creating management client")
		t.Cleanup(managementClient.CleanUpFunc(ctx))

		// we hit length limit of 63 chars, so we need a shorter name
		testName := "instsvc"

		providerAccount := testutil.NewProviderAccount(testName, rbacv1.Subject{
			Kind:     rbacv1.GroupKind,
			APIGroup: "rbac.authorization.k8s.io",
			Name:     "providerAccount",
		})
		require.NoError(t, managementClient.Create(ctx, providerAccount), "creating providerAccount")
		require.NoError(t, kubermatictestutil.WaitUntilReady(ctx, managementClient, providerAccount))

		require.NoError(t, kubermatictestutil.WaitUntilReady(ctx, managementClient, tenantAccount))

		serviceCluster := f.SetupServiceCluster(ctx, managementClient, t, "eu-west-1", providerAccount)

		catalogEntrySet := &catalogv1alpha1.CatalogEntrySet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testName,
				Namespace: providerAccount.Status.Namespace.Name,
			},
			Spec: catalogv1alpha1.CatalogEntrySetSpec{
				Metadata: catalogv1alpha1.CatalogEntrySetMetadata{
					CommonMetadata: catalogv1alpha1.CommonMetadata{
						DisplayName:      "FakeDB",
						Description:      "small database living near Tegel airport",
						ShortDescription: "some short description",
					},
				},
				Derive: &catalogv1alpha1.DerivedConfig{
					KindOverride: "DB",
					Expose: []catalogv1alpha1.VersionExposeConfig{
						{
							Versions: []string{
								"v1",
							},
							Fields: []catalogv1alpha1.FieldPath{
								{JSONPath: ".spec.databaseName"},
								{JSONPath: ".spec.databaseUser"},
								{JSONPath: ".spec.config.create"},
								{JSONPath: ".status.observedGeneration"},
							},
						},
					},
				},
				Discover: catalogv1alpha1.CustomResourceDiscoverySetConfig{
					CRD: catalogv1alpha1.ObjectReference{
						Name: "dbs.fake.kubecarrier.io",
					},
					ServiceClusterSelector: metav1.LabelSelector{},
					KindOverride:           "DBInternal",
					WebhookStrategy:        corev1alpha1.WebhookStrategyTypeServiceCluster,
				},
			},
		}
		require.NoError(t, managementClient.Create(ctx, catalogEntrySet))
		require.NoError(t, kubermatictestutil.WaitUntilReady(ctx, managementClient, catalogEntrySet, kubermatictestutil.WithTimeout(time.Minute*2)))

		catalog := testutil.NewCatalog("test-catalog", providerAccount.Status.Namespace.Name, &metav1.LabelSelector{}, &metav1.LabelSelector{})
		require.NoError(t, managementClient.Create(ctx, catalog), "creating Catalog error")

		// Check the status of the Catalog.
		assert.NoError(t, managementClient.WaitUntil(ctx, catalog, func() (b bool, err error) {
			return len(catalog.Status.Entries) == 1 && len(catalog.Status.Tenants) > 0, nil
		}))

		offering := &catalogv1alpha1.Offering{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: tenantAccount.Status.Namespace.Name,
				Name:      strings.Join([]string{"dbs", serviceCluster.Name, providerAccount.Name}, "."),
			},
		}

		require.NoError(t, kubermatictestutil.WaitUntilFound(ctx, managementClient, offering))
		require.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Namespace: tenantAccount.Status.Namespace.Name,
			Name:      strings.Join([]string{"dbs", serviceCluster.Name, providerAccount.Name}, "."),
		}, offering), "getting Offering error")

		serviceClusterAssignment := &corev1alpha1.ServiceClusterAssignment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tenantAccount.Status.Namespace.Name + "." + serviceCluster.Name,
				Namespace: providerAccount.Status.Namespace.Name,
			},
		}
		require.NoError(t, kubermatictestutil.WaitUntilReady(ctx, managementClient, serviceClusterAssignment), "service cluster assignment not ready")

		// TODO: replace someday, wait until the admin@kubecarrier.io user gets the required create permissions
		time.Sleep(5 * time.Second)
		client := apiserverv1.NewInstancesServiceClient(conn)
		// watch instances
		watchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		t.Cleanup(cancel)
		watchClient, err := client.Watch(watchCtx, &apiserverv1.InstanceWatchRequest{
			Offering: offering.Name,
			Version:  "v1",
			Account:  tenantAccount.Status.Namespace.Name,
		})
		require.NoError(t, err)
		createReq := &apiserverv1.InstanceCreateRequest{
			Offering: offering.Name,
			Version:  "v1",
			Spec: &apiserverv1.Instance{
				Metadata: &apiserverv1.ObjectMeta{Name: "fakedb"},
				Spec:     apiserverv1.NewJSONRawObject([]byte("{\"databaseName\":\"coolDB\",\"databaseUser\":\"username\"}")),
			},
			Account: tenantAccount.Status.Namespace.Name,
		}
		_, err = client.Create(ctx, createReq)
		require.NoError(t, err, "creating instance")
		nextEventType(t, watchClient, watch.Added)

		fakeDB := testutil.NewFakeDB("fakedb", serviceClusterAssignment.Status.ServiceClusterNamespace.Name)
		require.NoError(t, kubermatictestutil.WaitUntilFound(ctx, serviceClient, fakeDB))

		getReq := &apiserverv1.InstanceGetRequest{
			Offering: offering.Name,
			Version:  "v1",
			Name:     "fakedb",
			Account:  tenantAccount.Status.Namespace.Name,
		}
		_, err = client.Get(ctx, getReq)
		require.NoError(t, err, "getting instance")
		listReq := &apiserverv1.InstanceListRequest{
			Offering: offering.Name,
			Version:  "v1",
			Account:  tenantAccount.Status.Namespace.Name,
		}
		instances, err := client.List(ctx, listReq)
		require.NoError(t, err, "listing instance")
		assert.Len(t, instances.Items, 1)

		delReq := &apiserverv1.InstanceDeleteRequest{
			Offering: offering.Name,
			Version:  "v1",
			Name:     "fakedb",
			Account:  tenantAccount.Status.Namespace.Name,
		}
		_, err = client.Delete(ctx, delReq)
		require.NoError(t, err, "deleting instance")
		require.NoError(t, kubermatictestutil.WaitUntilNotFound(ctx, serviceClient, fakeDB))
		nextEventType(t, watchClient, watch.Deleted)
	}
}

func accountService(ctx context.Context, conn *grpc.ClientConn, account *catalogv1alpha1.Account, f *testutil.Framework) func(t *testing.T) {
	return func(t *testing.T) {
		managementClient, err := f.ManagementClient(t)
		require.NoError(t, err, "creating management client")
		t.Cleanup(managementClient.CleanUpFunc(ctx))

		testName := strings.Replace(strings.ToLower(t.Name()), "/", "-", -1)
		providerAccount := testutil.NewProviderAccount(testName, rbacv1.Subject{
			Kind:     rbacv1.GroupKind,
			APIGroup: "rbac.authorization.k8s.io",
			Name:     "admin@kubecarrier.io",
		})
		tenantAccount := testutil.NewTenantAccount(testName, rbacv1.Subject{
			Kind:     rbacv1.GroupKind,
			APIGroup: "rbac.authorization.k8s.io",
			Name:     "user",
		})
		require.NoError(t, managementClient.Create(ctx, providerAccount))
		require.NoError(t, managementClient.Create(ctx, tenantAccount))

		client := apiserverv1.NewAccountServiceClient(conn)
		accountCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		t.Cleanup(cancel)
		// list account for "admin@kubecarrier.io".
		require.NoError(t, wait.PollUntil(time.Second, func() (done bool, err error) {
			accounts, err := client.List(accountCtx, &apiserverv1.AccountListRequest{
				LabelSelector: "test-case=integration-apiserver-account-service",
			})
			if err != nil {
				return false, err
			}
			assert.Len(t, accounts.Items, 1)
			assert.True(t, accounts.Items[0].Metadata.Name == providerAccount.Name)
			accounts, err = client.List(accountCtx, &apiserverv1.AccountListRequest{})
			if err != nil {
				return false, err
			}
			// Another account is the one that passed to this test.
			assert.Len(t, accounts.Items, 2)
			return true, nil
		}, accountCtx.Done()))
	}
}

func providerService(ctx context.Context, conn *grpc.ClientConn, account *catalogv1alpha1.Account, f *testutil.Framework) func(t *testing.T) {
	return func(t *testing.T) {
		managementClient, err := f.ManagementClient(t)
		require.NoError(t, err, "creating management client")
		t.Cleanup(managementClient.CleanUpFunc(ctx))

		namespaceName := account.Status.Namespace.Name
		// Create providers objects in the management cluster.
		provider1 := &catalogv1alpha1.Provider{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-provider-1",
				Namespace: namespaceName,
				Labels: map[string]string{
					"test-label": "provider1",
				},
			},
			Spec: catalogv1alpha1.ProviderSpec{
				Metadata: catalogv1alpha1.AccountMetadata{
					CommonMetadata: catalogv1alpha1.CommonMetadata{
						ShortDescription: "Test Provider",
						DisplayName:      "Test Provider",
					},
				},
			},
		}
		provider2 := &catalogv1alpha1.Provider{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-provider-2",
				Namespace: namespaceName,
				Labels: map[string]string{
					"test-label": "provider2",
				},
			},
			Spec: catalogv1alpha1.ProviderSpec{
				Metadata: catalogv1alpha1.AccountMetadata{
					CommonMetadata: catalogv1alpha1.CommonMetadata{
						ShortDescription: "Test Provider",
						DisplayName:      "Test Provider",
					},
				},
			},
		}
		require.NoError(t, managementClient.Create(ctx, provider1))
		require.NoError(t, managementClient.Create(ctx, provider2))

		client := apiserverv1.NewProviderServiceClient(conn)
		providerCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		t.Cleanup(cancel)
		// list providers with limit and continuation token.
		require.NoError(t, wait.PollUntil(time.Second, func() (done bool, err error) {
			providers, err := client.List(providerCtx, &apiserverv1.ListRequest{
				Account: account.Name,
				Limit:   1,
			})
			if err != nil {
				return false, err
			}
			assert.Len(t, providers.Items, 1)
			providers, err = client.List(providerCtx, &apiserverv1.ListRequest{
				Account:  account.Name,
				Limit:    1,
				Continue: providers.Metadata.Continue,
			})
			if err != nil {
				return false, err
			}
			assert.Len(t, providers.Items, 1)
			return true, nil
		}, providerCtx.Done()))

		// get provider
		require.NoError(t, wait.PollUntil(time.Second, func() (done bool, err error) {
			provider, err := client.Get(providerCtx, &apiserverv1.GetRequest{
				Account: account.Name,
				Name:    "test-provider-1",
			})
			if err != nil {
				return false, err
			}
			creationTimestamp, err := ptypes.TimestampProto(provider1.CreationTimestamp.Time)
			if err != nil {
				return true, err
			}
			expectedResult := &apiserverv1.Provider{
				Metadata: &apiserverv1.ObjectMeta{
					Name:    "test-provider-1",
					Account: account.Name,
					Labels: map[string]string{
						"test-label": "provider1",
					},
					Uid:               string(provider1.UID),
					CreationTimestamp: creationTimestamp,
					ResourceVersion:   provider1.ResourceVersion,
					Generation:        provider1.Generation,
				},
				Spec: &apiserverv1.ProviderSpec{
					Metadata: &apiserverv1.ProviderMetadata{
						ShortDescription: "Test Provider",
						DisplayName:      "Test Provider",
					},
				},
			}
			assert.EqualValues(t, provider, expectedResult)
			return true, nil
		}, providerCtx.Done()))

		// watch providers
		watchClient, err := client.Watch(providerCtx, &apiserverv1.WatchRequest{
			Account:       namespaceName,
			LabelSelector: "test-label==provider1",
		})
		require.NoError(t, err)
		nextEventType(t, watchClient, watch.Added)

		// Update an provider object to get Modified event.
		provider1.Spec.Metadata.ShortDescription = "test provider update"
		require.NoError(t, managementClient.Update(ctx, provider1))
		nextEventType(t, watchClient, watch.Modified)

		// Delete an provider object to get Delete event.
		require.NoError(t, managementClient.Delete(ctx, provider1))
		nextEventType(t, watchClient, watch.Deleted)

	}
}

func offeringService(ctx context.Context, conn *grpc.ClientConn, account *catalogv1alpha1.Account, f *testutil.Framework) func(t *testing.T) {
	return func(t *testing.T) {
		managementClient, err := f.ManagementClient(t)
		require.NoError(t, err, "creating management client")
		t.Cleanup(managementClient.CleanUpFunc(ctx))

		namespaceName := account.Status.Namespace.Name
		// Create offering objects in the management cluster.
		offering1 := &catalogv1alpha1.Offering{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-offering-1",
				Namespace: namespaceName,
				Labels: map[string]string{
					"test-label": "offering1",
				},
			},
			Spec: catalogv1alpha1.OfferingSpec{
				Metadata: catalogv1alpha1.OfferingMetadata{
					CommonMetadata: catalogv1alpha1.CommonMetadata{
						ShortDescription: "Test Offering",
						DisplayName:      "Test Offering",
					},
				},
				Provider: catalogv1alpha1.ObjectReference{
					Name: "test-provider",
				},
				CRD: catalogv1alpha1.CRDInformation{
					Name:     "test-crd",
					APIGroup: "test-crd-group",
					Kind:     "test-kind",
					Plural:   "test-plural",
					Versions: []catalogv1alpha1.CRDVersion{
						{
							Name: "test-version",
							Schema: &apiextensionsv1.CustomResourceValidation{
								OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
									Properties: map[string]apiextensionsv1.JSONSchemaProps{
										"apiVersion": {Type: "string"},
									},
									Type: "object",
								},
							},
						},
					},
					Region: catalogv1alpha1.ObjectReference{
						Name: "test-region",
					},
				},
			},
		}
		offering2 := &catalogv1alpha1.Offering{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-offering-2",
				Namespace: namespaceName,
				Labels: map[string]string{
					"test-label": "offering2",
				},
			},
			Spec: catalogv1alpha1.OfferingSpec{
				Metadata: catalogv1alpha1.OfferingMetadata{
					CommonMetadata: catalogv1alpha1.CommonMetadata{
						ShortDescription: "Test Offering",
						DisplayName:      "Test Offering",
					},
				},
				Provider: catalogv1alpha1.ObjectReference{
					Name: "test-provider",
				},
				CRD: catalogv1alpha1.CRDInformation{
					Name:     "test-crd",
					APIGroup: "test-crd-group",
					Kind:     "test-kind",
					Plural:   "test-plural",
					Versions: []catalogv1alpha1.CRDVersion{
						{
							Name: "test-version",
							Schema: &apiextensionsv1.CustomResourceValidation{
								OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
									Properties: map[string]apiextensionsv1.JSONSchemaProps{
										"apiVersion": {Type: "string"},
									},
									Type: "object",
								},
							},
						},
					},
					Region: catalogv1alpha1.ObjectReference{
						Name: "test-region",
					},
				},
			},
		}
		require.NoError(t, managementClient.Create(ctx, offering1))
		require.NoError(t, managementClient.Create(ctx, offering2))

		client := apiserverv1.NewOfferingServiceClient(conn)
		offeringCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
		t.Cleanup(cancel)
		// list offerings with limit and continuation token.
		require.NoError(t, wait.PollUntil(time.Second, func() (done bool, err error) {
			offerings, err := client.List(offeringCtx, &apiserverv1.ListRequest{
				Account: account.Name,
				Limit:   1,
			})
			if err != nil {
				return false, err
			}
			assert.Len(t, offerings.Items, 1)
			offerings, err = client.List(offeringCtx, &apiserverv1.ListRequest{
				Account:  account.Name,
				Limit:    1,
				Continue: offerings.Metadata.Continue,
			})
			if err != nil {
				return false, err
			}
			assert.Len(t, offerings.Items, 1)
			return true, nil
		}, offeringCtx.Done()))

		// get offering
		require.NoError(t, wait.PollUntil(time.Second, func() (done bool, err error) {
			offering, err := client.Get(offeringCtx, &apiserverv1.GetRequest{
				Account: account.Name,
				Name:    "test-offering-1",
			})
			if err != nil {
				return false, err
			}
			creationTimestamp, err := ptypes.TimestampProto(offering1.CreationTimestamp.Time)
			if err != nil {
				return true, err
			}
			expectedResult := &apiserverv1.Offering{
				Metadata: &apiserverv1.ObjectMeta{
					Name:    "test-offering-1",
					Account: namespaceName,
					Labels: map[string]string{
						"test-label": "offering1",
					},
					Uid:               string(offering1.UID),
					CreationTimestamp: creationTimestamp,
					ResourceVersion:   offering1.ResourceVersion,
					Generation:        offering1.Generation,
				},
				Spec: &apiserverv1.OfferingSpec{
					Metadata: &apiserverv1.OfferingMetadata{
						ShortDescription: "Test Offering",
						DisplayName:      "Test Offering",
					},
					Provider: &apiserverv1.ObjectReference{
						Name: "test-provider",
					},
					Crd: &apiserverv1.CRDInformation{
						Name:     "test-crd",
						ApiGroup: "test-crd-group",
						Kind:     "test-kind",
						Plural:   "test-plural",
						Versions: []*apiserverv1.CRDVersion{
							{
								Name:   "test-version",
								Schema: `{"openAPIV3Schema":{"type":"object","properties":{"apiVersion":{"type":"string"}}}}`,
							},
						},
						Region: &apiserverv1.ObjectReference{
							Name: "test-region",
						},
					},
				},
			}
			assert.Equal(t, expectedResult, offering)
			return true, nil
		}, offeringCtx.Done()))

		// watch offerings
		watchClient, err := client.Watch(offeringCtx, &apiserverv1.WatchRequest{
			Account:       namespaceName,
			LabelSelector: "test-label==offering1",
		})
		require.NoError(t, err)
		nextEventType(t, watchClient, watch.Added)

		// Update an offering object to get Modified event.
		offering1.Spec.Metadata.ShortDescription = "test offering update"
		require.NoError(t, managementClient.Update(ctx, offering1))
		nextEventType(t, watchClient, watch.Modified)

		// Delete an offering object to get Delete event.
		require.NoError(t, managementClient.Delete(ctx, offering1))
		nextEventType(t, watchClient, watch.Deleted)
	}
}

func nextEventType(t *testing.T, watchClient apiserverv1.OfferingService_WatchClient, eventType watch.EventType) {
	event, err := watchClient.Recv()
	require.NoError(t, err)
	assert.Equal(t, string(eventType), event.Type)
}

func regionService(ctx context.Context, conn *grpc.ClientConn, account *catalogv1alpha1.Account, f *testutil.Framework) func(t *testing.T) {
	return func(t *testing.T) {
		managementClient, err := f.ManagementClient(t)
		require.NoError(t, err, "creating management client")
		t.Cleanup(managementClient.CleanUpFunc(ctx))

		namespaceName := account.Status.Namespace.Name
		// Create region objects in the management cluster.
		region1 := &catalogv1alpha1.Region{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-region-1",
				Namespace: namespaceName,
				Labels: map[string]string{
					"test-label": "region1",
				},
			},
			Spec: catalogv1alpha1.RegionSpec{
				Metadata: corev1alpha1.ServiceClusterMetadata{
					Description: "Test Region",
					DisplayName: "Test Region",
				},
				Provider: catalogv1alpha1.ObjectReference{
					Name: "test-provider",
				},
			},
		}
		region2 := &catalogv1alpha1.Region{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-region-2",
				Namespace: namespaceName,
				Labels: map[string]string{
					"test-label": "region2",
				},
			},
			Spec: catalogv1alpha1.RegionSpec{
				Metadata: corev1alpha1.ServiceClusterMetadata{
					Description: "Test Region",
					DisplayName: "Test Region",
				},
				Provider: catalogv1alpha1.ObjectReference{
					Name: "test-provider",
				},
			},
		}
		require.NoError(t, managementClient.Create(ctx, region1))
		require.NoError(t, managementClient.Create(ctx, region2))

		client := apiserverv1.NewRegionServiceClient(conn)
		regionCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		t.Cleanup(cancel)
		// list regions with limit and continuation token.
		require.NoError(t, wait.PollUntil(time.Second, func() (done bool, err error) {
			regions, err := client.List(regionCtx, &apiserverv1.ListRequest{
				Account: account.Name,
				Limit:   1,
			})
			if err != nil {
				return false, err
			}
			assert.Len(t, regions.Items, 1)
			regions, err = client.List(regionCtx, &apiserverv1.ListRequest{
				Account:  account.Name,
				Limit:    1,
				Continue: regions.Metadata.Continue,
			})
			if err != nil {
				return false, err
			}
			assert.Len(t, regions.Items, 1)
			return true, nil
		}, regionCtx.Done()))

		// get region
		require.NoError(t, wait.PollUntil(time.Second, func() (done bool, err error) {
			region, err := client.Get(regionCtx, &apiserverv1.GetRequest{
				Account: account.Name,
				Name:    "test-region-1",
			})
			if err != nil {
				return false, err
			}
			creationTimestamp, err := ptypes.TimestampProto(region1.CreationTimestamp.Time)
			if err != nil {
				return true, err
			}
			expectedResult := &apiserverv1.Region{
				Metadata: &apiserverv1.ObjectMeta{
					Name:    "test-region-1",
					Account: namespaceName,
					Labels: map[string]string{
						"test-label": "region1",
					},
					Uid:               string(region1.UID),
					CreationTimestamp: creationTimestamp,
					ResourceVersion:   region1.ResourceVersion,
					Generation:        region1.Generation,
				},
				Spec: &apiserverv1.RegionSpec{
					Metadata: &apiserverv1.RegionMetadata{
						Description: "Test Region",
						DisplayName: "Test Region",
					},
					Provider: &apiserverv1.ObjectReference{
						Name: "test-provider",
					},
				},
			}
			assert.Equal(t, expectedResult, region)
			return true, nil
		}, regionCtx.Done()))

		// watch regions
		watchClient, err := client.Watch(regionCtx, &apiserverv1.WatchRequest{
			Account:       namespaceName,
			LabelSelector: "test-label==region1",
		})
		require.NoError(t, err)
		nextEventType(t, watchClient, watch.Added)

		// Update an region object to get Modified event.
		region1.Spec.Metadata.Description = "test region update"
		require.NoError(t, managementClient.Update(ctx, region1))
		nextEventType(t, watchClient, watch.Modified)

		// Delete an region object to get Delete event.
		require.NoError(t, managementClient.Delete(ctx, region1))
		nextEventType(t, watchClient, watch.Deleted)

	}
}

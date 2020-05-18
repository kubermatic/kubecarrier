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
	"crypto/x509"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	certmanagerv1alpha3 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha3"
	v1 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	apiserverv1 "github.com/kubermatic/kubecarrier/pkg/apiserver/api/v1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func newAPIServer(f *testutil.Framework) func(t *testing.T) {
	return func(t *testing.T) {
		t.Log("testing how API Server works")
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)
		managementClient, err := f.ManagementClient(t)
		require.NoError(t, err, "creating management client")
		t.Cleanup(managementClient.CleanUpFunc(ctx))

		testName := strings.Replace(strings.ToLower(t.Name()), "/", "-", -1)

		ns := &corev1.Namespace{}
		ns.Name = testName
		require.NoError(t, managementClient.Create(ctx, ns))
		const localPort = 9443

		servingTLSSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo-tls",
				Namespace: ns.GetName(),
			},
		}

		issuer := &certmanagerv1alpha3.Issuer{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: ns.GetName(),
			},
			Spec: certmanagerv1alpha3.IssuerSpec{
				IssuerConfig: certmanagerv1alpha3.IssuerConfig{
					SelfSigned: &certmanagerv1alpha3.SelfSignedIssuer{},
				},
			},
		}
		cert := &certmanagerv1alpha3.Certificate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: ns.GetName(),
			},
			Spec: certmanagerv1alpha3.CertificateSpec{
				SecretName: servingTLSSecret.GetName(),
				DNSNames: []string{
					strings.Join([]string{"foo", servingTLSSecret.GetNamespace(), "svc"}, "."),
					"localhost",
				},
				IssuerRef: v1.ObjectReference{
					Name: issuer.GetName(),
				},
			},
		}
		require.NoError(t, managementClient.Create(ctx, issuer))
		require.NoError(t, managementClient.Create(ctx, cert))
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, cert), "cert not ready")
		require.NoError(t, managementClient.WaitUntil(ctx, servingTLSSecret, func() (done bool, err error) {
			data, ok := servingTLSSecret.Data[corev1.TLSCertKey]
			return ok && len(data) > 0, nil
		}))

		apiServer := &operatorv1alpha1.APIServer{ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: ns.GetName(),
		},
			Spec: operatorv1alpha1.APIServerSpec{
				TLSSecretRef: operatorv1alpha1.ObjectReference{
					Name: servingTLSSecret.GetName(),
				},
			},
		}
		require.NoError(t, managementClient.Create(ctx, apiServer))
		assert.NoError(t, testutil.WaitUntilReady(ctx, managementClient, apiServer))

		ctx, cancel = context.WithCancel(ctx)
		t.Cleanup(cancel)

		pfCmd := exec.CommandContext(ctx,
			"kubectl",
			"--kubeconfig", f.Config().ManagementExternalKubeconfigPath,
			"--namespace", apiServer.GetNamespace(),
			"port-forward",
			// well known service name since it's assumed only one API server shall be deployed
			"service/kubecarrier-api-server-manager",
			fmt.Sprintf("%d:https", localPort),
		)
		pfCmd.Stdout = os.Stdout
		pfCmd.Stderr = os.Stderr
		require.NoError(t, pfCmd.Start())

		certPool := x509.NewCertPool()
		require.True(t, certPool.AppendCertsFromPEM(servingTLSSecret.Data["ca.crt"]))
		require.True(t, certPool.AppendCertsFromPEM(servingTLSSecret.Data[corev1.TLSCertKey]))

		conn, err := grpc.DialContext(
			ctx,
			fmt.Sprintf("localhost:%d", localPort),
			grpc.WithTransportCredentials(
				credentials.NewClientTLSFromCert(certPool, ""),
			),
		)
		require.NoError(t, err)
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
				testutil.LogObject(t, version)
				return true, nil
			}
			if grpcStatus, ok := err.(toGRPCStatus); ok {
				if grpcStatus.GRPCStatus().Code() == codes.Unavailable {
					t.Log("gRPC server temporary unavailable, retrying")
					return false, nil
				}
			}
			return false, err
		}, versionCtx.Done()), "client version gRPC call")

		for name, testFn := range map[string]func(ctx context.Context, conn *grpc.ClientConn, managementClient *testutil.RecordingClient) func(t *testing.T){
			"offering-service": offeringService,
			"region-service":   regionService,
			"provider-service": providerService,
		} {
			name := name
			testFn := testFn

			t.Run(name, func(t *testing.T) {
				t.Parallel()
				testFn(ctx, conn, managementClient)(t)
			})
		}
	}
}

func providerService(ctx context.Context, conn *grpc.ClientConn, managementClient *testutil.RecordingClient) func(t *testing.T) {
	return func(t *testing.T) {
		testName := strings.Replace(strings.ToLower(t.Name()), "/", "-", -1)
		ns := &corev1.Namespace{}
		fmt.Printf("creating namespace with name: %v\n", testName)
		ns.Name = testName
		require.NoError(t, managementClient.Create(ctx, ns))
		// Create tenants objects in the management cluster.
		provider1 := &catalogv1alpha1.Provider{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-provider-1",
				Namespace: testName,
				Labels: map[string]string{
					"test-label": "provider1",
				},
			},
			Spec: catalogv1alpha1.ProviderSpec{
				Metadata: catalogv1alpha1.AccountMetadata{
					Description: "Test Provider",
					DisplayName: "Test Provider",
				},
			},
		}
		provider2 := &catalogv1alpha1.Provider{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-provider-2",
				Namespace: testName,
				Labels: map[string]string{
					"test-label": "provider2",
				},
			},
			Spec: catalogv1alpha1.ProviderSpec{
				Metadata: catalogv1alpha1.AccountMetadata{
					Description: "Test Provider",
					DisplayName: "Test Provider",
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
			providers, err := client.List(providerCtx, &apiserverv1.ProviderListRequest{
				Tenant: testName,
				Limit:  1,
			})
			if err != nil {
				return false, err
			}
			assert.Len(t, providers.Items, 1)
			testutil.LogObject(t, providers)
			providers, err = client.List(providerCtx, &apiserverv1.ProviderListRequest{
				Tenant:   testName,
				Limit:    1,
				Continue: providers.Continue,
			})
			if err != nil {
				return false, err
			}
			assert.Len(t, providers.Items, 1)
			testutil.LogObject(t, providers)
			return true, nil
		}, providerCtx.Done()))

		// get provider
		require.NoError(t, wait.PollUntil(time.Second, func() (done bool, err error) {
			provider, err := client.Get(providerCtx, &apiserverv1.ProviderRequest{
				Tenant: testName,
				Name:   "test-provider-1",
			})
			if err != nil {
				return false, err
			}
			expectedResult := &apiserverv1.Provider{
				Name: "test-provider-1",
				Metadata: &apiserverv1.AccountMetadata{
					Description: "Test Provider",
					DisplayName: "Test Provider",
				},
			}
			assert.EqualValues(t, provider, expectedResult)
			testutil.LogObject(t, provider)
			return true, nil
		}, providerCtx.Done()))
	}
}

type toGRPCStatus interface {
	GRPCStatus() *status.Status
}

func offeringService(ctx context.Context, conn *grpc.ClientConn, managementClient *testutil.RecordingClient) func(t *testing.T) {
	return func(t *testing.T) {
		testName := strings.Replace(strings.ToLower(t.Name()), "/", "-", -1)
		ns := &corev1.Namespace{}
		ns.Name = testName
		require.NoError(t, managementClient.Create(ctx, ns))
		// Create offering objects in the management cluster.
		offering1 := &catalogv1alpha1.Offering{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-offering-1",
				Namespace: testName,
				Labels: map[string]string{
					"test-label": "offering1",
				},
			},
			Spec: catalogv1alpha1.OfferingSpec{
				Metadata: catalogv1alpha1.OfferingMetadata{
					Description: "Test Offering",
					DisplayName: "Test Offering",
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
				Namespace: testName,
				Labels: map[string]string{
					"test-label": "offering2",
				},
			},
			Spec: catalogv1alpha1.OfferingSpec{
				Metadata: catalogv1alpha1.OfferingMetadata{
					Description: "Test Offering",
					DisplayName: "Test Offering",
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
		offeringCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		t.Cleanup(cancel)
		// list offerings with limit and continuation token.
		require.NoError(t, wait.PollUntil(time.Second, func() (done bool, err error) {
			offerings, err := client.List(offeringCtx, &apiserverv1.OfferingListRequest{
				Account: testName,
				Limit:   1,
			})
			if err != nil {
				return false, err
			}
			assert.Len(t, offerings.Items, 1)
			offerings, err = client.List(offeringCtx, &apiserverv1.OfferingListRequest{
				Account:  testName,
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
			offering, err := client.Get(offeringCtx, &apiserverv1.OfferingGetRequest{
				Account: testName,
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
					Account: testName,
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
						Description: "Test Offering",
						DisplayName: "Test Offering",
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
	}
}

func regionService(ctx context.Context, conn *grpc.ClientConn, managementClient *testutil.RecordingClient) func(t *testing.T) {
	return func(t *testing.T) {
		testName := strings.Replace(strings.ToLower(t.Name()), "/", "-", -1)
		ns := &corev1.Namespace{}
		ns.Name = testName
		require.NoError(t, managementClient.Create(ctx, ns))
		// Create region objects in the management cluster.
		region1 := &catalogv1alpha1.Region{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-region-1",
				Namespace: testName,
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
				Namespace: testName,
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
			regions, err := client.List(regionCtx, &apiserverv1.RegionListRequest{
				Account: testName,
				Limit:   1,
			})
			if err != nil {
				return false, err
			}
			assert.Len(t, regions.Items, 1)
			regions, err = client.List(regionCtx, &apiserverv1.RegionListRequest{
				Account:  testName,
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
			region, err := client.Get(regionCtx, &apiserverv1.RegionGetRequest{
				Account: testName,
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
					Account: testName,
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
	}
}

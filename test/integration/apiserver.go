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
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	certmanagerv1alpha3 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha3"
	v1 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

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

		// testName := strings.Replace(strings.ToLower(t.Name()), "/", "-", -1)
		ns := &corev1.Namespace{}
		ns.Name = "kubecarrier-system"
		// require.NoError(t, managementClient.Create(ctx, ns))
		const localAPIServerPort = 9443

		token := fetchUserToken(ctx, t, managementClient, f.Config().ManagementExternalKubeconfigPath)
		t.Log("token", token)

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
					strings.Join([]string{"kubecarrier-api-server-manager", servingTLSSecret.GetNamespace(), "svc"}, "."),
					"kubecarrier-api-server-manager",
					"localhost",
				},
				IsCA: true,
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
				OIDC: operatorv1alpha1.APIServerOIDCConfig{
					// from test/testdata/dex_values.yaml
					IssuerURL:     "https://dex.kubecarrier-system.svc",
					ClientID:      "e2e-client-id",
					UsernameClaim: "name",
					CertificateAuthority: operatorv1alpha1.ObjectReference{
						Name: "dex-web-server",
					},
				},
			},
		}
		require.NoError(t, managementClient.Create(ctx, apiServer))
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, apiServer))

		ctx, cancel = context.WithCancel(ctx)
		t.Cleanup(cancel)

		pfCmd := exec.CommandContext(ctx,
			"kubectl",
			"--kubeconfig", f.Config().ManagementExternalKubeconfigPath,
			"--namespace", apiServer.GetNamespace(),
			"port-forward",
			// well known service name since it's assumed only one API server shall be deployed
			"service/kubecarrier-api-server-manager",
			fmt.Sprintf("%d:https", localAPIServerPort),
		)
		pfCmd.Stdout = os.Stdout
		pfCmd.Stderr = os.Stderr
		require.NoError(t, pfCmd.Start())

		certPool := x509.NewCertPool()
		require.True(t, certPool.AppendCertsFromPEM(servingTLSSecret.Data["ca.crt"]))
		require.True(t, certPool.AppendCertsFromPEM(servingTLSSecret.Data[corev1.TLSCertKey]))

		conn, err := grpc.DialContext(
			ctx,
			fmt.Sprintf("localhost:%d", localAPIServerPort),
			grpc.WithTransportCredentials(
				credentials.NewClientTLSFromCert(certPool, ""),
			),
			grpc.WithPerRPCCredentials(gRPCWithAuthToken{token: token}),
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
				t.Logf("gRPC server errored out, retrying : %d %v %v",
					grpcStatus.GRPCStatus().Code(),
					grpcStatus.GRPCStatus().Message(),
					grpcStatus.GRPCStatus().Err(),
				)
				return false, nil
			}
			return false, err
		}, versionCtx.Done()), "client version gRPC call")

		userinfo, err := client.WhoAmI(ctx, &empty.Empty{})
		if assert.NoError(t, err, "whoami gRPC") {
			t.Log("User info:")
			testutil.LogObject(t, userinfo)
		}
	}
}

func fetchUserToken(ctx context.Context, t *testing.T, managementClient *testutil.RecordingClient, kubeconfig string) string {
	const localDexServerPort = 10443
	pfDex := exec.CommandContext(ctx,
		"kubectl",
		"--kubeconfig", kubeconfig,
		"--namespace", "kubecarrier-system",
		"port-forward",
		"service/dex",
		fmt.Sprintf("%d:https", localDexServerPort),
	)
	pfDex.Stdout = os.Stdout
	pfDex.Stderr = os.Stderr

	require.NoError(t, pfDex.Start())
	certPool := x509.NewCertPool()
	dexTLSSecret := &corev1.Secret{}
	require.NoError(t, managementClient.Get(ctx, types.NamespacedName{Name: "dex-web-server", Namespace: "kubecarrier-system"}, dexTLSSecret))
	require.True(t, certPool.AppendCertsFromPEM(dexTLSSecret.Data["ca.crt"]))
	require.True(t, certPool.AppendCertsFromPEM(dexTLSSecret.Data[corev1.TLSCertKey]))
	token, err := testutil.DexFakeClientCredentialsGrant(
		ctx,
		testutil.NewLogger(t),
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

type toGRPCStatus interface {
	GRPCStatus() *status.Status
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

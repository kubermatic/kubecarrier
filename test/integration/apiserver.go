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
	"k8s.io/apimachinery/pkg/util/wait"

	apiserverv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/apiserver/v1alpha1"
	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
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
				OIDC: operatorv1alpha1.APIServerOIDCConfig{
					// from test/testdata/dex_values.yaml
					IssuerURL: "https://dex.kubecarrier-system.svc",
					ClientID:  "e2e-client-id",
					CertificateAuthority: operatorv1alpha1.ObjectReference{
						Name: "dex-web-server-ca",
					},
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

		token, err := testutil.DexFakeClientCredentialsGrant(
			ctx,
			testutil.NewLogger(t),
			"https://localhost:10443",
			"admin@example.com",
			"password",
		)
		require.NoError(t, err, "getting token from internal dex instance")
		t.Log("token", token)

		conn, err := grpc.DialContext(
			ctx,
			fmt.Sprintf("localhost:%d", localPort),
			grpc.WithTransportCredentials(
				credentials.NewClientTLSFromCert(certPool, ""),
			),
			grpc.WithPerRPCCredentials(withAuth{token: "token"}),
		)
		client := apiserverv1alpha1.NewKubecarrierClient(conn)
		versionCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		t.Cleanup(cancel)
		require.NoError(t, wait.PollUntil(time.Second, func() (done bool, err error) {
			version, err := client.Version(versionCtx, &apiserverv1alpha1.VersionRequest{})
			if err == nil {
				assert.NotEmpty(t, version.Version)
				assert.NotEmpty(t, version.Branch)
				assert.NotEmpty(t, version.BuildDate)
				assert.NotEmpty(t, version.GoVersion)
				t.Log("got response for gRPC server version")
				testutil.LogObject(t, version)
				return true, nil
			}
			if grpcStatus, ok := err.(ToGRPCStatus); ok {
				if grpcStatus.GRPCStatus().Code() == codes.Unavailable {
					t.Log("gRPC server temporary unavailable, retrying")
					return false, nil
				}
			}
			return false, err
		}, versionCtx.Done()), "client version gRPC call")
	}
}

type ToGRPCStatus interface {
	GRPCStatus() *status.Status
}

type withAuth struct {
	token string
}

func (w withAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"Authorization": "Bearer " + w.token,
	}, nil
}

func (w withAuth) RequireTransportSecurity() bool {
	return true
}

var _ credentials.PerRPCCredentials = withAuth{}

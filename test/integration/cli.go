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
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1alpha1 "k8c.io/kubecarrier/pkg/apis/core/v1alpha1"
	"k8c.io/kubecarrier/pkg/testutil"

	kubermatictestutil "k8c.io/utils/pkg/testutil"
)

func newCLI(f *testutil.Framework) func(t *testing.T) {
	return func(t *testing.T) {
		testName := strings.Replace(strings.ToLower(t.Name()), "/", "-", -1)
		ctx := context.Background()

		managementClient, err := f.ManagementClient(t)
		require.NoError(t, err, "creating management client")
		t.Cleanup(managementClient.CleanUpFunc(ctx))

		account := testutil.NewTenantAccount(testName+"-force", rbacv1.Subject{
			Kind:     rbacv1.GroupKind,
			APIGroup: "rbac.authorization.k8s.io",
			Name:     "admin",
		})
		require.NoError(t, managementClient.Create(ctx, account))
		require.NoError(t, kubermatictestutil.WaitUntilReady(ctx, managementClient, account))

		serviceCluster := f.SetupServiceCluster(ctx, managementClient, t, "eu-west-1", account)

		sca := &corev1alpha1.ServiceClusterAssignment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      account.Status.Namespace.Name + "." + serviceCluster.Name,
				Namespace: account.Status.Namespace.Name,
			},
			Spec: corev1alpha1.ServiceClusterAssignmentSpec{
				ServiceCluster: corev1alpha1.ObjectReference{
					Name: serviceCluster.Name,
				},
				ManagementClusterNamespace: corev1alpha1.ObjectReference{
					Name: account.Status.Namespace.Name,
				},
			},
		}

		require.NoError(t, managementClient.Create(ctx, sca))
		require.NoError(t, kubermatictestutil.WaitUntilReady(ctx, managementClient, sca))
		require.Error(t, managementClient.Delete(ctx, account))

		c := exec.CommandContext(ctx, "kubectl", "kubecarrier", "delete", "--kubeconfig", f.Config().ManagementExternalKubeconfigPath, "account", "--force", account.Name)
		out, err := c.CombinedOutput()
		t.Log(string(out))
		require.NoError(t, err)

		require.NoError(t, kubermatictestutil.WaitUntilNotFound(ctx, managementClient, account, kubermatictestutil.WithTimeout(5*time.Second)))
	}
}

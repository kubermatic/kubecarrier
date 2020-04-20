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

package installation

import (
	"context"
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func newManagementCluster(f *testutil.Framework) func(t *testing.T) {
	return func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		t.Cleanup(cancel)

		masterClient, err := f.MasterClient(t)
		require.NoError(t, err, "creating master client")

		managementCluster := f.SetupManagementCluster(ctx, masterClient, t, "kubecarrier-1")
		assert.False(t, managementCluster.IsReady(), "management cluster should not become ready since the KubeCarrier haven't been installed in the management cluster")

		// Install KubeCarrier in the management cluster
		c := exec.CommandContext(ctx, "kubectl", "kubecarrier", "setup", "--kubeconfig", f.Config().ManagementExternalKubeconfigPath)
		out, err := c.CombinedOutput()
		t.Log(string(out))
		require.NoError(t, err)

		managementClient, err := f.ManagementClient(t)
		require.NoError(t, err, "creating management client")

		kubeCarrier := &operatorv1alpha1.KubeCarrier{ObjectMeta: metav1.ObjectMeta{
			Name: "kubecarrier1",
		}}

		err = managementClient.Create(ctx, kubeCarrier)
		if assert.Error(t, err,
			"KubeCarrier object with name other than 'kubecarrier' should not be allowed to be created",
		) {
			assert.Contains(
				t,
				err.Error(),
				"KubeCarrier object name should be 'kubecarrier', found: kubecarrier1",
				"KubeCarrier creation webhook should error out on incorrect KubeCarrier object name",
			)
		}
		kubeCarrier.Name = kubeCarrierName
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, kubeCarrier))

		// Check ManagementCluster in the master cluster is ready
		require.NoError(t, testutil.WaitUntilReady(ctx, masterClient, managementCluster))

		operatorDeployment := &appsv1.Deployment{}
		assert.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name:      "kubecarrier-operator-manager",
			Namespace: kubecarrierSystem,
		}, operatorDeployment), "getting the KubeCarrier operator deployment error")

		namespace := &corev1.Namespace{}
		assert.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name: kubecarrierSystem,
		}, namespace), "getting the KubeCarrier system namespace error")

		assert.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name:      "kubecarrier",
			Namespace: kubecarrierSystem,
		}, kubeCarrier), "getting the KubeCarrier object")

		// Check objects that owned by the KubeCarrier object.
		// Deployment
		deployment := &appsv1.Deployment{}
		assert.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-controller-manager", prefix),
			Namespace: kubecarrierSystem,
		}, deployment), "get the Deployment that owned by KubeCarrier object")

		// Webhook Service
		service := &corev1.Service{}
		assert.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-webhook-service", prefix),
			Namespace: kubecarrierSystem,
		}, service), "get the Webhook Service that owned by KubeCarrier object")

		// ClusterRoleBinding
		clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
		assert.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name: fmt.Sprintf("%s-manager-rolebinding", prefix),
		}, clusterRoleBinding), "get the ClusterRoleBinding that owned by KubeCarrier object")

		// ClusterRole
		clusterRole := &rbacv1.ClusterRole{}
		assert.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name: fmt.Sprintf("%s-manager-role", prefix),
		}, clusterRole), "get the ClusterRole that owned by KubeCarrier object")

		// RoleBinding
		roleBinding := &rbacv1.RoleBinding{}
		assert.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-leader-election-rolebinding", prefix),
			Namespace: kubecarrierSystem,
		}, roleBinding), "get the RoleBinding that owned by KubeCarrier object")

		// Role
		role := &rbacv1.Role{}
		assert.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-leader-election-role", prefix),
			Namespace: kubecarrierSystem,
		}, role), "get the Role that owned by KubeCarrier object")

		// CRD
		crd := &apiextensionsv1.CustomResourceDefinition{}
		assert.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name: "accounts.catalog.kubecarrier.io",
		}, crd), "get the CRD that owned by KubeCarrier object")
	}
}

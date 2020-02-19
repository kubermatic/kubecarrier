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
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
	"github.com/kubermatic/kubecarrier/test/framework"
)

const (
	kubecarrierSystem = "kubecarrier-system"
	prefix            = "kubecarrier-manager"
)

func NewInstallationSuite(f *framework.Framework) func(t *testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()
		kubeCarrier := &operatorv1alpha1.KubeCarrier{}

		if !t.Run("anchor setup", testAnchorSetup(ctx, f, kubeCarrier)) {
			t.FailNow()
		}
		if !t.Run("kubeCarrier teardown", testAnchorTeardown(ctx, f, kubeCarrier)) {
			t.FailNow()
		}

		var out bytes.Buffer
		c := exec.Command("anchor", "setup", "--kubeconfig", f.Config().MasterExternalKubeconfigPath)
		c.Stdout = &out
		c.Stderr = &out
		require.NoError(t, c.Run(), "\"anchor setup\" returned an error: %s", out.String())
	}
}

func testAnchorSetup(ctx context.Context, f *framework.Framework, kubeCarrier *operatorv1alpha1.KubeCarrier) func(t *testing.T) {
	return func(t *testing.T) {
		var out bytes.Buffer
		c := exec.Command("anchor", "setup", "--kubeconfig", f.Config().MasterExternalKubeconfigPath)
		c.Stdout = &out
		c.Stderr = &out
		require.NoError(t, c.Run(), "\"anchor setup\" returned an error: %s", out.String())

		// Create another client due to some issues about the restmapper.
		// The issue is that if you use the client that created before, and here try to create the kubeCarrier,
		// it will complain about: `no matches for kind "KubeCarrier" in version "operator.kubecarrier.io/v1alpha1"`,
		// but actually, the scheme is already added to the runtime scheme.
		// And in the following, reinitializing the client solves the issue.
		masterClient, err := f.MasterClient()
		require.NoError(t, err, "creating master client")

		operatorDeployment := &appsv1.Deployment{}
		assert.NoError(t, masterClient.Get(ctx, types.NamespacedName{
			Name:      "kubecarrier-operator-manager",
			Namespace: kubecarrierSystem,
		}, operatorDeployment), "getting the KubeCarrier operator deployment error")

		namespace := &corev1.Namespace{}
		assert.NoError(t, masterClient.Get(ctx, types.NamespacedName{
			Name: kubecarrierSystem,
		}, namespace), "getting the KubeCarrier system namespace error")

		assert.NoError(t, masterClient.Get(ctx, types.NamespacedName{
			Name:      "kubecarrier",
			Namespace: kubecarrierSystem,
		}, kubeCarrier), "getting the KubeCarrier object")

		// Check objects that owned by the KubeCarrier object.
		// Deployment
		deployment := &appsv1.Deployment{}
		assert.NoError(t, masterClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-controller-manager", prefix),
			Namespace: kubecarrierSystem,
		}, deployment), "get the Deployment that owned by KubeCarrier object")

		// Webhook Service
		service := &corev1.Service{}
		assert.NoError(t, masterClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-webhook-service", prefix),
			Namespace: kubecarrierSystem,
		}, service), "get the Webhook Service that owned by KubeCarrier object")

		// ClusterRoleBinding
		clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
		assert.NoError(t, masterClient.Get(ctx, types.NamespacedName{
			Name: fmt.Sprintf("%s-manager-rolebinding", prefix),
		}, clusterRoleBinding), "get the ClusterRoleBinding that owned by KubeCarrier object")

		// ClusterRole
		clusterRole := &rbacv1.ClusterRole{}
		assert.NoError(t, masterClient.Get(ctx, types.NamespacedName{
			Name: fmt.Sprintf("%s-manager-role", prefix),
		}, clusterRole), "get the ClusterRole that owned by KubeCarrier object")

		// RoleBinding
		roleBinding := &rbacv1.RoleBinding{}
		assert.NoError(t, masterClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-leader-election-rolebinding", prefix),
			Namespace: kubecarrierSystem,
		}, roleBinding), "get the RoleBinding that owned by KubeCarrier object")

		// Role
		role := &rbacv1.Role{}
		assert.NoError(t, masterClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-leader-election-role", prefix),
			Namespace: kubecarrierSystem,
		}, role), "get the Role that owned by KubeCarrier object")

		// CRD
		crd := &apiextensionsv1.CustomResourceDefinition{}
		assert.NoError(t, masterClient.Get(ctx, types.NamespacedName{
			Name: "accounts.catalog.kubecarrier.io",
		}, crd), "get the CRD that owned by KubeCarrier object")
	}
}

func testAnchorTeardown(ctx context.Context, f *framework.Framework, kubeCarrier *operatorv1alpha1.KubeCarrier) func(t *testing.T) {
	return func(t *testing.T) {
		masterClient, err := f.MasterClient()
		require.NoError(t, err, "creating master client")

		// Delete the KubeCarrier object.
		require.NoError(t, masterClient.Delete(ctx, kubeCarrier), "deleting the KubeCarrier object")

		// Deployment
		assert.NoError(t, testutil.WaitUntilNotFound(masterClient, &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-controller-manager", prefix),
				Namespace: kubecarrierSystem,
			},
		}))

		// Webhook Service
		service := &corev1.Service{}
		assert.True(t, errors.IsNotFound(masterClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-webhook-service", prefix),
			Namespace: kubecarrierSystem,
		}, service)), "get the Webhook Service that is owned by the KubeCarrier object")

		// ClusterRoleBinding
		clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
		assert.True(t, errors.IsNotFound(masterClient.Get(ctx, types.NamespacedName{
			Name: fmt.Sprintf("%s-manager-rolebinding", prefix),
		}, clusterRoleBinding)), "get the ClusterRoleBinding that owned by KubeCarrier object")

		// ClusterRole
		clusterRole := &rbacv1.ClusterRole{}
		assert.True(t, errors.IsNotFound(masterClient.Get(ctx, types.NamespacedName{
			Name: fmt.Sprintf("%s-manager-role", prefix),
		}, clusterRole)), "get the ClusterRole that owned by KubeCarrier object")

		// RoleBinding
		roleBinding := &rbacv1.RoleBinding{}
		assert.True(t, errors.IsNotFound(masterClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-leader-election-rolebinding", prefix),
			Namespace: kubecarrierSystem,
		}, roleBinding)), "get the RoleBinding that owned by KubeCarrier object")

		// Role
		role := &rbacv1.Role{}
		assert.True(t, errors.IsNotFound(masterClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-leader-election-role", prefix),
			Namespace: kubecarrierSystem,
		}, role)), "get the Role that owned by KubeCarrier object")

		// CRD
		crd := &apiextensionsv1.CustomResourceDefinition{}
		assert.True(t, errors.IsNotFound(masterClient.Get(ctx, types.NamespacedName{
			Name: "accounts.catalog.kubecarrier.io",
		}, crd)), "get the CRD that owned by KubeCarrier object")
	}
}

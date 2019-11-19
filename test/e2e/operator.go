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

package e2e

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
)

func (suite *E2ESuite) TestOperatorKubeCarrierReconcile() {
	ctx := context.Background()
	prefix := "kubecarrier-manager"
	nn := "kubecarrier-system"
	kubeCarrier := &operatorv1alpha1.KubeCarrier{}

	// Check the finalizers of the KubeCarrier.
	err := suite.masterClient.Get(ctx, types.NamespacedName{
		Name:      "kubecarrier",
		Namespace: nn,
	}, kubeCarrier)
	suite.Require().NoError(err, "getting KubeCarrier error")
	suite.Len(kubeCarrier.Finalizers, 1, "finalizer should be added.")

	// Check objects that owned by the KubeCarrier object.
	// ClusterRoleBinding
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	suite.NoError(suite.masterClient.Get(ctx, types.NamespacedName{
		Name: fmt.Sprintf("%s-proxy-rolebinding", prefix),
	}, clusterRoleBinding), "get the ClusterRoleBinding that owned by KubeCarrier object")

	// ClusterRole
	clusterRole := &rbacv1.ClusterRole{}
	suite.NoError(suite.masterClient.Get(ctx, types.NamespacedName{
		Name: fmt.Sprintf("%s-proxy-role", prefix),
	}, clusterRole), "get the ClusterRole that owned by KubeCarrier object")

	// RoleBinding
	roleBinding := &rbacv1.RoleBinding{}
	suite.NoError(suite.masterClient.Get(ctx, types.NamespacedName{
		Name:      fmt.Sprintf("%s-leader-election-rolebinding", prefix),
		Namespace: nn,
	}, roleBinding), "get the RoleBinding that owned by KubeCarrier object")

	// Role
	role := &rbacv1.Role{}
	suite.NoError(suite.masterClient.Get(ctx, types.NamespacedName{
		Name:      fmt.Sprintf("%s-leader-election-role", prefix),
		Namespace: nn,
	}, role), "get the Role that owned by KubeCarrier object")

	// Service
	service := &corev1.Service{}
	suite.NoError(suite.masterClient.Get(ctx, types.NamespacedName{
		Name:      fmt.Sprintf("%s-controller-manager-metrics-service", prefix),
		Namespace: nn,
	}, service), "get the Service that owned by KubeCarrier object")

	// Deployment
	deployment := &appsv1.Deployment{}
	suite.NoError(suite.masterClient.Get(ctx, types.NamespacedName{
		Name:      fmt.Sprintf("%s-controller-manager", prefix),
		Namespace: nn,
	}, deployment), "get the Deployment that owned by KubeCarrier object")

	// Delete the KubeCarrier object.
	suite.Require().NoError(suite.masterClient.Delete(ctx, kubeCarrier), "deleting KubeCarrier error")
	time.Sleep(10 * time.Second)

	// Check objects that owned by the KubeCarrier object.
	// ClusterRoleBinding
	suite.True(apierrors.IsNotFound(suite.masterClient.Get(ctx, types.NamespacedName{
		Name: fmt.Sprintf("%s-proxy-rolebinding", prefix),
	}, clusterRoleBinding)), "get the ClusterRoleBinding that owned by KubeCarrier object")

	// ClusterRole
	suite.True(apierrors.IsNotFound(suite.masterClient.Get(ctx, types.NamespacedName{
		Name: fmt.Sprintf("%s-proxy-role", prefix),
	}, clusterRole)), "get the ClusterRole that owned by KubeCarrier object")

	// RoleBinding
	suite.True(apierrors.IsNotFound(suite.masterClient.Get(ctx, types.NamespacedName{
		Name:      fmt.Sprintf("%s-leader-election-rolebinding", prefix),
		Namespace: nn,
	}, roleBinding)), "get the RoleBinding that owned by KubeCarrier object")

	// Role
	suite.True(apierrors.IsNotFound(suite.masterClient.Get(ctx, types.NamespacedName{
		Name:      fmt.Sprintf("%s-leader-election-role", prefix),
		Namespace: nn,
	}, role)), "get the Role that owned by KubeCarrier object")

	// Service
	suite.True(apierrors.IsNotFound(suite.masterClient.Get(ctx, types.NamespacedName{
		Name:      fmt.Sprintf("%s-controller-manager-metrics-service", prefix),
		Namespace: nn,
	}, service)), "get the Service that owned by KubeCarrier object")

	// Deployment
	suite.True(apierrors.IsNotFound(suite.masterClient.Get(ctx, types.NamespacedName{
		Name:      fmt.Sprintf("%s-controller-manager", prefix),
		Namespace: nn,
	}, deployment)), "get the Deployment that owned by KubeCarrier object")
}

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
	"time"

	"github.com/stretchr/testify/suite"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/test/framework"
)

var _ suite.SetupAllSuite = (*InstallationSuite)(nil)

// InstallationSuite verifies if the KubeCarrier operator and KubeCarrier can be deployed.
type InstallationSuite struct {
	suite.Suite
	*framework.Framework

	masterClient  client.Client
	serviceClient client.Client
}

func (s *InstallationSuite) SetupSuite() {
	var err error
	s.masterClient, err = s.MasterClient()
	s.Require().NoError(err, "creating master client")
	s.serviceClient, err = s.ServiceClient()
	s.Require().NoError(err, "creating service client")
}

func (s *InstallationSuite) TestInstallAndTeardown() {
	ctx := context.Background()
	nn := "kubecarrier-system"
	prefix := "kubecarrier-manager"
	kubeCarrier := &operatorv1alpha1.KubeCarrier{}
	s.Run("anchor setup", func() {
		s.T().Logf("running \"anchor setup\" to install KubeCarrier in the master cluster")
		var out bytes.Buffer
		c := exec.Command("anchor", "setup", "--kubeconfig", s.Framework.Config().MasterExternalKubeconfigPath)
		c.Stdout = &out
		c.Stderr = &out
		s.Require().NoError(c.Run(), "\"anchor setup\" returned an error: %s", out.String())

		// Create another client due to some issues about the restmapper.
		// The issue is that if you use the client that created before, and here try to create the kubeCarrier,
		// it will complain about: `no matches for kind "KubeCarrier" in version "operator.kubecarrier.io/v1alpha1"`,
		// but actually, the scheme is already added to the runtime scheme.
		// And in the following, reinitializing the client solves the issue.
		var err error
		s.masterClient, err = s.MasterClient()
		s.Require().NoError(err, "creating master client")

		operatorDeployment := &appsv1.Deployment{}
		s.NoError(s.masterClient.Get(ctx, types.NamespacedName{
			Name:      "kubecarrier-operator-manager",
			Namespace: nn,
		}, operatorDeployment), "getting the KubeCarrier operator deployment error")

		namespace := &corev1.Namespace{}
		s.NoError(s.masterClient.Get(ctx, types.NamespacedName{
			Name: nn,
		}, namespace), "getting the KubeCarrier system namespace error")

		s.NoError(s.masterClient.Get(ctx, types.NamespacedName{
			Name:      "kubecarrier",
			Namespace: nn,
		}, kubeCarrier), "getting the KubeCarrier error")

		// Check objects that owned by the KubeCarrier object.
		// Deployment
		deployment := &appsv1.Deployment{}
		s.NoError(s.masterClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-controller-manager", prefix),
			Namespace: nn,
		}, deployment), "get the Deployment that owned by KubeCarrier object")

		// ClusterRoleBinding
		clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
		s.NoError(s.masterClient.Get(ctx, types.NamespacedName{
			Name: fmt.Sprintf("%s-proxy-rolebinding", prefix),
		}, clusterRoleBinding), "get the ClusterRoleBinding that owned by KubeCarrier object")

		// ClusterRole
		clusterRole := &rbacv1.ClusterRole{}
		s.NoError(s.masterClient.Get(ctx, types.NamespacedName{
			Name: fmt.Sprintf("%s-proxy-role", prefix),
		}, clusterRole), "get the ClusterRole that owned by KubeCarrier object")

		// RoleBinding
		roleBinding := &rbacv1.RoleBinding{}
		s.NoError(s.masterClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-leader-election-rolebinding", prefix),
			Namespace: nn,
		}, roleBinding), "get the RoleBinding that owned by KubeCarrier object")

		// Role
		role := &rbacv1.Role{}
		s.NoError(s.masterClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-leader-election-role", prefix),
			Namespace: nn,
		}, role), "get the Role that owned by KubeCarrier object")

		// Service
		service := &corev1.Service{}
		s.NoError(s.masterClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-controller-manager-metrics-service", prefix),
			Namespace: nn,
		}, service), "get the Service that owned by KubeCarrier object")

	})

	s.Run("kubeCarrier teardown", func() {
		// Delete the KubeCarrier object.
		s.Require().NoError(wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
			if err = s.masterClient.Delete(ctx, kubeCarrier); err != nil {
				if errors.IsNotFound(err) {
					return true, nil
				}
				return false, err
			}
			return false, nil
		}), "deleting the KubeCarrier object")

		// Deployment
		deployment := &appsv1.Deployment{}
		s.NoError(wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
			if err = s.masterClient.Get(ctx, types.NamespacedName{
				Name:      fmt.Sprintf("%s-controller-manager", prefix),
				Namespace: nn,
			}, deployment); err != nil {
				if errors.IsNotFound(err) {
					return true, nil
				}
				return false, err
			}
			return false, nil
		}), "get the Deployment that owned by KubeCarrier object")

		// ClusterRoleBinding
		clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
		s.True(errors.IsNotFound(s.masterClient.Get(ctx, types.NamespacedName{
			Name: fmt.Sprintf("%s-proxy-rolebinding", prefix),
		}, clusterRoleBinding)), "get the ClusterRoleBinding that owned by KubeCarrier object")

		// ClusterRole
		clusterRole := &rbacv1.ClusterRole{}
		s.True(errors.IsNotFound(s.masterClient.Get(ctx, types.NamespacedName{
			Name: fmt.Sprintf("%s-proxy-role", prefix),
		}, clusterRole)), "get the ClusterRole that owned by KubeCarrier object")

		// RoleBinding
		roleBinding := &rbacv1.RoleBinding{}
		s.True(errors.IsNotFound(s.masterClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-leader-election-rolebinding", prefix),
			Namespace: nn,
		}, roleBinding)), "get the RoleBinding that owned by KubeCarrier object")

		// Role
		role := &rbacv1.Role{}
		s.True(errors.IsNotFound(s.masterClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-leader-election-role", prefix),
			Namespace: nn,
		}, role)), "get the Role that owned by KubeCarrier object")

		// Service
		service := &corev1.Service{}
		s.True(errors.IsNotFound(s.masterClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-controller-manager-metrics-service", prefix),
			Namespace: nn,
		}, service)), "get the Service that owned by KubeCarrier object")
	})
}

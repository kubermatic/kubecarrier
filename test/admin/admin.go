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

package admin

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	"github.com/kubermatic/kubecarrier/test/framework"
)

var _ suite.SetupAllSuite = (*AdminSuite)(nil)

// AdminSuite checks the Tenant/Provider creations and deletions.
type AdminSuite struct {
	suite.Suite
	*framework.Framework

	masterClient  client.Client
	serviceClient client.Client
}

func (s *AdminSuite) SetupSuite() {
	var err error
	s.serviceClient, err = s.ServiceClient()
	s.Require().NoError(err, "creating service client")

	s.T().Logf("running \"anchor setup\" to install KubeCarrier in the master cluster")
	var out bytes.Buffer
	c := exec.Command("anchor", "setup", "--kubeconfig", s.Framework.Config().MasterExternalKubeconfigPath)
	c.Stdout = &out
	c.Stderr = &out
	s.Require().NoError(c.Run(), "\"anchor setup\" returned an error: %s", out.String())

	s.masterClient, err = s.MasterClient()
	s.Require().NoError(err, "creating master client")
}

func (s *AdminSuite) TestTenantCreationAndDeletion() {
	ctx := context.Background()
	tenant := &catalogv1alpha1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-tenant1",
			Namespace: "kubecarrier-system",
		},
	}

	if !s.Run("Tenant creation", func() {
		s.Require().NoError(s.masterClient.Create(ctx, tenant), "creating tenant error")

		// Try to get the namespace that created for this tenant.
		namespace := &corev1.Namespace{}
		s.NoError(wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
			if err := s.masterClient.Get(ctx, types.NamespacedName{
				Name: fmt.Sprintf("tenant-%s", tenant.Name),
			}, namespace); err != nil {
				if errors.IsNotFound(err) {
					return false, nil
				}
				return true, err

			}
			return true, nil
		}), "getting the namespace for the Tenant error")
	}) {
		s.FailNow("Tenant creation e2e test failed.")
	}

	s.Run("Tenant deletion", func() {
		s.Require().NoError(wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
			if err = s.masterClient.Delete(ctx, tenant); err != nil {
				if errors.IsNotFound(err) {
					return true, nil
				}
				return false, err
			}
			return false, nil
		}), "deleting the Tenant error")

		// Try to delete the namespace that created for this tenant.
		namespace := &corev1.Namespace{}
		s.True(errors.IsNotFound(s.masterClient.Get(ctx, types.NamespacedName{
			Name: fmt.Sprintf("tenant-%s", tenant.Name),
		}, namespace)), "namespace should also be deleted.")
	})

}

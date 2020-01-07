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

package provider

import (
	"context"
	"time"

	"github.com/stretchr/testify/suite"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/test/framework"
)

var _ suite.SetupAllSuite = (*ProviderSuite)(nil)

// ProviderSuite verifies if the provider actions are working.
// - atm it just deploys a catapult instance
type ProviderSuite struct {
	suite.Suite
	*framework.Framework

	masterClient      client.Client
	serviceClient     client.Client
	providerNamespace string
}

func (s *ProviderSuite) SetupSuite() {
	var err error
	s.masterClient, err = s.MasterClient()
	s.Require().NoError(err, "creating master client")
	s.serviceClient, err = s.ServiceClient()
	s.Require().NoError(err, "creating service client")

	s.providerNamespace = "provider-test-provider1"
}

func (s *ProviderSuite) TestCatapultDeployAndTeardown() {
	ctx := context.Background()

	catapult := &operatorv1alpha1.Catapult{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-crd",
			Namespace: s.providerNamespace,
		},
	}

	s.Require().NoError(s.masterClient.Create(ctx, catapult), "creating Catapult error")

	// Wait for Catapult to be ready
	s.Require().NoError(wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
		if err := s.masterClient.Get(ctx, types.NamespacedName{
			Name:      catapult.Name,
			Namespace: catapult.Namespace,
		}, catapult); err != nil {
			return true, err
		}

		readyCond, _ := catapult.Status.GetCondition(operatorv1alpha1.CatapultReady)
		return readyCond.Status == operatorv1alpha1.ConditionTrue, nil
	}), "waiting for Catapult to become ready")

	// Check created objects
	catapultDeployment := &appsv1.Deployment{}
	s.Require().NoError(s.masterClient.Get(ctx, types.NamespacedName{
		Name:      "catapult-manager",
		Namespace: catapult.Namespace,
	}, catapultDeployment), "getting the Catapult manager deployment error")

	// Teardown
	s.Require().NoError(s.masterClient.Delete(ctx, catapult), "deleting the Catapult object")

	// Wait for Catapult to be gone
	s.Require().NoError(wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
		if err := s.masterClient.Get(ctx, types.NamespacedName{
			Name:      catapult.Name,
			Namespace: catapult.Namespace,
		}, catapult); err != nil {
			if errors.IsNotFound(err) {
				return true, nil
			}
			return true, err
		}

		return false, nil
	}), "waiting for Catapult to be gone")

	// check everything is gone
	s.Require().NoError(wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
		err = s.masterClient.Get(ctx, types.NamespacedName{
			Name:      catapultDeployment.Name,
			Namespace: catapultDeployment.Namespace,
		}, catapultDeployment)
		if errors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return true, err
		}
		return false, nil
	}), "waiting for catapult manager deployment to be gone")
}

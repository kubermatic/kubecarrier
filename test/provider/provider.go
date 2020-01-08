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

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/test/framework"
)

var (
	_ suite.SetupAllSuite    = (*ProviderSuite)(nil)
	_ suite.TearDownAllSuite = (*ProviderSuite)(nil)
)

// ProviderSuite verifies if the provider actions are working.
// - atm it just deploys a catapult instance
type ProviderSuite struct {
	suite.Suite
	*framework.Framework

	masterClient  client.Client
	serviceClient client.Client
	provider      *catalogv1alpha1.Provider
}

func (s *ProviderSuite) SetupSuite() {
	var err error
	s.masterClient, err = s.MasterClient()
	s.Require().NoError(err, "creating master client")
	s.serviceClient, err = s.ServiceClient()
	s.Require().NoError(err, "creating service client")

	// Create a Provider to execute our tests in
	s.provider = &catalogv1alpha1.Provider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-cloud",
			Namespace: "kubecarrier-system",
		},
	}
	ctx := context.Background()
	s.Require().NoError(s.masterClient.Create(ctx, s.provider), "could not create Provider")

	// wait for provider to be ready
	s.Require().NoError(wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
		if err := s.masterClient.Get(ctx, types.NamespacedName{
			Name:      s.provider.Name,
			Namespace: s.provider.Namespace,
		}, s.provider); err != nil {
			return true, err
		}

		cond, _ := s.provider.Status.GetCondition(catalogv1alpha1.ProviderReady)
		return cond.Status == catalogv1alpha1.ConditionTrue, nil
	}), "waiting for provider to be ready")
}

func (s *ProviderSuite) TearDownSuite() {
	ctx := context.Background()
	s.Require().NoError(s.masterClient.Delete(ctx, s.provider), "could not delete Provider")
}

func (s *ProviderSuite) TestCatapultDeployAndTeardown() {
	ctx := context.Background()

	catapult := &operatorv1alpha1.Catapult{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db.eu-west-1",
			Namespace: s.provider.Status.NamespaceName,
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
	s.NoError(s.masterClient.Get(ctx, types.NamespacedName{
		Name:      "db-eu-west-1-catapult-manager",
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

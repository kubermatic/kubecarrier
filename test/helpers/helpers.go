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

package helpers

import (
	"context"
	"time"

	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
)

type ProviderMixin struct {
	suite.Suite
	Client client.Client
}

func (s *ProviderMixin) CreateProvider(name string) *catalogv1alpha1.Provider {
	// Create a Provider to execute our tests in
	provider := &catalogv1alpha1.Provider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "kubecarrier-system",
		},
	}
	ctx := context.Background()
	s.Require().NoError(s.Client.Create(ctx, provider), "could not create Provider")

	// wait for provider to be ready
	s.Require().NoError(wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
		if err := s.Client.Get(ctx, types.NamespacedName{
			Name:      provider.Name,
			Namespace: provider.Namespace,
		}, provider); err != nil {
			return true, err
		}

		cond, _ := provider.Status.GetCondition(catalogv1alpha1.ProviderReady)
		return cond.Status == catalogv1alpha1.ConditionTrue, nil
	}), "waiting for provider to be ready")

	return provider
}

func (s *ProviderMixin) DeleteProvider(provider *catalogv1alpha1.Provider) {
	ctx := context.Background()
	s.Require().NoError(wait.Poll(time.Second, 30*time.Second, func() (done bool, err error) {
		if err = s.Client.Delete(ctx, provider); err != nil {
			if errors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	}), "could not delete the Provider")
}

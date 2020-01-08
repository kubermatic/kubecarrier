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
	"fmt"
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

var _ suite.SetupAllSuite = (*ProviderSuite)(nil)
var _ suite.TearDownAllSuite = (*ProviderSuite)(nil)

// ProviderSuite verify Provider related operations (ServiceCluster, Catalog, etc).
type ProviderSuite struct {
	suite.Suite
	*framework.Framework

	masterClient  client.Client
	serviceClient client.Client
}

func (s *ProviderSuite) SetupSuite() {
	var err error
	ctx := context.Background()
	s.masterClient, err = s.MasterClient()
	s.Require().NoError(err, "creating master client")
	s.serviceClient, err = s.ServiceClient()
	s.Require().NoError(err, "creating service client")

	tenant := &catalogv1alpha1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-tenant1",
			Namespace: "kubecarrier-system",
		},
	}

	provider := &catalogv1alpha1.Provider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-provider1",
			Namespace: "kubecarrier-system",
		},
	}

	catalogEntry := &catalogv1alpha1.CatalogEntry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "couchdbs",
			Namespace: "provider-test-provider1",
		},
		Spec: catalogv1alpha1.CatalogEntrySpec{
			Metadata: catalogv1alpha1.CatalogEntryMetadata{
				DisplayName: "Couch DB",
				Description: "The comfy nosql database",
			},
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

	if !s.Run("Provider creation", func() {
		s.Require().NoError(s.masterClient.Create(ctx, provider), "creating provider error")

		// Try to get the namespace that created for this provider.
		namespace := &corev1.Namespace{}
		s.NoError(wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
			if err := s.masterClient.Get(ctx, types.NamespacedName{
				Name: fmt.Sprintf("provider-%s", provider.Name),
			}, namespace); err != nil {
				if errors.IsNotFound(err) {
					return false, nil
				}
				return true, err

			}
			return true, nil
		}), "getting the namespace for the Provider error")

		// Try to get the TenantReference that created for this provider from the tenant.
		tenantReference := &catalogv1alpha1.TenantReference{}
		s.NoError(wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
			if err := s.masterClient.Get(ctx, types.NamespacedName{
				Name:      tenant.Name,
				Namespace: fmt.Sprintf("provider-%s", provider.Name),
			}, tenantReference); err != nil {
				if errors.IsNotFound(err) {
					return false, nil
				}
				return true, err

			}
			return true, nil
		}), "getting the tenantReference for the Provider error")
	}) {
		s.FailNow("Provider creation e2e test failed.")
	}

	if !s.Run("CatalogEntry creation", func() {
		s.Require().NoError(s.masterClient.Create(ctx, catalogEntry), "creating catalogEntry error")

		// Check the status of the CatalogEntry.
		catalogEntryFound := &catalogv1alpha1.CatalogEntry{}
		s.NoError(wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
			if err := s.masterClient.Get(ctx, types.NamespacedName{
				Name:      catalogEntry.Name,
				Namespace: catalogEntry.Namespace,
			}, catalogEntryFound); err != nil {
				if errors.IsNotFound(err) {
					return false, nil
				}
				return true, err
			}
			return true, nil
		}), "getting the CatalogEntry error")
	}) {
		s.FailNow("CatalogEntry creation e2e test failed.")
	}

}

func (s *ProviderSuite) TestCatalogCreationAndDeletion() {
	ctx := context.Background()
	catalog := &catalogv1alpha1.Catalog{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-catalog",
			Namespace: "provider-test-provider1",
		},
		Spec: catalogv1alpha1.CatalogSpec{
			CatalogEntrySelector:    &metav1.LabelSelector{},
			TenantReferenceSelector: &metav1.LabelSelector{},
		},
	}

	if !s.Run("Catalog creation", func() {
		s.Require().NoError(s.masterClient.Create(ctx, catalog), "creating Catalog error")

		// Check the status of the Catalog.
		catalogFound := &catalogv1alpha1.Catalog{}
		s.NoError(wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
			if err := s.masterClient.Get(ctx, types.NamespacedName{
				Name:      catalog.Name,
				Namespace: catalog.Namespace,
			}, catalogFound); err != nil {
				if errors.IsNotFound(err) {
					return false, nil
				}
				return true, err
			}
			return len(catalogFound.Status.Entries) == 1 && len(catalogFound.Status.Tenants) == 1, nil
		}), "getting the Catalog error")
	}) {
		s.FailNow("Catalog creation e2e test failed.")
	}

	if !s.Run("Catalog deletion", func() {
		s.Require().NoError(wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
			if err = s.masterClient.Delete(ctx, catalog); err != nil {
				if errors.IsNotFound(err) {
					return true, nil
				}
				return false, err
			}
			return false, nil
		}), "deleting the Catalog error")
	}) {
		s.FailNow("Catalog deletion e2e test failed.")
	}
}

func (s *ProviderSuite) TearDownSuite() {
	ctx := context.Background()
	// Remove the provider for testing.
	provider := &catalogv1alpha1.Provider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-provider1",
			Namespace: "kubecarrier-system",
		},
	}
	// Remove the tenant for testing.
	tenant := &catalogv1alpha1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-tenant1",
			Namespace: "kubecarrier-system",
		},
	}
	// Remove the catalogEntry for testing.
	catalogEntry := &catalogv1alpha1.CatalogEntry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "couchdbs",
			Namespace: "provider-test-provider1",
		},
	}

	s.Require().NoError(s.masterClient.Delete(ctx, provider), "deleting Provider error")
	s.Require().NoError(s.masterClient.Delete(ctx, tenant), "deleting Tenant error")
	s.Require().NoError(s.masterClient.Delete(ctx, catalogEntry), "deleting CatalogEntry error")
}

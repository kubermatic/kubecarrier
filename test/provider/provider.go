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
	appsv1 "k8s.io/api/apps/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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

	// objects that used for executing tests.
	provider     *catalogv1alpha1.Provider
	tenant       *catalogv1alpha1.Tenant
	catalogEntry *catalogv1alpha1.CatalogEntry
	crd          *apiextensionsv1.CustomResourceDefinition
}

func (s *ProviderSuite) SetupSuite() {
	var err error
	s.masterClient, err = s.MasterClient()
	s.Require().NoError(err, "creating master client")
	s.serviceClient, err = s.ServiceClient()
	s.Require().NoError(err, "creating service client")

	ctx := context.Background()

	// Create a Provider to execute our tests in
	s.provider = &catalogv1alpha1.Provider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-cloud",
			Namespace: "kubecarrier-system",
		},
	}
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

	s.setupSuiteCatalog()
}

func (s *ProviderSuite) TearDownSuite() {
	ctx := context.Background()
	s.Require().NoError(wait.Poll(time.Second, 30*time.Second, func() (done bool, err error) {
		if err = s.masterClient.Delete(ctx, s.provider); err != nil {
			if errors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	}), "could not delete the Provider")

	s.tearDownSuiteCatalog()
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

func (s *ProviderSuite) TestCatalogCreationAndDeletion() {
	ctx := context.Background()

	catalog := &catalogv1alpha1.Catalog{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-catalog",
			Namespace: s.provider.Status.NamespaceName,
		},
		Spec: catalogv1alpha1.CatalogSpec{
			CatalogEntrySelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"kubecarrier.io/test": "label",
				},
			},
			TenantReferenceSelector: &metav1.LabelSelector{},
		},
	}

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

	// Check the Offering object is created.
	offeringFound := &catalogv1alpha1.Offering{}
	s.NoError(wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
		if err := s.masterClient.Get(ctx, types.NamespacedName{
			Name:      s.catalogEntry.Name,
			Namespace: s.tenant.Status.NamespaceName,
		}, offeringFound); err != nil {
			if errors.IsNotFound(err) {
				return false, nil
			}
			return true, err
		}
		return len(offeringFound.Offering.CRDs) == len(s.catalogEntry.Status.CRDs) && offeringFound.Offering.Provider.Name == s.provider.Name, nil
	}), "getting the Offering error")

	// Check if the status will be updated when tenant is removed.
	s.Run("Catalog status updates when adding and removing Tenant", func() {
		// Remove the tenant
		s.Require().NoError(wait.Poll(time.Second, 30*time.Second, func() (done bool, err error) {
			if err = s.masterClient.Delete(ctx, s.tenant); err != nil {
				if errors.IsNotFound(err) {
					return true, nil
				}
				return false, err
			}
			return false, nil
		}), "could not delete the Tenant")

		catalogCheck := &catalogv1alpha1.Catalog{}
		s.NoError(wait.Poll(time.Second, 30*time.Second, func() (done bool, err error) {
			if err := s.masterClient.Get(ctx, types.NamespacedName{
				Name:      catalog.Name,
				Namespace: catalog.Namespace,
			}, catalogCheck); err != nil {
				if errors.IsNotFound(err) {
					return false, nil
				}
				return true, err
			}
			return len(catalogCheck.Status.Tenants) == 0, nil
		}), len(catalogCheck.Status.Tenants))

		// Recreate the tenant
		s.tenant = &catalogv1alpha1.Tenant{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-tenant2",
				Namespace: "kubecarrier-system",
			},
		}
		s.Require().NoError(s.masterClient.Create(ctx, s.tenant), "creating tenant error")

		s.NoError(wait.Poll(time.Second, 30*time.Second, func() (done bool, err error) {
			if err := s.masterClient.Get(ctx, types.NamespacedName{
				Name:      catalog.Name,
				Namespace: catalog.Namespace,
			}, catalogCheck); err != nil {
				if errors.IsNotFound(err) {
					return false, nil
				}
				return true, err
			}
			return len(catalogCheck.Status.Tenants) == 1 && catalogCheck.Status.Tenants[0].Name == s.tenant.Name, nil
		}), "getting the Catalog error")
	})

	s.Require().NoError(wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
		if err = s.masterClient.Delete(ctx, catalog); err != nil {
			if errors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	}), "deleting the Catalog error")

	// Offering object should also be removed
	offeringCheck := &catalogv1alpha1.Offering{}
	s.True(errors.IsNotFound(s.masterClient.Get(ctx, types.NamespacedName{
		Name:      offeringFound.Name,
		Namespace: offeringFound.Namespace,
	}, offeringCheck)), "offering object should also be deleted.")
}

func (s *ProviderSuite) setupSuiteCatalog() {
	ctx := context.Background()

	// Create a Tenant to execute our tests in
	s.tenant = &catalogv1alpha1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-tenant1",
			Namespace: "kubecarrier-system",
		},
	}
	s.Require().NoError(s.masterClient.Create(ctx, s.tenant), "creating tenant error")

	// wait for tenant to be ready
	s.Require().NoError(wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
		if err := s.masterClient.Get(ctx, types.NamespacedName{
			Name:      s.tenant.Name,
			Namespace: s.tenant.Namespace,
		}, s.tenant); err != nil {
			return true, err
		}

		cond, _ := s.tenant.Status.GetCondition(catalogv1alpha1.TenantReady)
		return cond.Status == catalogv1alpha1.ConditionTrue, nil
	}), "waiting for tenant to be ready")
	// wait for the TenantReference to be created.
	tenantReference := &catalogv1alpha1.TenantReference{}
	s.NoError(wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
		if err := s.masterClient.Get(ctx, types.NamespacedName{
			Name:      s.tenant.Name,
			Namespace: s.provider.Status.NamespaceName,
		}, tenantReference); err != nil {
			if errors.IsNotFound(err) {
				return false, nil
			}
			return true, err

		}
		return true, nil
	}), "waiting for the tenantReference to be created")

	// Create CRDs to execute tests
	s.crd = &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "couchdbs.eu-west-1.example.cloud",
			Annotations: map[string]string{
				"kubecarrier.io/service-cluster": "eu-west-1",
			},
			Labels: map[string]string{
				"kubecarrier.io/provider": s.provider.Name,
			},
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "eu-west-1.example.cloud",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Plural: "couchdbs",
				Kind:   "CouchDB",
			},
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name:    "v1alpha1",
					Storage: true,
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
							Type: "object",
						},
					},
				},
			},
			Scope: apiextensionsv1.ClusterScoped,
		},
	}
	s.Require().NoError(s.masterClient.Create(ctx, s.crd), fmt.Sprintf("creating CRD: %s error", s.crd.Name))

	// Create a CatalogEntry to execute our tests in
	s.catalogEntry = &catalogv1alpha1.CatalogEntry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "couchdbs",
			Namespace: s.provider.Status.NamespaceName,
			Labels: map[string]string{
				"kubecarrier.io/test": "label",
			},
		},
		Spec: catalogv1alpha1.CatalogEntrySpec{
			Metadata: catalogv1alpha1.CatalogEntryMetadata{
				DisplayName: "Couch DB",
				Description: "The comfy nosql database",
			},
		},
	}
	s.Require().NoError(s.masterClient.Create(ctx, s.catalogEntry), "could not create CatalogEntry")

	// wait for catalogEntry to be ready
	s.Require().NoError(wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
		if err := s.masterClient.Get(ctx, types.NamespacedName{
			Name:      s.catalogEntry.Name,
			Namespace: s.catalogEntry.Namespace,
		}, s.catalogEntry); err != nil {
			return true, err
		}

		cond, _ := s.catalogEntry.Status.GetCondition(catalogv1alpha1.CatalogEntryReady)
		return cond.Status == catalogv1alpha1.ConditionTrue, nil
	}), "waiting for catalogEntry to be ready")

}

func (s *ProviderSuite) tearDownSuiteCatalog() {
	ctx := context.Background()
	s.Require().NoError(wait.Poll(time.Second, 30*time.Second, func() (done bool, err error) {
		if err = s.masterClient.Delete(ctx, s.tenant); err != nil {
			if errors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	}), "could not delete the Tenant")

	s.Require().NoError(wait.Poll(time.Second, 30*time.Second, func() (done bool, err error) {
		if err = s.masterClient.Delete(ctx, s.catalogEntry); err != nil {
			if errors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	}), "could not delete the CatalogEntry")

	s.Require().NoError(wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
		if err = s.masterClient.Delete(ctx, s.crd); err != nil {
			if errors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	}), fmt.Sprintf("deleting CRD: %s error", s.crd.Name))
}

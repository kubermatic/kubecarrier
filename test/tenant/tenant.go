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

package tenant

import (
	"context"
	"fmt"
	"time"

	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	"github.com/kubermatic/kubecarrier/test/framework"
)

var (
	_ suite.SetupAllSuite    = (*TenantSuite)(nil)
	_ suite.TearDownAllSuite = (*TenantSuite)(nil)
)

// TenantSuite verify Tenant related operations (CatalogEntries, creating service instances).
type TenantSuite struct {
	suite.Suite
	*framework.Framework

	masterClient  client.Client
	serviceClient client.Client

	// Objects that used in this test suite.
	provider *catalogv1alpha1.Provider
	crds     []*apiextensionsv1.CustomResourceDefinition
}

func (s *TenantSuite) SetupSuite() {
	var err error
	ctx := context.Background()
	s.masterClient, err = s.MasterClient()
	s.Require().NoError(err, "creating master client")
	s.serviceClient, err = s.ServiceClient()
	s.Require().NoError(err, "creating service client")

	// Create a Provider to execute tests
	s.provider = &catalogv1alpha1.Provider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-cloud",
			Namespace: "kubecarrier-system",
		},
	}
	s.Require().NoError(s.masterClient.Create(ctx, s.provider), "creating provider error")

	// Try to get the namespace that created for this provider.
	namespace := &corev1.Namespace{}
	s.NoError(wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
		if err := s.masterClient.Get(ctx, types.NamespacedName{
			Name: fmt.Sprintf("provider-%s", s.provider.Name),
		}, namespace); err != nil {
			if errors.IsNotFound(err) {
				return false, nil
			}
			return true, err

		}
		return true, nil
	}), "getting the namespace for the Provider error")

	// Create CRDs to execute tests
	s.crds = append(s.crds,
		&apiextensionsv1.CustomResourceDefinition{
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
		},
		&apiextensionsv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "couchdbs.us-east-1.example.cloud",
				Annotations: map[string]string{
					"kubecarrier.io/service-cluster": "us-east-1",
				},
				Labels: map[string]string{
					"kubecarrier.io/provider": s.provider.Name,
				},
			},
			Spec: apiextensionsv1.CustomResourceDefinitionSpec{
				Group: "us-east-1.example.cloud",
				Names: apiextensionsv1.CustomResourceDefinitionNames{
					Plural: "couchdbs",
					Kind:   "CouchDB",
				},
				Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
					{
						Name:    "v1",
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
		})

	for _, crd := range s.crds {
		s.Require().NoError(s.masterClient.Create(ctx, crd), fmt.Sprintf("creating CRD: %s error", crd.Name))
	}

}

func (s *TenantSuite) TestCatalogEntryCreationAndDeletion() {
	ctx := context.Background()
	catalogEntry := &catalogv1alpha1.CatalogEntry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "couchdbs",
			Namespace: fmt.Sprintf("provider-%s", s.provider.Name),
		},
		Spec: catalogv1alpha1.CatalogEntrySpec{
			Metadata: catalogv1alpha1.CatalogEntryMetadata{
				DisplayName: "Couch DB",
				Description: "The comfy nosql database",
			},
		},
	}

	// Create the CatalogEntry
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
		return len(catalogEntryFound.Status.CRDs) == 2, nil
	}), "getting the CatalogEntry error")

	// Delete the CatalogEntry
	s.Require().NoError(wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
		if err = s.masterClient.Delete(ctx, catalogEntry); err != nil {
			if errors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	}), "deleting the CatalogEntry error")
}

func (s *TenantSuite) TearDownSuite() {
	ctx := context.Background()
	// Remove the CRDs for testing.
	for _, crd := range s.crds {
		s.Require().NoError(s.masterClient.Delete(ctx, crd), fmt.Sprintf("deleting CRD: %s error", crd.Name))
	}

	// Remove the provider for testing.
	s.Require().NoError(s.masterClient.Delete(ctx, s.provider), "deleting Provider error")
}

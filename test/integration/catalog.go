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

package integration

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func newCatalogSuite(
	f *testutil.Framework,
) func(t *testing.T) {
	return func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		t.Cleanup(cancel)
		managementClient, err := f.ManagementClient(t)
		require.NoError(t, err, "creating management client")
		t.Cleanup(managementClient.CleanUpFunc(ctx))
		serviceClient, err := f.ServiceClient(t)
		require.NoError(t, err, "creating service client")
		t.Cleanup(serviceClient.CleanUpFunc(ctx))
		testName := strings.Replace(strings.ToLower(t.Name()), "/", "-", -1)

		// Create a Tenant to execute our tests in
		tenant := f.NewTenantAccount(testName)
		provider := f.NewProviderAccount(testName)
		require.NoError(t, managementClient.Create(ctx, tenant))
		require.NoError(t, managementClient.Create(ctx, provider))
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, tenant))
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, provider))

		tenantReference := &catalogv1alpha1.TenantReference{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tenant.Name,
				Namespace: provider.Status.Namespace.Name,
			},
		}
		require.NoError(t, testutil.WaitUntilFound(ctx, managementClient, tenantReference))

		// Create CRDs to execute tests
		crd := &apiextensionsv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "couchdbs.eu-west-1.example.cloud",
				Labels: map[string]string{
					"kubecarrier.io/origin-namespace": provider.Status.Namespace.Name,
					"kubecarrier.io/service-cluster":  "eu-west-1",
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
								Properties: map[string]apiextensionsv1.JSONSchemaProps{
									"apiVersion": {Type: "string"},
									"kind":       {Type: "string"},
									"metadata":   {Type: "object"},
									"spec": {
										Type: "object",
										Properties: map[string]apiextensionsv1.JSONSchemaProps{
											"prop1": {Type: "string"},
											"prop2": {Type: "string"},
										},
									},
									"status": {
										Type: "object",
										Properties: map[string]apiextensionsv1.JSONSchemaProps{
											"observedGeneration": {Type: "integer"},
											"prop1":              {Type: "string"},
											"prop2":              {Type: "string"},
										},
									},
								},
								Type: "object",
							},
						},
					},
				},
				Scope: apiextensionsv1.NamespaceScoped,
			},
		}
		require.NoError(
			t, managementClient.Create(ctx, crd), fmt.Sprintf("creating CRD: %s error", crd.Name))

		// Create a CatalogEntry to execute our tests in
		catalogEntry := &catalogv1alpha1.CatalogEntry{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "couchdbs",
				Namespace: provider.Status.Namespace.Name,
				Labels: map[string]string{
					"kubecarrier.io/test": "label",
				},
			},
			Spec: catalogv1alpha1.CatalogEntrySpec{
				Metadata: catalogv1alpha1.CatalogEntryMetadata{
					DisplayName: "Couch DB",
					Description: "The comfy nosql database",
				},
				BaseCRD: catalogv1alpha1.ObjectReference{
					Name: crd.Name,
				},
			},
		}
		require.NoError(
			t, managementClient.Create(ctx, catalogEntry), "could not create CatalogEntry")
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, catalogEntry))

		// Create a ServiceCluster to execute our tests in
		serviceCluster := &corev1alpha1.ServiceCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "eu-west-1",
				Namespace: provider.Status.Namespace.Name,
			},
			Spec: corev1alpha1.ServiceClusterSpec{
				Metadata: corev1alpha1.ServiceClusterMetadata{
					DisplayName: "eu-west-1",
					Description: "eu-west-1 service cluster!",
				},
			},
		}
		require.NoError(
			t, managementClient.Create(ctx, serviceCluster), "could not create ServiceCluster")

		// Catalog
		// Test case
		catalog := &catalogv1alpha1.Catalog{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-catalog",
				Namespace: provider.Status.Namespace.Name,
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
		require.NoError(t, managementClient.Create(ctx, catalog), "creating Catalog error")

		// Check the status of the Catalog.
		catalogFound := &catalogv1alpha1.Catalog{}
		assert.NoError(t, wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
			if err := managementClient.Get(ctx, types.NamespacedName{
				Name:      catalog.Name,
				Namespace: catalog.Namespace,
			}, catalogFound); err != nil {
				if errors.IsNotFound(err) {
					return false, nil
				}
				return true, err
			}
			return len(catalogFound.Status.Entries) == 1 && len(catalogFound.Status.Tenants) > 0, nil
		}), "getting the Catalog error")

		// Check the Offering object is created.
		offeringFound := &catalogv1alpha1.Offering{
			ObjectMeta: metav1.ObjectMeta{
				Name:      catalogEntry.Name,
				Namespace: tenant.Status.Namespace.Name,
			},
		}
		assert.NoError(t, testutil.WaitUntilFound(ctx, managementClient, offeringFound))

		// Check the ProviderReference object is created.
		providerReferenceFound := &catalogv1alpha1.ProviderReference{
			ObjectMeta: metav1.ObjectMeta{
				Name:      provider.Name,
				Namespace: tenant.Status.Namespace.Name,
			},
		}
		require.NoError(t, testutil.WaitUntilFound(ctx, managementClient, providerReferenceFound), "getting the ProviderReference error")
		assert.Equal(t, providerReferenceFound.Spec.Metadata.DisplayName, provider.Spec.Metadata.DisplayName)
		assert.Equal(t, providerReferenceFound.Spec.Metadata.Description, provider.Spec.Metadata.Description)

		// Check the ServiceClusterReference object is created.
		serviceClusterReferenceFound := &catalogv1alpha1.ServiceClusterReference{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s.%s", serviceCluster.Name, provider.Name),
				Namespace: tenant.Status.Namespace.Name,
			},
		}
		require.NoError(t, testutil.WaitUntilFound(ctx, managementClient, serviceClusterReferenceFound), "getting the ServiceClusterReference error")
		assert.Equal(t, serviceClusterReferenceFound.Spec.Provider.Name, provider.Name)
		assert.Equal(t, serviceClusterReferenceFound.Spec.Metadata.Description, serviceCluster.Spec.Metadata.Description)

		// Check the ServiceClusterAssignment object is created.
		serviceClusterAssignmentFound := &corev1alpha1.ServiceClusterAssignment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s.%s", tenant.Status.Namespace.Name, serviceCluster.Name),
				Namespace: provider.Status.Namespace.Name,
			},
		}
		require.NoError(t, testutil.WaitUntilFound(ctx, managementClient, serviceClusterAssignmentFound), "getting the ServiceClusterAssignment error")
		assert.Equal(t, serviceClusterAssignmentFound.Spec.ServiceCluster.Name, serviceCluster.Name)
		assert.Equal(t, serviceClusterAssignmentFound.Spec.ManagementClusterNamespace.Name, tenant.Status.Namespace.Name)

		// Check if the status will be updated when tenant is removed.
		t.Run("Catalog status updates when adding and removing Tenant", func(t *testing.T) {
			// Remove the tenant
			require.NoError(t, managementClient.Delete(ctx, tenant), "deleting Tenant")
			require.NoError(t, testutil.WaitUntilNotFound(ctx, managementClient, tenant))

			catalogCheck := &catalogv1alpha1.Catalog{}
			assert.NoError(t, wait.Poll(time.Second, 30*time.Second, func() (done bool, err error) {
				if err := managementClient.Get(ctx, types.NamespacedName{
					Name:      catalog.Name,
					Namespace: catalog.Namespace,
				}, catalogCheck); err != nil {
					if errors.IsNotFound(err) {
						return false, nil
					}
					return true, err
				}

				for _, t := range catalogCheck.Status.Tenants {
					if t.Name == tenant.Name {
						return false, nil
					}
				}

				return true, nil
			}), catalogCheck.Status.Tenants)

			// Recreate the tenant
			tenant = &catalogv1alpha1.Account{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-tenant2",
				},
				Spec: catalogv1alpha1.AccountSpec{
					Metadata: catalogv1alpha1.AccountMetadata{
						DisplayName: "test tenant 2",
						Description: "A lovely perky tenant from the German capital",
					},
					Roles: []catalogv1alpha1.AccountRole{
						catalogv1alpha1.TenantRole,
					},
				},
			}
			require.NoError(t, managementClient.Create(ctx, tenant), "creating tenant error")
			require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, tenant))

			require.NoError(t, wait.Poll(time.Second, 30*time.Second, func() (done bool, err error) {
				if err := managementClient.Get(ctx, types.NamespacedName{
					Name:      catalog.Name,
					Namespace: catalog.Namespace,
				}, catalogCheck); err != nil {
					if errors.IsNotFound(err) {
						return false, nil
					}
					return true, err
				}
				return len(catalogCheck.Status.Tenants) == 1 && catalogCheck.Status.Tenants[0].Name == tenant.Name, nil
			}), "getting the Catalog error")
		})

		t.Run("cleanup", func(t *testing.T) {
			require.NoError(t, managementClient.Delete(ctx, catalog), "deleting Catalog")
			require.NoError(t, testutil.WaitUntilNotFound(ctx, managementClient, catalog))

			// Offering object should also be removed
			offeringCheck := &catalogv1alpha1.Offering{}
			assert.True(t, errors.IsNotFound(managementClient.Get(ctx, types.NamespacedName{
				Name:      offeringFound.Name,
				Namespace: offeringFound.Namespace,
			}, offeringCheck)), "offering object should also be deleted.")

			// ProviderReference object should also be removed
			providerReferenceCheck := &catalogv1alpha1.ProviderReference{}
			assert.True(t, errors.IsNotFound(managementClient.Get(ctx, types.NamespacedName{
				Name:      providerReferenceFound.Name,
				Namespace: providerReferenceFound.Namespace,
			}, providerReferenceCheck)), "providerReference object should also be deleted.")

			// ServiceClusterReference object should also be removed
			serviceClusterReferenceCheck := &catalogv1alpha1.ServiceClusterReference{}
			assert.True(t, errors.IsNotFound(managementClient.Get(ctx, types.NamespacedName{
				Name:      serviceClusterReferenceFound.Name,
				Namespace: serviceClusterReferenceFound.Namespace,
			}, serviceClusterReferenceCheck)), "serviceClusterReference object should also be deleted.")

			// ServiceClusterAssignment object should also be removed
			serviceClusterAssignmentCheck := &corev1alpha1.ServiceClusterAssignment{}
			assert.True(t, errors.IsNotFound(managementClient.Get(ctx, types.NamespacedName{
				Name:      serviceClusterAssignmentFound.Name,
				Namespace: serviceClusterAssignmentFound.Namespace,
			}, serviceClusterAssignmentCheck)), "serviceClusterAssignment object should also be deleted.")
		})
	}
}

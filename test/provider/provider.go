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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
	"github.com/kubermatic/kubecarrier/test/framework"
)

func NewProviderSuite(f *framework.Framework) func(t *testing.T) {
	return func(t *testing.T) {
		managementClient, err := f.ManagementClient()
		require.NoError(t, err, "creating management client")
		defer managementClient.CleanUp(t)

		ctx := context.Background()

		t.Run("", func(t *testing.T) {
			// parallel-group
			suites := []struct {
				name  string
				suite func(*framework.Framework, *catalogv1alpha1.Account) func(t *testing.T)
			}{
				{
					name:  "DerivedCR",
					suite: NewDerivedCRSuite,
				},
				{
					name:  "Catalog",
					suite: NewCatalogSuite,
				},
				{
					name:  "ServiceCluster",
					suite: NewServiceClusterSuite,
				},
			}

			for _, s := range suites {
				t.Run(s.name, func(t *testing.T) {
					// "for" will reassign s to the next item in the loop
					// so we have to same the value for the current index.
					name := s.name
					suite := s.suite
					t.Parallel()

					provider := &catalogv1alpha1.Account{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-" + strings.ToLower(name),
						},
						Spec: catalogv1alpha1.AccountSpec{
							Metadata: catalogv1alpha1.AccountMetadata{
								DisplayName: "provider",
								Description: "provider test description",
							},
							Roles: []catalogv1alpha1.AccountRole{
								catalogv1alpha1.ProviderRole,
							},
							Subjects: []rbacv1.Subject{
								{
									Kind:     rbacv1.GroupKind,
									APIGroup: "rbac.authorization.k8s.io",
									Name:     "admin",
								},
							},
						},
					}

					require.NoError(t, managementClient.Create(ctx, provider))
					require.NoError(t, testutil.WaitUntilReady(managementClient, provider))

					suite(f, provider)(t)
				})
			}
		})
	}
}

func NewCatalogSuite(
	f *framework.Framework,
	provider *catalogv1alpha1.Account,
) func(t *testing.T) {
	return func(t *testing.T) {
		// Catalog
		// Setup
		managementClient, err := f.ManagementClient()
		require.NoError(t, err, "creating management client")
		defer managementClient.CleanUp(t)

		ctx := context.Background()

		// Create a Tenant to execute our tests in
		tenant := &catalogv1alpha1.Account{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-tenant-catalog",
			},
			Spec: catalogv1alpha1.AccountSpec{
				Metadata: catalogv1alpha1.AccountMetadata{
					DisplayName: "test tenant",
					Description: "A simple, humble test tenant from Berlin",
				},
				Roles: []catalogv1alpha1.AccountRole{
					catalogv1alpha1.TenantRole,
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:     rbacv1.GroupKind,
						APIGroup: "rbac.authorization.k8s.io",
						Name:     "admin",
					},
				},
			},
		}
		require.NoError(
			t, managementClient.Create(ctx, tenant), "creating Tenant error")
		require.NoError(t, testutil.WaitUntilReady(managementClient, tenant))

		// wait for the TenantReference to be created.
		tenantReference := &catalogv1alpha1.TenantReference{}
		require.NoError(
			t, wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
				if err := managementClient.Get(ctx, types.NamespacedName{
					Name:      tenant.Name,
					Namespace: provider.Status.Namespace.Name,
				}, tenantReference); err != nil {
					if errors.IsNotFound(err) {
						return false, nil
					}
					return true, err
				}
				return true, nil
			}), "waiting for the TenantReference to be created")

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
		require.NoError(t, testutil.WaitUntilReady(managementClient, catalogEntry))

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
		offeringFound := &catalogv1alpha1.Offering{}
		assert.NoError(t, wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
			if err := managementClient.Get(ctx, types.NamespacedName{
				Name:      catalogEntry.Name,
				Namespace: tenant.Status.Namespace.Name,
			}, offeringFound); err != nil {
				if errors.IsNotFound(err) {
					return false, nil
				}
				return true, err
			}
			return offeringFound.Offering.CRD.Name == catalogEntry.Status.CRD.Name && offeringFound.Offering.Provider.Name == provider.Name, nil
		}), "getting the Offering error")

		// Check the Provider object is created.
		providerFound := &catalogv1alpha1.Provider{
			ObjectMeta: metav1.ObjectMeta{
				Name:      provider.Name,
				Namespace: tenant.Status.Namespace.Name,
			},
		}
		require.NoError(t, testutil.WaitUntilFound(managementClient, providerFound), "getting the Provider error")
		assert.Equal(t, providerFound.Spec.Metadata.DisplayName, provider.Spec.Metadata.DisplayName)
		assert.Equal(t, providerFound.Spec.Metadata.Description, provider.Spec.Metadata.Description)

		// Check the ServiceClusterReference object is created.
		serviceClusterReferenceFound := &catalogv1alpha1.ServiceClusterReference{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s.%s", serviceCluster.Name, provider.Name),
				Namespace: tenant.Status.Namespace.Name,
			},
		}
		require.NoError(t, testutil.WaitUntilFound(managementClient, serviceClusterReferenceFound), "getting the ServiceClusterReference error")
		assert.Equal(t, serviceClusterReferenceFound.Spec.Provider.Name, provider.Name)
		assert.Equal(t, serviceClusterReferenceFound.Spec.Metadata.Description, serviceCluster.Spec.Metadata.Description)

		// Check the ServiceClusterAssignment object is created.
		serviceClusterAssignmentFound := &corev1alpha1.ServiceClusterAssignment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s.%s", tenant.Status.Namespace.Name, serviceCluster.Name),
				Namespace: provider.Status.Namespace.Name,
			},
		}
		require.NoError(t, testutil.WaitUntilFound(managementClient, serviceClusterAssignmentFound), "getting the ServiceClusterAssignment error")
		assert.Equal(t, serviceClusterAssignmentFound.Spec.ServiceCluster.Name, serviceCluster.Name)
		assert.Equal(t, serviceClusterAssignmentFound.Spec.ManagementClusterNamespace.Name, tenant.Status.Namespace.Name)

		// Check if the status will be updated when tenant is removed.
		t.Run("Catalog status updates when adding and removing Tenant", func(t *testing.T) {
			// Remove the tenant
			require.NoError(t, managementClient.Delete(ctx, tenant), "deleting Tenant")
			require.NoError(t, testutil.WaitUntilNotFound(managementClient, tenant))

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
					Subjects: []rbacv1.Subject{
						{
							Kind:     rbacv1.GroupKind,
							APIGroup: "rbac.authorization.k8s.io",
							Name:     "admin",
						},
					},
				},
			}
			require.NoError(t, managementClient.Create(ctx, tenant), "creating tenant error")
			require.NoError(t, testutil.WaitUntilReady(managementClient, tenant))

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
			require.NoError(t, testutil.WaitUntilNotFound(managementClient, catalog))

			// Offering object should also be removed
			offeringCheck := &catalogv1alpha1.Offering{}
			assert.True(t, errors.IsNotFound(managementClient.Get(ctx, types.NamespacedName{
				Name:      offeringFound.Name,
				Namespace: offeringFound.Namespace,
			}, offeringCheck)), "offering object should also be deleted.")

			// Provider object should also be removed
			providerCheck := &catalogv1alpha1.Provider{}
			assert.True(t, errors.IsNotFound(managementClient.Get(ctx, types.NamespacedName{
				Name:      providerFound.Name,
				Namespace: providerFound.Namespace,
			}, providerCheck)), "provider object should also be deleted.")

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

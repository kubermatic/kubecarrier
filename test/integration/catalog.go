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
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func newCatalogSuite(
	f *testutil.Framework,
) func(t *testing.T) {
	return func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		t.Cleanup(cancel)
		managementClient, err := f.ManagementClient(t)
		require.NoError(t, err, "creating management client")
		t.Cleanup(managementClient.CleanUpFunc(ctx))
		serviceClient, err := f.ServiceClient(t)
		require.NoError(t, err, "creating service client")
		t.Cleanup(serviceClient.CleanUpFunc(ctx))
		testName := strings.Replace(strings.ToLower(t.Name()), "/", "-", -1)

		// Create a Tenant to execute our tests in
		tenantAccount := f.NewTenantAccount(testName, rbacv1.Subject{
			Kind:     rbacv1.GroupKind,
			APIGroup: "rbac.authorization.k8s.io",
			Name:     "admin",
		})
		provider := f.NewProviderAccount(testName)
		require.NoError(t, managementClient.Create(ctx, tenantAccount), "creating Tenant error")
		require.NoError(t, managementClient.Create(ctx, provider), "creating Tenant error")

		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, tenantAccount))
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, provider))

		// wait for the Tenant to be created.
		require.NoError(t, testutil.WaitUntilFound(ctx, managementClient, &catalogv1alpha1.Tenant{ObjectMeta: metav1.ObjectMeta{
			Name:      tenantAccount.Name,
			Namespace: provider.Status.Namespace.Name,
		}}))

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
				TenantSelector: &metav1.LabelSelector{},
			},
		}
		require.NoError(t, managementClient.Create(ctx, catalog), "creating Catalog error")

		// Check the status of the Catalog.
		assert.NoError(t, managementClient.WaitUntil(ctx, catalog, func() (b bool, err error) {
			return len(catalog.Status.Entries) == 1 && len(catalog.Status.Tenants) > 0, nil
		}))

		// Check the Offering object is created.
		offeringFound := &catalogv1alpha1.Offering{
			ObjectMeta: metav1.ObjectMeta{
				Name:      catalogEntry.Status.TenantCRD.Name,
				Namespace: tenantAccount.Status.Namespace.Name,
			},
		}
		assert.NoError(t, testutil.WaitUntilFound(ctx, managementClient, offeringFound))

		// Check the Provider object is created.
		providerFound := &catalogv1alpha1.Provider{
			ObjectMeta: metav1.ObjectMeta{
				Name:      provider.Name,
				Namespace: tenantAccount.Status.Namespace.Name,
			},
		}
		require.NoError(t, testutil.WaitUntilFound(ctx, managementClient, providerFound), "getting the Provider error")
		assert.Equal(t, providerFound.Spec.Metadata.DisplayName, provider.Spec.Metadata.DisplayName)
		assert.Equal(t, providerFound.Spec.Metadata.Description, provider.Spec.Metadata.Description)

		// Check the ServiceClusterReference object is created.
		serviceClusterReferenceFound := &catalogv1alpha1.ServiceClusterReference{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s.%s", serviceCluster.Name, provider.Name),
				Namespace: tenantAccount.Status.Namespace.Name,
			},
		}
		require.NoError(t, testutil.WaitUntilFound(ctx, managementClient, serviceClusterReferenceFound), "getting the ServiceClusterReference error")
		assert.Equal(t, serviceClusterReferenceFound.Spec.Provider.Name, provider.Name)
		assert.Equal(t, serviceClusterReferenceFound.Spec.Metadata.Description, serviceCluster.Spec.Metadata.Description)

		// Check the ServiceClusterAssignment object is created.
		serviceClusterAssignmentFound := &corev1alpha1.ServiceClusterAssignment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s.%s", tenantAccount.Status.Namespace.Name, serviceCluster.Name),
				Namespace: provider.Status.Namespace.Name,
			},
		}
		require.NoError(t, testutil.WaitUntilFound(ctx, managementClient, serviceClusterAssignmentFound), "getting the ServiceClusterAssignment error")
		assert.Equal(t, serviceClusterAssignmentFound.Spec.ServiceCluster.Name, serviceCluster.Name)
		assert.Equal(t, serviceClusterAssignmentFound.Spec.ManagementClusterNamespace.Name, tenantAccount.Status.Namespace.Name)

		// Check Provider Role
		providerRoleFound := &rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("kubecarrier:provider:%s", catalogEntry.Name),
				Namespace: tenantAccount.Status.Namespace.Name,
			},
		}
		require.NoError(t, testutil.WaitUntilFound(ctx, managementClient, providerRoleFound), "getting Provider Role error")
		assert.Contains(t, providerRoleFound.Rules, rbacv1.PolicyRule{
			Verbs:     []string{rbacv1.VerbAll},
			APIGroups: []string{catalogEntry.Status.ProviderCRD.APIGroup},
			Resources: []string{catalogEntry.Status.ProviderCRD.Plural},
		}, "Missing the PolicyRule")

		// Check Provider RoleBinding
		providerRoleBindingFound := &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("kubecarrier:provider:%s", catalogEntry.Name),
				Namespace: tenantAccount.Status.Namespace.Name,
			},
		}
		require.NoError(t, testutil.WaitUntilFound(ctx, managementClient, providerRoleBindingFound), "getting Provider RoleBinding error")
		assert.Equal(t, providerRoleBindingFound.Subjects, provider.Spec.Subjects, "Subjects is different")

		// Check Tenant Role
		tenantRoleFound := &rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("kubecarrier:tenant:%s", catalogEntry.Name),
				Namespace: tenantAccount.Status.Namespace.Name,
			},
		}
		require.NoError(t, testutil.WaitUntilFound(ctx, managementClient, tenantRoleFound), "getting Tenant Role error")
		assert.Contains(t, tenantRoleFound.Rules, rbacv1.PolicyRule{
			Verbs:     []string{rbacv1.VerbAll},
			APIGroups: []string{catalogEntry.Status.TenantCRD.APIGroup},
			Resources: []string{catalogEntry.Status.TenantCRD.Plural},
		}, "Missing the PolicyRule")

		// Check Tenant RoleBinding
		tenantRoleBindingFound := &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("kubecarrier:tenant:%s", catalogEntry.Name),
				Namespace: tenantAccount.Status.Namespace.Name,
			},
		}
		require.NoError(t, testutil.WaitUntilFound(ctx, managementClient, tenantRoleBindingFound), "getting Tenant RoleBinding error")
		assert.Equal(t, tenantRoleBindingFound.Subjects, tenantAccount.Spec.Subjects, "Subjects is different")

		// Check if the status will be updated when tenant is removed.
		t.Log("===== Catalog status updates when adding and removing Tenant =====")
		// Remove the tenant
		require.NoError(t, managementClient.Delete(ctx, tenantAccount), "deleting Tenant")
		require.NoError(t, testutil.WaitUntilNotFound(ctx, managementClient, tenantAccount))

		assert.NoError(t, managementClient.WaitUntil(ctx, catalog, func() (b bool, err error) {
			for _, t := range catalog.Status.Tenants {
				if t.Name == tenantAccount.Name {
					return false, nil
				}
			}
			return true, nil
		}))

		// Recreate the tenant
		tenantAccount = f.NewTenantAccount(testName, rbacv1.Subject{
			Kind:     rbacv1.GroupKind,
			APIGroup: "rbac.authorization.k8s.io",
			Name:     "admin",
		})
		require.NoError(t, managementClient.Create(ctx, tenantAccount), "creating tenant error")
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, tenantAccount))

		assert.NoError(t, managementClient.WaitUntil(ctx, catalog, func() (done bool, err error) {
			for _, t := range catalog.Status.Tenants {
				if t.Name == tenantAccount.Name {
					return true, nil
				}
			}
			return
		}))

		t.Log("===== cleanup =====")

		require.NoError(t, managementClient.Delete(ctx, catalog), "deleting Catalog")
		require.NoError(t, testutil.WaitUntilNotFound(ctx, managementClient, catalog))

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

		// Check Provider Role
		providerRoleCheck := &rbacv1.Role{}
		assert.True(t, errors.IsNotFound(managementClient.Get(ctx, types.NamespacedName{
			Name:      providerRoleFound.Name,
			Namespace: providerRoleFound.Namespace,
		}, providerRoleCheck)), "provider Role should be deleted")

		// Check Provider RoleBinding
		providerRoleBindingCheck := &rbacv1.RoleBinding{}
		assert.True(t, errors.IsNotFound(managementClient.Get(ctx, types.NamespacedName{
			Name:      providerRoleBindingFound.Name,
			Namespace: providerRoleBindingFound.Namespace,
		}, providerRoleBindingCheck)), "provider RoleBinding should be deleted")

		// Check Tenant Role
		tenantRoleCheck := &rbacv1.Role{}
		assert.True(t, errors.IsNotFound(managementClient.Get(ctx, types.NamespacedName{
			Name:      tenantRoleFound.Name,
			Namespace: tenantRoleFound.Namespace,
		}, tenantRoleCheck)), "tenant Role should be deleted")

		// Check Tenant RoleBinding
		tenantRoleBindingCheck := &rbacv1.RoleBinding{}
		assert.True(t, errors.IsNotFound(managementClient.Get(ctx, types.NamespacedName{
			Name:      tenantRoleBindingFound.Name,
			Namespace: tenantRoleBindingFound.Namespace,
		}, tenantRoleBindingCheck)), "tenant RoleBinding should be deleted")
	}
}

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
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func newDerivedCR(
	f *testutil.Framework,
) func(t *testing.T) {
	return func(t *testing.T) {
		managementClient, err := f.ManagementClient(t)
		require.NoError(t, err, "creating management client")
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		t.Cleanup(cancel)
		t.Cleanup(managementClient.CleanUpFunc(ctx))

		testName := strings.Replace(strings.ToLower(t.Name()), "/", "-", -1)
		provider := f.NewProviderAccount(testName, rbacv1.Subject{
			Kind:     rbacv1.GroupKind,
			APIGroup: "rbac.authorization.k8s.io",
			Name:     "provider",
		})
		require.NoError(t, managementClient.Create(ctx, provider))
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, provider))

		baseCRD := &apiextensionsv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "catapults.test.kubecarrier.io",
				Labels: map[string]string{
					"kubecarrier.io/service-cluster":  "eu-west-1",
					"kubecarrier.io/origin-namespace": provider.Status.Namespace.Name,
				},
			},
			Spec: apiextensionsv1.CustomResourceDefinitionSpec{
				Group: "test.kubecarrier.io",
				Names: apiextensionsv1.CustomResourceDefinitionNames{
					Kind:     "Catapult",
					ListKind: "CatapultList",
					Plural:   "catapults",
					Singular: "catapult",
				},
				Scope: apiextensionsv1.NamespaceScoped,
				Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
					{
						Name:    "v1alpha1",
						Served:  true,
						Storage: true,
						Subresources: &apiextensionsv1.CustomResourceSubresources{
							Status: &apiextensionsv1.CustomResourceSubresourceStatus{},
						},
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
			},
		}
		// create base CRD
		require.NoError(t, managementClient.Create(ctx, baseCRD), "creating base CRD")

		// Test
		// Create a CatalogEntry to execute our tests in
		catalogEntry := &catalogv1alpha1.CatalogEntry{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: provider.Status.Namespace.Name,
			},
			Spec: catalogv1alpha1.CatalogEntrySpec{
				Metadata: catalogv1alpha1.CatalogEntryMetadata{
					DisplayName: "Catapult",
					Description: "Catapult",
				},
				BaseCRD: catalogv1alpha1.ObjectReference{
					Name: baseCRD.Name,
				},
				Derive: &catalogv1alpha1.DerivedConfig{
					KindOverride: "TestResource",
					Expose: []catalogv1alpha1.VersionExposeConfig{
						{
							Versions: []string{
								"v1alpha1",
							},
							Fields: []catalogv1alpha1.FieldPath{
								{JSONPath: ".spec.prop1"},
								{JSONPath: ".status.observedGeneration"},
								{JSONPath: ".status.prop1"},
							},
						},
					},
				},
			},
		}

		require.NoError(
			t, managementClient.Create(ctx, catalogEntry), "creating CatalogEntry")

		// Wait for the CatalogEntry to be ready, it takes more time since it requires the
		// DerivedCustomResource object and Elevator get ready
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, catalogEntry))

		// Check the DerivedCustomResource Object
		dcr := &catalogv1alpha1.DerivedCustomResource{}
		require.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name:      catalogEntry.Name,
			Namespace: catalogEntry.Namespace,
		}, dcr), "getting derived CRD")
		// Check reported status
		if assert.NotNil(t, dcr.Status.DerivedCR, ".status.derivedCR should be set") &&
			assert.NotNil(t, catalogEntry.Status.TenantCRD, ".status.CRD should be set") {
			assert.Equal(t, catalogEntry.Status.TenantCRD.Name, dcr.Status.DerivedCR.Name)
			assert.Equal(t, catalogEntry.Status.TenantCRD.APIGroup, dcr.Status.DerivedCR.Group)
			assert.Equal(t, catalogEntry.Status.TenantCRD.Kind, dcr.Status.DerivedCR.Kind)
			assert.Equal(t, catalogEntry.Status.TenantCRD.Plural, dcr.Status.DerivedCR.Plural)
		}
		err = managementClient.Delete(ctx, provider)
		if assert.Error(t, err, "dirty provider %s deletion should error out", provider.Name) {
			assert.Equal(t,
				fmt.Sprintf(`admission webhook "vaccount.kubecarrier.io" denied the request: deletion blocking objects found:
DerivedCustomResource.catalog.kubecarrier.io/v1alpha1: %s
`, dcr.Name),
				err.Error(),
				"deleting dirty provider %s", provider.Name)
		}

		// Check created CRD
		crd := &apiextensionsv1.CustomResourceDefinition{}
		require.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name: dcr.Status.DerivedCR.Name,
		}, crd), "getting derived CRD")

		schemaYaml, _ := yaml.Marshal(crd.Spec.Versions[0].Schema.OpenAPIV3Schema)
		assert.Equal(t, `properties:
  apiVersion:
    type: string
  kind:
    type: string
  metadata:
    type: object
  spec:
    properties:
      prop1:
        type: string
    type: object
  status:
    properties:
      observedGeneration:
        type: integer
      prop1:
        type: string
    type: object
type: object
`, string(schemaYaml))

		// Create a Tenant obj
		someNamespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: testName + "derived-crd-test-namespace",
			},
		}
		require.NoError(
			t, managementClient.Create(ctx, someNamespace), "creating a Namespace")

		// to be able to work with the new CRD, we have to re-create the client
		managementClient, err = f.ManagementClient(t)
		require.NoError(t, err, "recreating management client")
		t.Cleanup(managementClient.CleanUpFunc(ctx))

		// Check Tenant -> Provider
		tenantObj := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": fmt.Sprintf("eu-west-1.%s/v1alpha1", provider.Status.Namespace.Name),
				"kind":       "TestResource",
				"metadata": map[string]interface{}{
					"name":      "test-instance-1",
					"namespace": someNamespace.Name,
				},
				"spec": map[string]interface{}{
					"prop1": "test1",
				},
			},
		}
		require.NoError(
			t, managementClient.Create(ctx, tenantObj), "creating a TestResource")

		providerObj := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "test.kubecarrier.io/v1alpha1",
				"kind":       "Catapult",
				"metadata": map[string]interface{}{
					"name":      "test-instance-1",
					"namespace": someNamespace.Name,
				},
			},
		}
		require.NoError(t, testutil.WaitUntilFound(ctx, managementClient, providerObj))

		// Check Provider -> Tenant
		providerObj2 := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "test.kubecarrier.io/v1alpha1",
				"kind":       "Catapult",
				"metadata": map[string]interface{}{
					"name":      "test-instance-2",
					"namespace": someNamespace.Name,
				},
				"spec": map[string]interface{}{
					"prop1": "test1",
					"prop2": "test1",
				},
			},
		}
		require.NoError(
			t, managementClient.Create(ctx, providerObj2), "creating a Catapult")

		tenantObj2 := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "test.kubecarrier.io/v1alpha1",
				"kind":       "Catapult",
				"metadata": map[string]interface{}{
					"name":      "test-instance-2",
					"namespace": someNamespace.Name,
				},
				"spec": map[string]interface{}{
					"prop1": "test1",
				},
			},
		}
		require.NoError(t, testutil.WaitUntilFound(ctx, managementClient, tenantObj2))
	}
}

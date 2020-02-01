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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
	"github.com/kubermatic/kubecarrier/test/framework"
)

func NewDerivedCRDSuite(
	f *framework.Framework,
	provider *catalogv1alpha1.Provider,
) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		// Setup
		//
		masterClient, err := f.MasterClient()
		require.NoError(t, err, "creating master client")
		defer masterClient.CleanUp(t)

		ctx := context.Background()

		baseCRD := &apiextensionsv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "catapults.test.kubecarrier.io",
				Labels: map[string]string{
					"kubecarrier.io/service-cluster": "eu-west-1",
					"kubecarrier.io/provider":        provider.Name,
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
		require.NoError(t, masterClient.Create(ctx, baseCRD), "creating base CRD")

		// Test
		//
		dcrd := &catalogv1alpha1.DerivedCustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: provider.Status.NamespaceName,
			},
			Spec: catalogv1alpha1.DerivedCustomResourceDefinitionSpec{
				BaseCRD: catalogv1alpha1.ObjectReference{
					Name: baseCRD.Name,
				},
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
		}
		require.NoError(
			t, masterClient.Create(ctx, dcrd), "creating DerivedCustomResourceDefinition")

		// Wait for DCRD to be ready
		require.NoError(t, testutil.WaitUntilReady(masterClient, dcrd))

		// Check reported status
		if assert.NotNil(t, dcrd.Status.DerivedCRD, ".status.derivedCRD should be set") {
			assert.Equal(t, "testresources.eu-west-1.example-cloud", dcrd.Status.DerivedCRD.Name)
			assert.Equal(t, "eu-west-1.example-cloud", dcrd.Status.DerivedCRD.Group)
			assert.Equal(t, "TestResource", dcrd.Status.DerivedCRD.Kind)
			assert.Equal(t, "testresources", dcrd.Status.DerivedCRD.Plural)
			assert.Equal(t, "testresource", dcrd.Status.DerivedCRD.Singular)
		}

		// Check created CRD
		crd := &apiextensionsv1.CustomResourceDefinition{}
		require.NoError(t, masterClient.Get(ctx, types.NamespacedName{
			Name: dcrd.Status.DerivedCRD.Name,
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

		// Create Elevator instance for new CRD
		elevator := &operatorv1alpha1.Elevator{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "db.eu-west-1",
				Namespace: provider.Status.NamespaceName,
			},
			Spec: operatorv1alpha1.ElevatorSpec{
				ProviderCRD: operatorv1alpha1.CRDReference{
					Kind:    "Catapult",
					Version: "v1alpha1",
					Group:   "test.kubecarrier.io",
					Plural:  "catapults",
				},
				TenantCRD: operatorv1alpha1.CRDReference{
					Kind:    "TestResource",
					Version: "v1alpha1",
					Group:   "eu-west-1.example-cloud",
					Plural:  "testresources",
				},
				DerivedCRD: operatorv1alpha1.ObjectReference{
					Name: dcrd.Name,
				},
			},
		}

		require.NoError(
			t, masterClient.Create(ctx, elevator), "creating Elevator error")
		require.NoError(t, testutil.WaitUntilReady(masterClient, elevator))

		// Check created objects
		elevatorDeployment := &appsv1.Deployment{}
		assert.NoError(t, masterClient.Get(ctx, types.NamespacedName{
			Name:      "db-eu-west-1-elevator-manager",
			Namespace: elevator.Namespace,
		}, elevatorDeployment), "getting Elevator manager deployment error")

		// Create a Tenant obj
		someNamespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "derived-crd-test-namespace",
			},
		}
		require.NoError(
			t, masterClient.Create(ctx, someNamespace), "creating a Namespace")

		// to be able to work with the new CRD, we have to re-create the client
		masterClient, err = f.MasterClient()
		require.NoError(t, err, "recreating master client")

		// Check Tenant -> Provider
		tenantObj := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "eu-west-1.example-cloud/v1alpha1",
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
			t, masterClient.Create(ctx, tenantObj), "creating a TestResource")

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
		require.NoError(t, testutil.WaitUntilFound(masterClient, providerObj))

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
			t, masterClient.Create(ctx, providerObj2), "creating a Catapult")

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
		require.NoError(t, testutil.WaitUntilFound(masterClient, tenantObj2))

		// Teardown
		//
		if _, noCleanup := os.LookupEnv("NO_CLEANUP"); noCleanup {
			return
		}

		// Cleanup Elevator
		require.NoError(
			t, masterClient.Delete(ctx, elevator), "deleting Elevator object")
		require.NoError(t, testutil.WaitUntilNotFound(masterClient, elevator))
		require.NoError(t, testutil.WaitUntilNotFound(masterClient, elevatorDeployment))

		// Cleanup DerivedCRD
		require.NoError(t, masterClient.Delete(ctx, dcrd), "deleting the DerivedCustomResourceDefinition object")
		require.NoError(t, testutil.WaitUntilNotFound(masterClient, dcrd))

		// Cleanup base CRD
		require.NoError(t, masterClient.Delete(ctx, baseCRD), "deleting base CRD")
	}
}

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
		ctx, cancel := context.WithCancel(context.Background())
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

		baseCRD := f.NewFakeCouchDBCRD(testName + "test.kubecarrier.io")
		baseCRD.Labels = map[string]string{
			"kubecarrier.io/service-cluster":  "eu-west-1",
			"kubecarrier.io/origin-namespace": provider.Status.Namespace.Name,
		}
		// create base CRD
		require.NoError(t, managementClient.Create(ctx, baseCRD), "creating base CRD")

		// Check the DerivedCustomResource Object
		dcr := &catalogv1alpha1.DerivedCustomResource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: provider.Status.Namespace.Name,
			},
			Spec: catalogv1alpha1.DerivedCustomResourceSpec{
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

		require.NoError(t, managementClient.Create(ctx, dcr))
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, dcr))

		// Check the Elevator dynamic webhook service is deployed.
		webhookService := &corev1.Service{}
		assert.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-elevator-webhook-service", dcr.Name),
			Namespace: provider.Status.Namespace.Name,
		}, webhookService), "get the Webhook Service that owned by Elevator object")

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
				"apiVersion": fmt.Sprintf("%s/v1alpha1", baseCRD.Spec.Group),
				"kind":       baseCRD.Spec.Names.Kind,
				"metadata": map[string]interface{}{
					"name":      "test-instance-1",
					"namespace": someNamespace.Name,
				},
			},
		}
		require.NoError(t, testutil.WaitUntilFound(ctx, managementClient, providerObj))

		err = managementClient.Delete(ctx, dcr)
		if assert.Error(t, err,
			"derived custom resource must not be allowed to delete if derived CRD instances are present",
		) {
			assert.Contains(
				t,
				err.Error(),
				"derived CRD instances are still present in the cluster",
				"derivedCR deletion webhook should error out on derived CRD instance presence",
			)
		}

		// Check Provider -> Tenant
		providerObj2 := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": fmt.Sprintf("%s/v1alpha1", baseCRD.Spec.Group),
				"kind":       baseCRD.Spec.Names.Kind,
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
				"apiVersion": fmt.Sprintf("%s/v1alpha1", baseCRD.Spec.Group),
				"kind":       baseCRD.Spec.Names.Kind,
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

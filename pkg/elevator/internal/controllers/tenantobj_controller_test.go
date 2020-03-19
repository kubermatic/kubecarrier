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

package controllers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func TestPatch(t *testing.T) {
	tenantObj := &unstructured.Unstructured{Object: map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "test-1",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"test1": "spec2000",
		},
	}}

	log := testutil.NewLogger(t)
	specFields := []catalogv1alpha1.FieldPath{{JSONPath: ".spec.test1"}}
	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"test2": "test2",
		},
	}
	client := fakeclient.NewFakeClientWithScheme(testScheme)
	r := TenantObjReconciler{
		Client:           client,
		Log:              log,
		Scheme:           testScheme,
		NamespacedClient: client,

		ProviderGVK: providerGVK,
		TenantGVK:   tenantGVK,

		DerivedCRName:     dcr.Name,
		ProviderNamespace: providerNamespace,
	}
	tenantObj.SetGroupVersionKind(r.TenantGVK)
	providerObj, err := r.buildProviderObj(tenantObj, specFields, patch)
	require.NoError(t, err)

	wantedProviderObj := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "eu-west-1.provider/v1alpha1",
		"kind":       "CouchDBInternal",
		"metadata": map[string]interface{}{
			"name":      "test-1",
			"namespace": "default",
			"ownerReferences": []interface{}{
				map[string]interface{}{
					"apiVersion":         "eu-west-1.provider/v1alpha1",
					"blockOwnerDeletion": true,
					"controller":         true,
					"kind":               "CouchDB",
					"name":               "test-1",
					"uid":                "",
				},
			},
		},
		"spec": map[string]interface{}{
			"test1": "spec2000",
			"test2": "test2",
		},
	}}
	assert.Equal(t, wantedProviderObj, providerObj)
}

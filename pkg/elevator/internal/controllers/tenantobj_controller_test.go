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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kubermatic/utils/pkg/testutil"
)

func TestTenantObjReconciler(t *testing.T) {
	tenantObj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":       "test-1",
				"namespace":  "another-namespace",
				"generation": int64(2),
			},
			"spec": map[string]interface{}{
				"test1": "spec2000",
			},
		},
	}
	tenantObj.SetGroupVersionKind(tenantGVK)

	t.Run("updates existing obj", func(t *testing.T) {
		providerObj := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":       tenantObj.GetName(),
					"namespace":  tenantObj.GetNamespace(),
					"generation": int64(10),
				},
				"spec": map[string]interface{}{
					"test1": "spec1000",
				},
				"status": map[string]interface{}{
					"test1":              "status2000",
					"observedGeneration": int64(10),
				},
			},
		}
		providerObj.SetGroupVersionKind(providerGVK)

		log := testutil.NewLogger(t)
		client := fakeclient.NewFakeClientWithScheme(
			testScheme, dcr, tenantObj, providerObj)

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

		_, err := r.Reconcile(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      tenantObj.GetName(),
				Namespace: tenantObj.GetNamespace(),
			},
		})
		require.NoError(t, err)

		// Check Provider Obj
		ctx := context.Background()
		checkProviderObj := &unstructured.Unstructured{}
		checkProviderObj.SetGroupVersionKind(providerGVK)
		err = client.Get(ctx, types.NamespacedName{
			Name:      tenantObj.GetName(),
			Namespace: tenantObj.GetNamespace(),
		}, checkProviderObj)
		require.NoError(t, err)

		assert.Equal(t, map[string]interface{}{
			"apiVersion": "eu-west-1.provider/v1alpha1",
			"kind":       "CouchDBInternal",
			"metadata": map[string]interface{}{
				"name":            "test-1",
				"namespace":       "another-namespace",
				"generation":      int64(10),
				"resourceVersion": "1",
				"ownerReferences": []interface{}{
					// owner reference is added by our controler
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
				// should be updated from "spec1000"
				"test1": "spec2000",
			},
			"status": map[string]interface{}{
				"test1":              "status2000",
				"observedGeneration": int64(10),
			},
		}, checkProviderObj.Object)

		// Check Tenant Obj
		checkTenantObj := &unstructured.Unstructured{}
		checkTenantObj.SetGroupVersionKind(tenantGVK)
		err = client.Get(ctx, types.NamespacedName{
			Name:      tenantObj.GetName(),
			Namespace: tenantObj.GetNamespace(),
		}, checkTenantObj)
		require.NoError(t, err)

		assert.Equal(t, map[string]interface{}{
			"apiVersion": "eu-west-1.provider/v1alpha1",
			"kind":       "CouchDB",
			"metadata": map[string]interface{}{
				"name":            "test-1",
				"namespace":       "another-namespace",
				"resourceVersion": "1",
				"generation":      int64(2),
			},
			"spec": map[string]interface{}{
				"test1": "spec2000",
			},
			"status": map[string]interface{}{
				"test1": "status2000",
				// is set to this objects generation,
				// because generation == observedGeneration in provider obj
				"observedGeneration": int64(2),
			},
		}, checkTenantObj.Object)
	})

	t.Run("creates provider obj", func(t *testing.T) {
		log := testutil.NewLogger(t)
		client := fakeclient.NewFakeClientWithScheme(testScheme, dcr, tenantObj)

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

		_, err := r.Reconcile(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      tenantObj.GetName(),
				Namespace: tenantObj.GetNamespace(),
			},
		})
		require.NoError(t, err)

		// Check Provider obj
		ctx := context.Background()
		checkProviderObj := &unstructured.Unstructured{}
		checkProviderObj.SetGroupVersionKind(providerGVK)
		err = client.Get(ctx, types.NamespacedName{
			Name:      tenantObj.GetName(),
			Namespace: tenantObj.GetNamespace(),
		}, checkProviderObj)
		require.NoError(t, err)

		assert.Equal(t, map[string]interface{}{
			"apiVersion": "eu-west-1.provider/v1alpha1",
			"kind":       "CouchDBInternal",
			"metadata": map[string]interface{}{
				"name":            "test-1",
				"namespace":       "another-namespace",
				"resourceVersion": "1",
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
			},
		}, checkProviderObj.Object)
	})
}

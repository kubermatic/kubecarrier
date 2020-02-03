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

	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func TestAdoptionReconciler(t *testing.T) {
	providerObj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":       "test-1",
				"namespace":  "another-namespace",
				"generation": int64(4),
			},
			"spec": map[string]interface{}{
				"test1": "spec2000",
				"test2": "spec2000",
				"test3": "spec2000",
			},
		},
	}
	providerObj.SetGroupVersionKind(providerGVK)

	t.Run("creates tenant object", func(t *testing.T) {
		log := testutil.NewLogger(t)
		client := fakeclient.NewFakeClientWithScheme(
			testScheme, dcrd, providerObj)

		r := AdoptionReconciler{
			Client:           client,
			Log:              log,
			Scheme:           testScheme,
			NamespacedClient: client,

			ProviderGVK: providerGVK,
			TenantGVK:   tenantGVK,

			DerivedCRDName:    dcrd.Name,
			ProviderNamespace: providerNamespace,
		}

		_, err := r.Reconcile(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      providerObj.GetName(),
				Namespace: providerObj.GetNamespace(),
			},
		})
		require.NoError(t, err)

		ctx := context.Background()
		checkTenantObj := &unstructured.Unstructured{}
		checkTenantObj.SetGroupVersionKind(tenantGVK)
		err = client.Get(ctx, types.NamespacedName{
			Name:      providerObj.GetName(),
			Namespace: providerObj.GetNamespace(),
		}, checkTenantObj)
		require.NoError(t, err)

		assert.Equal(t, map[string]interface{}{
			"apiVersion": "eu-west-1.provider/v1alpha1",
			"kind":       "CouchDB",
			"metadata": map[string]interface{}{
				"name":            "test-1",
				"namespace":       "another-namespace",
				"resourceVersion": "1",
			},
			"spec": map[string]interface{}{
				"test1": "spec2000",
			},
		}, checkTenantObj.Object)
	})
}

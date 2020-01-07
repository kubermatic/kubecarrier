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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func TestCatalogEntryReconciler(t *testing.T) {
	catalogEntry := &catalogv1alpha1.CatalogEntry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-catalogEntry",
			Namespace: "kubecarrier-system",
		},
		Spec: catalogv1alpha1.CatalogEntrySpec{
			Metadata: catalogv1alpha1.CatalogEntryMetadata{
				DisplayName: "Test CatalogEntry",
				Description: "Test CatalogEntry",
			},
			CRDSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"kubecarrier.io/provider": "example.provider",
				},
			},
		},
	}

	crd1 := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-crd-1.test-crd-group-1.test",
			Annotations: map[string]string{
				"kubecarrier.io/service-cluster": "test-service-cluster-1",
			},
			Labels: map[string]string{
				"kubecarrier.io/provider": "example.provider",
			},
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "test-crd-group-1",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind: "TestCRD1",
			},
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name: "v1alpha1",
				},
			},
		},
	}

	crd2 := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-crd-2.test-crd-group-2.test",
			Annotations: map[string]string{
				"kubecarrier.io/service-cluster": "test-service-cluster-2",
			},
			Labels: map[string]string{
				"kubecarrier.io/provider": "example.provider",
			},
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "test-crd-group-2",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind: "TestCRD2",
			},
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name: "v1",
				},
			},
		},
	}

	// This CRD is used to test that one crd can only be referenced by one CatalogEntry.
	crd3 := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-crd-3.test-crd-group-3.test",
			Annotations: map[string]string{
				"kubecarrier.io/service-cluster": "test-service-cluster-3",
				catalogEntryReferenceAnnotation:  "test-catalogentry",
			},
			Labels: map[string]string{
				"kubecarrier.io/provider": "example.provider",
			},
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "test-crd-group-3",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind: "TestCRD3",
			},
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name: "v1",
				},
			},
		},
	}

	client := fakeclient.NewFakeClientWithScheme(testScheme, catalogEntry, crd1, crd2, crd3)
	log := testutil.NewLogger(t)
	r := &CatalogEntryReconciler{
		Client: client,
		Log:    log,
		Scheme: testScheme,
	}
	ctx := context.Background()

	catalogEntryFound := &catalogv1alpha1.CatalogEntry{}
	crdFound := &apiextensionsv1.CustomResourceDefinition{}
	if !t.Run("create/update CatalogEntry", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			// Run Reconcile multiple times, because
			// the reconcilation stops after changing the CatalogEntry
			_, err := r.Reconcile(ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      catalogEntry.Name,
					Namespace: catalogEntry.Namespace,
				},
			})
			require.NoError(t, err, "unexpected error returned by Reconcile")
		}

		// Check CatalogEntry
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name:      catalogEntry.Name,
			Namespace: catalogEntry.Namespace,
		}, catalogEntryFound), "error when getting catalogEntry")
		assert.Len(t, catalogEntryFound.Finalizers, 1, "finalizer should be added to CatalogEntry")

		// Check CatalogEntry Status
		assert.Len(t, catalogEntryFound.Status.CRDs, 2, "CRD information should be added to the CatalogEntry.Status.CRDs")
		assert.Contains(t, catalogEntryFound.Status.CRDs[0].Kind, "TestCRD", "CRD Kind is wrong")
		assert.Contains(t, catalogEntryFound.Status.CRDs[0].Name, "test-crd", "CRD Name is wrong")
		assert.Contains(t, catalogEntryFound.Status.CRDs[0].ServiceCluster.Name, "test-service-cluster", "CRD ServiceCluster is wrong")
		assert.Contains(t, catalogEntryFound.Status.CRDs[0].APIGroup, "test-crd-group", "CRD APIGroup is wrong")

		// Check CatalogEntry Conditions
		readyCondition, readyConditionExists := catalogEntryFound.Status.GetCondition(catalogv1alpha1.CatalogEntryReady)
		assert.True(t, readyConditionExists, "Ready Condition is not set")
		assert.Equal(t, catalogv1alpha1.ConditionTrue, readyCondition.Status, "Wrong Ready condition.Status")

		// Check CRDs Annotation
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name: crd1.Name,
		}, crdFound), "error when getting crd")
		assert.Contains(t, crdFound.Annotations, catalogEntryReferenceAnnotation, "the catalogEntry annotation should be added to the CRD")
	}) {
		t.FailNow()
	}

	t.Run("delete CatalogEntry", func(t *testing.T) {
		// Setup
		ts := metav1.Now()
		catalogEntryFound.DeletionTimestamp = &ts
		require.NoError(t, client.Update(ctx, catalogEntryFound), "unexpected error updating catalogEntry")

		for i := 0; i < 5; i++ {
			// Run Reconcile multiple times, because
			// the reconcilation stops after changing the CatalogEntry
			_, err := r.Reconcile(ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      catalogEntryFound.Name,
					Namespace: catalogEntryFound.Namespace,
				},
			})
			require.NoError(t, err, "unexpected error returned by Reconcile")
		}

		catalogEntryCheck := &catalogv1alpha1.CatalogEntry{}
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name:      catalogEntryFound.Name,
			Namespace: catalogEntryFound.Namespace,
		}, catalogEntryCheck), "cannot check CatalogEntry")
		assert.Len(t, catalogEntryCheck.Finalizers, 0, "finalizers should have been removed")

		// Check CatalogEntry Conditions
		readyCondition, readyConditionExists := catalogEntryCheck.Status.GetCondition(catalogv1alpha1.CatalogEntryReady)
		assert.True(t, readyConditionExists, "Ready Condition is not set")
		assert.Equal(t, catalogv1alpha1.ConditionFalse, readyCondition.Status, "Wrong Ready condition.Status")
		assert.Equal(t, catalogv1alpha1.CatalogEntryTerminatingReason, readyCondition.Reason, "Wrong Reason condition.Status")

		// Check CRD Annotation
		crdCheck := &apiextensionsv1.CustomResourceDefinition{}
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name: crd1.Name,
		}, crdCheck), "error when getting crd")
		assert.NotContains(t, crdCheck.Annotations, catalogEntryReferenceAnnotation, "the catalogEntry annotation should be removed from the CRD")
	})
}

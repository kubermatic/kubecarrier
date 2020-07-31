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

	"k8c.io/utils/pkg/testutil"

	catalogv1alpha1 "k8c.io/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "k8c.io/kubecarrier/pkg/apis/core/v1alpha1"
)

func TestCatalogEntrySetReconciler(t *testing.T) {
	baseCRD := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "catapults.stable.com",
			Labels: map[string]string{
				"kubecarrier.io/service-cluster": "eu-west-1",
				"kubecarrier.io/provider":        "example.provider",
			},
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "stable.com",
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
							},
							Type: "object",
						},
					},
				},
			},
		},
	}

	catalogEntrySet := &catalogv1alpha1.CatalogEntrySet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "catapult",
			Namespace: "hans",
		},
		Spec: catalogv1alpha1.CatalogEntrySetSpec{
			Metadata: catalogv1alpha1.CatalogEntrySetMetadata{
				CommonMetadata: catalogv1alpha1.CommonMetadata{
					DisplayName: "Test CatalogEntrySet",
					Description: "Test CatalogEntrySet",
				},
			},
			Derive: &catalogv1alpha1.DerivedConfig{
				Expose: []catalogv1alpha1.VersionExposeConfig{
					{
						Versions: []string{
							"v1alpha1",
						},
						Fields: []catalogv1alpha1.FieldPath{
							{JSONPath: ".spec.prop1"},
						},
					},
				},
			},
			Discover: catalogv1alpha1.CustomResourceDiscoverySetConfig{
				CRD: catalogv1alpha1.ObjectReference{
					Name: baseCRD.Name,
				},
				ServiceClusterSelector: metav1.LabelSelector{},
				WebhookStrategy:        corev1alpha1.WebhookStrategyTypeServiceCluster,
			},
		},
	}
	catalogEntrySetNN := types.NamespacedName{
		Name:      "catapult",
		Namespace: "hans",
	}
	r := &CatalogEntrySetReconciler{
		Log:    testutil.NewLogger(t),
		Client: fakeclient.NewFakeClientWithScheme(testScheme, baseCRD, catalogEntrySet),
		Scheme: testScheme,
	}
	ctx := context.Background()
	reconcileLoop := func() {
		for i := 0; i < 3; i++ {
			_, err := r.Reconcile(ctrl.Request{
				NamespacedName: catalogEntrySetNN,
			})
			require.NoError(t, err)
			require.NoError(t, r.Client.Get(ctx, catalogEntrySetNN, catalogEntrySet))
		}
	}

	reconcileLoop()
	customResourceDiscoverySet := &corev1alpha1.CustomResourceDiscoverySet{}
	require.NoError(t, r.Get(ctx, types.NamespacedName{
		Name:      catalogEntrySet.Name,
		Namespace: catalogEntrySet.Namespace,
	}, customResourceDiscoverySet))
	assert.Equal(t, corev1alpha1.WebhookStrategyTypeServiceCluster, customResourceDiscoverySet.Spec.WebhookStrategy)

	ready, ok := catalogEntrySet.Status.GetCondition(catalogv1alpha1.CatalogEntrySetReady)
	if assert.True(t, ok) {
		assert.Equal(t, catalogv1alpha1.ConditionFalse, ready.Status)
		assert.Equal(t, "CustomResourceDiscoverySetUnready", ready.Reason)
	}

	internalCRDName := "catapultinternals.eu-west-1.example-provider"
	customResourceDiscoverySet.Status.Conditions = []corev1alpha1.CustomResourceDiscoverySetCondition{
		{
			Type:   corev1alpha1.CustomResourceDiscoverySetReady,
			Status: corev1alpha1.ConditionTrue,
		},
	}
	customResourceDiscoverySet.Status.ManagementClusterCRDs = []corev1alpha1.CustomResourceDiscoverySetCRDReference{
		{
			CRD: corev1alpha1.ObjectReference{
				Name: internalCRDName,
			},
			ServiceCluster: corev1alpha1.ObjectReference{
				Name: "eu-west-1",
			},
		},
	}
	require.NoError(t, r.Status().Update(ctx, customResourceDiscoverySet))

	reconcileLoop() // should update status

	catalogEntry := &catalogv1alpha1.CatalogEntry{}
	require.NoError(t, r.Get(ctx, types.NamespacedName{
		Name:      catalogEntrySet.Name + "." + "eu-west-1",
		Namespace: catalogEntrySet.Namespace,
	}, catalogEntry))

	ready, ok = catalogEntrySet.Status.GetCondition(catalogv1alpha1.CatalogEntrySetReady)
	if assert.True(t, ok) {
		assert.Equal(t, catalogv1alpha1.ConditionFalse, ready.Status)
		assert.Equal(t, "CatalogEntriesUnready", ready.Reason)
	}

	catalogEntry.Status.Conditions = []catalogv1alpha1.CatalogEntryCondition{
		{
			Type:   catalogv1alpha1.CatalogEntryReady,
			Status: catalogv1alpha1.ConditionTrue,
		},
	}
	require.NoError(t, r.Status().Update(ctx, catalogEntry))

	reconcileLoop()

	ready, ok = catalogEntrySet.Status.GetCondition(catalogv1alpha1.CatalogEntrySetReady)
	if assert.True(t, ok) {
		assert.Equal(t, catalogv1alpha1.ConditionTrue, ready.Status)
	}
}

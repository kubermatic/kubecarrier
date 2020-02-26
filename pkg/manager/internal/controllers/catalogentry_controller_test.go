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
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func TestCatalogEntryReconciler(t *testing.T) {
	provider := &catalogv1alpha1.Account{
		ObjectMeta: metav1.ObjectMeta{
			Name: "example.provider",
		},
		Status: catalogv1alpha1.AccountStatus{
			NamespaceName: "example.provider",
		},
	}
	providerNS := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "example.provider"}}

	baseCRD := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "catapults.test.kubecarrier.io",
			Labels: map[string]string{
				ServiceClusterLabel:  "eu-west-1",
				OriginNamespaceLabel: "example.provider",
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
	baseCRD.Status.AcceptedNames = baseCRD.Spec.Names

	catalogEntry := &catalogv1alpha1.CatalogEntry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-catalogEntry",
			Namespace: providerNS.Name,
		},
		Spec: catalogv1alpha1.CatalogEntrySpec{
			Metadata: catalogv1alpha1.CatalogEntryMetadata{
				DisplayName: "Test CatalogEntry",
				Description: "Test CatalogEntry",
			},
			BaseCRD: catalogv1alpha1.ObjectReference{
				Name: baseCRD.Name,
			},
			DerivedConfig: &catalogv1alpha1.DerivedConfig{
				KindOverride: "TestResource",
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
		},
	}

	derivedCRD := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "testresources.test.kubecarrier.io",
			Labels: map[string]string{
				ServiceClusterLabel:  "eu-west-1",
				OriginNamespaceLabel: "example.provider",
			},
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "test.kubecarrier.io",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind:     "TestResource",
				ListKind: "TestResourceList",
				Plural:   "testresources",
				Singular: "testresource",
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
	derivedCRD.Status.AcceptedNames = derivedCRD.Spec.Names

	client := fakeclient.NewFakeClientWithScheme(testScheme, catalogEntry, baseCRD, provider, derivedCRD, providerNS)
	log := testutil.NewLogger(t)
	r := &CatalogEntryReconciler{
		Client: client,
		Log:    log,
		Scheme: testScheme,
	}
	ctx := context.Background()
	reconcileLoop := func() {
		for i := 0; i < 5; i++ {
			_, err := r.Reconcile(ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      catalogEntry.Name,
					Namespace: catalogEntry.Namespace,
				},
			})
			require.NoError(t, err, "unexpected error returned by Reconcile")
			require.NoError(t, client.Get(ctx, types.NamespacedName{
				Name:      catalogEntry.Name,
				Namespace: catalogEntry.Namespace,
			}, catalogEntry))
		}
	}

	crdFound := &apiextensionsv1.CustomResourceDefinition{}
	derivedCustomResource := &catalogv1alpha1.DerivedCustomResource{}
	if !t.Run("create/update CatalogEntry", func(t *testing.T) {
		reconcileLoop()
		// Check CatalogEntry
		assert.Len(t, catalogEntry.Finalizers, 1, "finalizer should be added to CatalogEntry")

		// Check DerivedCustomResource
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name:      catalogEntry.Name,
			Namespace: catalogEntry.Namespace,
		}, derivedCustomResource), "error when getting DerivedCustomResource")
		derivedCustomResource.Status.Conditions = []catalogv1alpha1.DerivedCustomResourceCondition{
			{
				Type:   catalogv1alpha1.DerivedCustomResourceReady,
				Status: catalogv1alpha1.ConditionTrue,
			},
		}
		derivedCustomResource.Status.DerivedCR = &catalogv1alpha1.DerivedCustomResourceReference{
			Name: derivedCRD.Name,
		}
		require.NoError(t, client.Status().Update(ctx, derivedCustomResource))

		reconcileLoop()
		// Check CatalogEntry Conditions
		readyCondition, readyConditionExists := catalogEntry.Status.GetCondition(catalogv1alpha1.CatalogEntryReady)
		assert.True(t, readyConditionExists, "Ready Condition is not set")
		assert.Equal(t, catalogv1alpha1.ConditionTrue, readyCondition.Status, "Wrong Ready condition.Status")

		// Check CatalogEntry Status
		assert.Equal(t, catalogEntry.Status.CRD.Kind, derivedCRD.Spec.Names.Kind, "CRD Kind is wrong")
		assert.Equal(t, catalogEntry.Status.CRD.Name, derivedCRD.Name, "CRD Name is wrong")
		assert.Equal(t, catalogEntry.Status.CRD.ServiceCluster.Name, derivedCRD.Labels[ServiceClusterLabel], "CRD ServiceCluster is wrong")
		assert.Equal(t, catalogEntry.Status.CRD.APIGroup, derivedCRD.Spec.Group, "CRD APIGroup is wrong")

		// Check CRDs Annotation
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name: baseCRD.Name,
		}, crdFound), "error when getting crd")
		assert.Contains(t, crdFound.Annotations, catalogEntryReferenceAnnotation, "the catalogEntry annotation should be added to the CRD")

	}) {
		t.FailNow()
	}

	t.Run("delete CatalogEntry", func(t *testing.T) {
		// Setup
		ts := metav1.Now()
		catalogEntry.DeletionTimestamp = &ts
		require.NoError(t, client.Update(ctx, catalogEntry), "unexpected error updating catalogEntry")

		reconcileLoop()
		catalogEntryCheck := &catalogv1alpha1.CatalogEntry{}
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name:      catalogEntry.Name,
			Namespace: catalogEntry.Namespace,
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
			Name: baseCRD.Name,
		}, crdCheck), "error when getting crd")
		assert.NotContains(t, crdCheck.Annotations, catalogEntryReferenceAnnotation, "the catalogEntry annotation should be removed from the CRD")
	})

	// Check TenantReference
	derivedCustomResourceCheck := &catalogv1alpha1.TenantReference{}
	assert.True(t, errors.IsNotFound(client.Get(ctx, types.NamespacedName{
		Name:      derivedCustomResource.Name,
		Namespace: derivedCustomResource.Namespace,
	}, derivedCustomResourceCheck)), "DerivedCustomResource should be gone")
}

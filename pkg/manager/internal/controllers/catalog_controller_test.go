/*
Copyright 2020 The KubeCarrier Authors.

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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func TestCatalogReconciler(t *testing.T) {

	provider := &catalogv1alpha1.Provider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-provider",
			Namespace: "kubecarrier-system",
		},
	}

	providerNamespaceName := fmt.Sprintf("provider-%s", provider.Name)

	providerNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: providerNamespaceName,
		},
	}

	tenant := &catalogv1alpha1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-tenant",
			Namespace: "kubecarrier-system",
		},
	}
	tenant.Status.SetCondition(catalogv1alpha1.TenantCondition{
		Type:    catalogv1alpha1.TenantReady,
		Status:  catalogv1alpha1.ConditionTrue,
		Reason:  "SetupComplete",
		Message: "Tenant setup is complete.",
	})
	tenantNamespaceName := fmt.Sprintf("tenant-%s", tenant.Name)
	tenant.Status.NamespaceName = tenantNamespaceName

	tenantReference := &catalogv1alpha1.TenantReference{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tenant.Name,
			Namespace: providerNamespaceName,
		},
	}

	tenantNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: tenantNamespaceName,
		},
	}

	catalogEntry := &catalogv1alpha1.CatalogEntry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-catalogEntry",
			Namespace: providerNamespaceName,
		},
		Spec: catalogv1alpha1.CatalogEntrySpec{
			Metadata: catalogv1alpha1.CatalogEntryMetadata{
				DisplayName: "Test CatalogEntry",
				Description: "Test CatalogEntry",
			},
		},
	}

	catalog := &catalogv1alpha1.Catalog{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-catalog",
			Namespace: providerNamespaceName,
		},
		Spec: catalogv1alpha1.CatalogSpec{
			CatalogEntrySelector:    &metav1.LabelSelector{},
			TenantReferenceSelector: &metav1.LabelSelector{},
		},
	}

	client := fakeclient.NewFakeClientWithScheme(testScheme, catalogEntry, catalog, provider, providerNamespace, tenant, tenantReference, tenantNamespace)
	log := testutil.NewLogger(t)
	r := &CatalogReconciler{
		Client: client,
		Log:    log,
		Scheme: testScheme,
	}
	ctx := context.Background()

	catalogFound := &catalogv1alpha1.Catalog{}
	offeringFound := &catalogv1alpha1.Offering{}
	providerReferenceFound := &catalogv1alpha1.ProviderReference{}
	if !t.Run("create/update Catalog", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			// Run Reconcile multiple times, because
			// the reconcilation stops after changing the Catalog
			_, err := r.Reconcile(ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      catalog.Name,
					Namespace: catalog.Namespace,
				},
			})
			require.NoError(t, err, "unexpected error returned by Reconcile")
		}

		// Check Catalog
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name:      catalog.Name,
			Namespace: catalog.Namespace,
		}, catalogFound), "error when getting Catalog")
		assert.Len(t, catalogFound.Finalizers, 1, "finalizer should be added to Catalog")

		// Check Catalog Status
		assert.Len(t, catalogFound.Status.Entries, 1, "CatalogEntry is not added to the Catalog.Status.Entries")
		assert.Equal(t, catalogFound.Status.Entries[0].Name, catalogEntry.Name, "CatalogEntry name is wrong")
		assert.Len(t, catalogFound.Status.Tenants, 1, "TenantReference is not added to the Catalog.Status.Tenants")
		assert.Equal(t, catalogFound.Status.Tenants[0].Name, tenantReference.Name, "TenantReference name is wrong")

		// Check CatalogEntry Conditions
		readyCondition, readyConditionExists := catalogFound.Status.GetCondition(catalogv1alpha1.CatalogReady)
		assert.True(t, readyConditionExists, "Ready Condition is not set")
		assert.Equal(t, catalogv1alpha1.ConditionTrue, readyCondition.Status, "Wrong Ready condition.Status")

		// Check Offering
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name:      catalogEntry.Name,
			Namespace: tenantNamespaceName,
		}, offeringFound), "getting Offering error")
		assert.Equal(t, offeringFound.Offering.Provider.Name, provider.Name, "Wrong Offering provider name")
		assert.Equal(t, offeringFound.Offering.Metadata.Description, catalogEntry.Spec.Metadata.Description, "Wrong Offering description")
		assert.Len(t, offeringFound.Offering.CRDs, len(catalogEntry.Status.CRDs), "Wrong Offering description")

		// Check ProviderReference
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name:      provider.Name,
			Namespace: tenantNamespaceName,
		}, providerReferenceFound), "getting ProviderReference error")
		assert.Equal(t, providerReferenceFound.Spec.Metadata.Description, provider.Spec.Metadata.Description, "Wrong ProviderReference Metadata.Description")
		assert.Equal(t, providerReferenceFound.Spec.Metadata.DisplayName, provider.Spec.Metadata.DisplayName, "Wrong ProviderReference Metadata.DisplayName")
	}) {
		t.FailNow()
	}

	t.Run("delete Catalog", func(t *testing.T) {
		// Setup
		ts := metav1.Now()
		catalogFound.DeletionTimestamp = &ts
		require.NoError(t, client.Update(ctx, catalogFound), "unexpected error updating catalog")

		for i := 0; i < 5; i++ {
			// Run Reconcile multiple times, because
			// the reconcilation stops after changing the Catalog
			_, err := r.Reconcile(ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      catalogFound.Name,
					Namespace: catalogFound.Namespace,
				},
			})
			require.NoError(t, err, "unexpected error returned by Reconcile")
		}

		catalogCheck := &catalogv1alpha1.Catalog{}
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name:      catalogFound.Name,
			Namespace: catalogFound.Namespace,
		}, catalogCheck), "cannot check Catalog")
		assert.Len(t, catalogCheck.Finalizers, 0, "finalizers should have been removed")

		// Check Catalog Conditions
		readyCondition, readyConditionExists := catalogCheck.Status.GetCondition(catalogv1alpha1.CatalogReady)
		assert.True(t, readyConditionExists, "Ready Condition is not set")
		assert.Equal(t, catalogv1alpha1.ConditionFalse, readyCondition.Status, "Wrong Ready condition.Status")
		assert.Equal(t, catalogv1alpha1.CatalogTerminatingReason, readyCondition.Reason, "Wrong Reason condition.Status")

		// Check Offering
		offeringCheck := &catalogv1alpha1.Offering{}
		assert.True(t, errors.IsNotFound(client.Get(ctx, types.NamespacedName{
			Name:      offeringFound.Name,
			Namespace: offeringFound.Namespace,
		}, offeringCheck)), "Offering should be gone")

		// Check ProviderReference
		providerReferenceCheck := &catalogv1alpha1.ProviderReference{}
		assert.True(t, errors.IsNotFound(client.Get(ctx, types.NamespacedName{
			Name:      providerReferenceFound.Name,
			Namespace: providerReferenceFound.Namespace,
		}, providerReferenceCheck)), "ProviderReference should be gone")
	})
}

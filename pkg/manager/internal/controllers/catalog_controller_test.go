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
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/owner"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func TestCatalogReconciler(t *testing.T) {

	provider := &catalogv1alpha1.Account{
		ObjectMeta: metav1.ObjectMeta{
			Name: "example-provider",
		},
		Spec: catalogv1alpha1.AccountSpec{
			Roles: []catalogv1alpha1.AccountRole{
				catalogv1alpha1.ProviderRole,
			},
		},
	}

	provider.Status.Namespace.Name = provider.Name

	providerNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: provider.Name,
		},
	}
	owner.SetOwnerReference(provider, providerNamespace, testScheme)

	tenant := &catalogv1alpha1.Account{
		ObjectMeta: metav1.ObjectMeta{
			Name: "example-tenant",
		},
		Spec: catalogv1alpha1.AccountSpec{
			Roles: []catalogv1alpha1.AccountRole{
				catalogv1alpha1.TenantRole,
			},
		},
	}
	tenant.Status.SetCondition(catalogv1alpha1.AccountCondition{
		Type:    catalogv1alpha1.AccountReady,
		Status:  catalogv1alpha1.ConditionTrue,
		Reason:  "SetupComplete",
		Message: "Tenant setup is complete.",
	})
	tenant.Status.Namespace.Name = tenant.Name

	tenantReference := &catalogv1alpha1.TenantReference{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tenant.Name,
			Namespace: providerNamespace.Name,
		},
	}

	tenantNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: tenant.Name,
		},
	}
	owner.SetOwnerReference(tenant, tenantNamespace, testScheme)

	catalogEntry := &catalogv1alpha1.CatalogEntry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-catalogEntry",
			Namespace: providerNamespace.Name,
		},
		Spec: catalogv1alpha1.CatalogEntrySpec{
			Metadata: catalogv1alpha1.CatalogEntryMetadata{
				DisplayName: "Test CatalogEntry",
				Description: "Test CatalogEntry",
			},
		},
		Status: catalogv1alpha1.CatalogEntryStatus{
			CRD: &catalogv1alpha1.CRDInformation{
				ServiceCluster: catalogv1alpha1.ObjectReference{
					Name: "test-service-cluster",
				},
			},
			Conditions: []catalogv1alpha1.CatalogEntryCondition{
				{
					Type:   catalogv1alpha1.CatalogEntryReady,
					Status: catalogv1alpha1.ConditionTrue,
				},
			},
		},
	}

	serviceCluster := &corev1alpha1.ServiceCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service-cluster",
			Namespace: providerNamespace.Name,
		},
		Spec: corev1alpha1.ServiceClusterSpec{
			Metadata: corev1alpha1.ServiceClusterMetadata{
				DisplayName: "test-service-cluster",
				Description: "a service cluster for testing",
			},
		},
	}

	catalog := &catalogv1alpha1.Catalog{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-catalog",
			Namespace: providerNamespace.Name,
		},
		Spec: catalogv1alpha1.CatalogSpec{
			CatalogEntrySelector:    &metav1.LabelSelector{},
			TenantReferenceSelector: &metav1.LabelSelector{},
		},
	}

	client := fakeclient.NewFakeClientWithScheme(testScheme, catalogEntry, catalog, provider, providerNamespace, tenant, tenantReference, tenantNamespace, serviceCluster)
	log := testutil.NewLogger(t)
	r := &CatalogReconciler{
		Client: client,
		Log:    log,
		Scheme: testScheme,
	}
	ctx := context.Background()

	catalogFound := &catalogv1alpha1.Catalog{}
	offeringFound := &catalogv1alpha1.Offering{}
	providerFound := &catalogv1alpha1.Provider{}
	serviceClusterReferenceFound := &catalogv1alpha1.ServiceClusterReference{}
	serviceClusterAssignmentFound := &corev1alpha1.ServiceClusterAssignment{}
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
			Namespace: tenantNamespace.Name,
		}, offeringFound), "getting Offering error")
		assert.Equal(t, offeringFound.Offering.Provider.Name, provider.Name, "Wrong Offering provider name")
		assert.Equal(t, offeringFound.Offering.Metadata.Description, catalogEntry.Spec.Metadata.Description, "Wrong Offering description")
		assert.Equal(t, offeringFound.Offering.CRD, *catalogEntry.Status.CRD, "Wrong Offering description")

		// Check Provider
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name:      provider.Name,
			Namespace: tenantNamespace.Name,
		}, providerFound), "getting Provider error")
		assert.Equal(t, providerFound.Spec.Metadata.Description, provider.Spec.Metadata.Description, "Wrong Provider Metadata.Description")
		assert.Equal(t, providerFound.Spec.Metadata.DisplayName, provider.Spec.Metadata.DisplayName, "Wrong Provider Metadata.DisplayName")

		// Check ServiceClusterReference
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s.%s", serviceCluster.Name, provider.Name),
			Namespace: tenantNamespace.Name,
		}, serviceClusterReferenceFound), "getting ServiceClusterReference error")
		assert.Equal(t, serviceClusterReferenceFound.Spec.Provider.Name, provider.Name, "Wrong ServiceClusterReference provider name")
		assert.Equal(t, serviceClusterReferenceFound.Spec.Metadata.Description, serviceCluster.Spec.Metadata.Description, "Wrong ServiceClusterReference description")
		assert.Equal(t, serviceClusterReferenceFound.Spec.Metadata.DisplayName, serviceCluster.Spec.Metadata.DisplayName, "Wrong ServiceClusterReference display name")

		// Check ServiceClusterAssignment
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s.%s", tenantNamespace.Name, serviceCluster.Name),
			Namespace: providerNamespace.Name,
		}, serviceClusterAssignmentFound), "getting ServiceClusterAssignment error")
		assert.Equal(t, serviceClusterAssignmentFound.Spec.ServiceCluster.Name, serviceCluster.Name, "Wrong ServiceCluster name")
		assert.Equal(t, serviceClusterAssignmentFound.Spec.ManagementClusterNamespace.Name, tenantNamespace.Name, "Wrong ManagementCluster Namespace name.")
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

		// Check Provider
		providerCheck := &catalogv1alpha1.Provider{}
		assert.True(t, errors.IsNotFound(client.Get(ctx, types.NamespacedName{
			Name:      providerFound.Name,
			Namespace: providerFound.Namespace,
		}, providerCheck)), "Provider should be gone")

		// Check ServiceClusterReference
		serviceClusterReferenceCheck := &catalogv1alpha1.ServiceClusterReference{}
		assert.True(t, errors.IsNotFound(client.Get(ctx, types.NamespacedName{
			Name:      serviceClusterReferenceFound.Name,
			Namespace: serviceClusterReferenceFound.Namespace,
		}, serviceClusterReferenceCheck)), "ServiceClusterReference should be gone")

		// Check ServiceClusterAssignment
		serviceClusterAssignmentCheck := &corev1alpha1.ServiceClusterAssignment{}
		assert.True(t, errors.IsNotFound(client.Get(ctx, types.NamespacedName{
			Name:      serviceClusterAssignmentFound.Name,
			Namespace: serviceClusterAssignmentFound.Namespace,
		}, serviceClusterAssignmentCheck)), "ServiceClusterAssignment should be gone")
	})
}

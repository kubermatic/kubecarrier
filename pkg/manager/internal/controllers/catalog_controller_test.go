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
	rbacv1 "k8s.io/api/rbac/v1"
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
			Subjects: []rbacv1.Subject{
				{
					Kind:     rbacv1.GroupKind,
					APIGroup: "rbac.authorization.k8s.io",
					Name:     "example-provider",
				},
			},
		},
	}

	provider.Status.Namespace = &catalogv1alpha1.ObjectReference{
		Name: provider.Name,
	}

	providerNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: provider.Name,
		},
	}
	owner.SetOwnerReference(provider, providerNamespace, testScheme)

	tenantAccount := &catalogv1alpha1.Account{
		ObjectMeta: metav1.ObjectMeta{
			Name: "example-tenant",
		},
		Spec: catalogv1alpha1.AccountSpec{
			Roles: []catalogv1alpha1.AccountRole{
				catalogv1alpha1.TenantRole,
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:     rbacv1.GroupKind,
					APIGroup: "rbac.authorization.k8s.io",
					Name:     "example-tenant",
				},
			},
		},
	}
	tenantAccount.Status.SetCondition(catalogv1alpha1.AccountCondition{
		Type:    catalogv1alpha1.AccountReady,
		Status:  catalogv1alpha1.ConditionTrue,
		Reason:  "SetupComplete",
		Message: "Tenant setup is complete.",
	})
	tenantAccount.Status.Namespace = &catalogv1alpha1.ObjectReference{
		Name: tenantAccount.Name,
	}

	tenant := &catalogv1alpha1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tenantAccount.Name,
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
			TenantCRD: &catalogv1alpha1.CRDInformation{
				APIGroup: "tenant.apigroup",
				Plural:   "tenant.plural",
				ServiceCluster: catalogv1alpha1.ObjectReference{
					Name: "test-service-cluster",
				},
			},
			ProviderCRD: &catalogv1alpha1.CRDInformation{
				APIGroup: "provider.apigroup",
				Plural:   "provider.plural",
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
			CatalogEntrySelector: &metav1.LabelSelector{},
			TenantSelector:       &metav1.LabelSelector{},
		},
	}

	client := fakeclient.NewFakeClientWithScheme(testScheme, catalogEntry, catalog, provider, providerNamespace, tenant, tenantAccount, tenantNamespace, serviceCluster)
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
	providerRoleFound := &rbacv1.Role{}
	providerRoleBindingFound := &rbacv1.RoleBinding{}
	tenantRoleFound := &rbacv1.Role{}
	tenantRoleBindingFound := &rbacv1.RoleBinding{}
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
		assert.Len(t, catalogFound.Status.Tenants, 1, "Tenant is not added to the Catalog.Status.Tenants")
		assert.Equal(t, catalogFound.Status.Tenants[0].Name, tenant.Name, "Tenant name is wrong")

		// Check CatalogEntry Conditions
		readyCondition, readyConditionExists := catalogFound.Status.GetCondition(catalogv1alpha1.CatalogReady)
		assert.True(t, readyConditionExists, "Ready Condition is not set")
		assert.Equal(t, catalogv1alpha1.ConditionTrue, readyCondition.Status, "Wrong Ready condition.Status")

		// Check Offering
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name:      catalogEntry.Status.TenantCRD.Name,
			Namespace: tenantNamespace.Name,
		}, offeringFound), "getting Offering error")
		assert.Equal(t, offeringFound.Spec.Provider.Name, provider.Name, "Wrong Offering provider name")
		assert.Equal(t, offeringFound.Spec.Metadata.Description, catalogEntry.Spec.Metadata.Description, "Wrong Offering description")
		assert.Equal(t, offeringFound.Spec.CRD, *catalogEntry.Status.TenantCRD, "Wrong Offering description")

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

		// Check Provider Role
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("kubecarrier:provider:%s", catalogEntry.Name),
			Namespace: tenantAccount.Status.Namespace.Name,
		}, providerRoleFound), "getting Role error")
		assert.Contains(t, providerRoleFound.Rules, rbacv1.PolicyRule{
			Verbs:     []string{rbacv1.VerbAll},
			APIGroups: []string{catalogEntry.Status.ProviderCRD.APIGroup},
			Resources: []string{catalogEntry.Status.ProviderCRD.Plural},
		}, "Missing the PolicyRule")

		// Check Provider RoleBinding
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("kubecarrier:provider:%s", catalogEntry.Name),
			Namespace: tenantAccount.Status.Namespace.Name,
		}, providerRoleBindingFound), "getting RoleBinding error")
		assert.Equal(t, providerRoleBindingFound.Subjects, provider.Spec.Subjects, "Subjects is different")

		// Check Tenant Role
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("kubecarrier:tenant:%s", catalogEntry.Name),
			Namespace: tenantAccount.Status.Namespace.Name,
		}, tenantRoleFound), "getting Role error")
		assert.Contains(t, tenantRoleFound.Rules, rbacv1.PolicyRule{
			Verbs:     []string{rbacv1.VerbAll},
			APIGroups: []string{catalogEntry.Status.TenantCRD.APIGroup},
			Resources: []string{catalogEntry.Status.TenantCRD.Plural},
		}, "Missing the PolicyRule")

		// Check Tenant RoleBinding
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("kubecarrier:tenant:%s", catalogEntry.Name),
			Namespace: tenantAccount.Status.Namespace.Name,
		}, tenantRoleBindingFound), "getting RoleBinding error")
		assert.Equal(t, tenantRoleBindingFound.Subjects, tenantAccount.Spec.Subjects, "Subjects is different")
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

		// Check Provider Role
		providerRoleCheck := &rbacv1.Role{}
		assert.True(t, errors.IsNotFound(client.Get(ctx, types.NamespacedName{
			Name:      providerRoleFound.Name,
			Namespace: providerRoleFound.Namespace,
		}, providerRoleCheck)), "provider Role should be gone")

		// Check Provider RoleBinding
		providerRoleBindingCheck := &rbacv1.RoleBinding{}
		assert.True(t, errors.IsNotFound(client.Get(ctx, types.NamespacedName{
			Name:      providerRoleBindingFound.Name,
			Namespace: providerRoleBindingFound.Namespace,
		}, providerRoleBindingCheck)), "provider RoleBinding should be gone")

		// Check Tenant Role
		tenantRoleCheck := &rbacv1.Role{}
		assert.True(t, errors.IsNotFound(client.Get(ctx, types.NamespacedName{
			Name:      tenantRoleFound.Name,
			Namespace: tenantRoleFound.Namespace,
		}, tenantRoleCheck)), "tenant Role should be gone")

		// Check Tenant RoleBinding
		tenantRoleBindingCheck := &rbacv1.RoleBinding{}
		assert.True(t, errors.IsNotFound(client.Get(ctx, types.NamespacedName{
			Name:      tenantRoleBindingFound.Name,
			Namespace: tenantRoleBindingFound.Namespace,
		}, tenantRoleBindingCheck)), "tenant RoleBinding should be gone")
	})
}

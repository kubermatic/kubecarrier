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

	"github.com/kubermatic/kubecarrier/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
)

func TestTenantReconciler(t *testing.T) {
	tenant := &catalogv1alpha1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-tenant",
			Namespace: "test-ns",
		},
	}

	client := fakeclient.NewFakeClientWithScheme(testScheme, tenant)
	log := testutil.NewLogger(t)
	r := &TenantReconciler{
		Client: client,
		Log:    log,
		Scheme: testScheme,
	}
	ctx := context.Background()

	tenantFound := &catalogv1alpha1.Tenant{}
	namespaceFound := &corev1.Namespace{}
	t.Run("create/update Tenant", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			// Run Reconcile multiple times, because
			// the reconcilation stops after changing the Tenant
			_, err := r.Reconcile(ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      tenant.Name,
					Namespace: tenant.Namespace,
				},
			})
			require.NoError(t, err, "unexpected error returned by Reconcile")
		}

		// Check Tenant
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name:      tenant.Name,
			Namespace: tenant.Namespace,
		}, tenantFound), "error when getting tenant")
		assert.Len(t, tenantFound.Finalizers, 1, "finalizer should be added to Tenant")
		assert.NotEmpty(t, tenantFound.Status.NamespaceName, ".Status.NamespaceName should be set")

		// Check Tenant Conditions
		readyCondition, readyConditionExists := tenantFound.Status.GetCondition(catalogv1alpha1.TenantReady)
		assert.True(t, readyConditionExists, "Ready Condition is not set")
		assert.Equal(t, catalogv1alpha1.ConditionTrue, readyCondition.Status, "Wrong Ready condition.Status")

		// Check Namespace
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name: "tenant-" + tenant.Name,
		}, namespaceFound), "getting namespace error")
	})

	t.Run("delete Tenant", func(t *testing.T) {
		// Setup
		ts := metav1.Now()
		tenant.DeletionTimestamp = &ts
		require.NoError(t, client.Update(ctx, tenant), "unexpected error updating tenant")

		for i := 0; i < 5; i++ {
			// Run Reconcile multiple times, because
			// the reconcilation stops after changing the Tenant
			_, err := r.Reconcile(ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      tenant.Name,
					Namespace: tenant.Namespace,
				},
			})
			require.NoError(t, err, "unexpected error returned by Reconcile")
		}

		namespaceCheck := &corev1.Namespace{}
		assert.True(t, errors.IsNotFound(client.Get(ctx, types.NamespacedName{
			Name: namespaceFound.Name,
		}, namespaceCheck)), "Namespace should be gone")

		tenantCheck := &catalogv1alpha1.Tenant{}
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name: tenantFound.Name,
		}, tenantCheck), "cannot check Tenant")
		assert.Len(t, tenantCheck.Finalizers, 0, "finalizers should have been removed")
	})
}

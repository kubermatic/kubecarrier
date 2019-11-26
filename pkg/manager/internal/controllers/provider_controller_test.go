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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func TestProviderReconciler(t *testing.T) {
	provider := &catalogv1alpha1.Provider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-provider",
			Namespace: "kubecarrier-system",
		},
	}

	client := fakeclient.NewFakeClientWithScheme(testScheme, provider)
	log := testutil.NewLogger(t)
	r := &ProviderReconciler{
		Client: client,
		Log:    log,
		Scheme: testScheme,
	}
	ctx := context.Background()

	providerFound := &catalogv1alpha1.Provider{}
	namespaceFound := &corev1.Namespace{}
	if !t.Run("create/update Provider", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			// Run Reconcile multiple times, because
			// the reconcilation stops after changing the Provider
			_, err := r.Reconcile(ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      provider.Name,
					Namespace: provider.Namespace,
				},
			})
			require.NoError(t, err, "unexpected error returned by Reconcile")
		}

		// Check Provider
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name:      provider.Name,
			Namespace: provider.Namespace,
		}, providerFound), "error when getting provider")
		assert.Len(t, providerFound.Finalizers, 1, "finalizer should be added to Provider")
		assert.NotEmpty(t, providerFound.Status.NamespaceName, ".Status.NamespaceName should be set")

		// Check Provider Conditions
		readyCondition, readyConditionExists := providerFound.Status.GetCondition(catalogv1alpha1.ProviderReady)
		assert.True(t, readyConditionExists, "Ready Condition is not set")
		assert.Equal(t, catalogv1alpha1.ConditionTrue, readyCondition.Status, "Wrong Ready condition.Status")

		// Check Namespace
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name: "provider-" + provider.Name,
		}, namespaceFound), "getting namespace error")
	}) {
		t.FailNow()
	}

	t.Run("delete Provider", func(t *testing.T) {
		// Setup
		ts := metav1.Now()
		providerFound.DeletionTimestamp = &ts
		require.NoError(t, client.Update(ctx, providerFound), "unexpected error updating provider")

		for i := 0; i < 5; i++ {
			// Run Reconcile multiple times, because
			// the reconcilation stops after changing the Provider
			_, err := r.Reconcile(ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      providerFound.Name,
					Namespace: providerFound.Namespace,
				},
			})
			require.NoError(t, err, "unexpected error returned by Reconcile")
		}

		namespaceCheck := &corev1.Namespace{}
		assert.True(t, errors.IsNotFound(client.Get(ctx, types.NamespacedName{
			Name: namespaceFound.Name,
		}, namespaceCheck)), "Namespace should be gone")

		providerCheck := &catalogv1alpha1.Provider{}
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name:      providerFound.Name,
			Namespace: providerFound.Namespace,
		}, providerCheck), "cannot check Provider")
		assert.Len(t, providerCheck.Finalizers, 0, "finalizers should have been removed")
	})
}

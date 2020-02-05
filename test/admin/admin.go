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

package admin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
	"github.com/kubermatic/kubecarrier/test/framework"
)

// AdminSuite tests administrator operations - notably the management of Tenants and Providers.
func NewAdminSuite(f *framework.Framework) func(t *testing.T) {
	return func(t *testing.T) {
		// Setup
		masterClient, err := f.MasterClient()
		require.NoError(t, err, "creating master client")
		defer masterClient.CleanUp(t)

		serviceClient, err := f.ServiceClient()
		require.NoError(t, err, "creating service client")
		defer serviceClient.CleanUp(t)

		ctx := context.Background()

		// Create a Tenant
		tenant := &catalogv1alpha1.Tenant{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-tenant1",
				Namespace: "kubecarrier-system",
			},
		}
		require.NoError(t, masterClient.Create(ctx, tenant), "creating tenant error")
		require.NoError(t, testutil.WaitUntilReady(masterClient, tenant))

		tenantNamespaceName := tenant.Status.NamespaceName
		tenantNamespace := &corev1.Namespace{}
		assert.NoError(t, masterClient.Get(ctx, types.NamespacedName{
			Name: tenantNamespaceName,
		}, tenantNamespace))

		// Create a Provider
		provider := &catalogv1alpha1.Provider{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-provider1",
				Namespace: "kubecarrier-system",
			},
			Spec: catalogv1alpha1.ProviderSpec{
				Metadata: catalogv1alpha1.ProviderMetadata{
					DisplayName: "provider",
					Description: "provider test description",
				},
			},
		}

		require.NoError(t, masterClient.Create(ctx, provider), "creating provider error")
		require.NoError(t, testutil.WaitUntilReady(masterClient, provider))

		providerNamespaceName := provider.Status.NamespaceName
		providerNamespace := &corev1.Namespace{}
		assert.NoError(t, masterClient.Get(ctx, types.NamespacedName{
			Name: providerNamespaceName,
		}, providerNamespace))

		tenantReference := &catalogv1alpha1.TenantReference{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tenant.Name,
				Namespace: providerNamespaceName,
			},
		}
		require.NoError(t, testutil.WaitUntilFound(masterClient, tenantReference))

		// Delete Tenant
		require.NoError(t, testutil.DeleteAndWaitUntilNotFound(masterClient, tenant))

		assert.True(t, errors.IsNotFound(masterClient.Get(ctx, types.NamespacedName{
			Name: tenantNamespaceName,
		}, tenantNamespace)), "namespace should also be deleted.")

		assert.True(t, errors.IsNotFound(masterClient.Get(ctx, types.NamespacedName{
			Name:      tenant.Name,
			Namespace: tenantNamespaceName,
		}, tenantReference)), "TenantReference should also be deleted.")

		// Delete Provider
		require.NoError(t, testutil.DeleteAndWaitUntilNotFound(masterClient, provider))
		assert.True(t, errors.IsNotFound(masterClient.Get(ctx, types.NamespacedName{
			Name: providerNamespaceName,
		}, providerNamespace)), "namespace should also be deleted.")
	}
}

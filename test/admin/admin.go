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
	"sigs.k8s.io/controller-runtime/pkg/client"

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

		ctx := context.Background()

		var (
			mdata = catalogv1alpha1.AccountMetadata{
				DisplayName: "metadata name",
				Description: "metadata desc",
			}
			provider = &catalogv1alpha1.Account{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-provider1",
				},
				Spec: catalogv1alpha1.AccountSpec{
					Metadata: mdata,
					Roles: []catalogv1alpha1.AccountRole{
						catalogv1alpha1.ProviderRole,
					},
				},
			}
			tenant = &catalogv1alpha1.Account{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-tenant",
				},
				Spec: catalogv1alpha1.AccountSpec{
					Metadata: mdata,
					Roles: []catalogv1alpha1.AccountRole{
						catalogv1alpha1.TenantRole,
					},
				},
			}
			providerTenant = &catalogv1alpha1.Account{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-tenantprovider",
				},
				Spec: catalogv1alpha1.AccountSpec{
					Metadata: mdata,
					Roles: []catalogv1alpha1.AccountRole{
						catalogv1alpha1.TenantRole,
						catalogv1alpha1.ProviderRole,
					},
				},
			}
		)
		// simple single account operations
		t.Log("creating single provider")
		require.NoError(t, masterClient.Create(ctx, provider), "creating provider")
		require.NoError(t, testutil.WaitUntilReady(masterClient, provider))
		ns := &corev1.Namespace{}
		assert.NoError(t, masterClient.Get(ctx, types.NamespacedName{
			Name: provider.Status.NamespaceName,
		}, ns))

		t.Log("adding single tenant")
		require.NoError(t, masterClient.Create(ctx, tenant), "creating tenant")
		require.NoError(t, testutil.WaitUntilReady(masterClient, tenant))
		assert.NoError(t, masterClient.Get(ctx, types.NamespacedName{
			Name: tenant.Status.NamespaceName,
		}, ns))

		tenantReferencePresent(t, masterClient, ctx, tenant, provider, true)
		tenantReferencePresent(t, masterClient, ctx, tenant, tenant, false)

		t.Log("adding providerTenant")
		require.NoError(t, masterClient.Create(ctx, providerTenant), "creating providerTenant")
		require.NoError(t, testutil.WaitUntilReady(masterClient, providerTenant))
		assert.NoError(t, masterClient.Get(ctx, types.NamespacedName{
			Name: providerTenant.Status.NamespaceName,
		}, ns))

		tenantReferencePresent(t, masterClient, ctx, tenant, provider, true)
		tenantReferencePresent(t, masterClient, ctx, tenant, providerTenant, true)
		tenantReferencePresent(t, masterClient, ctx, tenant, tenant, false)

		tenantReferencePresent(t, masterClient, ctx, provider, provider, false)
		tenantReferencePresent(t, masterClient, ctx, provider, providerTenant, false)
		tenantReferencePresent(t, masterClient, ctx, provider, tenant, false)

		tenantReferencePresent(t, masterClient, ctx, providerTenant, provider, true)
		tenantReferencePresent(t, masterClient, ctx, providerTenant, providerTenant, true)
		tenantReferencePresent(t, masterClient, ctx, providerTenant, tenant, false)

		t.Log("deleting tenant")
		require.NoError(t, testutil.DeleteAndWaitUntilNotFound(masterClient, tenant))
		assert.True(t, errors.IsNotFound(masterClient.Get(ctx, types.NamespacedName{
			Name: tenant.Status.NamespaceName,
		}, ns)), "namespace should also be deleted.")

		tenantReferencePresent(t, masterClient, ctx, tenant, provider, false)
		tenantReferencePresent(t, masterClient, ctx, tenant, providerTenant, false)
		tenantReferencePresent(t, masterClient, ctx, tenant, tenant, false)

		tenantReferencePresent(t, masterClient, ctx, provider, provider, false)
		tenantReferencePresent(t, masterClient, ctx, provider, providerTenant, false)
		tenantReferencePresent(t, masterClient, ctx, provider, tenant, false)

		tenantReferencePresent(t, masterClient, ctx, providerTenant, provider, true)
		tenantReferencePresent(t, masterClient, ctx, providerTenant, providerTenant, true)
		tenantReferencePresent(t, masterClient, ctx, providerTenant, tenant, false)

		t.Log("deleting provider")
		require.NoError(t, testutil.DeleteAndWaitUntilNotFound(masterClient, provider))
		assert.True(t, errors.IsNotFound(masterClient.Get(ctx, types.NamespacedName{
			Name: provider.Status.NamespaceName,
		}, ns)), "namespace should also be deleted.")

		t.Log("deleting providerTenant")
		require.NoError(t, testutil.DeleteAndWaitUntilNotFound(masterClient, providerTenant))
		assert.True(t, errors.IsNotFound(masterClient.Get(ctx, types.NamespacedName{
			Name: providerTenant.Status.NamespaceName,
		}, ns)), "namespace should also be deleted.")
	}
}

func tenantReferencePresent(t *testing.T, masterClient client.Client, ctx context.Context, tenant *catalogv1alpha1.Account, provider *catalogv1alpha1.Account, expected bool) {
	trefs := &catalogv1alpha1.TenantReferenceList{}
	require.NoError(t, masterClient.List(ctx, trefs, client.InNamespace(provider.Status.NamespaceName)))
	var found bool
	for _, tref := range trefs.Items {
		if tref.Name == tenant.Name {
			found = true
		}
	}
	assert.Equalf(t, expected, found, "tenantReference %s presence in provider %s", tenant.Name, provider.Name)
}

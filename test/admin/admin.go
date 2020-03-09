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
	rbacv1 "k8s.io/api/rbac/v1"
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
		managementClient, err := f.ManagementClient()
		require.NoError(t, err, "creating management client")
		defer managementClient.CleanUp(t)

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
					Subjects: []rbacv1.Subject{
						{
							Kind:     rbacv1.GroupKind,
							APIGroup: "rbac.authorization.k8s.io",
							Name:     "provider1",
						},
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
					Subjects: []rbacv1.Subject{
						{
							Kind:     rbacv1.GroupKind,
							APIGroup: "rbac.authorization.k8s.io",
							Name:     "tenant",
						},
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
					Subjects: []rbacv1.Subject{
						{
							Kind:     rbacv1.GroupKind,
							APIGroup: "rbac.authorization.k8s.io",
							Name:     "tenantprovider",
						},
					},
				},
			}
		)
		// simple single account operations
		t.Log("creating single provider")
		require.NoError(t, managementClient.Create(ctx, provider), "creating provider")
		require.NoError(t, testutil.WaitUntilReady(managementClient, provider))
		ns := &corev1.Namespace{}
		assert.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name: provider.Status.Namespace.Name,
		}, ns))
		providerRoleAndRoleBindingPresent(t, managementClient, ctx, provider, true)
		tenantRoleAndRoleBindingPresent(t, managementClient, ctx, provider, false)

		t.Log("adding single tenant")
		require.NoError(t, managementClient.Create(ctx, tenant), "creating tenant")
		require.NoError(t, testutil.WaitUntilReady(managementClient, tenant))
		assert.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name: tenant.Status.Namespace.Name,
		}, ns))
		providerRoleAndRoleBindingPresent(t, managementClient, ctx, tenant, false)
		tenantRoleAndRoleBindingPresent(t, managementClient, ctx, tenant, true)

		tenantPresent(t, managementClient, ctx, tenant, provider, true)
		tenantPresent(t, managementClient, ctx, tenant, tenant, false)

		t.Log("adding providerTenant")
		require.NoError(t, managementClient.Create(ctx, providerTenant), "creating providerTenant")
		require.NoError(t, testutil.WaitUntilReady(managementClient, providerTenant))
		assert.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name: providerTenant.Status.Namespace.Name,
		}, ns))
		providerRoleAndRoleBindingPresent(t, managementClient, ctx, providerTenant, true)
		tenantRoleAndRoleBindingPresent(t, managementClient, ctx, providerTenant, true)

		tenantPresent(t, managementClient, ctx, tenant, provider, true)
		tenantPresent(t, managementClient, ctx, tenant, providerTenant, true)
		tenantPresent(t, managementClient, ctx, tenant, tenant, false)

		tenantPresent(t, managementClient, ctx, provider, provider, false)
		tenantPresent(t, managementClient, ctx, provider, providerTenant, false)
		tenantPresent(t, managementClient, ctx, provider, tenant, false)

		tenantPresent(t, managementClient, ctx, providerTenant, provider, true)
		tenantPresent(t, managementClient, ctx, providerTenant, providerTenant, true)
		tenantPresent(t, managementClient, ctx, providerTenant, tenant, false)

		t.Log("deleting tenant")
		require.NoError(t, testutil.DeleteAndWaitUntilNotFound(managementClient, tenant))
		assert.True(t, errors.IsNotFound(managementClient.Get(ctx, types.NamespacedName{
			Name: tenant.Status.Namespace.Name,
		}, ns)), "namespace should also be deleted.")
		providerRoleAndRoleBindingPresent(t, managementClient, ctx, tenant, false)
		tenantRoleAndRoleBindingPresent(t, managementClient, ctx, tenant, false)

		tenantPresent(t, managementClient, ctx, tenant, provider, false)
		tenantPresent(t, managementClient, ctx, tenant, providerTenant, false)
		tenantPresent(t, managementClient, ctx, tenant, tenant, false)

		tenantPresent(t, managementClient, ctx, provider, provider, false)
		tenantPresent(t, managementClient, ctx, provider, providerTenant, false)
		tenantPresent(t, managementClient, ctx, provider, tenant, false)

		tenantPresent(t, managementClient, ctx, providerTenant, provider, true)
		tenantPresent(t, managementClient, ctx, providerTenant, providerTenant, true)
		tenantPresent(t, managementClient, ctx, providerTenant, tenant, false)

		t.Log("deleting provider")
		require.NoError(t, testutil.DeleteAndWaitUntilNotFound(managementClient, provider))
		assert.True(t, errors.IsNotFound(managementClient.Get(ctx, types.NamespacedName{
			Name: provider.Status.Namespace.Name,
		}, ns)), "namespace should also be deleted.")
		providerRoleAndRoleBindingPresent(t, managementClient, ctx, provider, false)
		tenantRoleAndRoleBindingPresent(t, managementClient, ctx, provider, false)

		t.Log("deleting providerTenant")
		require.NoError(t, testutil.DeleteAndWaitUntilNotFound(managementClient, providerTenant))
		assert.True(t, errors.IsNotFound(managementClient.Get(ctx, types.NamespacedName{
			Name: providerTenant.Status.Namespace.Name,
		}, ns)), "namespace should also be deleted.")
		providerRoleAndRoleBindingPresent(t, managementClient, ctx, providerTenant, false)
		tenantRoleAndRoleBindingPresent(t, managementClient, ctx, providerTenant, false)
	}
}

func tenantPresent(t *testing.T, managementClient client.Client, ctx context.Context, tenant *catalogv1alpha1.Account, provider *catalogv1alpha1.Account, expected bool) {
	trefs := &catalogv1alpha1.TenantList{}
	require.NoError(t, managementClient.List(ctx, trefs, client.InNamespace(provider.Status.Namespace.Name)))
	var found bool
	for _, tref := range trefs.Items {
		if tref.Name == tenant.Name {
			found = true
		}
	}
	assert.Equalf(t, expected, found, "tenant %s presence in provider %s", tenant.Name, provider.Name)
}

func providerRoleAndRoleBindingPresent(t *testing.T, managementClient client.Client, ctx context.Context, account *catalogv1alpha1.Account, expected bool) {
	var found bool
	role := &rbacv1.Role{}
	roleBinding := &rbacv1.RoleBinding{}
	if err := managementClient.Get(ctx, types.NamespacedName{
		Name:      "kubecarrier:provider",
		Namespace: account.Status.Namespace.Name,
	}, role); err == nil {
		found = true
	}
	assert.Equalf(t, expected, found, "provider Role of account %s", account.Name)
	found = false
	if err := managementClient.Get(ctx, types.NamespacedName{
		Name:      "kubecarrier:provider",
		Namespace: account.Status.Namespace.Name,
	}, roleBinding); err == nil {
		found = true
	}
	assert.Equalf(t, expected, found, "provider RoleBinding of account %s", account.Name)
}

func tenantRoleAndRoleBindingPresent(t *testing.T, managementClient client.Client, ctx context.Context, account *catalogv1alpha1.Account, expected bool) {
	var found bool
	role := &rbacv1.Role{}
	roleBinding := &rbacv1.RoleBinding{}
	if err := managementClient.Get(ctx, types.NamespacedName{
		Name:      "kubecarrier:tenant",
		Namespace: account.Status.Namespace.Name,
	}, role); err == nil {
		found = true
	}
	assert.Equalf(t, expected, found, "tenant Role of account %s", account.Name)
	found = false
	if err := managementClient.Get(ctx, types.NamespacedName{
		Name:      "kubecarrier:tenant",
		Namespace: account.Status.Namespace.Name,
	}, roleBinding); err == nil {
		found = true
	}
	assert.Equalf(t, expected, found, "tenant RoleBinding of account %s", account.Name)
}

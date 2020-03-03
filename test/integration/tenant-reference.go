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

package integration

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func newAccountRefs(f *testutil.Framework) func(t *testing.T) {
	return func(t *testing.T) {
		t.Log("testing how account handles tenant references")
		logger := testutil.NewLogger(t)
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		t.Cleanup(cancel)
		managementClient, err := f.ManagementClient(logger)
		require.NoError(t, err, "creating management client")
		t.Cleanup(managementClient.CleanUpFunc(ctx, t, f.Config().CleanUpStrategy))

		testName := strings.Replace(strings.ToLower(t.Name()), "/", "-", -1)

		var (
			mdata = catalogv1alpha1.AccountMetadata{
				DisplayName: "metadata name",
				Description: "metadata desc",
			}
			provider = &catalogv1alpha1.Account{
				ObjectMeta: metav1.ObjectMeta{
					Name: testName + "-provider",
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
					Name: testName + "-tenant",
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
					Name: testName + "-providertenant",
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
		require.NoError(t, managementClient.Create(ctx, provider), "creating provider")
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, provider))
		ns := &corev1.Namespace{}
		assert.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name: provider.Status.Namespace.Name,
		}, ns))

		t.Log("adding single tenant")
		require.NoError(t, managementClient.Create(ctx, tenant), "creating tenant")
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, tenant))
		assert.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name: tenant.Status.Namespace.Name,
		}, ns))

		tenantReferencePresent(t, managementClient, ctx, tenant, provider, true)
		tenantReferencePresent(t, managementClient, ctx, tenant, tenant, false)

		t.Log("adding providerTenant")
		require.NoError(t, managementClient.Create(ctx, providerTenant), "creating providerTenant")
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, providerTenant))
		assert.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name: providerTenant.Status.Namespace.Name,
		}, ns))

		tenantReferencePresent(t, managementClient, ctx, tenant, provider, true)
		tenantReferencePresent(t, managementClient, ctx, tenant, providerTenant, true)
		tenantReferencePresent(t, managementClient, ctx, tenant, tenant, false)

		tenantReferencePresent(t, managementClient, ctx, provider, provider, false)
		tenantReferencePresent(t, managementClient, ctx, provider, providerTenant, false)
		tenantReferencePresent(t, managementClient, ctx, provider, tenant, false)

		tenantReferencePresent(t, managementClient, ctx, providerTenant, provider, true)
		tenantReferencePresent(t, managementClient, ctx, providerTenant, providerTenant, true)
		tenantReferencePresent(t, managementClient, ctx, providerTenant, tenant, false)

		t.Log("deleting tenant")
		require.NoError(t, testutil.DeleteAndWaitUntilNotFound(ctx, managementClient, tenant))
		assert.True(t, errors.IsNotFound(managementClient.Get(ctx, types.NamespacedName{
			Name: tenant.Status.Namespace.Name,
		}, ns)), "namespace should also be deleted.")

		tenantReferencePresent(t, managementClient, ctx, tenant, provider, false)
		tenantReferencePresent(t, managementClient, ctx, tenant, providerTenant, false)
		tenantReferencePresent(t, managementClient, ctx, tenant, tenant, false)

		tenantReferencePresent(t, managementClient, ctx, provider, provider, false)
		tenantReferencePresent(t, managementClient, ctx, provider, providerTenant, false)
		tenantReferencePresent(t, managementClient, ctx, provider, tenant, false)

		tenantReferencePresent(t, managementClient, ctx, providerTenant, provider, true)
		tenantReferencePresent(t, managementClient, ctx, providerTenant, providerTenant, true)
		tenantReferencePresent(t, managementClient, ctx, providerTenant, tenant, false)

		t.Log("deleting provider")
		require.NoError(t, testutil.DeleteAndWaitUntilNotFound(ctx, managementClient, provider))
		assert.True(t, errors.IsNotFound(managementClient.Get(ctx, types.NamespacedName{
			Name: provider.Status.Namespace.Name,
		}, ns)), "namespace should also be deleted.")

		t.Log("deleting providerTenant")
		require.NoError(t, testutil.DeleteAndWaitUntilNotFound(ctx, managementClient, providerTenant))
		assert.True(t, errors.IsNotFound(managementClient.Get(ctx, types.NamespacedName{
			Name: providerTenant.Status.Namespace.Name,
		}, ns)), "namespace should also be deleted.")
	}
}

func tenantReferencePresent(t *testing.T, cl *testutil.RecordingClient, ctx context.Context, tenant *catalogv1alpha1.Account, provider *catalogv1alpha1.Account, expected bool) {
	tref := &catalogv1alpha1.TenantReference{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tenant.Name,
			Namespace: provider.Status.Namespace.Name,
		},
	}

	if expected {
		assert.NoError(t, testutil.WaitUntilFound(ctx, cl, tref), "tenantReference %s not found in provider %s", tenant.Name, provider.Name)
	} else {
		assert.NoError(t, testutil.WaitUntilNotFound(ctx, cl, tref), "tenantReference %s found in provider %s", tenant.Name, provider.Name)
	}
}

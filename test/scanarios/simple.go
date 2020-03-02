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

package scanarios

import (
	"context"
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

func newSimpleScenario(f *testutil.Framework) func(t *testing.T) {
	return func(t *testing.T) {
		// Setup
		logger := testutil.NewLogger(t)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		t.Cleanup(cancel)
		managementClient, err := f.ManagementClient(logger)
		require.NoError(t, err, "creating management client")
		t.Cleanup(managementClient.CleanUpFunc(ctx, t))

		serviceClient, err := f.ServiceClient(logger)
		require.NoError(t, err, "creating service client")
		t.Cleanup(serviceClient.CleanUpFunc(ctx, t))

		// Create a Tenant
		tenant := &catalogv1alpha1.Account{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "simple-tenant-",
			},
			Spec: catalogv1alpha1.AccountSpec{
				Metadata: catalogv1alpha1.AccountMetadata{
					DisplayName: "tenant display name",
					Description: "tenant desc",
				},
				Roles: []catalogv1alpha1.AccountRole{
					catalogv1alpha1.TenantRole,
				},
			},
		}
		require.NoError(t, managementClient.Create(ctx, tenant), "creating tenant error")
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, tenant))

		tenantNamespaceName := tenant.Status.Namespace.Name
		tenantNamespace := &corev1.Namespace{}
		assert.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name: tenantNamespaceName,
		}, tenantNamespace))

		// Create a Provider
		provider := &catalogv1alpha1.Account{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "simple-provider-",
			},
			Spec: catalogv1alpha1.AccountSpec{
				Metadata: catalogv1alpha1.AccountMetadata{
					DisplayName: "provider",
					Description: "provider test description",
				},
				Roles: []catalogv1alpha1.AccountRole{
					catalogv1alpha1.ProviderRole,
				},
			},
		}
		require.NoError(t, managementClient.Create(ctx, provider), "creating provider error")
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, provider))

		providerNamespaceName := provider.Status.Namespace.Name
		providerNamespace := &corev1.Namespace{}
		assert.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name: providerNamespaceName,
		}, providerNamespace))

		tenantReference := &catalogv1alpha1.TenantReference{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tenant.Name,
				Namespace: providerNamespaceName,
			},
		}
		t.Log("checking tenant reference")
		require.NoError(t, testutil.WaitUntilFound(ctx, managementClient, tenantReference))

		t.Log("deleting tenant")
		require.NoError(t, testutil.DeleteAndWaitUntilNotFound(ctx, managementClient, tenant))
		assert.True(t, errors.IsNotFound(managementClient.Get(ctx, types.NamespacedName{
			Name: tenantNamespaceName,
		}, tenantNamespace)), "namespace should also be deleted.")

		assert.True(t, errors.IsNotFound(managementClient.Get(ctx, types.NamespacedName{
			Name:      tenant.Name,
			Namespace: tenantNamespaceName,
		}, tenantReference)), "TenantReference should also be deleted.")

		t.Log("deleting provider")
		require.NoError(t, testutil.DeleteAndWaitUntilNotFound(ctx, managementClient, provider))
		assert.True(t, errors.IsNotFound(managementClient.Get(ctx, types.NamespacedName{
			Name: providerNamespaceName,
		}, providerNamespace)), "namespace should also be deleted.")
	}
}

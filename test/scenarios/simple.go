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

package scenarios

import (
	"context"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func newSimpleScenario(f *testutil.Framework) func(t *testing.T) {
	return func(t *testing.T) {
		// Setup
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		t.Cleanup(cancel)
		ctx = context.Background()

		managementClient, err := f.ManagementClient(t)
		require.NoError(t, err, "creating management client")
		t.Cleanup(managementClient.CleanUpFunc(ctx))

		serviceClient, err := f.ServiceClient(t)
		require.NoError(t, err, "creating service client")
		t.Cleanup(serviceClient.CleanUpFunc(ctx))
		testName := strings.Replace(strings.ToLower(t.Name()), "/", "-", -1)

		// Creating account
		t.Log("===== creating necessary accounts =====")
		var (
			tenantUser   = testName + "-tenant"
			providerUser = testName + "-provider"
		)
		tenant := f.NewTenantAccount(testName, rbacv1.Subject{
			Kind:     rbacv1.UserKind,
			APIGroup: "rbac.authorization.k8s.io",
			Name:     tenantUser,
		})
		provider := f.NewProviderAccount(testName, rbacv1.Subject{
			Kind:     rbacv1.UserKind,
			APIGroup: "rbac.authorization.k8s.io",
			Name:     providerUser,
		})

		require.NoError(t, managementClient.Create(ctx, tenant), "creating tenant error")
		require.NoError(t, managementClient.Create(ctx, provider), "creating provider error")
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, tenant))
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, provider))
		require.NotEmpty(t, tenant.Status.Namespace.Name)
		require.NotEmpty(t, provider.Status.Namespace.Name)

		t.Log("===== checking tenant =====")
		tenantReference := &catalogv1alpha1.Tenant{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tenant.Name,
				Namespace: provider.Status.Namespace.Name,
			},
		}
		require.NoError(t, testutil.WaitUntilFound(ctx, managementClient, tenantReference))

		t.Log("===== creating service cluster =====")
		serviceKubeconfig, err := ioutil.ReadFile(f.Config().ServiceInternalKubeconfigPath)
		require.NoError(t, err, "cannot read service internal kubeconfig")
		serviceClusterSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "eu-west-1",
				Namespace: provider.Status.Namespace.Name,
			},
			Data: map[string][]byte{
				"kubeconfig": serviceKubeconfig,
			},
		}
		serviceCluster := &corev1alpha1.ServiceCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "eu-west-1",
				Namespace: provider.Status.Namespace.Name,
			},
			Spec: corev1alpha1.ServiceClusterSpec{
				Metadata: corev1alpha1.ServiceClusterMetadata{
					DisplayName: "eu-west-1",
					Description: "eu-west-1 service cluster in German's capital",
				},
				KubeconfigSecret: corev1alpha1.ObjectReference{
					Name: "eu-west-1",
				},
			},
		}
		require.NoError(t, managementClient.Create(ctx, serviceClusterSecret))
		require.NoError(t, managementClient.Create(ctx, serviceCluster))
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, serviceCluster))

		t.Log("===== creating CRD on the service cluster =====")
		baseCRD := f.NewFakeCouchDBCRD(testName + ".test.kubecarrier.io")
		require.NoError(t, serviceClient.Create(ctx, baseCRD))

		catalogEntrySet := &catalogv1alpha1.CatalogEntrySet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "couchdb",
				Namespace: provider.Status.Namespace.Name,
			},
			Spec: catalogv1alpha1.CatalogEntrySetSpec{
				Metadata: catalogv1alpha1.CatalogEntrySetMetadata{
					DisplayName: "CouchDB",
					Description: "small database living near Tegel airport",
				},
				Derive: &catalogv1alpha1.DerivedConfig{
					KindOverride: "CouchDB",
					Expose: []catalogv1alpha1.VersionExposeConfig{
						{
							Versions: []string{
								"v1alpha1",
							},
							Fields: []catalogv1alpha1.FieldPath{
								{JSONPath: ".spec.prop1"},
								{JSONPath: ".status.observedGeneration"},
								{JSONPath: ".status.prop1"},
							},
						},
					},
				},
				DiscoverySet: catalogv1alpha1.CustomResourceDiscoverySetConfig{
					CRD: catalogv1alpha1.ObjectReference{
						Name: baseCRD.Name,
					},
					ServiceClusterSelector: metav1.LabelSelector{},
					KindOverride:           "CouchDBInternal",
				},
			},
		}
		require.NoError(t, managementClient.Create(ctx, catalogEntrySet))
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, catalogEntrySet))

		internalCRD := &apiextensions.CustomResourceDefinition{}
		require.NotEmpty(t, managementClient.Get(ctx, types.NamespacedName{
			Name: strings.Join([]string{"couchdbinternals", serviceCluster.Name, provider.Name}, "."),
		}, internalCRD))
		externalCRD := &apiextensions.CustomResourceDefinition{}
		require.NotEmpty(t, managementClient.Get(ctx, types.NamespacedName{
			Name: strings.Join([]string{"couchdb", serviceCluster.Name, provider.Name}, "."),
		}, externalCRD))

		catalog := &catalogv1alpha1.Catalog{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "default",
				Namespace: provider.Status.Namespace.Name,
			},
			Spec: catalogv1alpha1.CatalogSpec{
				CatalogEntrySelector: &metav1.LabelSelector{},
				TenantSelector:       &metav1.LabelSelector{},
			},
		}
		require.NoError(t, managementClient.Create(ctx, catalog))
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, catalog))

		tenantClient, err := f.ManagementClient(t, func(config *rest.Config) error {
			config.Impersonate = rest.ImpersonationConfig{
				UserName: tenantUser,
			}
			return nil
		})
		require.NoError(t, err)
		t.Cleanup(tenantClient.CleanUpFunc(ctx))

		{
			offeringList := &catalogv1alpha1.OfferingList{}
			require.NoError(t, tenantClient.List(ctx, offeringList, client.InNamespace(tenant.Status.Namespace.Name)))
			assert.NotEmpty(t, offeringList.Items, "no offerings found")
			for _, it := range offeringList.Items {
				t.Logf("tenant %s has offerring %s", tenant.Name, it.Name)
			}
			offering := &catalogv1alpha1.Offering{}
			if assert.NoError(t, tenantClient.Get(ctx, types.NamespacedName{
				Namespace: tenant.Status.Namespace.Name,
				Name:      strings.Join([]string{"couchdbs", serviceCluster.Name, provider.Name}, "."),
			}, offering), "tenant %s doesn't have the required offering", tenant.Name) {
				// TODO: create the off
			}

			providerList := &catalogv1alpha1.ProviderList{}
			require.NoError(t, tenantClient.List(ctx, providerList, client.InNamespace(tenant.Status.Namespace.Name)))
			assert.NotEmpty(t, providerList.Items, "no offerings found")
			for _, it := range providerList.Items {
				t.Logf("tenant %s has provider %s", tenant.Name, it.Name)
			}
		}

		{
			providerClient, err := f.ManagementClient(t, func(config *rest.Config) error {
				config.Impersonate = rest.ImpersonationConfig{
					UserName: providerUser,
				}
				return nil
			})
			require.NoError(t, err)
			t.Cleanup(providerClient.CleanUpFunc(ctx))

			tenantList := &catalogv1alpha1.TenantList{}
			require.NoError(t, providerClient.List(ctx, tenantList, client.InNamespace(provider.Status.Namespace.Name)))
			assert.NotEmpty(t, tenantList.Items, "no tenants found")
			var tenantFound bool
			for _, it := range tenantList.Items {
				t.Logf("provider %s has tenant %s", provider.Name, it.Name)
				if it.Name == tenant.Name {
					tenantFound = true
				}
			}
			assert.True(t, tenantFound, "cannot find tenant %s for the provider %s", tenant.Name, provider.Name)
		}
	}
}

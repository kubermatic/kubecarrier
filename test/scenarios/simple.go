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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

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
		tenantAccount := f.NewTenantAccount(testName, rbacv1.Subject{
			Kind:     rbacv1.UserKind,
			APIGroup: "rbac.authorization.k8s.io",
			Name:     tenantUser,
		})
		provider := f.NewProviderAccount(testName, rbacv1.Subject{
			Kind:     rbacv1.UserKind,
			APIGroup: "rbac.authorization.k8s.io",
			Name:     providerUser,
		})

		require.NoError(t, managementClient.Create(ctx, tenantAccount), "creating tenant error")
		require.NoError(t, managementClient.Create(ctx, provider), "creating provider error")
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, tenantAccount))
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, provider))
		require.NotEmpty(t, tenantAccount.Status.Namespace.Name)
		require.NotEmpty(t, provider.Status.Namespace.Name)

		t.Log("===== checking tenant =====")
		tenant := &catalogv1alpha1.Tenant{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tenantAccount.Name,
				Namespace: provider.Status.Namespace.Name,
			},
		}
		require.NoError(t, testutil.WaitUntilFound(ctx, managementClient, tenant))

		t.Log("===== creating service cluster =====")
		serviceCluster := f.SetupServiceCluster(ctx, managementClient, t, "eu-east-1", provider)

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
				Discover: catalogv1alpha1.CustomResourceDiscoverySetConfig{
					CRD: catalogv1alpha1.ObjectReference{
						Name: baseCRD.Name,
					},
					ServiceClusterSelector: metav1.LabelSelector{},
					KindOverride:           "CouchDBInternal",
				},
			},
		}
		require.NoError(t, managementClient.Create(ctx, catalogEntrySet))
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, catalogEntrySet, testutil.WithTimeout(time.Minute)))

		internalCRD := &apiextensionsv1.CustomResourceDefinition{}
		require.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name: strings.Join([]string{"couchdbinternals", serviceCluster.Name, provider.Name}, "."),
		}, internalCRD))
		externalCRD := &apiextensionsv1.CustomResourceDefinition{}
		require.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name: strings.Join([]string{"couchdbs", serviceCluster.Name, provider.Name}, "."),
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

		providerClient, err := f.ManagementClient(t, func(config *rest.Config) error {
			config.Impersonate = rest.ImpersonationConfig{
				UserName: providerUser,
			}
			return nil
		})
		require.NoError(t, err)
		t.Cleanup(providerClient.CleanUpFunc(ctx))
		{
			offeringList := &catalogv1alpha1.OfferingList{}
			require.NoError(t, tenantClient.List(ctx, offeringList, client.InNamespace(tenantAccount.Status.Namespace.Name)))
			assert.NotEmpty(t, offeringList.Items, "no offerings found")
			for _, it := range offeringList.Items {
				t.Logf("tenant %s has offerring %s", tenantAccount.Name, it.Name)
			}
			offering := &catalogv1alpha1.Offering{}
			if assert.NoError(t, tenantClient.Get(ctx, types.NamespacedName{
				Namespace: tenantAccount.Status.Namespace.Name,
				Name:      strings.Join([]string{"couchdbs", serviceCluster.Name, provider.Name}, "."),
			}, offering), "tenant %s doesn't have the required offering", tenantAccount.Name) {
				assert.Equal(t, externalCRD.Name, offering.Spec.CRD.Name)
				externalObj := &unstructured.Unstructured{}
				externalObj.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   offering.Spec.CRD.APIGroup,
					Version: offering.Spec.CRD.Versions[0].Name,
					Kind:    offering.Spec.CRD.Kind,
				})
				externalObj.SetNamespace(tenantAccount.Status.Namespace.Name)
				externalObj.SetName("db1")
				externalObj.Object["spec"] = map[string]interface{}{
					"prop1": "dummy value",
				}
				require.NoError(t, tenantClient.Create(ctx, externalObj))

				t.Log("checking internal object existance")
				internalObj := &unstructured.Unstructured{}
				internalObj.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   internalCRD.Spec.Group,
					Version: internalCRD.Spec.Versions[0].Name,
					Kind:    internalCRD.Spec.Names.Kind,
				})
				internalObj.SetNamespace(tenantAccount.Status.Namespace.Name)
				internalObj.SetName(externalObj.GetName())
				assert.NoError(t,
					testutil.WaitUntilFound(ctx, providerClient, internalObj, testutil.WithTimeout(15*time.Second)),
					"cannot find the CRD on the service cluster within the time limit",
				)

				sca := &corev1alpha1.ServiceClusterAssignment{}
				require.NoError(t, managementClient.Get(ctx, types.NamespacedName{
					Namespace: provider.Status.Namespace.Name,
					Name:      tenantAccount.Name + "." + serviceCluster.Name,
				}, sca))

				t.Log("checking internal object")
				svcObj := &unstructured.Unstructured{}
				svcObj.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   baseCRD.Spec.Group,
					Version: baseCRD.Spec.Versions[0].Name,
					Kind:    baseCRD.Spec.Names.Kind,
				})
				svcObj.SetName(externalObj.GetName())
				svcObj.SetNamespace(sca.Status.ServiceClusterNamespace.Name)

				t.Log("checking service cluster object")
				assert.NoError(t,
					testutil.WaitUntilFound(ctx, serviceClient, svcObj, testutil.WithTimeout(15*time.Second)),
					"cannot find the CRD on the service cluster within the time limit",
				)
				assert.NoError(t, testutil.DeleteAndWaitUntilNotFound(ctx, tenantClient, externalObj))
				assert.NoError(t, testutil.WaitUntilNotFound(ctx, serviceClient, svcObj))
			}

			providerList := &catalogv1alpha1.ProviderList{}
			require.NoError(t, tenantClient.List(ctx, providerList, client.InNamespace(tenantAccount.Status.Namespace.Name)))
			assert.NotEmpty(t, providerList.Items, "no offerings found")
			for _, it := range providerList.Items {
				t.Logf("tenant %s has provider %s", tenantAccount.Name, it.Name)
			}
		}

		{

			tenantList := &catalogv1alpha1.TenantList{}
			require.NoError(t, providerClient.List(ctx, tenantList, client.InNamespace(provider.Status.Namespace.Name)))
			assert.NotEmpty(t, tenantList.Items, "no tenants found")
			var tenantFound bool
			for _, it := range tenantList.Items {
				t.Logf("provider %s has tenant %s", provider.Name, it.Name)
				if it.Name == tenantAccount.Name {
					tenantFound = true
				}
			}
			assert.True(t, tenantFound, "cannot find tenant %s for the provider %s", tenantAccount.Name, provider.Name)
		}
	}
}

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
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func newSimpleScenario(f *testutil.Framework) func(t *testing.T) {
	return func(t *testing.T) {
		// Setup
		logger := testutil.NewLogger(t)
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		t.Cleanup(cancel)
		managementClient, err := f.ManagementClient(logger)
		require.NoError(t, err, "creating management client")
		t.Cleanup(managementClient.CleanUpFunc(ctx, t, f.Config().CleanUpStrategy))

		serviceClient, err := f.ServiceClient(logger)
		require.NoError(t, err, "creating service client")
		t.Cleanup(serviceClient.CleanUpFunc(ctx, t, f.Config().CleanUpStrategy))
		testName := strings.Replace(strings.ToLower(t.Name()), "/", "-", -1)

		// Create a Tenant
		tenant := &catalogv1alpha1.Account{
			ObjectMeta: metav1.ObjectMeta{
				Name: testName + "-tenant",
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
				Name: testName + "-provider",
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

		serviceKubeconfig, err := ioutil.ReadFile(f.Config().ServiceInternalKubeconfigPath)
		require.NoError(t, err, "cannot read service internal kubeconfig")
		t.Log("creating service cluster")
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

		t.Log("creating CRD on the service cluster")
		baseCRD := &apiextensionsv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "redis.test.kubecarrier.io",
			},
			Spec: apiextensionsv1.CustomResourceDefinitionSpec{
				Group: "test.kubecarrier.io",
				Names: apiextensionsv1.CustomResourceDefinitionNames{
					Kind:     "Redis",
					ListKind: "RedisList",
					Plural:   "redis",
					Singular: "redis",
				},
				Scope: apiextensionsv1.NamespaceScoped,
				Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
					{
						Name:    "v1alpha1",
						Served:  true,
						Storage: true,
						Subresources: &apiextensionsv1.CustomResourceSubresources{
							Status: &apiextensionsv1.CustomResourceSubresourceStatus{},
						},
						Schema: &apiextensionsv1.CustomResourceValidation{
							OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
								Properties: map[string]apiextensionsv1.JSONSchemaProps{
									"apiVersion": {Type: "string"},
									"kind":       {Type: "string"},
									"metadata":   {Type: "object"},
									"spec": {
										Type: "object",
										Properties: map[string]apiextensionsv1.JSONSchemaProps{
											"prop1": {Type: "string"},
											"prop2": {Type: "string"},
										},
									},
									"status": {
										Type: "object",
										Properties: map[string]apiextensionsv1.JSONSchemaProps{
											"observedGeneration": {Type: "integer"},
											"prop1":              {Type: "string"},
											"prop2":              {Type: "string"},
										},
									},
								},
								Type: "object",
							},
						},
					},
				},
			},
		}
		require.NoError(t, serviceClient.Create(ctx, baseCRD))
		t.Log("creating Custrom Resouce Discovery")
		crDiscovery := &corev1alpha1.CustomResourceDiscovery{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "redis",
				Namespace: provider.Status.Namespace.Name,
			},
			Spec: corev1alpha1.CustomResourceDiscoverySpec{
				CRD: corev1alpha1.ObjectReference{
					baseCRD.Name,
				},
				ServiceCluster: corev1alpha1.ObjectReference{
					Name: serviceCluster.Name,
				},
				KindOverride: "",
			},
		}
		require.NoError(t, managementClient.Create(ctx, crDiscovery))
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, crDiscovery))
	}
}

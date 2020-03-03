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

package provider

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
	"github.com/kubermatic/kubecarrier/test/framework"
)

// ServiceClusterSuite registers a ServiceCluster and tests apis interacting with it.
func NewServiceClusterSuite(
	f *framework.Framework,
	provider *catalogv1alpha1.Account,
) func(t *testing.T) {
	return func(t *testing.T) {
		managementClient, err := f.ManagementClient()
		require.NoError(t, err, "creating management client")
		defer managementClient.CleanUp(t)

		serviceClient, err := f.ServiceClient()
		require.NoError(t, err, "creating service client")
		defer serviceClient.CleanUp(t)

		// Setup
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
					Description: "eu-west-1 service cluster",
				},
				KubeconfigSecret: corev1alpha1.ObjectReference{
					Name: "eu-west-1",
				},
			},
		}

		crd := &apiextensionsv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "redis.test.kubecarrier.io",
				Labels: map[string]string{
					"kubecarrier.io/service-cluster":  serviceCluster.Name,
					"kubecarrier.io/origin-namespace": provider.Status.Namespace.Name,
				},
			},
			Spec: apiextensionsv1.CustomResourceDefinitionSpec{
				Group: "test.kubecarrier.io",
				Names: apiextensionsv1.CustomResourceDefinitionNames{
					Singular: "redis",
					Plural:   "redis",
					Kind:     "Redis",
					ListKind: "RedisList",
				},
				Scope: apiextensionsv1.NamespaceScoped,
				Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
					{
						Name:    "v1alpha1",
						Served:  true,
						Storage: true,
						Schema: &apiextensionsv1.CustomResourceValidation{
							OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
								Type: "object",
							},
						},
						Subresources: &apiextensionsv1.CustomResourceSubresources{
							Status: &apiextensionsv1.CustomResourceSubresourceStatus{},
						},
					},
				},
			},
		}

		serviceNamespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "servicecluster-svc-test",
			},
		}

		serviceClusterAssignment := &corev1alpha1.ServiceClusterAssignment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceNamespace.Name + "." + serviceCluster.Name,
				Namespace: provider.Status.Namespace.Name,
			},
			Spec: corev1alpha1.ServiceClusterAssignmentSpec{
				ServiceCluster: corev1alpha1.ObjectReference{
					Name: serviceCluster.Name,
				},
				ManagementClusterNamespace: corev1alpha1.ObjectReference{
					Name: serviceNamespace.Name,
				},
			},
		}

		ctx := context.Background()
		require.NoError(t, managementClient.Create(ctx, serviceClusterSecret))
		require.NoError(t, managementClient.Create(ctx, serviceCluster))
		require.NoError(t, managementClient.Create(ctx, serviceNamespace))
		require.NoError(t, managementClient.Create(ctx, serviceClusterAssignment))
		require.NoError(t, testutil.WaitUntilReady(managementClient, serviceCluster))
		require.NoError(t, testutil.WaitUntilReady(managementClient, serviceClusterAssignment))
		require.NoError(t, serviceClient.Create(ctx, crd))

		// Test CatalogEntrySet
		catalogEntrySet := &catalogv1alpha1.CatalogEntrySet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "redis",
				Namespace: provider.Status.Namespace.Name,
			},
			Spec: catalogv1alpha1.CatalogEntrySetSpec{
				Metadata: catalogv1alpha1.CatalogEntrySetMetadata{
					DisplayName: "Test CatalogEntrySet",
					Description: "Test CatalogEntrySet",
				},
				DiscoverySet: catalogv1alpha1.CustomResourceDiscoverySetConfig{
					CRD: catalogv1alpha1.ObjectReference{
						Name: crd.Name,
					},
					ServiceClusterSelector: metav1.LabelSelector{},
					KindOverride:           "RedisInternal",
					WebhookStrategy:        corev1alpha1.WebhookStrategyTypeServiceCluster,
				},
			},
		}
		require.NoError(t, managementClient.Create(ctx, catalogEntrySet))
		require.NoError(t, testutil.WaitUntilReady(managementClient, catalogEntrySet))

		// Check the CustomResourceDiscoverySet
		crDiscoverySet := &corev1alpha1.CustomResourceDiscoverySet{}
		require.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name:      catalogEntrySet.Name,
			Namespace: catalogEntrySet.Namespace,
		}, crDiscoverySet), "getting CustomResourceDiscoverySet")

		// Check the CatalogEntry Object
		catalogEntry := &catalogv1alpha1.CatalogEntry{}
		require.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name:      "redisinternals.eu-west-1.test-servicecluster",
			Namespace: catalogEntrySet.Namespace,
		}, catalogEntry), "getting CatalogEntry")

		err = managementClient.Delete(ctx, provider)
		if assert.Error(t, err, "dirty provider %s deletion should error out", provider.Name) {
			assert.Equal(t,
				`admission webhook "vaccount.kubecarrier.io" denied the request: deletion blocking objects found:
CustomResourceDiscovery.kubecarrier.io/v1alpha1: redis.eu-west-1
CustomResourceDiscoverySet.kubecarrier.io/v1alpha1: redis
ServiceClusterAssignment.kubecarrier.io/v1alpha1: servicecluster-svc-test.eu-west-1
`,
				err.Error(),
				"deleting dirty provider %s", provider.Name)
		}

		// We have created/registered new CRD's, so we need a new client
		managementClient, err = f.ManagementClient()
		require.NoError(t, err, "creating management client")
		defer managementClient.CleanUp(t)

		serviceClient, err = f.ServiceClient()
		require.NoError(t, err, "creating service client")
		defer serviceClient.CleanUp(t)

		// management cluster -> service cluster
		//
		managementClusterObj := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": fmt.Sprintf("%s.%s/v1alpha1", serviceCluster.Name, provider.Name),
				"kind":       "RedisInternal",
				"metadata": map[string]interface{}{
					"name":      "test-instance-1",
					"namespace": serviceNamespace.Name,
				},
				"spec": map[string]interface{}{
					"prop1": "test1",
				},
			},
		}
		require.NoError(t, managementClient.Create(ctx, managementClusterObj))

		// a object on the service cluster should have been created
		serviceClusterObj := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "test.kubecarrier.io/v1alpha1",
				"kind":       "Redis",
				"metadata": map[string]interface{}{
					"name":      managementClusterObj.GetName(),
					"namespace": serviceClusterAssignment.Status.ServiceClusterNamespace.Name,
				},
			},
		}
		require.NoError(
			t, testutil.WaitUntilFound(serviceClient, serviceClusterObj))

		// service cluster -> management cluster
		//
		serviceClusterObj2 := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "test.kubecarrier.io/v1alpha1",
				"kind":       "Redis",
				"metadata": map[string]interface{}{
					"name":      "test-instance-2",
					"namespace": serviceClusterAssignment.Status.ServiceClusterNamespace.Name,
				},
				"spec": map[string]interface{}{
					"prop1": "test1",
				},
			},
		}
		require.NoError(t, serviceClient.Create(ctx, serviceClusterObj2))
		// we need to unregister this object,
		// as the management cluster takes control and will just recreate it.
		serviceClient.UnregisterForCleanup(serviceClusterObj2)

		// a object on the management cluster should have been created
		managementClusterObj2 := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": fmt.Sprintf("%s.%s/v1alpha1", serviceCluster.Name, provider.Name),
				"kind":       "RedisInternal",
				"metadata": map[string]interface{}{
					"name":      serviceClusterObj2.GetName(),
					"namespace": serviceNamespace.Name,
				},
			},
		}
		managementClient.RegisterForCleanup(managementClusterObj2)
		require.NoError(
			t, testutil.WaitUntilFound(managementClient, managementClusterObj2))
	}
}

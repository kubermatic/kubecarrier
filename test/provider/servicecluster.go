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
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
	"github.com/kubermatic/kubecarrier/test/framework"
)

// ServiceClusterSuite registers a ServiceCluster and tests apis interacting with it.
func NewServiceClusterSuite(
	f *framework.Framework,
	provider *catalogv1alpha1.Provider,
) func(t *testing.T) {
	return func(t *testing.T) {
		masterClient, err := f.MasterClient()
		require.NoError(t, err, "creating master client")
		defer masterClient.CleanUp(t)

		serviceClient, err := f.ServiceClient()
		require.NoError(t, err, "creating service client")
		defer serviceClient.CleanUp(t)

		// Setup
		serviceKubeconfig, err := ioutil.ReadFile(f.Config().ServiceInternalKubeconfigPath)
		require.NoError(t, err, "cannot read service internal kubeconfig")

		serviceClusterSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "eu-west-1",
				Namespace: provider.Status.NamespaceName,
			},
			Data: map[string][]byte{
				"kubeconfig": serviceKubeconfig,
			},
		}

		serviceClusterRegistration := &operatorv1alpha1.ServiceClusterRegistration{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "eu-west-1",
				Namespace: provider.Status.NamespaceName,
			},
			Spec: operatorv1alpha1.ServiceClusterRegistrationSpec{
				KubeconfigSecret: operatorv1alpha1.ObjectReference{
					Name: "eu-west-1",
				},
			},
		}

		crd := &apiextensionsv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "redis.test.kubecarrier.io",
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

		ctx := context.Background()
		require.NoError(t, masterClient.Create(ctx, serviceClusterSecret))
		require.NoError(t, masterClient.Create(ctx, serviceClusterRegistration))
		require.NoError(t, testutil.WaitUntilReady(masterClient, serviceClusterRegistration))
		require.NoError(t, serviceClient.Create(ctx, crd))

		// Check if the ServiceCluster becomes ready
		serviceCluster := &corev1alpha1.ServiceCluster{}
		serviceCluster.SetName(serviceClusterRegistration.Name)
		serviceCluster.SetNamespace(provider.Status.NamespaceName)
		require.NoError(t, testutil.WaitUntilReady(masterClient, serviceCluster))

		// Test CustomResourceDefinitionDiscoverySet
		crdds := &corev1alpha1.CustomResourceDefinitionDiscoverySet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "redis",
				Namespace: provider.Status.NamespaceName,
			},
			Spec: corev1alpha1.CustomResourceDefinitionDiscoverySetSpec{
				KindOverride: "RedisInternal",
				CRD: corev1alpha1.ObjectReference{
					Name: crd.Name,
				},
			},
		}
		require.NoError(t, masterClient.Create(ctx, crdds))
		require.NoError(t, testutil.WaitUntilReady(masterClient, crdds))

		// We have created/registered new CRD's, so we need a new client
		masterClient, err = f.MasterClient()
		require.NoError(t, err, "creating master client")
		defer masterClient.CleanUp(t)

		serviceClient, err = f.ServiceClient()
		require.NoError(t, err, "creating service client")
		defer serviceClient.CleanUp(t)

		// makes sure we delete the ServiceClusterAssignment object BEFORE
		// the ServiceClusterRegistration - or the finalizer will block cleanup of the
		// NOTE: we need to register this BEFORE creating the master cluster obj,
		// so it will be cleaned up AFTER it.
		// otherwise the controller will just recreate it.
		serviceClusterAssignment := &corev1alpha1.ServiceClusterAssignment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      provider.Status.NamespaceName + "." + serviceCluster.Name,
				Namespace: provider.Status.NamespaceName,
			},
		}
		masterClient.RegisterForCleanup(serviceClusterAssignment)

		// master cluster -> service cluster
		//
		masterClusterObj := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "eu-west-1.test-servicecluster/v1alpha1",
				"kind":       "RedisInternal",
				"metadata": map[string]interface{}{
					"name":      "test-instance-1",
					"namespace": provider.Status.NamespaceName,
				},
				"spec": map[string]interface{}{
					"prop1": "test1",
				},
			},
		}
		require.NoError(t, masterClient.Create(ctx, masterClusterObj))

		// there should be a ServiceClusterAssignment for this namespace
		require.NoError(
			t, testutil.WaitUntilReady(masterClient, serviceClusterAssignment))

		// a object on the service cluster should have been created
		serviceClusterObj := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "test.kubecarrier.io/v1alpha1",
				"kind":       "Redis",
				"metadata": map[string]interface{}{
					"name":      masterClusterObj.GetName(),
					"namespace": serviceClusterAssignment.Status.ServiceClusterNamespace.Name,
				},
			},
		}
		require.NoError(
			t, testutil.WaitUntilFound(serviceClient, serviceClusterObj))

		// service cluster -> master cluster
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
		// as the master cluster takes control and will just recreate it.
		serviceClient.UnregisterForCleanup(serviceClusterObj2)

		// a object on the master cluster should have been created
		masterClusterObj2 := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "eu-west-1.test-servicecluster/v1alpha1",
				"kind":       "RedisInternal",
				"metadata": map[string]interface{}{
					"name":      serviceClusterObj2.GetName(),
					"namespace": provider.Status.NamespaceName,
				},
			},
		}
		masterClient.RegisterForCleanup(masterClusterObj2)
		require.NoError(
			t, testutil.WaitUntilFound(masterClient, masterClusterObj2))
	}
}

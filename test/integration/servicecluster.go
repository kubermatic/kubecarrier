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
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	fakev1 "github.com/kubermatic/kubecarrier/pkg/apis/fake/v1"
	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

// ServiceClusterSuite registers a ServiceCluster and tests apis interacting with it.
func newServiceClusterSuite(
	f *testutil.Framework,
) func(t *testing.T) {
	return func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		managementClient, err := f.ManagementClient(t)
		require.NoError(t, err, "creating management client")
		t.Cleanup(managementClient.CleanUpFunc(ctx))

		serviceClient, err := f.ServiceClient(t)
		require.NoError(t, err, "creating service client")
		t.Cleanup(serviceClient.CleanUpFunc(ctx))

		testName := strings.Replace(strings.ToLower(t.Name()), "/", "-", -1)

		provider := testutil.NewProviderAccount(testName, rbacv1.Subject{
			Kind:     rbacv1.GroupKind,
			APIGroup: "rbac.authorization.k8s.io",
			Name:     "provider",
		})
		require.NoError(t, managementClient.Create(ctx, provider))
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, provider))

		// Setup
		serviceCluster := f.SetupServiceCluster(ctx, managementClient, t, "eu-west-1", provider)

		crd := &apiextensionsv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "dbs.fake.kubecarrier.io",
			},
		}
		require.NoError(t, serviceClient.Get(ctx, types.NamespacedName{
			Name: crd.Name,
		}, crd), "getting fake DB crd in service cluster")

		serviceNamespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: testName + "-svc-test",
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

		require.NoError(t, managementClient.Create(ctx, serviceNamespace))
		require.NoError(t, managementClient.Create(ctx, serviceClusterAssignment))
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, serviceClusterAssignment))
		t.Log("service cluster successfully created")

		// Test CatalogEntrySet
		catalogEntrySet := &catalogv1alpha1.CatalogEntrySet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testName,
				Namespace: provider.Status.Namespace.Name,
			},
			Spec: catalogv1alpha1.CatalogEntrySetSpec{
				Metadata: catalogv1alpha1.CatalogEntrySetMetadata{
					CommonMetadata: catalogv1alpha1.CommonMetadata{
						DisplayName:      "Test CatalogEntrySet",
						ShortDescription: "Test CatalogEntrySet",
					},
				},
				Discover: catalogv1alpha1.CustomResourceDiscoverySetConfig{
					CRD: catalogv1alpha1.ObjectReference{
						Name: crd.Name,
					},
					ServiceClusterSelector: metav1.LabelSelector{},
					KindOverride:           "DBInternal",
					WebhookStrategy:        corev1alpha1.WebhookStrategyTypeServiceCluster,
				},
			},
		}
		require.NoError(t, managementClient.Create(ctx, catalogEntrySet))
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, catalogEntrySet, testutil.WithTimeout(2*time.Minute)))
		t.Log("CatalogEntrySet successfully created")

		// Check the CustomResourceDiscoverySet
		crDiscoverySet := &corev1alpha1.CustomResourceDiscoverySet{}
		require.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name:      catalogEntrySet.Name,
			Namespace: catalogEntrySet.Namespace,
		}, crDiscoverySet), "getting CustomResourceDiscoverySet")
		assert.Equal(t, crDiscoverySet.Spec.WebhookStrategy, catalogEntrySet.Spec.Discover.WebhookStrategy)

		// Check the CustomResourceDiscovery object
		customResourceDiscovery := &corev1alpha1.CustomResourceDiscovery{}
		require.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name:      crDiscoverySet.Name + "." + serviceCluster.Name,
			Namespace: catalogEntrySet.Namespace,
		}, customResourceDiscovery), "getting CustomResourceDiscovery")

		// Check the CatalogEntry Object
		catalogEntry := &catalogv1alpha1.CatalogEntry{}
		require.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name:      catalogEntrySet.Name + "." + serviceCluster.Name,
			Namespace: catalogEntrySet.Namespace,
		}, catalogEntry), "getting CatalogEntry")

		t.Log("CatalogEntry & CustomResourceDiscoverySet exists")

		// Check Catapult
		catapult := &operatorv1alpha1.Catapult{}
		require.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name:      catalogEntrySet.Name + "." + serviceCluster.Name,
			Namespace: catalogEntrySet.Namespace,
		}, catapult), "getting catapult")

		t.Log("Catapult exists")
		catapult.Spec.Paused = true
		require.NoError(t, managementClient.Update(ctx, catapult), "set catapult paused flag to true")
		require.NoError(t, testutil.WaitUntilCondition(ctx, managementClient, catapult, operatorv1alpha1.CatapultPaused, operatorv1alpha1.ConditionTrue))
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, catapult))
		require.Equal(t, operatorv1alpha1.CatapultPhasePaused, catapult.Status.Phase)
		catapult.Spec.Paused = false
		require.NoError(t, managementClient.Update(ctx, catapult), "set catapult paused flag to true")
		require.NoError(t, testutil.WaitUntilCondition(ctx, managementClient, catapult, operatorv1alpha1.CatapultPaused, operatorv1alpha1.ConditionFalse))
		require.Equal(t, operatorv1alpha1.CatapultPhaseReady, catapult.Status.Phase)

		// Check the Catapult dynamic webhook service is deployed.
		webhookService := &corev1.Service{}
		assert.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-%s-catapult-webhook-service", catalogEntrySet.Name, serviceCluster.Name),
			Namespace: catalogEntrySet.Namespace,
		}, webhookService), "get the Webhook Service that owned by Catapult object")

		err = managementClient.Delete(ctx, provider)
		if assert.Error(t, err, "dirty provider %s deletion should error out", provider.Name) {
			assert.Equal(t,
				fmt.Sprintf(`admission webhook "vaccount.kubecarrier.io" denied the request: deletion blocking objects found:
CustomResourceDiscovery.kubecarrier.io/v1alpha1: %s.eu-west-1
CustomResourceDiscoverySet.kubecarrier.io/v1alpha1: %s
ServiceClusterAssignment.kubecarrier.io/v1alpha1: %s.eu-west-1
`, catalogEntrySet.Name, catalogEntrySet.Name, serviceNamespace.Name),
				err.Error(),
				"deleting dirty provider %s", provider.Name)
		}

		// management cluster -> service cluster
		//
		managementClusterObj := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": fmt.Sprintf("%s.%s/v1", serviceCluster.Name, provider.Name),
				"kind":       "DBInternal",
				"metadata": map[string]interface{}{
					"name":      "test-instance-1",
					"namespace": serviceNamespace.Name,
				},
				"spec": map[string]interface{}{
					"databaseName": "test-instance-1",
					"databaseUser": "test-instance-user",
					"config": map[string]interface{}{
						"create": "Enabled",
					},
				},
			},
		}
		require.NoError(t, managementClient.Create(ctx, managementClusterObj))

		err = managementClient.Delete(ctx, customResourceDiscovery)
		if assert.Error(t, err,
			"CRDiscovery object must not be allowed to delete if ManagementClusterCRD instances are present",
		) {
			assert.Contains(
				t,
				err.Error(),
				"management cluster CRD instances are still present in the management cluster",
				"CRDiscovery deletion webhook should error out on ManagementClusterCRD instance presence",
			)
		}

		// a object on the service cluster should have been created
		serviceClusterObj := &fakev1.DB{
			ObjectMeta: metav1.ObjectMeta{
				Name:      managementClusterObj.GetName(),
				Namespace: serviceClusterAssignment.Status.ServiceClusterNamespace.Name,
			},
		}
		require.NoError(
			t, testutil.WaitUntilFound(ctx, serviceClient, serviceClusterObj))

		// service cluster -> management cluster
		//
		serviceClusterObj2 := testutil.NewFakeDB("test-instance-2", serviceClusterAssignment.Status.ServiceClusterNamespace.Name)
		require.NoError(t, serviceClient.Create(ctx, serviceClusterObj2))
		// we need to unregister this object,
		// as the management cluster takes control and will just recreate it.
		serviceClient.UnregisterForCleanup(serviceClusterObj2)

		// a object on the management cluster should have been created
		managementClusterObj2 := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": fmt.Sprintf("%s.%s/v1", serviceCluster.Name, provider.Name),
				"kind":       "DBInternal",
				"metadata": map[string]interface{}{
					"name":      serviceClusterObj2.GetName(),
					"namespace": serviceNamespace.Name,
				},
			},
		}
		managementClient.RegisterForCleanup(managementClusterObj2)
		require.NoError(
			t, testutil.WaitUntilFound(ctx, managementClient, managementClusterObj2))
	}
}

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

package controllers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func TestManagementClusterObjReconciler(t *testing.T) {
	managementClusterObj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":      "test-1",
				"namespace": "another-namespace",
				"labels": map[string]interface{}{
					"l1": "v1",
				},
				"generation": int64(2),
			},
			"spec": map[string]interface{}{
				"test1": "spec2000",
			},
		},
	}
	managementClusterObj.SetGroupVersionKind(managementClusterGVK)

	sca := &corev1alpha1.ServiceClusterAssignment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "another-namespace.eu-west-1",
			Namespace: providerNamespace,
		},
		Spec: corev1alpha1.ServiceClusterAssignmentSpec{
			ServiceCluster: corev1alpha1.ObjectReference{
				Name: "eu-west-1",
			},
			ManagementClusterNamespace: corev1alpha1.ObjectReference{
				Name: "another-namespace",
			},
		},
		Status: corev1alpha1.ServiceClusterAssignmentStatus{
			Conditions: []corev1alpha1.ServiceClusterAssignmentCondition{
				{
					Type:   corev1alpha1.ServiceClusterAssignmentReady,
					Status: corev1alpha1.ConditionTrue,
				},
			},
			ServiceClusterNamespace: &corev1alpha1.ObjectReference{
				Name: "sc-test-123",
			},
		},
	}

	ctx := context.Background()

	t.Run("updates existing obj", func(t *testing.T) {
		serviceClusterObj := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":       "test-1",
					"namespace":  "sc-test-123",
					"generation": int64(11),
				},
				"spec": map[string]interface{}{
					"test1": "spec1000",
				},
				"status": map[string]interface{}{
					"test1":              "status3000",
					"observedGeneration": int64(11),
				},
			},
		}
		serviceClusterObj.SetGroupVersionKind(serviceClusterGVK)

		log := testutil.NewLogger(t)
		managementClient := fakeclient.NewFakeClientWithScheme(testScheme, managementClusterObj, sca)
		serviceClient := fakeclient.NewFakeClientWithScheme(testScheme, serviceClusterObj)

		r := ManagementClusterObjReconciler{
			Client:               managementClient,
			Log:                  log,
			ServiceClusterClient: serviceClient,
			NamespacedClient:     managementClient,

			ManagementClusterGVK: managementClusterGVK,
			ServiceClusterGVK:    serviceClusterGVK,

			ServiceCluster:    "eu-west-1",
			ProviderNamespace: providerNamespace,
		}

		_, err := r.Reconcile(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      managementClusterObj.GetName(),
				Namespace: managementClusterObj.GetNamespace(),
			},
		})
		require.NoError(t, err)

		// Check service cluster instance
		checkServiceClusterObj := &unstructured.Unstructured{}
		checkServiceClusterObj.SetGroupVersionKind(serviceClusterGVK)

		require.NoError(t, serviceClient.Get(ctx, types.NamespacedName{
			Name:      managementClusterObj.GetName(),
			Namespace: sca.Status.ServiceClusterNamespace.Name,
		}, checkServiceClusterObj))

		assert.Equal(t, map[string]interface{}{
			"apiVersion": "couchdb.io/v1alpha1",
			"kind":       "CouchDB",
			"metadata": map[string]interface{}{
				"labels": map[string]interface{}{
					"owner.kubecarrier.io/name":      "test-1",
					"owner.kubecarrier.io/namespace": "another-namespace",
					"owner.kubecarrier.io/type":      "CouchDBInternal.eu-west-1.provider",
				},
				"name":            "test-1",
				"namespace":       "sc-test-123",
				"resourceVersion": "1",
				"generation":      int64(11),
			},
			"spec": map[string]interface{}{
				"test1": "spec2000",
			},
			"status": map[string]interface{}{
				"test1":              "status3000",
				"observedGeneration": int64(11),
			},
		}, checkServiceClusterObj.Object)

		// Check management cluster instance
		checkManagementClusterObj := &unstructured.Unstructured{}
		checkManagementClusterObj.SetGroupVersionKind(managementClusterGVK)
		require.NoError(t, managementClient.Get(ctx, types.NamespacedName{
			Name:      managementClusterObj.GetName(),
			Namespace: managementClusterObj.GetNamespace(),
		}, checkManagementClusterObj))

		assert.Equal(t, map[string]interface{}{
			"apiVersion": "eu-west-1.provider/v1alpha1",
			"kind":       "CouchDBInternal",
			"metadata": map[string]interface{}{
				"name":            managementClusterObj.GetName(),
				"namespace":       managementClusterObj.GetNamespace(),
				"resourceVersion": "2",
				"generation":      int64(2),
				"labels": map[string]interface{}{
					"l1": "v1",
				},
				"finalizers": []interface{}{
					catapultControllerFinalizer,
				},
			},
			"spec": map[string]interface{}{
				"test1": "spec2000",
			},
			"status": map[string]interface{}{
				"test1":              "status3000",
				"observedGeneration": int64(2),
			},
		}, checkManagementClusterObj.Object)
	})

	t.Run("creates service cluster obj", func(t *testing.T) {
		log := testutil.NewLogger(t)
		managementClient := fakeclient.NewFakeClientWithScheme(testScheme, managementClusterObj, sca)
		serviceClient := fakeclient.NewFakeClientWithScheme(testScheme)

		r := ManagementClusterObjReconciler{
			Client:               managementClient,
			Log:                  log,
			NamespacedClient:     managementClient,
			ServiceClusterClient: serviceClient,

			ManagementClusterGVK: managementClusterGVK,
			ServiceClusterGVK:    serviceClusterGVK,

			ServiceCluster:    "eu-west-1",
			ProviderNamespace: providerNamespace,
		}

		// Creates service cluster instance
		sca.Status.Conditions = []corev1alpha1.ServiceClusterAssignmentCondition{
			{
				Type:   corev1alpha1.ServiceClusterAssignmentReady,
				Status: corev1alpha1.ConditionTrue,
			},
		}
		sca.Status.ServiceClusterNamespace = &corev1alpha1.ObjectReference{
			Name: "sc-test-123",
		}
		require.NoError(t, managementClient.Status().Update(ctx, sca))

		_, err := r.Reconcile(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      managementClusterObj.GetName(),
				Namespace: managementClusterObj.GetNamespace(),
			},
		})
		require.NoError(t, err)

		serviceClusterObj := &unstructured.Unstructured{}
		serviceClusterObj.SetGroupVersionKind(serviceClusterGVK)
		require.NoError(t, serviceClient.Get(ctx, types.NamespacedName{
			Name:      managementClusterObj.GetName(),
			Namespace: sca.Status.ServiceClusterNamespace.Name,
		}, serviceClusterObj))

		assert.Equal(t, map[string]interface{}{
			"apiVersion": "couchdb.io/v1alpha1",
			"kind":       "CouchDB",
			"metadata": map[string]interface{}{
				"labels": map[string]interface{}{
					"owner.kubecarrier.io/name":      "test-1",
					"owner.kubecarrier.io/namespace": "another-namespace",
					"owner.kubecarrier.io/type":      "CouchDBInternal.eu-west-1.provider",
				},
				"name":            "test-1",
				"namespace":       "sc-test-123",
				"resourceVersion": "1",
			},
			"spec": map[string]interface{}{
				"test1": "spec2000",
			},
		}, serviceClusterObj.Object)
	})
}

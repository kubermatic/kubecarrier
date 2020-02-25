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

func TestAdoptionReconciler(t *testing.T) {
	serviceClusterObj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":       "test-1",
				"namespace":  "another-namespace",
				"generation": int64(4),
			},
			"spec": map[string]interface{}{
				"test1": "spec2000",
				"test2": "spec2000",
				"test3": "spec2000",
			},
		},
	}
	serviceClusterObj.SetGroupVersionKind(serviceClusterGVK)

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
			ServiceClusterNamespace: corev1alpha1.ObjectReference{
				Name: "sc-test-123",
			},
		},
	}

	t.Run("creates management cluster object", func(t *testing.T) {
		log := testutil.NewLogger(t)
		managementClient := fakeclient.NewFakeClientWithScheme(testScheme, sca)
		serviceClient := fakeclient.NewFakeClientWithScheme(
			testScheme, serviceClusterObj)

		r := AdoptionReconciler{
			Client:               managementClient,
			NamespacedClient:     managementClient,
			Log:                  log,
			ServiceClusterClient: serviceClient,

			ServiceClusterGVK:    serviceClusterGVK,
			ManagementClusterGVK: managementClusterGVK,
			ProviderNamespace:    providerNamespace,
		}

		_, err := r.Reconcile(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      serviceClusterObj.GetName(),
				Namespace: serviceClusterObj.GetNamespace(),
			},
		})
		require.NoError(t, err)

		ctx := context.Background()
		checkManagementClusterObj := &unstructured.Unstructured{}
		checkManagementClusterObj.SetGroupVersionKind(managementClusterGVK)
		err = managementClient.Get(ctx, types.NamespacedName{
			Name:      serviceClusterObj.GetName(),
			Namespace: sca.Spec.ManagementClusterNamespace.Name,
		}, checkManagementClusterObj)
		require.NoError(t, err)

		assert.Equal(t, map[string]interface{}{
			"apiVersion": "eu-west-1.provider/v1alpha1",
			"kind":       "CouchDBInternal",
			"metadata": map[string]interface{}{
				"name":            "test-1",
				"namespace":       sca.Spec.ManagementClusterNamespace.Name,
				"resourceVersion": "1",
			},
			"spec": map[string]interface{}{
				"test1": "spec2000",
				"test2": "spec2000",
				"test3": "spec2000",
			},
		}, checkManagementClusterObj.Object)
	})
}

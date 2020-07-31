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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"k8c.io/utils/pkg/testutil"

	corev1alpha1 "k8c.io/kubecarrier/pkg/apis/core/v1alpha1"
	operatorv1alpha1 "k8c.io/kubecarrier/pkg/apis/operator/v1alpha1"
)

func TestServiceClusterReconciler(t *testing.T) {
	serviceCluster := &corev1alpha1.ServiceCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "eu-west-1",
			Namespace: "my-provider",
		},
	}
	scc := &ServiceClusterReconciler{
		Log:                testutil.NewLogger(t),
		Client:             fakeclient.NewFakeClientWithScheme(testScheme, serviceCluster),
		Scheme:             testScheme,
		MonitorGracePeriod: 40 * time.Second,
	}

	ctx := context.Background()
	reconcileLoop := func() {
		for i := 0; i < 3; i++ {
			_, err := scc.Reconcile(ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      serviceCluster.Name,
					Namespace: serviceCluster.Namespace,
				},
			})
			require.NoError(t, err)
			require.NoError(t, scc.Client.Get(ctx, types.NamespacedName{
				Name:      serviceCluster.Name,
				Namespace: serviceCluster.Namespace,
			}, serviceCluster))
		}
	}

	if !t.Run("ferry controller", func(t *testing.T) {
		reconcileLoop()
		ferry := &operatorv1alpha1.Ferry{}
		require.NoError(t, scc.Client.Get(ctx, types.NamespacedName{
			Name:      serviceCluster.Name,
			Namespace: serviceCluster.Namespace,
		}, ferry))
		ferry.Status.Conditions = []operatorv1alpha1.FerryCondition{
			{
				Type:   operatorv1alpha1.FerryReady,
				Status: operatorv1alpha1.ConditionTrue,
			},
		}
		require.NoError(t, scc.Client.Status().Update(ctx, ferry))

		reconcileLoop()
		assert.Equal(t, serviceCluster.Generation, serviceCluster.Status.ObservedGeneration)
		cond, present := serviceCluster.Status.GetCondition(corev1alpha1.ServiceClusterControllerReady)
		if assert.True(t, present, "service cluster controller condition missing") {
			assert.Equal(t, corev1alpha1.ConditionTrue, cond.Status)
		}
		cond, present = serviceCluster.Status.GetCondition(corev1alpha1.ServiceClusterReady)
		if assert.True(t, present, "service cluster ready condition missing") {
			assert.Equal(t, corev1alpha1.ConditionFalse, cond.Status)
		}
	}) {
		t.FailNow()
	}

	if !t.Run("service cluster registration", func(t *testing.T) {
		serviceCluster.Status.SetCondition(corev1alpha1.ServiceClusterCondition{
			Type:    corev1alpha1.ServiceClusterReachable,
			Status:  corev1alpha1.ConditionTrue,
			Reason:  "ServiceClusterReachable",
			Message: "service cluster is posting ready status",
		})
		require.NoError(t, scc.Client.Status().Update(ctx, serviceCluster))

		reconcileLoop()
		assert.Equal(t, serviceCluster.Generation, serviceCluster.Status.ObservedGeneration)
		cond, present := serviceCluster.Status.GetCondition(corev1alpha1.ServiceClusterReachable)
		if assert.True(t, present, "service cluster rechable condition missing") {
			assert.Equal(t, corev1alpha1.ConditionTrue, cond.Status)
		}
		cond, present = serviceCluster.Status.GetCondition(corev1alpha1.ServiceClusterReady)
		if assert.True(t, present, "service cluster ready condition missing") {
			assert.Equal(t, corev1alpha1.ConditionTrue, cond.Status)
		}
	}) {
		t.FailNow()
	}

	if !t.Run("monitor grace period", func(t *testing.T) {
		minuteAgo := metav1.Time{Time: time.Now().Add(-time.Minute)}
		serviceCluster.Status.SetCondition(corev1alpha1.ServiceClusterCondition{
			Type:               corev1alpha1.ServiceClusterReachable,
			Status:             corev1alpha1.ConditionTrue,
			Reason:             "ServiceClusterReachable",
			Message:            "service cluster is posting ready status",
			LastHeartbeatTime:  minuteAgo,
			LastTransitionTime: minuteAgo,
		})
		require.NoError(t, scc.Client.Status().Update(ctx, serviceCluster))
		reconcileLoop()
		cond, ok := serviceCluster.Status.GetCondition(corev1alpha1.ServiceClusterReachable)
		require.True(t, ok)
		assert.Equal(t, corev1alpha1.ConditionUnknown, cond.Status)
	}) {
		t.FailNow()
	}
}

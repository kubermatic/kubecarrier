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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func TestServiceClusterReconciler(t *testing.T) {
	clusterInfo := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-info",
			Namespace: "kube-public",
		},
	}
	scc := &ServiceClusterReconciler{
		Log:                testutil.NewLogger(t),
		MasterClient:       fakeclient.NewFakeClientWithScheme(testScheme),
		ServiceClient:      fakeclient.NewFakeClientWithScheme(testScheme, clusterInfo),
		ServiceClusterName: "eu-west-1",
		ProviderNamespace:  "my-provider",
		StatusUpdatePeriod: time.Second,
	}
	serviceClusterNN := types.NamespacedName{
		Name:      "eu-west-1",
		Namespace: "my-provider",
	}

	if !t.Run("service cluster registration", func(t *testing.T) {
		scc.Log = testutil.NewLogger(t)
		res, err := scc.Reconcile(ctrl.Request{
			NamespacedName: serviceClusterNN,
		})
		require.NoError(t, err, "error reconciling ServiceCluster")
		assert.Equal(t, scc.StatusUpdatePeriod, res.RequeueAfter)

		ctx := context.Background()
		serviceCluster := &corev1alpha1.ServiceCluster{}
		require.NoError(t, scc.MasterClient.Get(ctx, serviceClusterNN, serviceCluster))
		assert.Equal(t, serviceCluster.Generation, serviceCluster.Status.ObservedGeneration)

		cond, present := serviceCluster.Status.GetCondition(corev1alpha1.ServiceClusterReady)
		if assert.True(t, present, "cluster ready condition missing") {
			assert.Equal(t, corev1alpha1.ConditionTrue, cond.Status)
		}
	}) {
		t.FailNow()
	}

	if !t.Run("service cluster unreachable", func(t *testing.T) {
		scc.Log = testutil.NewLogger(t)
		ctx := context.Background()
		require.NoError(t, scc.ServiceClient.Delete(ctx, clusterInfo))
		_, err := scc.Reconcile(ctrl.Request{
			NamespacedName: serviceClusterNN,
		})
		require.Error(t, err)

		serviceCluster := &corev1alpha1.ServiceCluster{}
		require.NoError(t, scc.MasterClient.Get(ctx, serviceClusterNN, serviceCluster))
		assert.Equal(t, serviceCluster.Generation, serviceCluster.Status.ObservedGeneration)

		cond, present := serviceCluster.Status.GetCondition(corev1alpha1.ServiceClusterReady)
		if assert.True(t, present, "cluster ready condition missing") {
			assert.Equal(t, corev1alpha1.ConditionFalse, cond.Status)
			assert.Equal(t, "NotFound", cond.Reason)
			assert.Equal(t, `configmaps "cluster-info" not found`, cond.Message)
		}
	}) {
		t.FailNow()
	}

}

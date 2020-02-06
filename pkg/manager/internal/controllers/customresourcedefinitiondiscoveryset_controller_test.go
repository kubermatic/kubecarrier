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

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func TestCustomResourceDefinitionDiscoverySetReconciler(t *testing.T) {
	serviceCluster1 := &corev1alpha1.ServiceCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "svc-1",
			Namespace: "hans",
		},
		Spec: corev1alpha1.ServiceClusterSpec{},
	}
	serviceCluster2 := &corev1alpha1.ServiceCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "svc-2",
			Namespace: "hans",
		},
		Spec: corev1alpha1.ServiceClusterSpec{},
	}

	crddsnn := types.NamespacedName{
		Name:      "couchdb",
		Namespace: "hans",
	}
	crdds := &corev1alpha1.CustomResourceDefinitionDiscoverySet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "couchdb",
			Namespace: "hans",
		},
		Spec: corev1alpha1.CustomResourceDefinitionDiscoverySetSpec{
			CRD: corev1alpha1.ObjectReference{
				Name: "couchdbs.couchdb.io",
			},
			KindOverride: "CouchDBInternal",
		},
	}
	crds := &corev1alpha1.CustomResourceDefinitionDiscovery{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "couchdb.us-east-1",
			Namespace: "hans",
			Labels: map[string]string{
				crddsLabel: crdds.Name,
			},
		},
	}
	r := &CustomResourceDefinitionDiscoverySetReconciler{
		Log:    testutil.NewLogger(t),
		Client: fakeclient.NewFakeClientWithScheme(testScheme, serviceCluster1, serviceCluster2, crdds, crds),
		Scheme: testScheme,
	}
	ctx := context.Background()
	reconcileLoop := func() {
		for i := 0; i < 3; i++ {
			_, err := r.Reconcile(ctrl.Request{
				NamespacedName: crddsnn,
			})
			require.NoError(t, err)
			require.NoError(t, r.Client.Get(ctx, crddsnn, crdds))
		}
	}

	reconcileLoop() // should create two CustomResourceDefinitionDiscovery objects

	crddServicCluster1 := &corev1alpha1.CustomResourceDefinitionDiscovery{}
	require.NoError(t, r.Get(ctx, types.NamespacedName{
		Name:      "couchdb.svc-1",
		Namespace: crdds.Namespace,
	}, crddServicCluster1))

	crddServicCluster2 := &corev1alpha1.CustomResourceDefinitionDiscovery{}
	require.NoError(t, r.Get(ctx, types.NamespacedName{
		Name:      "couchdb.svc-2",
		Namespace: crdds.Namespace,
	}, crddServicCluster2))

	ready, ok := crdds.Status.GetCondition(corev1alpha1.CustomResourceDefinitionDiscoverySetReady)
	if assert.True(t, ok) {
		assert.Equal(t, corev1alpha1.ConditionFalse, ready.Status)
		assert.Equal(t, "ComponentsUnready", ready.Reason)
	}

	crddServicCluster1.Status.Conditions = []corev1alpha1.CustomResourceDefinitionDiscoveryCondition{
		{
			Type:   corev1alpha1.CustomResourceDefinitionDiscoveryReady,
			Status: corev1alpha1.ConditionTrue,
		},
	}
	require.NoError(t, r.Status().Update(ctx, crddServicCluster1))

	crddServicCluster2.Status.Conditions = []corev1alpha1.CustomResourceDefinitionDiscoveryCondition{
		{
			Type:   corev1alpha1.CustomResourceDefinitionDiscoveryReady,
			Status: corev1alpha1.ConditionTrue,
		},
	}
	require.NoError(t, r.Status().Update(ctx, crddServicCluster2))

	reconcileLoop() // should update status

	ready, ok = crdds.Status.GetCondition(corev1alpha1.CustomResourceDefinitionDiscoverySetReady)
	if assert.True(t, ok) {
		assert.Equal(t, corev1alpha1.ConditionTrue, ready.Status)
		assert.Equal(t, "ComponentsReady", ready.Reason)
	}
}

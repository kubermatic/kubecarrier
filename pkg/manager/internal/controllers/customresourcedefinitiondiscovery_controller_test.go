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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func TestCustomResourceDefinitionDiscoveryReconciler(t *testing.T) {
	const (
		serviceClusterName = "eu-west-1"
	)
	provider := &v1alpha1.Provider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "extreme-cloud",
			Namespace: "kubecarrier-system",
		},
	}

	crdDiscovery := &corev1alpha1.CustomResourceDefinitionDiscovery{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "redis.cloud",
			Namespace: "tenant-hans",
		},
		Spec: corev1alpha1.CustomResourceDefinitionDiscoverySpec{
			CRD:            corev1alpha1.ObjectReference{Name: "redis.cloud"},
			ServiceCluster: corev1alpha1.ObjectReference{Name: serviceClusterName},
		},
	}
	crdDiscoveryNN := types.NamespacedName{
		Namespace: crdDiscovery.Namespace,
		Name:      crdDiscovery.Name,
	}

	crd := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "redis.cloud",
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "cloud",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Plural:   "redis",
				Singular: "redis",
				Kind:     "Redis",
				ListKind: "RedisList",
			},
			Scope: "Namespaced",
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{Name: "corev1alpha1"},
			},
		},
	}
	r := &CustomResourceDefinitionDiscoveryReconciler{
		Log:                        testutil.NewLogger(t),
		Client:                     fakeclient.NewFakeClientWithScheme(testScheme, crdDiscovery, provider),
		Scheme:                     testScheme,
		KubeCarrierSystemNamespace: provider.Namespace,
	}
	ctx := context.Background()
	reconcileLoop := func() {
		for i := 0; i < 3; i++ {
			_, err := r.Reconcile(ctrl.Request{
				NamespacedName: crdDiscoveryNN,
			})
			require.NoError(t, err)
			require.NoError(t, r.Client.Get(ctx, crdDiscoveryNN, crdDiscovery))
		}
	}

	reconcileLoop() // should not panic on undiscovered instances

	crdDiscovery.Status.CRD = crd
	crdDiscovery.Status.SetCondition(corev1alpha1.CustomResourceDefinitionDiscoveryCondition{
		Type:   corev1alpha1.CustomResourceDefinitionDiscoveryDiscovered,
		Status: corev1alpha1.ConditionTrue,
	})
	require.NoError(t, r.Client.Status().Update(ctx, crdDiscovery))

	reconcileLoop() // creates the CRD in the master cluster

	establishedCondition, ok := crdDiscovery.Status.GetCondition(corev1alpha1.CustomResourceDefinitionDiscoveryEstablished)
	if assert.True(t, ok) {
		assert.Equal(t, corev1alpha1.ConditionFalse, establishedCondition.Status)
		assert.Equal(t, "Establishing", establishedCondition.Reason)
	}

	internalCRD := &apiextensionsv1.CustomResourceDefinition{}
	require.NoError(t, r.Client.Get(ctx, types.NamespacedName{
		Name: strings.Join([]string{"redis", serviceClusterName, provider.Name}, "."),
	}, internalCRD))

	internalCRD.Status.Conditions = []apiextensionsv1.CustomResourceDefinitionCondition{
		{
			Type:   apiextensionsv1.Established,
			Status: apiextensionsv1.ConditionTrue,
		},
	}
	require.NoError(t, r.Client.Status().Update(ctx, internalCRD))

	reconcileLoop() // updates the status to established and launches Catapult

	establishedCondition, ok = crdDiscovery.Status.GetCondition(corev1alpha1.CustomResourceDefinitionDiscoveryEstablished)
	if assert.True(t, ok) {
		assert.Equal(t, corev1alpha1.ConditionTrue, establishedCondition.Status)
		assert.Equal(t, "Established", establishedCondition.Reason)
	}

	catapult := &operatorv1alpha1.Catapult{}
	require.NoError(t, r.Client.Get(ctx, types.NamespacedName{
		Name:      crdDiscovery.Name,
		Namespace: crdDiscovery.Namespace,
	}, catapult))
	catapult.Status.Conditions = []operatorv1alpha1.CatapultCondition{
		{
			Type:   operatorv1alpha1.CatapultReady,
			Status: operatorv1alpha1.ConditionTrue,
		},
	}
	require.NoError(t, r.Client.Status().Update(ctx, catapult))

	reconcileLoop() // updates status to ready

	readyCondition, ok := crdDiscovery.Status.GetCondition(corev1alpha1.CustomResourceDefinitionDiscoveryReady)
	if assert.True(t, ok) {
		assert.Equal(t, corev1alpha1.ConditionTrue, readyCondition.Status)
		assert.Equal(t, "ComponentsReady", readyCondition.Reason)
	}
}

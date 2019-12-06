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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func TestCRDReferenceReconciler(t *testing.T) {
	const (
		serviceClusterName = "eu-west-1"
	)

	crdRef := &corev1alpha1.CRDReference{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster.redis",
			Namespace: "tenant-hans",
		},
		Spec: corev1alpha1.CRDReferenceSpec{
			CRD:            corev1alpha1.ObjectReference{Name: "cluster.redis"},
			ServiceCluster: corev1alpha1.ObjectReference{Name: serviceClusterName},
		},
	}
	crdRefNN := types.NamespacedName{
		Namespace: crdRef.Namespace,
		Name:      crdRef.Name,
	}

	crd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster.redis",
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   "redis",
			Version: "corev1alpha1",
			Scope:   "Namespaced",
		},
	}

	r := &CRDReferenceReconciler{
		MasterClient:       fakeclient.NewFakeClientWithScheme(testScheme, crdRef),
		ServiceClient:      fakeclient.NewFakeClientWithScheme(testScheme, crd),
		MasterScheme:       testScheme,
		ServiceClusterName: serviceClusterName,
	}

	if !t.Run("creation", func(t *testing.T) {
		r.Log = testutil.NewLogger(t)
		for i := 0; i < 2; i++ {
			res, err := r.Reconcile(ctrl.Request{
				NamespacedName: crdRefNN,
			})
			_ = res
			require.NoError(t, err, "cannot reconcile the CRD Reference")
		}
		crdRef = &corev1alpha1.CRDReference{}
		require.NoError(t, r.MasterClient.Get(context.Background(), crdRefNN, crdRef))
		require.NoError(t, testutil.ConditionStatusEqual(crdRef, corev1alpha1.CRDReferenceReady, corev1alpha1.ConditionTrue))

		crdSpec := &apiextensionsv1beta1.CustomResourceDefinitionSpec{}
		b, err := crdRef.Status.CRDSpec.MarshalJSON()
		require.NoError(t, err, "cannot marshall CRD spec")
		require.NoError(t, json.Unmarshal(b, crdSpec), "cannot convert .Status.CRDSpec to proper CustromResourceDefinition Object")

		assert.Equal(t, crd.Spec.Group, crdSpec.Group)
		assert.Equal(t, crd.Spec.Version, crdSpec.Version)
		assert.Equal(t, crd.Spec.Scope, crdSpec.Scope)
	}) {
		t.FailNow()
	}
	if !t.Run("CRD deletion", func(t *testing.T) {
		r.Log = testutil.NewLogger(t)
		require.NoError(t, r.ServiceClient.Delete(context.Background(), crd))
		for i := 0; i < 5; i++ {
			res, err := r.Reconcile(ctrl.Request{
				NamespacedName: crdRefNN,
			})
			require.NoError(t, err, "cannot reconcile the CRD Reference")
			assert.True(t, res.Requeue)
		}
		crdRef = &corev1alpha1.CRDReference{}
		require.NoError(t, r.MasterClient.Get(context.Background(), crdRefNN, crdRef))
		assert.NoError(t, testutil.ConditionStatusEqual(crdRef, corev1alpha1.CRDReferenceReady, corev1alpha1.ConditionFalse))
		assert.Equal(t, (*runtime.RawExtension)(nil), crdRef.Status.CRDSpec)
	}) {
		t.FailNow()
	}
	if !t.Run("CRDRef deletion", func(t *testing.T) {
		r.Log = testutil.NewLogger(t)
		require.NoError(t, r.ServiceClient.Create(context.Background(), crd))
		for i := 0; i < 2; i++ {
			_, err := r.Reconcile(ctrl.Request{
				NamespacedName: crdRefNN,
			})
			require.NoError(t, err, "cannot reconcile the CRD Reference")
		}
		crdRef = &corev1alpha1.CRDReference{}
		require.NoError(t, r.MasterClient.Get(context.Background(), crdRefNN, crdRef))
		require.NoError(t, testutil.ConditionStatusEqual(crdRef, corev1alpha1.CRDReferenceReady, corev1alpha1.ConditionTrue))

		require.NoError(t, r.MasterClient.Delete(context.Background(), crdRef))
		for i := 0; i < 2; i++ {
			_, err := r.Reconcile(ctrl.Request{
				NamespacedName: crdRefNN,
			})
			require.NoError(t, err, "cannot reconcile the CRD Reference")
		}
		require.True(t, errors.IsNotFound(r.MasterClient.Get(context.Background(), crdRefNN, crdRef)))
	}) {
		t.FailNow()
	}
}

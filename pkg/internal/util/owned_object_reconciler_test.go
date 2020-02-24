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

package util

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/kubermatic/kubecarrier/pkg/internal/owner"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func TestListObjects(t *testing.T) {
	cmA := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cma",
			Namespace: "a",
		},
		Data: map[string]string{
			"cm": "a",
		},
	}
	cmB := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cmb",
			Namespace: "b",
		},
		Data: map[string]string{
			"cm": "b",
		},
	}
	secA := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secA",
			Namespace: "a",
		},
		Data: map[string][]byte{
			"sec": []byte("a"),
		},
	}
	secB := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secB",
			Namespace: "b",
		},
		Data: map[string][]byte{
			"sec": []byte("b"),
		},
	}

	for name, testCase := range map[string]struct {
		wantsObj []runtime.Object
		types    []runtime.Object
		listOpts []client.ListOption
	}{
		"configMaps": {
			types:    []runtime.Object{&corev1.ConfigMap{}},
			wantsObj: []runtime.Object{cmA, cmB},
		},
		"multi": {
			types:    []runtime.Object{&corev1.ConfigMap{}, &corev1.Secret{}},
			wantsObj: []runtime.Object{cmA, cmB, secA, secB},
		},
		"ns-only": {
			types:    []runtime.Object{&corev1.ConfigMap{}, &corev1.Secret{}},
			wantsObj: []runtime.Object{cmA, secA},
			listOpts: []client.ListOption{client.InNamespace("a")},
		},
	} {
		t.Run(name, func(t *testing.T) {
			cl := fakeclient.NewFakeClientWithScheme(testScheme, cmA, cmB, secA, secB)
			ctx := context.Background()
			objs, err := ListObjects(ctx, cl, testScheme, testCase.types, testCase.listOpts...)
			require.NoError(t, err)

			wants := make(map[ObjectReference]struct{})
			for _, obj := range testCase.wantsObj {
				wants[ToObjectReference(obj.(Object), testScheme)] = struct{}{}
			}

			got := make(map[ObjectReference]struct{})
			for _, obj := range objs {
				got[ToObjectReference(obj.(Object), testScheme)] = struct{}{}
			}
			assert.Equal(t, wants, got)
		})
	}
}

func TestOwnedObjectReconciler_Reconcile(t *testing.T) {
	cmA := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cma",
			Namespace: "default",
		},
		Data: map[string]string{
			"cm": "a",
		},
	}
	cmB := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cmb",
			Namespace: "default",
		},
		Data: map[string]string{
			"cm": "b",
		},
	}

	for name, testCase := range map[string]struct {
		existingState []runtime.Object
		wantedState   []Object
		muateFn       func(oldObj, wantedObj runtime.Object) error
		change        bool
	}{
		"clearing": {
			existingState: []runtime.Object{cmA.DeepCopy(), cmB.DeepCopy()},
			change:        true,
		},
		"creating": {
			wantedState: []Object{cmA.DeepCopy(), cmB.DeepCopy()},
			change:      true,
		},
		"delete & create": {
			existingState: []runtime.Object{cmA.DeepCopy()},
			wantedState:   []Object{cmB.DeepCopy()},
			change:        true,
		},
		"nothing-changes": {
			existingState: []runtime.Object{cmA.DeepCopy()},
			wantedState:   []Object{cmA.DeepCopy()},
			change:        false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			ownerObj := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
				Name:      "ownerObj",
				Namespace: "default",
			}}
			cl := fakeclient.NewFakeClientWithScheme(testScheme, ownerObj)
			ctx := context.Background()
			for _, obj := range testCase.existingState {
				owner.SetOwnerReference(ownerObj, obj.(Object), testScheme)
				require.NoError(t, cl.Create(ctx, obj))
			}

			changed, err := (&OwnedObjectReconciler{
				Scheme:      testScheme,
				Log:         testutil.NewLogger(t),
				Owner:       ownerObj,
				TypeFilter:  []runtime.Object{&corev1.ConfigMap{}},
				WantedState: testCase.wantedState,
				MutateFn:    testCase.muateFn,
			}).Reconcile(ctx, cl)
			assert.NoError(t, err)
			assert.Equal(t, testCase.change, changed)
		})
	}
}

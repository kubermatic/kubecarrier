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

package multiowner

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/kubermatic/kubecarrier/pkg/internal/util"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func TestOwnedObjectReconcile(t *testing.T) {
	cmA := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cma",
			Namespace: "default",
		},
		Data: map[string]string{
			"cm": "a",
		},
	}
	cmAPrime := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cma",
			Namespace: "default",
		},
		Data: map[string]string{
			"cm": "a prime",
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

	ownerObj1 := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
		Name:      "ownerObj1",
		Namespace: "default",
	}}
	ownerObj2 := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
		Name:      "ownerObj2",
		Namespace: "default",
	}}

	for name, testCase := range map[string]struct {
		owners        []runtime.Object
		existingState []runtime.Object
		wantedState   []runtime.Object
		finalState    []*corev1.ConfigMap
		updateFn      func(obj, wantedObj runtime.Object) error
	}{
		"single owner cleaning": {
			owners:        []runtime.Object{ownerObj1},
			existingState: []runtime.Object{cmA.DeepCopy(), cmB.DeepCopy()},
		},
		"double owner cleaning": {
			owners:        []runtime.Object{ownerObj1, ownerObj2},
			existingState: []runtime.Object{cmA.DeepCopy(), cmB.DeepCopy()},
			finalState:    []*corev1.ConfigMap{cmA.DeepCopy(), cmB.DeepCopy()},
		},
		"creating": {
			owners:      []runtime.Object{ownerObj1, ownerObj2},
			wantedState: []runtime.Object{cmA.DeepCopy(), cmB.DeepCopy()},
			finalState:  []*corev1.ConfigMap{cmA.DeepCopy(), cmB.DeepCopy()},
		},
		"single owner delete & create": {
			owners:        []runtime.Object{ownerObj1},
			existingState: []runtime.Object{cmA.DeepCopy()},
			wantedState:   []runtime.Object{cmB.DeepCopy()},
			finalState:    []*corev1.ConfigMap{cmB.DeepCopy()},
		},
		"double owner delete & create": {
			owners:        []runtime.Object{ownerObj1, ownerObj2},
			existingState: []runtime.Object{cmA.DeepCopy()},
			wantedState:   []runtime.Object{cmB.DeepCopy()},
			finalState:    []*corev1.ConfigMap{cmA.DeepCopy(), cmB.DeepCopy()},
		},
		"nothing-changes": {
			owners:        []runtime.Object{ownerObj1, ownerObj2},
			existingState: []runtime.Object{cmA.DeepCopy()},
			wantedState:   []runtime.Object{cmA.DeepCopy()},
			finalState:    []*corev1.ConfigMap{cmA.DeepCopy()},
		},
		"mutation": {
			owners:        []runtime.Object{ownerObj1, ownerObj2},
			existingState: []runtime.Object{cmA.DeepCopy()},
			wantedState:   []runtime.Object{cmA.DeepCopy()},
			finalState:    []*corev1.ConfigMap{cmAPrime.DeepCopy()},
			updateFn: func(obj, wantedObj runtime.Object) error {
				objCM := obj.(*corev1.ConfigMap)
				objCM.Data["cm"] = "a prime"
				return nil
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			cl := fakeclient.NewFakeClientWithScheme(testScheme, ownerObj1, ownerObj2)
			ctx := context.Background()
			for _, obj := range testCase.existingState {
				for _, owner := range testCase.owners {
					changed, err := insertOwnerReference(owner, obj, testScheme)
					require.NoError(t, err)
					assert.True(t, changed)
				}
				require.NoError(t, cl.Create(ctx, obj))
			}
			assert.NoError(t, ReconcileOwnedObjects(ctx, testutil.NewLogger(t), cl, testScheme, ownerObj1, testCase.wantedState, &corev1.ConfigMap{}, testCase.updateFn))

			cmLst := &corev1.ConfigMapList{}
			require.NoError(t, cl.List(ctx, cmLst))
			wants := make(map[util.ObjectReference]struct{})
			for _, obj := range testCase.finalState {
				wants[util.ToObjectReference(obj, testScheme)] = struct{}{}
				cm := &corev1.ConfigMap{}
				if assert.NoError(t, cl.Get(ctx, types.NamespacedName{
					Namespace: obj.Namespace,
					Name:      obj.Name,
				}, cm)) {
					assert.Equal(t, obj.Data, cm.Data)
				}
			}
			got := make(map[util.ObjectReference]struct{})
			for _, obj := range cmLst.Items {
				got[util.ToObjectReference(&obj, testScheme)] = struct{}{}
			}
			assert.Equal(t, wants, got, "some object exist and shouldn't or vice versa")
		})
	}
}

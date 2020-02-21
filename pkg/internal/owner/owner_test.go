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

package owner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var testScheme = runtime.NewScheme()

func TestSetOwnerReference(t *testing.T) {
	owner := &unstructured.Unstructured{}
	owner.SetName("hans")
	owner.SetNamespace("hans-playground")
	owner.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "test.kubecarrier.io",
		Kind:    "Test",
		Version: "v1alpha1",
	})

	obj := &unstructured.Unstructured{}

	changed := SetOwnerReference(owner, obj, testScheme)
	assert.True(t, changed)
	assert.Equal(t, map[string]string{
		OwnerNameLabel:      "hans",
		OwnerNamespaceLabel: "hans-playground",
		OwnerTypeLabel:      "Test.test.kubecarrier.io",
	}, obj.GetLabels())
	// this is the kubernetes regex that validates label values
	assert.Regexp(
		t, `(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?`,
		obj.GetLabels()[OwnerTypeLabel])
}

func Test_requestHandlerForOwner(t *testing.T) {
	owner := &unstructured.Unstructured{}
	owner.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "test.kubecarrier.io",
		Kind:    "Test",
		Version: "v1alpha1",
	})

	matchingObj := &unstructured.Unstructured{}
	matchingObj.SetName("test")
	matchingObj.SetName("test-ns")
	matchingObj.SetLabels(map[string]string{
		OwnerNameLabel:      "test12",
		OwnerNamespaceLabel: "hans3000",
		OwnerTypeLabel:      "Test.test.kubecarrier.io",
	})

	nonMatchingTypeObj := &unstructured.Unstructured{}
	nonMatchingTypeObj.SetName("test")
	nonMatchingTypeObj.SetName("test-ns")
	nonMatchingTypeObj.SetLabels(map[string]string{
		OwnerNameLabel:      "test12",
		OwnerNamespaceLabel: "hans3000",
		OwnerTypeLabel:      "Test.example.io",
	})

	nonMatchingNoLabelObj := &unstructured.Unstructured{}
	nonMatchingNoLabelObj.SetName("test")
	nonMatchingNoLabelObj.SetName("test-ns")

	tests := []struct {
		name     string
		obj      *unstructured.Unstructured
		requests []reconcile.Request
	}{
		{
			name: "should match",
			obj:  matchingObj,
			requests: []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{
						Name:      "test12",
						Namespace: "hans3000",
					},
				},
			},
		},
		{
			name: "filtered by type",
			obj:  nonMatchingTypeObj,
		},
		{
			name: "no label",
			obj:  nonMatchingNoLabelObj,
		},
	}

	handlerFn := requestHandlerForOwner(owner, testScheme)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requests := handlerFn(handler.MapObject{
				Meta:   test.obj,
				Object: test.obj,
			})
			assert.Equal(t, test.requests, requests)
		})
	}
}

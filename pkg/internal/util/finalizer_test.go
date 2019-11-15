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
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	finalizerA Finalizer = "kubecarrier.io/finalizer-a"
	finalizerB Finalizer = "kubecarrier.io/finalizer-b"
)

func TestFinalizer(t *testing.T) {
	u := &unstructured.Unstructured{}

	t.Run("add finalizer", func(t *testing.T) {
		assert.True(t, finalizerA.Insert(u))
		// Add the same finalizer twice to make sure the finalizer is only added once.
		assert.False(t, finalizerA.Insert(u))
		assert.Contains(t, u.GetFinalizers(), string(finalizerA))
		assert.Len(t, u.GetFinalizers(), 1)
	})

	t.Run("add another finalizer", func(t *testing.T) {
		assert.True(t, finalizerB.Insert(u))
		assert.Contains(t, u.GetFinalizers(), string(finalizerA), "should contain finalizerA")
		assert.Contains(t, u.GetFinalizers(), string(finalizerB), "should contain finalizerB")
		assert.Len(t, u.GetFinalizers(), 2)
	})

	t.Run("remove finalizers", func(t *testing.T) {
		assert.True(t, finalizerB.Delete(u))
		// Remove the same finalizer twice to make sure nothing errors.
		assert.False(t, finalizerB.Delete(u))
		assert.Len(t, u.GetFinalizers(), 1)
		assert.Contains(t, u.GetFinalizers(), string(finalizerA), "should contain finalizerA")
	})
}

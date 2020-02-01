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
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// UpdateObservedGeneration sets dest.Status.ObservedGeneration=dest.Generation,
// if src.Generation == src.Status.ObservedGeneration
func UpdateObservedGeneration(src, dest *unstructured.Unstructured) error {
	// check if source status is "up to date", by checking ObservedGeneration
	srcObservedGeneration, found, err := unstructured.NestedInt64(src.Object, "status", "observedGeneration")
	if err != nil {
		return fmt.Errorf("reading observedGeneration from %s: %w", src.GetKind(), err)
	}
	if !found {
		// observedGeneration field not present -> nothing to do
		return nil
	}

	if srcObservedGeneration != src.GetGeneration() {
		// observedGeneration is set, but it does not match
		// this means the status is not up to date
		// and we don't want to update observedGeneration on dest
		return nil
	}

	return unstructured.SetNestedField(
		dest.Object, dest.GetGeneration(), "status", "observedGeneration")
}

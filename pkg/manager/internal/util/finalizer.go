/*
Copyright 2019 The Kubecarrier Authors.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

// Finalizer implements helper methods to add/remove a finalizer to/from a metav1.Object.
type Finalizer string

// Insert inserts the Finalizer to the metav1.Object if the Finalizer is not present.
func (c Finalizer) Insert(object metav1.Object) (changed bool) {
	finalizers := sets.NewString(object.GetFinalizers()...)
	if finalizers.Has(string(c)) {
		return false
	}
	finalizers.Insert(string(c))
	object.SetFinalizers(finalizers.List())
	return true
}

// Delete deletes the finalizer from the metav1.Object if the Finalizer is present.
func (c Finalizer) Delete(object metav1.Object) (changed bool) {
	finalizers := sets.NewString(object.GetFinalizers()...)
	if !finalizers.Has(string(c)) {
		return false
	}
	finalizers.Delete(string(c))
	object.SetFinalizers(finalizers.List())
	return true
}

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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// objectReference references an owning/controlling object across Namespaces.
type ObjectReference struct {
	// Name of the referent.
	Name string `json:"name"`
	// Namespace that the referent object lives in.
	Namespace string `json:"namespace"`
	// The API Group of the referent.
	Group string `json:"group"`
	// The Kind of the referent.
	Kind string `json:"kind"`
}

type Object interface {
	runtime.Object
	metav1.Object
}

// toOwnerReference converts the given object into an ownerReference.
func ToObjectReference(owner Object, scheme *runtime.Scheme) ObjectReference {
	gvk, err := apiutil.GVKForObject(owner, scheme)
	if err != nil {
		// if this panic occurs many, many other stuff has gone wrong as well
		// by owner type's safety ensures this is somewhat well formed k8s object
		// When using client-go API, it needs to be able to deduce GVK in the same manner
		// thus get/create/update/patch/delete shall error out long before this is called
		// This massively simplifies the function interface and allows OwnedBy to be a
		// one-liner instead of 3 line check which never errors
		// this is error is completely under our control, users of kubecarrier cannot
		// change cluster state to cause it.
		panic(fmt.Sprintf("cannot deduce GVK for owner (type %T)", owner))
	}

	return ObjectReference{
		Name:      owner.GetName(),
		Namespace: owner.GetNamespace(),
		Kind:      gvk.Kind,
		Group:     gvk.Group,
	}
}

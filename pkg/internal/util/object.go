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
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// ObjectReference references an owning/controlling object across Namespaces.
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

// ToOwnerReference converts the given object into an ownerReference.
func ToObjectReference(object runtime.Object, scheme *runtime.Scheme) ObjectReference {
	gvk, err := apiutil.GVKForObject(object, scheme)
	if err != nil {
		// if this panic occurs many, many other stuff has gone wrong as well
		// by object type's safety ensures this is somewhat well formed k8s object
		// When using client-go API, it needs to be able to deduce GVK in the same manner
		// thus get/create/update/patch/delete shall error out long before this is called
		// This massively simplifies the function interface and allows OwnedBy to be a
		// one-liner instead of 3 line check which never errors
		// this is error is completely under our control, users of kubecarrier cannot
		// change cluster state to cause it.
		panic(fmt.Sprintf("cannot deduce GVK for object (type %T)", object))
	}
	accessor, err := meta.Accessor(object)
	if err != nil {
		panic(fmt.Errorf("cannot get meta accessor for %T %w", object, err))
	}

	return ObjectReference{
		Name:      accessor.GetName(),
		Namespace: accessor.GetNamespace(),
		Kind:      gvk.Kind,
		Group:     gvk.Group,
	}
}

// ListObjects lists all object of given types adhering to additional ListOptions
func ListObjects(ctx context.Context, cl client.Client, scheme *runtime.Scheme, listTypes []runtime.Object, options ...client.ListOption) ([]runtime.Object, error) {
	objs := make([]runtime.Object, 0)
	for _, objType := range listTypes {
		gvk, err := apiutil.GVKForObject(objType, scheme)
		if err != nil {
			return nil, fmt.Errorf("cannot get GVK for %T: %w", objType, err)
		}
		if _, isList := objType.(metav1.ListInterface); isList {
			return nil, fmt.Errorf("should not pass ListInterface as listTypes, got %v", gvk)
		}

		ListGVK := gvk
		ListGVK.Kind = gvk.Kind + "List"
		ListObjType, err := scheme.New(ListGVK)
		if err != nil {
			return nil, fmt.Errorf("cannot make a list out of a types: %v", gvk)
		}
		if !meta.IsListType(ListObjType) {
			return nil, fmt.Errorf("cannot make a list out of a types: %v", gvk)
		}

		if err := cl.List(ctx, ListObjType, options...); err != nil {
			return nil, fmt.Errorf("listing %s.%s: %w", strings.ToLower(gvk.Kind), gvk.Group, err)
		}

		lstObjs, err := meta.ExtractList(ListObjType)
		if err != nil {
			return nil, fmt.Errorf("extracting list: %w", err)
		}
		objs = append(objs, lstObjs...)
	}
	return objs, nil
}

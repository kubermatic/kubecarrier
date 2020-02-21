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
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	// OwnerAnnotation is the annotation key that references the owner of this object.
	OwnerAnnotation = "kubecarrier.io/owner"
)

// objectReference references an owning/controlling object across Namespaces.
type objectReference struct {
	// Name of the referent.
	Name string `json:"name"`
	// Namespace that the referent object lives in.
	Namespace string `json:"namespace"`
	// The API Group of the referent.
	Group string `json:"group"`
	// The Kind of the referent.
	Kind string `json:"kind"`
}

type GeneralizedListOption interface {
	client.ListOption
	client.DeleteAllOfOption
}

// InsertOwnerReference adds an OwnerReference to the given object.
func InsertOwnerReference(owner, object Object, scheme *runtime.Scheme) (changed bool, err error) {
	ownerReference := toObjectReference(owner, scheme)

	refs, err := getRefs(object)
	if err != nil {
		return false, err
	}

	for _, ref := range refs {
		if ref == ownerReference {
			// already inserted, early stop
			return false, nil
		}
	}
	refs = append(refs, ownerReference)
	err = setRefs(object, refs)
	if err != nil {
		return false, err
	}
	return true, nil
}

// DeleteOwnerReference removes an owner from the given object.
func DeleteOwnerReference(owner, object Object, scheme *runtime.Scheme) (changed bool, err error) {
	reference := toObjectReference(owner, scheme)

	refs, err := getRefs(object)
	if err != nil {
		return false, err
	}

	var newRefs []objectReference
	for _, ref := range refs {
		if ref != reference {
			newRefs = append(newRefs, ref)
		} else {
			changed = true
		}
	}

	// early stopping
	if !changed {
		return false, nil
	}

	err = setRefs(object, newRefs)
	if err != nil {
		return false, err
	}
	return changed, nil
}

// IsUnowned checks if any owners claim ownership of this object.
func IsUnowned(object metav1.Object) (unowned bool, err error) {
	refs, err := getRefs(object)
	if err != nil {
		return
	}
	return len(refs) == 0, nil
}

// EnqueueRequestForOwner enqueues requests for all owners of an object.
//
// It implements the same behavior as handler.EnqueueRequestForOwner, but for our custom objectReference.
func EnqueueRequestForOwner(ownerType Object, scheme *runtime.Scheme) handler.EventHandler {
	ownerTypeRef := toObjectReference(ownerType, scheme)
	ownerKind, ownerGroup := ownerTypeRef.Kind, ownerTypeRef.Group

	return &handler.EnqueueRequestsFromMapFunc{
		ToRequests: handler.ToRequestsFunc(func(obj handler.MapObject) []reconcile.Request {
			obj.Object.GetObjectKind().GroupVersionKind()
			refs, err := getRefs(obj.Meta)
			if err != nil {
				utilruntime.HandleError(
					fmt.Errorf("parsing owner references name=%s namespace=%s gvk=%s: %w",
						obj.Meta.GetName(),
						obj.Meta.GetNamespace(),
						obj.Object.GetObjectKind().GroupVersionKind().String(),
						err,
					))
				return nil
			}
			var req []reconcile.Request
			for _, r := range refs {
				if ownerKind == r.Kind && ownerGroup == r.Group {
					req = append(req, reconcile.Request{
						NamespacedName: types.NamespacedName{
							Namespace: r.Namespace,
							Name:      r.Name,
						},
					})
				}
			}
			return req
		}),
	}
}

// AddOwnerReverseFieldIndex adds a reverse index for OwnerReferences.
//
// The created index allows listing all owned objects of a given type by the owner object.
// Keep in mind this function should be called for each owned object type separately.
//
// See also: OwnedBy
func AddOwnerReverseFieldIndex(indexer client.FieldIndexer, log logr.Logger, object runtime.Object) error {
	_, ok := object.(metav1.Object)
	if !ok {
		return fmt.Errorf("%T is not a metav1.Object", object)
	}
	return indexer.IndexField(
		object,
		OwnerAnnotation,
		func(object runtime.Object) (values []string) {
			// this should not panic due to previous casting check
			obj := object.(metav1.Object)

			refs, err := getRefs(obj)
			if err != nil {
				log.Error(err, "cannot list owner references", "name", obj.GetName(), "namespace", obj.GetNamespace())
				return
			}

			for _, r := range refs {
				values = append(values, r.fieldIndexValue())
			}
			return
		})
}

// OwnedBy returns owner filter for listing objects.
//
// See also: AddOwnerReverseFieldIndex
func OwnedBy(owner Object, sc *runtime.Scheme) GeneralizedListOption {
	return client.MatchingFields{
		OwnerAnnotation: toObjectReference(owner, sc).fieldIndexValue(),
	}
}

// fieldIndexValue converts the objectReference into a simple value for a client.FieldIndexer.
// to be used as key for indexing structure.
func (n objectReference) fieldIndexValue() string {
	b, err := json.Marshal(n)
	if err != nil {
		// this should never ever happen
		panic(err)
	}
	return string(b)
}

// Object generic k8s object with metav1 and runtime Object interfaces implemented
type Object interface {
	runtime.Object
	metav1.Object
}

// toObjectReference converts the given object into an objectReference.
func toObjectReference(owner Object, scheme *runtime.Scheme) objectReference {
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

	return objectReference{
		Name:      owner.GetName(),
		Namespace: owner.GetNamespace(),
		Kind:      gvk.Kind,
		Group:     gvk.Group,
	}
}

func getRefs(object metav1.Object) (refs []objectReference, err error) {
	annotations := object.GetAnnotations()
	if annotations == nil {
		return nil, nil
	}

	if data, present := annotations[OwnerAnnotation]; present {
		err = json.Unmarshal([]byte(data), &refs)
		return
	}
	return
}

func setRefs(object metav1.Object, refs []objectReference) error {
	annotations := object.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	if len(refs) == 0 {
		delete(annotations, OwnerAnnotation)
		return nil
	}

	b, err := json.Marshal(refs)
	if err != nil {
		return err
	}
	annotations[OwnerAnnotation] = string(b)
	object.SetAnnotations(annotations)
	return nil
}

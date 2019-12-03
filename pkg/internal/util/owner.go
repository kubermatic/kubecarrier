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

// ownerReference references an owning/controlling object across Namespaces.
type ownerReference struct {
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
func InsertOwnerReference(owner, object metav1.Object, scheme *runtime.Scheme) (changed bool, err error) {
	ownerReference, err := toOwnerReference(owner, scheme)
	if err != nil {
		return false, err
	}

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

// EnqueueRequestForOwner enqueues requests for all owners of an object.
//
// It implements the same behavior as handler.EnqueueRequestForOwner, but for our custom ownerReference.
func EnqueueRequestForOwner(ownerType metav1.Object, scheme *runtime.Scheme) (handler.EventHandler, error) {
	ownerTypeRef, err := toOwnerReference(ownerType, scheme)
	if err != nil {
		return nil, err
	}
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
	}, nil
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
func OwnedBy(owner metav1.Object, sc *runtime.Scheme) (GeneralizedListOption, error) {
	ref, err := toOwnerReference(owner, sc)
	if err != nil {
		return nil, err
	}
	return client.MatchingFields{
		OwnerAnnotation: ref.fieldIndexValue(),
	}, nil
}

// fieldIndexValue converts the ownerReference into a simple value for a client.FieldIndexer.
// to be used as key for indexing structure.
func (n ownerReference) fieldIndexValue() string {
	b, err := json.Marshal(n)
	if err != nil {
		// this should never ever happen
		panic(err)
	}
	return string(b)
}

// toOwnerReference converts the given object into an ownerReference.
func toOwnerReference(owner metav1.Object, scheme *runtime.Scheme) (ownerReference, error) {
	object, ok := owner.(runtime.Object)
	if !ok {
		return ownerReference{}, fmt.Errorf("%T is not a runtime.Object", owner)
	}

	gvk, err := apiutil.GVKForObject(object, scheme)
	if err != nil {
		return ownerReference{}, err
	}

	return ownerReference{
		Name:      owner.GetName(),
		Namespace: owner.GetNamespace(),
		Kind:      gvk.Kind,
		Group:     gvk.Group,
	}, nil
}

func getRefs(object metav1.Object) (refs []ownerReference, err error) {
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

func setRefs(object metav1.Object, refs []ownerReference) error {
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

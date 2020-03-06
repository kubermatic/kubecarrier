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
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

const (
	// ownerAnnotation is the annotation key that references the owner of this object.
	ownerAnnotation = "kubecarrier.io/owner"
)

type generalizedListOption interface {
	client.ListOption
	client.DeleteAllOfOption
}

// EnqueueRequestForOwner enqueues requests for all owners of an object.
//
// It implements the same behavior as handler.EnqueueRequestForOwner, but for our custom objectReference.
func EnqueueRequestForOwner(ownerType runtime.Object, scheme *runtime.Scheme) handler.EventHandler {
	ownerTypeRef := util.ToObjectReference(ownerType, scheme)
	ownerKind, ownerGroup := ownerTypeRef.Kind, ownerTypeRef.Group

	h := func(obj handler.MapObject) []reconcile.Request {
		refs, err := getRefs(obj.Object)
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
	}

	return &handler.EnqueueRequestsFromMapFunc{
		ToRequests: handler.ToRequestsFunc(h),
	}
}

// AddOwnerReverseFieldIndex adds a reverse index for OwnerReferences.
//
// The created index allows listing all owned objects of a given type by the owner object.
// Keep in mind this function should be called for each owned object type separately.
//
// See also: OwnedBy
func AddOwnerReverseFieldIndex(log logr.Logger, indexer client.FieldIndexer, object runtime.Object) error {
	return indexer.IndexField(
		object,
		ownerAnnotation,
		func(object runtime.Object) (values []string) {

			refs, err := getRefs(object)
			if err != nil {
				log.Error(err, "cannot list owner references")
				return
			}

			for _, r := range refs {
				values = append(values, fieldIndexValue(r))
			}
			return
		})
}

// OwnedBy returns owner filter for listing objects.
//
// See also: AddOwnerReverseFieldIndex
func OwnedBy(owner runtime.Object, sc *runtime.Scheme) generalizedListOption {
	return client.MatchingFields{
		ownerAnnotation: fieldIndexValue(util.ToObjectReference(owner, sc)),
	}
}

// insertOwnerReference adds an OwnerReference to the given object.
func insertOwnerReference(owner, object runtime.Object, scheme *runtime.Scheme) (changed bool, err error) {
	ownerReference := util.ToObjectReference(owner, scheme)

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

// deleteOwnerReference removes an owner from the given object.
func deleteOwnerReference(owner, object runtime.Object, scheme *runtime.Scheme) (changed bool, err error) {
	reference := util.ToObjectReference(owner, scheme)

	refs, err := getRefs(object)
	if err != nil {
		return false, err
	}

	var newRefs []util.ObjectReference
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

// isOwned checks if any owners claim ownership of this object.
func isOwned(object runtime.Object) (owned bool, err error) {
	refs, err := getRefs(object)
	if err != nil {
		return false, err
	}
	return len(refs) > 0, nil
}

// fieldIndexValue converts the objectReference into a simple value for a client.FieldIndexer.
// to be used as key for indexing structure.
func fieldIndexValue(n util.ObjectReference) string {
	b, err := json.Marshal(n)
	if err != nil {
		// this should never ever happen
		panic(err)
	}
	return string(b)
}

func getRefs(object runtime.Object) (refs []util.ObjectReference, err error) {
	objectAccessor, err := meta.Accessor(object)
	if err != nil {
		panic(fmt.Errorf("cannot get accessor for %T :%w", object, err))
	}
	annotations := objectAccessor.GetAnnotations()
	if annotations == nil {
		return nil, nil
	}

	if data, present := annotations[ownerAnnotation]; present {
		err = json.Unmarshal([]byte(data), &refs)
		return
	}
	return
}

func setRefs(object runtime.Object, refs []util.ObjectReference) error {
	objectAccessor, err := meta.Accessor(object)
	if err != nil {
		panic(fmt.Errorf("cannot get accessor for %T :%w", object, err))
	}
	annotations := objectAccessor.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	if len(refs) == 0 {
		delete(annotations, ownerAnnotation)
		return nil
	}

	b, err := json.Marshal(refs)
	if err != nil {
		return err
	}
	annotations[ownerAnnotation] = string(b)
	objectAccessor.SetAnnotations(annotations)
	return nil
}

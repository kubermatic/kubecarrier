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
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	// ownerNameLabel references the name of the owner of this object.
	ownerNameLabel = "owner.kubecarrier.io/name"
	// ownerNamespaceLabel references the namespace of the owner of this object.
	ownerNamespaceLabel = "owner.kubecarrier.io/namespace"
	// ownerTypeLabel references the type of the owner of this object.
	ownerTypeLabel = "owner.kubecarrier.io/type"
)

type generalizedListOption interface {
	client.ListOption
	client.DeleteAllOfOption
}

// SetOwnerReference sets a the owner as owner of object.
func SetOwnerReference(owner, object runtime.Object, scheme *runtime.Scheme) (changed bool) {
	objectAccessor, err := meta.Accessor(object)
	if err != nil {
		panic(fmt.Errorf("cannot get accessor for %T :%w", object, err))
	}

	labels := objectAccessor.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}

	ownerLabels := labelsForOwner(owner, scheme)
	for k, v := range ownerLabels {
		if labels[k] != v {
			changed = true
		}
		labels[k] = v
	}
	objectAccessor.SetLabels(labels)
	return
}

// RemoveOwnerReference removes an owner from the given object.
func RemoveOwnerReference(owner, object runtime.Object) (changed bool) {
	objectAccessor, err := meta.Accessor(object)
	if err != nil {
		panic(fmt.Errorf("cannot get accessor for %T :%w", object, err))
	}

	labels := objectAccessor.GetLabels()
	if labels == nil {
		return
	}

	if labels[ownerNameLabel] != "" || labels[ownerNamespaceLabel] != "" || labels[ownerTypeLabel] != "" {
		changed = true
	}
	delete(labels, ownerNameLabel)
	delete(labels, ownerNamespaceLabel)
	delete(labels, ownerTypeLabel)
	objectAccessor.SetLabels(labels)
	return
}

func requestHandlerForOwner(ownerType runtime.Object, scheme *runtime.Scheme) handler.ToRequestsFunc {
	gvk, err := apiutil.GVKForObject(ownerType, scheme)
	if err != nil {
		// if this panic occurs many, many other stuff has gone wrong as well
		// by owner type's safety ensures this is somewhat well formed k8s object
		// When using client-go API, it needs to be able to deduce GVK in the same manner
		// thus get/create/update/patch/delete shall error out long before this is called
		// This massively simplifies the function interface and allows OwnedBy to be a
		// one-liner instead of 3 line check which never errors
		// this is error is completely under our control, users of kubecarrier cannot
		// change cluster state to cause it.
		panic(fmt.Sprintf("cannot deduce GVK for owner (type %T)", ownerType))
	}

	gk := gvk.GroupKind().String()

	return func(obj handler.MapObject) (requests []reconcile.Request) {
		labels := obj.Meta.GetLabels()
		if labels == nil {
			return
		}

		ownerName, ok := labels[ownerNameLabel]
		if !ok {
			return
		}
		ownerNamespace, ok := labels[ownerNamespaceLabel]
		if !ok {
			return
		}
		ownerType, ok := labels[ownerTypeLabel]
		if !ok {
			return
		}

		if ownerType != gk {
			return
		}

		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      ownerName,
				Namespace: ownerNamespace,
			},
		})
		return
	}
}

// EnqueueRequestForOwner enqueues a request for the owner of an object
func EnqueueRequestForOwner(ownerType runtime.Object, scheme *runtime.Scheme) handler.EventHandler {
	return &handler.EnqueueRequestsFromMapFunc{
		ToRequests: requestHandlerForOwner(ownerType, scheme),
	}
}

// OwnedBy returns a list filter to fetch owned objects.
func OwnedBy(owner runtime.Object, scheme *runtime.Scheme) generalizedListOption {
	return client.MatchingLabels(labelsForOwner(owner, scheme))
}

// IsOwned checks if any owners claim ownership of this object.
func IsOwned(object metav1.Object) (owned bool) {
	l := object.GetLabels()
	if l == nil {
		return false
	}

	return l[ownerNameLabel] != "" && l[ownerNamespaceLabel] != "" && l[ownerTypeLabel] != ""
}

func labelsForOwner(obj runtime.Object, scheme *runtime.Scheme) map[string]string {
	gvk, err := apiutil.GVKForObject(obj, scheme)
	if err != nil {
		// if this panic occurs many, many other stuff has gone wrong as well
		// by owner type's safety ensures this is somewhat well formed k8s object
		// When using client-go API, it needs to be able to deduce GVK in the same manner
		// thus get/create/update/patch/delete shall error out long before this is called
		// This massively simplifies the function interface and allows OwnedBy to be a
		// one-liner instead of 3 line check which never errors
		// this is error is completely under our control, users of kubecarrier cannot
		// change cluster state to cause it.
		panic(fmt.Sprintf("cannot deduce GVK for owner (type %T)", obj))
	}

	metaAccessor, err := meta.Accessor(obj)
	if err != nil {
		panic(err)
	}

	return map[string]string{
		ownerNameLabel:      metaAccessor.GetName(),
		ownerNamespaceLabel: metaAccessor.GetNamespace(),
		ownerTypeLabel:      gvk.GroupKind().String(),
	}
}

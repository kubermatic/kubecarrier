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
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

// updateFunc is called to update the current existing object (actual) to the desired state.
type updateFunc func(actual, desired runtime.Object) error

// ReconcileOwnedObjects ensures that desired objects are up to date and
// other objects of the same type and owned by the same owner are either:
// 1. be removed or
// 2. the owner reference of this owner is removed from objects if they are owned by any other owners.
// It works as following. We have an object, the Owner, owning multiple objects in the kubernetes cluster. And we want
// to ensure that after this Reconciliation of owned objects finishes the only owned objects existing are those that
// are wanted. Also this would only operate on the kubernetes objects objectType GroupKind.
// In case object already exists in the kubernetes cluster the updateFn function is called allowing the user fixing
// between found and wanted object. In case the function is nil it's ignored.
func ReconcileOwnedObjects(
	ctx context.Context, log logr.Logger,
	client client.Client, scheme *runtime.Scheme,
	ownerObj runtime.Object,
	desired []runtime.Object, objectType runtime.Object,
	updateFn updateFunc,
) (err error) {
	existing, err := util.ListObjects(ctx, client, scheme, []runtime.Object{objectType}, OwnedBy(ownerObj, scheme))
	if err != nil {
		return fmt.Errorf("list Objects: %w", err)
	}

	wantedMap := make(map[util.ObjectReference]runtime.Object)
	for _, it := range desired {
		wantedMap[util.ToObjectReference(it, scheme)] = it
	}

	for _, obj := range existing {
		key := util.ToObjectReference(obj, scheme)
		if _, shouldExists := wantedMap[key]; !shouldExists {
			_, err := cleanupOutdatedReference(ctx, log, client, scheme, ownerObj, obj)
			if err != nil {
				return fmt.Errorf("cleanup unneeded references: %w", err)
			}
		}
	}

	for _, obj := range desired {
		// ctrl.CreateOrUpdate shall override obj with the current k8s value, thus we're performing a
		// deep copy to preserve wanted object data
		wantedObj := obj.DeepCopyObject()
		op, err := controllerruntime.CreateOrUpdate(ctx, client, obj, func() error {
			if _, err := insertOwnerReference(ownerObj, obj, scheme); err != nil {
				return fmt.Errorf("inserting OwnerReference: %w", err)
			}
			if updateFn != nil {
				return updateFn(obj, wantedObj)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("create or updating %v: %w", obj, err)
		}

		key := util.ToObjectReference(obj, scheme)
		if log != nil {
			log.V(6).Info("object "+string(op), "group", key.Group, "kind", key.Kind, "name", key.Name, "namespace", key.Namespace)
		}
	}
	return nil
}

// DeleteOwnedObjects performs cleanupOutdatedReference for all object of given types which are owned by the owner.
func DeleteOwnedObjects(
	ctx context.Context, log logr.Logger,
	client client.Client,
	scheme *runtime.Scheme,
	owner runtime.Object,
	listTypes []runtime.Object,
) (cleanedUp bool, err error) {
	objs, err := util.ListObjects(ctx, client, scheme, listTypes, OwnedBy(owner, scheme))
	if err != nil {
		return false, fmt.Errorf("ListObjects: %w", err)
	}

	cleanedUp = true
	for _, obj := range objs {
		cleanedObj, err := cleanupOutdatedReference(ctx, log, client, scheme, owner, obj)
		if err != nil {
			return false, err
		}
		if !cleanedObj {
			cleanedUp = false
		}
	}
	return
}

// cleanupOutdatedReference cleans up outdated objects as follows:
// 1. If the object is not owned by any owners, it will be deleted.
// 2. If the object is owned by any other owners, it will just remove the owner reference of this owner from the object.
func cleanupOutdatedReference(
	ctx context.Context, log logr.Logger,
	client client.Client,
	scheme *runtime.Scheme,
	owner runtime.Object,
	object runtime.Object,
) (cleanedUp bool, err error) {
	ownerReferenceChanged, err := deleteOwnerReference(owner, object, scheme)
	if err != nil {
		return false, fmt.Errorf("deleting OwnerReference: %w", err)
	}

	owned, err := isOwned(object)
	if err != nil {
		return false, fmt.Errorf("checking object isUnowned: %w", err)
	}
	objKey := util.ToObjectReference(object, scheme)
	ownerKey := util.ToObjectReference(owner, scheme)
	switch {
	case !owned:
		// The Object object is unowned by any objects, it can be removed.
		err := client.Delete(ctx, object)
		if errors.IsNotFound(err) {
			cleanedUp = true
		} else if err != nil {
			return false, fmt.Errorf("deleting Object: %w", err)
		}
		log.Info("deleting unowned Object",
			"OwnerName", ownerKey.Name, "OwnerNamespace", ownerKey.Namespace, "OwnerKind", ownerKey.Kind,
			"ObjectKind", objKey.Kind, "ObjectName", objKey.Name, "ObjectNamespace", objKey.Namespace)
	case owned && ownerReferenceChanged:
		if err := client.Update(ctx, object); err != nil {
			return false, fmt.Errorf("updating Object: %w", err)
		}
		cleanedUp = true
		log.Info("removing owner from Object",
			"OwnerName", ownerKey.Name, "OwnerNamespace", ownerKey.Namespace, "OwnerKind", ownerKey.Kind,
			"ObjectKind", objKey.Kind, "ObjectName", objKey.Name, "ObjectNamespace", objKey.Namespace)
	}
	return
}

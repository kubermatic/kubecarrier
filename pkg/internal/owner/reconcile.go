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

// Package reconcile implements reconcile functions for common Kubernetes types.

package owner

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
// other objects of the same type and owned by the same owner are removed.
// It works as following. We have an object, the Owner, owning multiple objects in the kubernetes cluster. And we want
// to ensure that after this Reconciliation of owned objects finishes the only owned objects existing are those that
// are wanted. Also this would only operate on the kubernetes objects objectType GroupKind.
// In case object already exists in the kubernetes cluster the updateFn function is called allowing the user fixing
// between found and wanted object. In case the function is nil it's ignored.
func ReconcileOwnedObjects(ctx context.Context, log logr.Logger, cl client.Client, scheme *runtime.Scheme, ownerObj runtime.Object, desired []runtime.Object, objectType runtime.Object, updateFn updateFunc) error {
	existing, err := util.ListObjects(ctx, cl, scheme, []runtime.Object{objectType}, OwnedBy(ownerObj, scheme))
	if err != nil {
		return err
	}

	wantedMap := make(map[util.ObjectReference]runtime.Object)
	for _, it := range desired {
		wantedMap[util.ToObjectReference(it, scheme)] = it
	}

	for _, obj := range existing {
		key := util.ToObjectReference(obj, scheme)
		if _, shouldExists := wantedMap[key]; !shouldExists {
			if err := cl.Delete(ctx, obj); err != nil {
				return client.IgnoreNotFound(err)
			}
		}
	}

	for _, obj := range desired {
		// ctrl.CreateOrUpdate shall override obj with the current k8s value, thus we're performing a
		// deep copy to preserve wanted object data
		wantedObj := obj.DeepCopyObject()
		op, err := controllerruntime.CreateOrUpdate(ctx, cl, obj, func() error {
			SetOwnerReference(ownerObj, obj, scheme)
			if updateFn != nil {
				return updateFn(obj, wantedObj)
			}
			return nil
		})
		if err != nil {
			return err
		}

		key := util.ToObjectReference(obj, scheme)
		if log != nil {
			log.V(6).Info("object "+string(op), "group", key.Group, "kind", key.Kind, "name", key.Name, "namespace", key.Namespace)
		}
	}
	return nil
}

// DeleteOwnedObjects deletes all object of given types which are owned by the owner.
func DeleteOwnedObjects(ctx context.Context,
	cl client.Client,
	scheme *runtime.Scheme,
	owner runtime.Object,
	listTypes []runtime.Object,
) (cleanedUp bool, err error) {
	objs, err := util.ListObjects(ctx, cl, scheme, listTypes, OwnedBy(owner, scheme))
	if err != nil {
		return false, fmt.Errorf("ListObjects: %w", err)
	}

	cleanedUp = true
	for _, obj := range objs {
		err := cl.Delete(ctx, obj)
		switch {
		case err == nil:
			cleanedUp = false
		case errors.IsNotFound(err):
		default:
			return false, fmt.Errorf("deleting ")
		}
	}
	return cleanedUp, nil
}

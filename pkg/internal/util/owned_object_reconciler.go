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
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/kubermatic/kubecarrier/pkg/internal/owner"
)

type OwnedObjectReconciler struct {
	Scheme      *runtime.Scheme
	Log         logr.Logger
	Owner       Object
	TypeFilter  []runtime.Object
	WantedState []Object
	MutateFn    func(oldObj, wantedObj runtime.Object) error
}

func (r *OwnedObjectReconciler) Reconcile(ctx context.Context, cl client.Client) (changed bool, err error) {
	existing, err := ListObjects(ctx, cl, r.Scheme, r.TypeFilter, owner.OwnedBy(r.Owner, r.Scheme))
	if err != nil {
		return false, fmt.Errorf("ListObjects: %w", err)
	}
	return r.ensureCreatedObject(ctx, cl, existing)
}

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
		if _, isList := ListObjType.(metav1.ListInterface); !isList {
			return nil, fmt.Errorf("cannot make a list out of a types: %v", gvk)
		}

		if err := cl.List(ctx, ListObjType, options...); err != nil {
			return nil, fmt.Errorf("listing %s.%s: %w", strings.ToLower(gvk.Kind), gvk.Group, err)
		}

		// for some reason there's no function in the list object for getting all the items...
		// but they all have .Items struct field
		items := reflect.ValueOf(ListObjType).Elem().FieldByName("Items")
		for i := 0; i < items.Len(); i++ {
			objs = append(objs, items.Index(i).Addr().Interface().(runtime.Object))
		}
	}
	return objs, nil
}

func (r *OwnedObjectReconciler) ensureCreatedObject(ctx context.Context, cl client.Client, existing []runtime.Object) (changed bool, err error) {
	wantedMap := make(map[ObjectReference]runtime.Object)
	for _, it := range r.WantedState {
		wantedMap[ToObjectReference(it, r.Scheme)] = it
	}

	for _, obj := range existing {
		key := ToObjectReference(obj.(Object), r.Scheme)
		if _, shouldExists := wantedMap[key]; !shouldExists {
			err := cl.Delete(ctx, obj)
			switch {
			case err == nil:
				changed = true
				if r.Log != nil {
					r.Log.V(6).Info("object deleted", "group", key.Group, "kind", key.Kind, "name", key.Name, "namespace", key.Namespace)
				}
			case errors.IsNotFound(err):
				break
			default:
				return changed, fmt.Errorf("deleting %v: %w", obj, err)
			}
		}
	}

	for _, obj := range r.WantedState {
		owner.SetOwnerReference(r.Owner, obj, r.Scheme)

		// ctrl.CreateOrUpdate shall override obj with the current k8s value, thus we're performing a
		// deep copy to preserve wanted object data
		wantedObj := obj.DeepCopyObject()
		op, err := ctrl.CreateOrUpdate(ctx, cl, obj, func() error {
			owner.SetOwnerReference(r.Owner, obj.(Object), r.Scheme)
			if err != nil {
				return fmt.Errorf("inserting owner ref %v: %w", obj, err)
			}
			if r.MutateFn != nil {
				return r.MutateFn(obj, wantedObj)
			}
			return nil
		})
		if op != controllerutil.OperationResultNone {
			changed = true
		}

		if err != nil {
			return changed, fmt.Errorf("create or deleting %v: %w", obj, err)
		}

		key := ToObjectReference(obj, r.Scheme)
		if r.Log != nil {
			r.Log.V(6).Info("object "+string(op), "group", key.Group, "kind", key.Kind, "name", key.Name, "namespace", key.Namespace)
		}
	}
	return changed, nil
}

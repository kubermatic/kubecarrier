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

// TODO: Write unit tests for this functionality
package util

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// TODO: replace list with something easier on the eyes
type List interface {
	runtime.Object
	metav1.ListInterface
}

type OwnedObjectReconciler struct {
	Scheme      *runtime.Scheme
	Log         logr.Logger
	Owner       Object
	TypeFilter  []List
	WantedState []Object
	MutateFn    func(obj runtime.Object) error
}

func (r *OwnedObjectReconciler) Do(ctx context.Context, cl client.Client) (changed bool, err error) {
	existing, err := listOwnedObjects(ctx, cl, r.Scheme, r.Owner, r.TypeFilter)
	if err != nil {
		return false, fmt.Errorf("listOwnedObjects: %w", err)
	}
	return r.ensureCreatedObject(ctx, cl, existing)
}

func listOwnedObjects(ctx context.Context, cl client.Client, scheme *runtime.Scheme, owner Object, lsts []List) ([]runtime.Object, error) {
	objs := make([]runtime.Object, 0)
	for _, lst := range lsts {
		gvk, err := apiutil.GVKForObject(lst, scheme)
		if err != nil {
			return nil, fmt.Errorf("cannot get GVK for %T: %w", lst, err)
		}
		if err := cl.List(ctx, lst, OwnedBy(owner, scheme)); err != nil {
			return nil, fmt.Errorf("listing %s.%s: %w", strings.ToLower(gvk.Kind), gvk.Group, err)
		}

		// for some reason there's no function in the list object for getting all the items...
		// but they all have .Items struct field
		items := reflect.ValueOf(lst).Elem().FieldByName("Items")
		for i := 0; i < items.Len(); i++ {
			objs = append(objs, items.Index(i).Addr().Interface().(runtime.Object))
		}
	}
	return objs, nil
}

func (r *OwnedObjectReconciler) ensureCreatedObject(ctx context.Context, cl client.Client, existing []runtime.Object) (changed bool, err error) {
	wantedMap := make(map[objectReference]runtime.Object)
	for _, it := range r.WantedState {
		wantedMap[toObjectReference(it, r.Scheme)] = it
	}

	for _, obj := range existing {
		key := toObjectReference(obj.(Object), r.Scheme)
		if _, shouldExists := wantedMap[key]; !shouldExists {
			changed = true
			if err := cl.Delete(ctx, obj); client.IgnoreNotFound(err) != nil {
				return changed, fmt.Errorf("deleting %v: %w", obj, err)
			}
			if r.Log != nil {
				r.Log.V(6).Info("object deleted", "group", key.Group, "kind", key.Kind, "name", key.Name, "namespace", key.Namespace)
			}
		}
	}

	for _, obj := range r.WantedState {
		_, err := InsertOwnerReference(r.Owner, obj, r.Scheme)
		if err != nil {
			return changed, fmt.Errorf("inserting owner ref %v: %w", obj, err)
		}
		op, err := ctrl.CreateOrUpdate(ctx, cl, obj, func() error {
			_, err := InsertOwnerReference(r.Owner, obj.(Object), r.Scheme)
			if err != nil {
				return fmt.Errorf("inserting owner ref %v: %w", obj, err)
			}
			if r.MutateFn != nil {
				return r.MutateFn(obj)
			}
			return nil
		})
		if op != controllerutil.OperationResultNone {
			changed = true
		}

		if err != nil {
			return changed, fmt.Errorf("create or deleting %v: %w", obj, err)
		}

		key := toObjectReference(obj, r.Scheme)
		if r.Log != nil {
			r.Log.V(6).Info("object "+string(op), "group", key.Group, "kind", key.Kind, "name", key.Name, "namespace", key.Namespace)
		}
	}
	return changed, nil
}

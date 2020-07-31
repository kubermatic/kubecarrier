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

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8c.io/utils/pkg/owner"
	"k8c.io/utils/pkg/util"

	catalogv1alpha1 "k8c.io/kubecarrier/pkg/apis/catalog/v1alpha1"
	elevatorutil "k8c.io/kubecarrier/pkg/elevator/internal/util"
)

// AdoptionReconciler reconciles Provider objects, that are not owned by a Tenant by creating the Tenant instance.
// The TenantObjReconciler will then "adopt" (add a OwnerReference).
type AdoptionReconciler struct {
	client.Client
	Log              logr.Logger
	Scheme           *runtime.Scheme
	NamespacedClient client.Client

	// Dynamic types we work with
	ProviderGVK, TenantGVK schema.GroupVersionKind

	DerivedCRName, ProviderNamespace string
}

func (r *AdoptionReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var (
		ctx    = context.Background()
		result ctrl.Result
	)

	providerObj := r.newProviderObject()
	if err := r.Get(ctx, req.NamespacedName, providerObj); err != nil {
		return result, client.IgnoreNotFound(err)
	}
	if owner.IsOwned(providerObj) ||
		!providerObj.GetDeletionTimestamp().IsZero() {
		// the object is already owned or was deleted, so we don't want to do anything.
		// otherwise we might recreate the owning object preventing the deletion of this instance.
		return result, nil
	}

	// Get DerivedCustomResource field configs
	derivedCR := &catalogv1alpha1.DerivedCustomResource{}
	if err := r.NamespacedClient.Get(ctx, types.NamespacedName{
		Name:      r.DerivedCRName,
		Namespace: r.ProviderNamespace,
	}, derivedCR); err != nil {
		return result, fmt.Errorf("getting DerivedCustomResource: %w", err)
	}
	version := r.ProviderGVK.Version
	exposeConfig, ok := elevatorutil.VersionExposeConfigForVersion(derivedCR.Spec.Expose, version)
	if !ok {
		return result, fmt.Errorf("missing version expose config for version %q", version)
	}

	// Reconcile ProviderCRD
	err := r.reconcileProviderObj(
		ctx, providerObj, exposeConfig)
	if err != nil {
		return result, fmt.Errorf("reconciling %s: %w", r.ProviderGVK.Kind, err)
	}

	return result, nil
}

func (r *AdoptionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(r.newProviderObject()).
		WithEventFilter(util.PredicateFn(func(obj runtime.Object) bool {
			// we are only interested in unowned objects
			meta, ok := obj.(metav1.Object)
			if !ok {
				return false
			}

			return len(meta.GetOwnerReferences()) == 0
		})).
		Complete(r)
}

func (r *AdoptionReconciler) reconcileProviderObj(
	ctx context.Context, providerObj *unstructured.Unstructured,
	config catalogv1alpha1.VersionExposeConfig,
) error {
	desiredTenantObj := r.newTenantObject()
	desiredTenantObj.SetName(providerObj.GetName())
	desiredTenantObj.SetNamespace(providerObj.GetNamespace())

	// prepare config
	_, otherFields := elevatorutil.SplitStatusFields(config.Fields)

	// Lookup current instance
	currentTenantObj := r.newTenantObject()
	err := r.Get(ctx, types.NamespacedName{
		Name:      desiredTenantObj.GetName(),
		Namespace: desiredTenantObj.GetNamespace(),
	}, currentTenantObj)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("getting %s: %w", r.TenantGVK.Kind, err)
	}
	if errors.IsNotFound(err) {
		// Create the Tenant Obj
		if err = elevatorutil.CopyFields(providerObj, desiredTenantObj, otherFields); err != nil {
			return fmt.Errorf("copy fields: %w", err)
		}

		if err = r.Create(ctx, desiredTenantObj); err != nil {
			return fmt.Errorf("creating %s: %w", r.TenantGVK.Kind, err)
		}
	}

	return nil
}

func (r *AdoptionReconciler) newTenantObject() *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(r.TenantGVK)
	return obj
}

func (r *AdoptionReconciler) newProviderObject() *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(r.ProviderGVK)
	return obj
}

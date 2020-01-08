/*
Copyright 2020 The KubeCarrier Authors.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

const (
	catalogControllerFinalizer = "catalog.kubecarrier.io/controller"
)

// CatalogReconciler reconciles a Catalog object
type CatalogReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=catalogs,verbs=get;list;watch;update;
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=catalogs/status,verbs=get;update;patch

// Reconcile function reconciles the Catalog object which specified by the request. Currently, it does the follwoing:
// - Fetch the Catalog object.
// - Handle the deletion of the Catalog object.
// - Fetch the CatalogEntries and TenantReferences that selected by this Catalog object.
// - Update the Status of the Catalog object.
func (r *CatalogReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("catalog", req.NamespacedName)

	// Fetch the Catalog object.
	catalog := &catalogv1alpha1.Catalog{}
	if err := r.Get(ctx, req.NamespacedName, catalog); err != nil {
		// If the Catalog object is already gone, we just ignore the NotFound error.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle the deletion of the Catalog object.
	if !catalog.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, log, catalog); err != nil {
			return ctrl.Result{}, fmt.Errorf("handling deletion: %w", err)
		}
		return ctrl.Result{}, nil
	}

	// check/add the finalizer for the Catalog.
	if util.AddFinalizer(catalog, catalogControllerFinalizer) {
		// Update the Catalog with the finalizer
		if err := r.Update(ctx, catalog); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating finalizers: %w", err)
		}
	}

	// Get CatalogEntries.
	catalogEntries, err := r.listSelectedCatalogEntries(ctx, log, catalog)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("getting selected CatalogEntries: %w", err)
	}

	var entries []catalogv1alpha1.ObjectReference
	for _, catalogEntry := range catalogEntries {
		entries = append(entries,
			catalogv1alpha1.ObjectReference{
				Name: catalogEntry.Name,
			})
	}

	catalog.Status.Entries = entries

	// Get TenantReferences.
	tenantReferences, err := r.listSelectedTenantReferences(ctx, log, catalog)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("getting selected TenantReferences: %w", err)
	}

	var tenants []catalogv1alpha1.ObjectReference
	for _, tenantReference := range tenantReferences {
		tenants = append(tenants,
			catalogv1alpha1.ObjectReference{
				Name: tenantReference.Name,
			})
	}
	catalog.Status.Tenants = tenants

	// Update Catalog Status.
	catalog.Status.ObservedGeneration = catalog.Generation
	catalog.Status.SetCondition(catalogv1alpha1.CatalogCondition{
		Type:    catalogv1alpha1.CatalogReady,
		Status:  catalogv1alpha1.ConditionTrue,
		Reason:  "CatalogReady",
		Message: "Catalog is Ready.",
	})

	if err := r.Status().Update(ctx, catalog); err != nil {
		return ctrl.Result{}, fmt.Errorf("updating Catalog Status: %w", err)
	}

	return ctrl.Result{}, nil
}

func (r *CatalogReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&catalogv1alpha1.Catalog{}).
		Complete(r)
}

func (r *CatalogReconciler) handleDeletion(ctx context.Context, log logr.Logger, catalog *catalogv1alpha1.Catalog) error {
	// Update the Catalog Status to Terminating.
	readyCondition, _ := catalog.Status.GetCondition(catalogv1alpha1.CatalogReady)
	if readyCondition.Status != catalogv1alpha1.ConditionFalse ||
		readyCondition.Status == catalogv1alpha1.ConditionFalse && readyCondition.Reason != catalogv1alpha1.CatalogTerminatingReason {
		catalog.Status.ObservedGeneration = catalog.Generation
		catalog.Status.SetCondition(catalogv1alpha1.CatalogCondition{
			Type:    catalogv1alpha1.CatalogReady,
			Status:  catalogv1alpha1.ConditionFalse,
			Reason:  catalogv1alpha1.CatalogTerminatingReason,
			Message: "Catalog is being terminated",
		})
		if err := r.Status().Update(ctx, catalog); err != nil {
			return fmt.Errorf("updating Catalog status: %w", err)
		}
	}

	// Remove the finalizer.
	if util.RemoveFinalizer(catalog, catalogControllerFinalizer) {
		if err := r.Update(ctx, catalog); err != nil {
			return fmt.Errorf("updating Catalog: %w", err)
		}
	}
	return nil
}

func (r *CatalogReconciler) listSelectedCatalogEntries(ctx context.Context, log logr.Logger, catalog *catalogv1alpha1.Catalog) ([]catalogv1alpha1.CatalogEntry, error) {
	catalogEntrySelector, err := metav1.LabelSelectorAsSelector(catalog.Spec.CatalogEntrySelector)
	if err != nil {
		return nil, fmt.Errorf("parsing CatalogEntry selector: %w", err)
	}
	catalogEntries := &catalogv1alpha1.CatalogEntryList{}
	if err := r.List(ctx, catalogEntries, client.InNamespace(catalog.Namespace), client.MatchingLabelsSelector{Selector: catalogEntrySelector}); err != nil {
		return nil, fmt.Errorf("listing CatalogEntry: %w", err)
	}
	return catalogEntries.Items, nil
}

func (r *CatalogReconciler) listSelectedTenantReferences(ctx context.Context, log logr.Logger, catalog *catalogv1alpha1.Catalog) ([]catalogv1alpha1.TenantReference, error) {
	tenantReferenceSelector, err := metav1.LabelSelectorAsSelector(catalog.Spec.TenantReferenceSelector)
	if err != nil {
		return nil, fmt.Errorf("parsing TenantReference selector: %w", err)
	}
	tenantReferences := &catalogv1alpha1.TenantReferenceList{}
	if err := r.List(ctx, tenantReferences, client.InNamespace(catalog.Namespace), client.MatchingLabelsSelector{Selector: tenantReferenceSelector}); err != nil {
		return nil, fmt.Errorf("listing TenantReference: %w", err)
	}
	return tenantReferences.Items, nil
}

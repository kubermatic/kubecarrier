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
	"reflect"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

const (
	catalogControllerFinalizer = "catalog.kubecarrier.io/controller"
)

// CatalogReconciler reconciles a Catalog object
type CatalogReconciler struct {
	client.Client
	Log                        logr.Logger
	Scheme                     *runtime.Scheme
	KubeCarrierSystemNamespace string
}

// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=catalogs,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=catalogs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=catalogentries,verbs=list;watch
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=offerings,verbs=get;list;watch;create;update;delete

// Reconcile function reconciles the Catalog object which specified by the request. Currently, it does the following:
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

	// First update the entries and tenants to the status.
	if err := r.updateStatus(ctx, catalog, nil); err != nil {
		return ctrl.Result{}, fmt.Errorf("updating Catalog Status: %w", err)
	}

	// Get Provider
	provider, err := getProviderByProviderNamespace(ctx, r.Client, r.KubeCarrierSystemNamespace, req.Namespace)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("getting Provider: %w", err)
	}

	desiredOfferings, err := r.buildDesiredOfferings(ctx, log, catalog, provider, tenantReferences, catalogEntries)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("building Offering: %w", err)
	}

	if err := r.reconcileOfferings(ctx, log, catalog, desiredOfferings); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconcliing Offerings: %w", err)
	}

	// Update Catalog Status.
	if err := r.updateStatus(ctx, catalog, &catalogv1alpha1.CatalogCondition{
		Type:    catalogv1alpha1.CatalogReady,
		Status:  catalogv1alpha1.ConditionTrue,
		Reason:  "CatalogReady",
		Message: "Catalog is Ready.",
	}); err != nil {
		return ctrl.Result{}, fmt.Errorf("updating Catalog Status: %w", err)
	}

	return ctrl.Result{}, nil
}

func (r *CatalogReconciler) SetupWithManager(mgr ctrl.Manager) error {
	enqueueAllCatalogsInNamespace := &handler.EnqueueRequestsFromMapFunc{
		ToRequests: handler.ToRequestsFunc(func(mapObject handler.MapObject) (out []reconcile.Request) {
			catalogList := &catalogv1alpha1.CatalogList{}
			if err := r.List(context.Background(), catalogList, client.InNamespace(mapObject.Meta.GetNamespace())); err != nil {
				// This will makes the manager crashes, and it will restart and reconcile all objects again.
				panic(fmt.Errorf("listting Catalog: %w", err))
			}
			for _, catalog := range catalogList.Items {
				out = append(out, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      catalog.Name,
						Namespace: catalog.Namespace,
					},
				})
			}
			return
		}),
	}
	enqueuerForOwner, err := util.EnqueueRequestForOwner(&catalogv1alpha1.Catalog{}, mgr.GetScheme())
	if err != nil {
		return fmt.Errorf("cannot create enqueuerForOnwer for Catalog: %w", err)
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&catalogv1alpha1.Catalog{}).
		Watches(&source.Kind{Type: &catalogv1alpha1.TenantReference{}}, enqueueAllCatalogsInNamespace).
		Watches(&source.Kind{Type: &catalogv1alpha1.CatalogEntry{}}, enqueueAllCatalogsInNamespace).
		Watches(&source.Kind{Type: &catalogv1alpha1.Offering{}}, enqueuerForOwner).
		Complete(r)
}

func (r *CatalogReconciler) handleDeletion(ctx context.Context, log logr.Logger, catalog *catalogv1alpha1.Catalog) error {
	// Update the Catalog Status to Terminating.
	readyCondition, _ := catalog.Status.GetCondition(catalogv1alpha1.CatalogReady)
	if readyCondition.Status != catalogv1alpha1.ConditionFalse ||
		readyCondition.Status == catalogv1alpha1.ConditionFalse && readyCondition.Reason != catalogv1alpha1.CatalogTerminatingReason {
		if err := r.updateStatus(ctx, catalog, &catalogv1alpha1.CatalogCondition{
			Type:    catalogv1alpha1.CatalogReady,
			Status:  catalogv1alpha1.ConditionFalse,
			Reason:  catalogv1alpha1.CatalogTerminatingReason,
			Message: "Catalog is being terminated",
		}); err != nil {
			return fmt.Errorf("updating Catalog Status: %w", err)
		}
	}

	deletedOfferingsCounter, err := r.cleanupOffering(ctx, log, catalog, nil)
	if err != nil {
		return fmt.Errorf("cleaning up Offerings: %w", err)
	}
	if deletedOfferingsCounter != 0 {
		// move to the next reconcilation round.
		return nil
	}

	if util.RemoveFinalizer(catalog, catalogControllerFinalizer) {
		if err := r.Update(ctx, catalog); err != nil {
			return fmt.Errorf("updating Catalog: %w", err)
		}
	}
	return nil
}

func (r *CatalogReconciler) updateStatus(
	ctx context.Context,
	catalog *catalogv1alpha1.Catalog,
	condition *catalogv1alpha1.CatalogCondition,
) error {
	catalog.Status.ObservedGeneration = catalog.Generation
	if condition != nil {
		catalog.Status.SetCondition(*condition)
	}
	if err := r.Status().Update(ctx, catalog); err != nil {
		return fmt.Errorf("updating Catalog status: %w", err)
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

func (r *CatalogReconciler) buildDesiredOfferings(
	ctx context.Context, log logr.Logger,
	catalog *catalogv1alpha1.Catalog,
	provider *catalogv1alpha1.Provider,
	tenantReferences []catalogv1alpha1.TenantReference,
	catalogEntries []catalogv1alpha1.CatalogEntry,
) ([]catalogv1alpha1.Offering, error) {
	var desiredOfferings []catalogv1alpha1.Offering
	for _, tenantReference := range tenantReferences {
		tenant := &catalogv1alpha1.Tenant{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      tenantReference.Name,
			Namespace: r.KubeCarrierSystemNamespace,
		}, tenant); err != nil {
			return nil, fmt.Errorf("getting Tenant: %w", err)
		}
		for _, catalogEntry := range catalogEntries {
			desiredOfferings = append(desiredOfferings, catalogv1alpha1.Offering{
				ObjectMeta: metav1.ObjectMeta{
					Name:      catalogEntry.Name,
					Namespace: tenant.Status.NamespaceName,
				},
				Offering: catalogv1alpha1.OfferingData{
					Metadata: catalogv1alpha1.OfferingMetadata{
						DisplayName: catalogEntry.Spec.Metadata.DisplayName,
						Description: catalogEntry.Spec.Metadata.Description,
					},
					Provider: catalogv1alpha1.ObjectReference{
						Name: provider.Name,
					},
					CRDs: catalogEntry.Status.CRDs,
				},
			})
		}
	}
	return desiredOfferings, nil
}

func (r *CatalogReconciler) reconcileOfferings(
	ctx context.Context, log logr.Logger,
	catalog *catalogv1alpha1.Catalog,
	desiredOfferings []catalogv1alpha1.Offering,
) error {
	if _, err := r.cleanupOffering(ctx, log, catalog, desiredOfferings); err != nil {
		return fmt.Errorf("cleanup Offering: %w", err)
	}

	for _, desiredOffering := range desiredOfferings {
		if _, err := util.InsertOwnerReference(catalog, &desiredOffering, r.Scheme); err != nil {
			return fmt.Errorf("inserting OwnerRefernence: %w", err)
		}

		currentOffering := &catalogv1alpha1.Offering{}
		err := r.Get(ctx, types.NamespacedName{
			Namespace: desiredOffering.Name,
			Name:      desiredOffering.Namespace,
		}, currentOffering)
		if err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("getting Offering: %w", err)
		}
		if errors.IsNotFound(err) {
			// Create the Offering.
			if err := r.Create(ctx, &desiredOffering); err != nil && !errors.IsAlreadyExists(err) {
				return fmt.Errorf("creating Offering: %w", err)
			}
			currentOffering = &desiredOffering
		}

		ownerChanged, err := util.InsertOwnerReference(catalog, currentOffering, r.Scheme)
		if err != nil {
			return fmt.Errorf("inserting OwnerReference: %w", err)
		}
		if !reflect.DeepEqual(desiredOffering.Offering, currentOffering.Offering) || ownerChanged {
			currentOffering.Offering = desiredOffering.Offering
			if err := r.Update(ctx, currentOffering); err != nil {
				return fmt.Errorf("updaing Offering: %w", err)
			}
		}
	}
	return nil
}

func (r *CatalogReconciler) cleanupOffering(
	ctx context.Context, log logr.Logger,
	catalog *catalogv1alpha1.Catalog,
	desiredOfferings []catalogv1alpha1.Offering,
) (int, error) {
	// Fetch existing Offerings.
	ownerListFilter, err := util.OwnedBy(catalog, r.Scheme)
	if err != nil {
		return 0, fmt.Errorf("building OwnedBy filter: %w", err)
	}
	foundOfferingList := &catalogv1alpha1.OfferingList{}
	if err := r.List(ctx, foundOfferingList, ownerListFilter); err != nil {
		return 0, fmt.Errorf("listing Offerings: %w", err)
	}

	desiredIndces := map[string]struct{}{}
	for _, desiredOffering := range desiredOfferings {
		desiredIndces[types.NamespacedName{
			Name:      desiredOffering.Name,
			Namespace: desiredOffering.Namespace,
		}.String()] = struct{}{}
	}

	var deletedOfferings int
	// Delete Offerings that are no longer in the desiredOfferings list
	for _, currentOffering := range foundOfferingList.Items {
		if _, present := desiredIndces[types.NamespacedName{
			Name:      currentOffering.Name,
			Namespace: currentOffering.Namespace,
		}.String()]; present {
			continue
		}
		ownerReferenceChanged, err := util.DeleteOwnerReference(catalog, &currentOffering, r.Scheme)
		if err != nil {
			return 0, fmt.Errorf("deleting OwnerReference: %w", err)
		}

		unowned, err := util.IsUnowned(&currentOffering)
		if err != nil {
			return 0, fmt.Errorf("checking object isUnowned: %w", err)
		}

		switch {
		case unowned:
			// The Offering object is unowned by any Catalog objects, it can be removed.
			if err := r.Delete(ctx, &currentOffering); err != nil && !errors.IsNotFound(err) {
				return 0, fmt.Errorf("deleting Offering: %w", err)
			}
			log.Info("deleting unowned Offering", "OfferingName", currentOffering.Name, "OfferingNamespace", currentOffering.Namespace)
			deletedOfferings++
		case !unowned && ownerReferenceChanged:
			if err := r.Update(ctx, &currentOffering); err != nil {
				return 0, fmt.Errorf("updating Offering: %w", err)
			}
			log.Info("removing Catalog as owner from Offering", "OfferingName", currentOffering.Name, "OfferingNamespace", currentOffering.Namespace)
			deletedOfferings++
		}
	}
	return deletedOfferings, nil
}

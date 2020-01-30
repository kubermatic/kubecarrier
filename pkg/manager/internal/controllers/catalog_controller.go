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
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
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
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=serviceclusterreferences,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=providerreferences,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=kubecarrier.io,resources=serviceclusters,verbs=get;list;watch

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
	provider, err := GetProviderByProviderNamespace(ctx, r.Client, r.KubeCarrierSystemNamespace, req.Namespace)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("getting Provider: %w", err)
	}

	var (
		desiredProviderReferences       []catalogv1alpha1.ProviderReference
		desiredOfferings                []catalogv1alpha1.Offering
		desiredServiceClusterReferences []catalogv1alpha1.ServiceClusterReference
	)
	for _, tenantReference := range tenantReferences {
		tenant := &catalogv1alpha1.Tenant{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      tenantReference.Name,
			Namespace: r.KubeCarrierSystemNamespace,
		}, tenant); err != nil {
			return ctrl.Result{}, fmt.Errorf("getting Tenant: %w", err)
		}

		if !tenant.IsReady() {
			continue
		}

		desiredProviderReferences = append(desiredProviderReferences, r.buildDesiredProviderReference(provider, tenant))
		desiredOfferings = append(desiredOfferings, r.buildDesiredOfferings(provider, tenant, catalogEntries)...)
		desiredServiceClusterReferencesForTenant, err := r.buildDesiredServiceClusterReferences(ctx, log, provider, tenant, catalogEntries)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("building ServiceClusterReference: %w", err)
		}
		desiredServiceClusterReferences = append(desiredServiceClusterReferences, desiredServiceClusterReferencesForTenant...)
	}

	if err := r.reconcileProviderReferences(ctx, log, catalog, desiredProviderReferences); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconcliing ProviderReferences: %w", err)
	}

	if err := r.reconcileOfferings(ctx, log, catalog, desiredOfferings); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconcliing Offerings: %w", err)
	}

	if err := r.reconcileServiceClusterReferences(ctx, log, catalog, desiredServiceClusterReferences); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconcliing ServiceClusterReferences: %w", err)
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
	enqueuerForOwner := util.EnqueueRequestForOwner(&catalogv1alpha1.Catalog{}, mgr.GetScheme())
	return ctrl.NewControllerManagedBy(mgr).
		For(&catalogv1alpha1.Catalog{}).
		Watches(&source.Kind{Type: &catalogv1alpha1.TenantReference{}}, enqueueAllCatalogsInNamespace).
		Watches(&source.Kind{Type: &catalogv1alpha1.CatalogEntry{}}, enqueueAllCatalogsInNamespace).
		Watches(&source.Kind{Type: &corev1alpha1.ServiceCluster{}}, enqueueAllCatalogsInNamespace).
		Watches(&source.Kind{Type: &catalogv1alpha1.Offering{}}, enqueuerForOwner).
		Watches(&source.Kind{Type: &catalogv1alpha1.ProviderReference{}}, enqueuerForOwner).
		Watches(&source.Kind{Type: &catalogv1alpha1.ServiceClusterReference{}}, enqueuerForOwner).
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

	deletedOfferingsCounter, err := r.cleanupOfferings(ctx, log, catalog, nil)
	if err != nil {
		return fmt.Errorf("cleaning up Offerings: %w", err)
	}
	deletedProviderReferencesCounter, err := r.cleanupProviderReferences(ctx, log, catalog, nil)
	if err != nil {
		return fmt.Errorf("cleaning up ProviderReferences: %w", err)
	}
	deletedServiceClusterReferencesCounter, err := r.cleanupServiceClusterReferences(ctx, log, catalog, nil)
	if err != nil {
		return fmt.Errorf("cleaning up ServiceClusterReferences: %w", err)
	}
	if deletedOfferingsCounter != 0 || deletedProviderReferencesCounter != 0 || deletedServiceClusterReferencesCounter != 0 {
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
	provider *catalogv1alpha1.Provider,
	tenant *catalogv1alpha1.Tenant,
	catalogEntries []catalogv1alpha1.CatalogEntry,
) []catalogv1alpha1.Offering {
	var desiredOfferings []catalogv1alpha1.Offering
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
	return desiredOfferings
}

func (r *CatalogReconciler) buildDesiredProviderReference(
	provider *catalogv1alpha1.Provider,
	tenant *catalogv1alpha1.Tenant,
) catalogv1alpha1.ProviderReference {
	return catalogv1alpha1.ProviderReference{
		ObjectMeta: metav1.ObjectMeta{
			Name:      provider.Name,
			Namespace: tenant.Status.NamespaceName,
		},
		Spec: catalogv1alpha1.ProviderReferenceSpec{
			Metadata: provider.Spec.Metadata,
		},
	}
}

func (r *CatalogReconciler) buildDesiredServiceClusterReferences(
	ctx context.Context, log logr.Logger,
	provider *catalogv1alpha1.Provider,
	tenant *catalogv1alpha1.Tenant,
	catalogEntries []catalogv1alpha1.CatalogEntry,
) ([]catalogv1alpha1.ServiceClusterReference, error) {
	var desiredServiceClusterReferences []catalogv1alpha1.ServiceClusterReference
	serviceClusterNames := map[string]struct{}{}
	for _, catalogEntry := range catalogEntries {
		for _, crdInfo := range catalogEntry.Status.CRDs {
			serviceClusterNames[crdInfo.ServiceCluster.Name] = struct{}{}
		}
	}
	for serviceClusterName := range serviceClusterNames {
		serviceCluster := &corev1alpha1.ServiceCluster{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      serviceClusterName,
			Namespace: provider.Status.NamespaceName,
		}, serviceCluster); err != nil {
			return nil, fmt.Errorf("getting ServiceCluster: %w", err)
		}
		desiredServiceClusterReferences = append(desiredServiceClusterReferences, catalogv1alpha1.ServiceClusterReference{
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceClusterName + "." + provider.Name,
				Namespace: tenant.Status.NamespaceName,
			},
			Spec: catalogv1alpha1.ServiceClusterReferenceSpec{
				Metadata: serviceCluster.Spec.Metadata,
				Provider: catalogv1alpha1.ObjectReference{
					Name: provider.Name,
				},
			},
		})
	}
	return desiredServiceClusterReferences, nil
}

func (r *CatalogReconciler) reconcileOfferings(
	ctx context.Context, log logr.Logger,
	catalog *catalogv1alpha1.Catalog,
	desiredOfferings []catalogv1alpha1.Offering,
) error {
	if _, err := r.cleanupOfferings(ctx, log, catalog, desiredOfferings); err != nil {
		return fmt.Errorf("cleanup Offering: %w", err)
	}

	for _, desiredOffering := range desiredOfferings {
		if _, err := util.InsertOwnerReference(catalog, &desiredOffering, r.Scheme); err != nil {
			return fmt.Errorf("inserting OwnerRefernence: %w", err)
		}

		foundOffering := &catalogv1alpha1.Offering{}
		err := r.Get(ctx, types.NamespacedName{
			Namespace: desiredOffering.Name,
			Name:      desiredOffering.Namespace,
		}, foundOffering)
		if err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("getting Offering: %w", err)
		}
		if errors.IsNotFound(err) {
			// Create the Offering.
			if err := r.Create(ctx, &desiredOffering); err != nil && !errors.IsAlreadyExists(err) {
				return fmt.Errorf("creating Offering: %w", err)
			}
			foundOffering = &desiredOffering
		}

		ownerChanged, err := util.InsertOwnerReference(catalog, foundOffering, r.Scheme)
		if err != nil {
			return fmt.Errorf("inserting OwnerReference: %w", err)
		}
		if !reflect.DeepEqual(desiredOffering.Offering, foundOffering.Offering) || ownerChanged {
			foundOffering.Offering = desiredOffering.Offering
			if err := r.Update(ctx, foundOffering); err != nil {
				return fmt.Errorf("updaing Offering: %w", err)
			}
		}
	}
	return nil
}

func (r *CatalogReconciler) reconcileProviderReferences(
	ctx context.Context, log logr.Logger,
	catalog *catalogv1alpha1.Catalog,
	desiredProviderReferences []catalogv1alpha1.ProviderReference,
) error {
	if _, err := r.cleanupProviderReferences(ctx, log, catalog, desiredProviderReferences); err != nil {
		return fmt.Errorf("cleanup ProviderReference: %w", err)
	}

	for _, desiredProviderReference := range desiredProviderReferences {
		if _, err := util.InsertOwnerReference(catalog, &desiredProviderReference, r.Scheme); err != nil {
			return fmt.Errorf("inserting OwnerRefernence: %w", err)
		}

		foundProviderReference := &catalogv1alpha1.ProviderReference{}
		err := r.Get(ctx, types.NamespacedName{
			Namespace: desiredProviderReference.Name,
			Name:      desiredProviderReference.Namespace,
		}, foundProviderReference)
		if err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("getting ProviderReference: %w", err)
		}
		if errors.IsNotFound(err) {
			// Create the ProviderReference.
			if err := r.Create(ctx, &desiredProviderReference); err != nil && !errors.IsAlreadyExists(err) {
				return fmt.Errorf("creating ProviderReference: %w", err)
			}
			foundProviderReference = &desiredProviderReference
		}

		ownerChanged, err := util.InsertOwnerReference(catalog, foundProviderReference, r.Scheme)
		if err != nil {
			return fmt.Errorf("inserting OwnerReference: %w", err)
		}
		if !reflect.DeepEqual(desiredProviderReference.Spec, foundProviderReference.Spec) || ownerChanged {
			foundProviderReference.Spec = desiredProviderReference.Spec
			if err := r.Update(ctx, foundProviderReference); err != nil {
				return fmt.Errorf("updaing ProviderReference: %w", err)
			}
		}
	}
	return nil
}

func (r *CatalogReconciler) reconcileServiceClusterReferences(
	ctx context.Context, log logr.Logger,
	catalog *catalogv1alpha1.Catalog,
	desiredServiceClusterReferences []catalogv1alpha1.ServiceClusterReference,
) error {
	if _, err := r.cleanupServiceClusterReferences(ctx, log, catalog, desiredServiceClusterReferences); err != nil {
		return fmt.Errorf("cleanup ServiceClusterReference: %w", err)
	}

	for _, desiredServiceClusterReference := range desiredServiceClusterReferences {
		if _, err := util.InsertOwnerReference(catalog, &desiredServiceClusterReference, r.Scheme); err != nil {
			return fmt.Errorf("inserting OwnerRefernence: %w", err)
		}

		foundServiceClusterReference := &catalogv1alpha1.ServiceClusterReference{}
		err := r.Get(ctx, types.NamespacedName{
			Namespace: desiredServiceClusterReference.Name,
			Name:      desiredServiceClusterReference.Namespace,
		}, foundServiceClusterReference)
		if err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("getting ServiceClusterReference: %w", err)
		}
		if errors.IsNotFound(err) {
			// Create the ServiceClusterReference.
			if err := r.Create(ctx, &desiredServiceClusterReference); err != nil && !errors.IsAlreadyExists(err) {
				return fmt.Errorf("creating ServiceClusterReference: %w", err)
			}
			foundServiceClusterReference = &desiredServiceClusterReference
		}

		ownerChanged, err := util.InsertOwnerReference(catalog, foundServiceClusterReference, r.Scheme)
		if err != nil {
			return fmt.Errorf("inserting OwnerReference: %w", err)
		}
		if !reflect.DeepEqual(desiredServiceClusterReference.Spec, foundServiceClusterReference.Spec) || ownerChanged {
			foundServiceClusterReference.Spec = desiredServiceClusterReference.Spec
			if err := r.Update(ctx, foundServiceClusterReference); err != nil {
				return fmt.Errorf("updaing ServiceClusterReference: %w", err)
			}
		}
	}
	return nil
}

func (r *CatalogReconciler) cleanupOfferings(
	ctx context.Context, log logr.Logger,
	catalog *catalogv1alpha1.Catalog,
	desiredOfferings []catalogv1alpha1.Offering,
) (deletedOfferingCounter int, err error) {
	// Fetch existing Offerings.
	foundOfferingList := &catalogv1alpha1.OfferingList{}
	if err := r.List(ctx, foundOfferingList, util.OwnedBy(catalog, r.Scheme)); err != nil {
		return 0, fmt.Errorf("listing Offerings: %w", err)
	}
	return r.cleanupOutdatedReferences(ctx, log,
		catalog,
		offeringsToObjectArray(foundOfferingList.Items),
		offeringsToObjectArray(desiredOfferings))
}

func (r *CatalogReconciler) cleanupProviderReferences(
	ctx context.Context, log logr.Logger,
	catalog *catalogv1alpha1.Catalog,
	desiredProviderReferences []catalogv1alpha1.ProviderReference,
) (deletedProviderReferenceCounter int, err error) {
	// Fetch existing ProviderReferences.
	foundProviderReferenceList := &catalogv1alpha1.ProviderReferenceList{}
	if err := r.List(ctx, foundProviderReferenceList, util.OwnedBy(catalog, r.Scheme)); err != nil {
		return 0, fmt.Errorf("listing ProviderReferences: %w", err)
	}

	return r.cleanupOutdatedReferences(ctx, log,
		catalog,
		providerReferencesToObjectArray(foundProviderReferenceList.Items),
		providerReferencesToObjectArray(desiredProviderReferences))
}

func (r *CatalogReconciler) cleanupServiceClusterReferences(
	ctx context.Context, log logr.Logger,
	catalog *catalogv1alpha1.Catalog,
	desiredServiceClusterReferences []catalogv1alpha1.ServiceClusterReference,
) (deletedServiceClusterReferenceCounter int, err error) {
	// Fetch existing ServiceClusterReferences.
	foundServiceClusterReferenceList := &catalogv1alpha1.ServiceClusterReferenceList{}
	if err := r.List(ctx, foundServiceClusterReferenceList, util.OwnedBy(catalog, r.Scheme)); err != nil {
		return 0, fmt.Errorf("listing ServiceClusterReferences: %w", err)
	}
	return r.cleanupOutdatedReferences(ctx, log,
		catalog,
		serviceClusterReferencesToObjectArray(foundServiceClusterReferenceList.Items),
		serviceClusterReferencesToObjectArray(desiredServiceClusterReferences))
}

type object interface {
	metav1.Object
	runtime.Object
}

func offeringsToObjectArray(offerings []catalogv1alpha1.Offering) []object {
	out := make([]object, len(offerings))
	for i := range offerings {
		out[i] = &offerings[i]
	}
	return out
}

func providerReferencesToObjectArray(providerReferences []catalogv1alpha1.ProviderReference) []object {
	out := make([]object, len(providerReferences))
	for i := range providerReferences {
		out[i] = &providerReferences[i]
	}
	return out
}

func serviceClusterReferencesToObjectArray(serviceClusterReferences []catalogv1alpha1.ServiceClusterReference) []object {
	out := make([]object, len(serviceClusterReferences))
	for i := range serviceClusterReferences {
		out[i] = &serviceClusterReferences[i]
	}
	return out
}

func (r *CatalogReconciler) cleanupOutdatedReferences(
	ctx context.Context, log logr.Logger,
	catalog *catalogv1alpha1.Catalog,
	foundObjects []object,
	desiredObjects []object,
) (deletedObjectCounter int, err error) {
	desiredObjectMap := map[string]struct{}{}
	for _, desiredObject := range desiredObjects {
		desiredObjectMap[types.NamespacedName{
			Name:      desiredObject.GetName(),
			Namespace: desiredObject.GetNamespace(),
		}.String()] = struct{}{}
	}

	// Delete Objects that are no longer in the desiredObjects list
	for _, foundObject := range foundObjects {
		if _, present := desiredObjectMap[types.NamespacedName{
			Name:      foundObject.GetName(),
			Namespace: foundObject.GetNamespace(),
		}.String()]; present {
			continue
		}
		ownerReferenceChanged, err := util.DeleteOwnerReference(catalog, foundObject, r.Scheme)
		if err != nil {
			return 0, fmt.Errorf("deleting OwnerReference: %w", err)
		}

		unowned, err := util.IsUnowned(foundObject)
		if err != nil {
			return 0, fmt.Errorf("checking object isUnowned: %w", err)
		}

		switch {
		case unowned:
			// The Object object is unowned by any Catalog objects, it can be removed.
			if err := r.Delete(ctx, foundObject); err != nil && !errors.IsNotFound(err) {
				return 0, fmt.Errorf("deleting Object: %w", err)
			}
			log.Info("deleting unowned Object", "Kind", foundObject.GetObjectKind().GroupVersionKind().Kind,
				"ObjectName", foundObject.GetName(),
				"ObjectNamespace", foundObject.GetNamespace())
			deletedObjectCounter++
		case !unowned && ownerReferenceChanged:
			if err := r.Update(ctx, foundObject); err != nil {
				return 0, fmt.Errorf("updating Object: %w", err)
			}
			log.Info("removing Catalog as owner from Object", "Kind", foundObject.GetObjectKind().GroupVersionKind().Kind,
				"ObjectName", foundObject.GetName(),
				"ObjectNamespace", foundObject.GetNamespace())
			deletedObjectCounter++
		}
	}
	return
}

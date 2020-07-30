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
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/kubermatic/utils/pkg/multiowner"
	"github.com/kubermatic/utils/pkg/util"

	catalogv1alpha1 "k8c.io/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "k8c.io/kubecarrier/pkg/apis/core/v1alpha1"
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

// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=catalogs,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=catalogs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=catalogentries,verbs=list;watch
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=offerings,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=regions,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=providers,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=kubecarrier.io,resources=serviceclusters,verbs=get;list;watch
// +kubebuilder:rbac:groups=kubecarrier.io,resources=serviceclusterassignments,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete;escalate;bind
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete

// Reconcile function reconciles the Catalog object which specified by the request. Currently, it does the following:
// - Fetch the Catalog object.
// - Handle the deletion of the Catalog object.
// - Fetch the CatalogEntries and Tenants that selected by this Catalog object.
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
	readyCatalogEntries, err := r.listSelectedReadyCatalogEntries(ctx, log, catalog)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("getting selected CatalogEntries: %w", err)
	}

	var entries []catalogv1alpha1.ObjectReference
	for _, catalogEntry := range readyCatalogEntries {
		entries = append(entries,
			catalogv1alpha1.ObjectReference{
				Name: catalogEntry.Name,
			})
	}
	catalog.Status.Entries = entries

	// Get Tenants.
	readyTenants, err := r.listSelectedReadyTenants(ctx, log, catalog)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("getting selected Tenants: %w", err)
	}

	var tenants []catalogv1alpha1.ObjectReference
	for _, tenant := range readyTenants {
		tenants = append(tenants,
			catalogv1alpha1.ObjectReference{
				Name: tenant.Name,
			})
	}
	catalog.Status.Tenants = tenants

	// First update the entries and tenants to the status.
	if err := r.updateStatus(ctx, catalog, nil); err != nil {
		return ctrl.Result{}, fmt.Errorf("updating Catalog Status: %w", err)
	}

	// Get Provider
	provider, err := catalogv1alpha1.GetAccountByAccountNamespace(ctx, r.Client, req.Namespace)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("getting Provider: %w", err)
	}

	var (
		desiredProviders                 []catalogv1alpha1.Provider
		desiredOfferings                 []catalogv1alpha1.Offering
		desiredRegions                   []catalogv1alpha1.Region
		desiredServiceClusterAssignments []corev1alpha1.ServiceClusterAssignment
		desiredRoles                     []rbacv1.Role
		desiredRoleBindings              []rbacv1.RoleBinding
	)
	for _, catalogEntry := range readyCatalogEntries {
		desiredProviderRoles, desiredProviderRoleBindings := r.buildDesiredProviderRolesAndRoleBindings(readyTenants, provider, catalogEntry)
		desiredTenantRoles, desiredTenantRoleBindings := r.buildDesiredTenantRolesAndRoleBindings(readyTenants, catalogEntry)
		desiredRoles = append(desiredRoles, desiredTenantRoles...)
		desiredRoles = append(desiredRoles, desiredProviderRoles...)
		desiredRoleBindings = append(desiredRoleBindings, desiredProviderRoleBindings...)
		desiredRoleBindings = append(desiredRoleBindings, desiredTenantRoleBindings...)
	}

	for _, tenant := range readyTenants {

		desiredProviders = append(desiredProviders, r.buildDesiredProvider(provider, tenant))
		desiredOfferings = append(desiredOfferings, r.buildDesiredOfferings(provider, tenant, readyCatalogEntries)...)
		desiredRegionsForTenant, desiredServiceClusterAssignmentsForTenant, err := r.buildDesiredRegionsAndAssignments(ctx, log, provider, tenant, readyCatalogEntries)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("building RegionAndAssignment: %w", err)
		}
		desiredRegions = append(desiredRegions, desiredRegionsForTenant...)
		desiredServiceClusterAssignments = append(desiredServiceClusterAssignments, desiredServiceClusterAssignmentsForTenant...)
	}

	if err := r.reconcileProviders(ctx, log, catalog, desiredProviders); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconcliing Providers: %w", err)
	}

	if err := r.reconcileOfferings(ctx, log, catalog, desiredOfferings); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconcliing Offerings: %w", err)
	}

	if err := r.reconcileRegions(ctx, log, catalog, desiredRegions); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconcliing Regions: %w", err)
	}

	if err := r.reconcileServiceClusterAssignments(ctx, log, catalog, desiredServiceClusterAssignments); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconcliing ServiceClusterAssignments: %w", err)
	}

	if err := r.reconcileRoles(ctx, log, catalog, desiredRoles); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconciling Roles: %w", err)
	}

	if err := r.reconcileRoleBindings(ctx, log, catalog, desiredRoleBindings); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconciling RoleBindings: %w", err)
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
	enqueuerForOwner := multiowner.EnqueueRequestForOwner(&catalogv1alpha1.Catalog{}, mgr.GetScheme())
	return ctrl.NewControllerManagedBy(mgr).
		For(&catalogv1alpha1.Catalog{}).
		Watches(&source.Kind{Type: &catalogv1alpha1.Tenant{}}, enqueueAllCatalogsInNamespace).
		Watches(&source.Kind{Type: &catalogv1alpha1.CatalogEntry{}}, enqueueAllCatalogsInNamespace).
		Watches(&source.Kind{Type: &corev1alpha1.ServiceCluster{}}, enqueueAllCatalogsInNamespace).
		Watches(&source.Kind{Type: &catalogv1alpha1.Offering{}}, enqueuerForOwner).
		Watches(&source.Kind{Type: &catalogv1alpha1.Provider{}}, enqueuerForOwner).
		Watches(&source.Kind{Type: &catalogv1alpha1.Region{}}, enqueuerForOwner).
		Watches(&source.Kind{Type: &corev1alpha1.ServiceClusterAssignment{}}, enqueuerForOwner).
		Watches(&source.Kind{Type: &rbacv1.Role{}}, enqueuerForOwner).
		Watches(&source.Kind{Type: &rbacv1.RoleBinding{}}, enqueuerForOwner).
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
	deletedProvidersCounter, err := r.cleanupProviders(ctx, log, catalog, nil)
	if err != nil {
		return fmt.Errorf("cleaning up Providers: %w", err)
	}
	deletedRegionsCounter, err := r.cleanupRegions(ctx, log, catalog, nil)
	if err != nil {
		return fmt.Errorf("cleaning up Regions: %w", err)
	}
	deletedServiceClusterAssignmentsCounter, err := r.cleanupServiceClusterAssignments(ctx, log, catalog, nil)
	if err != nil {
		return fmt.Errorf("cleaning up Regions: %w", err)
	}
	deletedRolesCounter, err := r.cleanupRoles(ctx, log, catalog, nil)
	if err != nil {
		return fmt.Errorf("cleaning up Roles: %w", err)
	}
	deletedRoleBindingsCounter, err := r.cleanupRoleBindings(ctx, log, catalog, nil)
	if err != nil {
		return fmt.Errorf("cleaning up RoleBindings: %w", err)
	}

	if deletedOfferingsCounter != 0 ||
		deletedProvidersCounter != 0 ||
		deletedRegionsCounter != 0 ||
		deletedServiceClusterAssignmentsCounter != 0 ||
		deletedRolesCounter != 0 ||
		deletedRoleBindingsCounter != 0 {
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

func (r *CatalogReconciler) listSelectedReadyCatalogEntries(ctx context.Context, log logr.Logger, catalog *catalogv1alpha1.Catalog) ([]catalogv1alpha1.CatalogEntry, error) {
	catalogEntrySelector, err := metav1.LabelSelectorAsSelector(catalog.Spec.CatalogEntrySelector)
	if err != nil {
		return nil, fmt.Errorf("parsing CatalogEntry selector: %w", err)
	}
	catalogEntries := &catalogv1alpha1.CatalogEntryList{}
	if err := r.List(ctx, catalogEntries, client.InNamespace(catalog.Namespace), client.MatchingLabelsSelector{Selector: catalogEntrySelector}); err != nil {
		return nil, fmt.Errorf("listing CatalogEntry: %w", err)
	}
	var readyCatalogEntries []catalogv1alpha1.CatalogEntry
	for _, catalogEntry := range catalogEntries.Items {
		if catalogEntry.IsReady() {
			readyCatalogEntries = append(readyCatalogEntries, catalogEntry)
		}
	}
	return readyCatalogEntries, nil
}

func (r *CatalogReconciler) listSelectedReadyTenants(ctx context.Context, log logr.Logger, catalog *catalogv1alpha1.Catalog) ([]*catalogv1alpha1.Account, error) {
	tenantSelector, err := metav1.LabelSelectorAsSelector(catalog.Spec.TenantSelector)
	if err != nil {
		return nil, fmt.Errorf("parsing Tenant selector: %w", err)
	}
	tenants := &catalogv1alpha1.TenantList{}
	if err := r.List(ctx, tenants, client.InNamespace(catalog.Namespace), client.MatchingLabelsSelector{Selector: tenantSelector}); err != nil {
		return nil, fmt.Errorf("listing Tenant: %w", err)
	}
	var readyTenants []*catalogv1alpha1.Account
	for _, tenant := range tenants.Items {
		tenantAccount := &catalogv1alpha1.Account{}
		if err := r.Get(ctx, types.NamespacedName{
			Name: tenant.Name,
		}, tenantAccount); err != nil {
			return nil, fmt.Errorf("getting Tenant: %w", err)
		}

		if tenantAccount.IsReady() && tenantAccount.HasRole(catalogv1alpha1.TenantRole) {
			readyTenants = append(readyTenants, tenantAccount)
		}
	}
	return readyTenants, nil
}

func (r *CatalogReconciler) buildDesiredOfferings(
	provider *catalogv1alpha1.Account,
	tenant *catalogv1alpha1.Account,
	catalogEntries []catalogv1alpha1.CatalogEntry,
) []catalogv1alpha1.Offering {
	var desiredOfferings []catalogv1alpha1.Offering
	for _, catalogEntry := range catalogEntries {
		desiredOfferings = append(desiredOfferings, catalogv1alpha1.Offering{
			ObjectMeta: metav1.ObjectMeta{
				Name:      catalogEntry.Status.TenantCRD.Name,
				Namespace: tenant.Status.Namespace.Name,
			},
			Spec: catalogv1alpha1.OfferingSpec{
				Metadata: catalogv1alpha1.OfferingMetadata{
					CommonMetadata: catalogEntry.Spec.Metadata.CommonMetadata,
				},
				Provider: catalogv1alpha1.ObjectReference{
					Name: provider.Name,
				},
				CRD: *catalogEntry.Status.TenantCRD,
			},
		})
	}
	return desiredOfferings
}

func (r *CatalogReconciler) buildDesiredProvider(
	provider *catalogv1alpha1.Account,
	tenant *catalogv1alpha1.Account,
) catalogv1alpha1.Provider {
	return catalogv1alpha1.Provider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      provider.Name,
			Namespace: tenant.Status.Namespace.Name,
		},
		Spec: catalogv1alpha1.ProviderSpec{
			Metadata: provider.Spec.Metadata,
		},
	}
}

func (r *CatalogReconciler) buildDesiredRegionsAndAssignments(
	ctx context.Context, log logr.Logger,
	provider *catalogv1alpha1.Account,
	tenant *catalogv1alpha1.Account,
	catalogEntries []catalogv1alpha1.CatalogEntry,
) ([]catalogv1alpha1.Region, []corev1alpha1.ServiceClusterAssignment, error) {
	var desiredRegions []catalogv1alpha1.Region
	var desiredServiceClusterAssignments []corev1alpha1.ServiceClusterAssignment
	serviceClusterNames := map[string]struct{}{}
	for _, catalogEntry := range catalogEntries {
		serviceClusterNames[catalogEntry.Status.TenantCRD.Region.Name] = struct{}{}
	}
	for serviceClusterName := range serviceClusterNames {
		serviceCluster := &corev1alpha1.ServiceCluster{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      serviceClusterName,
			Namespace: provider.Status.Namespace.Name,
		}, serviceCluster); err != nil {
			return nil, nil, fmt.Errorf("getting ServiceCluster: %w", err)
		}
		desiredRegions = append(desiredRegions, catalogv1alpha1.Region{
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceClusterName + "." + provider.Name,
				Namespace: tenant.Status.Namespace.Name,
			},
			Spec: catalogv1alpha1.RegionSpec{
				Metadata: serviceCluster.Spec.Metadata,
				Provider: catalogv1alpha1.ObjectReference{
					Name: provider.Name,
				},
			},
		})

		desiredServiceClusterAssignments = append(desiredServiceClusterAssignments, corev1alpha1.ServiceClusterAssignment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tenant.Status.Namespace.Name + "." + serviceClusterName,
				Namespace: provider.Status.Namespace.Name,
			},
			Spec: corev1alpha1.ServiceClusterAssignmentSpec{
				ServiceCluster: corev1alpha1.ObjectReference{
					Name: serviceClusterName,
				},
				ManagementClusterNamespace: corev1alpha1.ObjectReference{
					Name: tenant.Status.Namespace.Name,
				},
			},
		})
	}
	return desiredRegions, desiredServiceClusterAssignments, nil
}

func (r *CatalogReconciler) buildDesiredTenantRolesAndRoleBindings(
	tenants []*catalogv1alpha1.Account,
	catalogEntry catalogv1alpha1.CatalogEntry,
) ([]rbacv1.Role, []rbacv1.RoleBinding) {
	var desiredRoles []rbacv1.Role
	var desiredRoleBindings []rbacv1.RoleBinding
	tenantCRDInfo := catalogEntry.Status.TenantCRD
	for _, tenant := range tenants {
		role := rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("kubecarrier:tenant:%s", catalogEntry.Name),
				Namespace: tenant.Status.Namespace.Name,
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{tenantCRDInfo.APIGroup},
					Resources: []string{tenantCRDInfo.Plural},
					Verbs:     []string{rbacv1.VerbAll},
				},
			},
		}
		desiredRoles = append(desiredRoles, role)

		roleBinding := rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("kubecarrier:tenant:%s", catalogEntry.Name),
				Namespace: tenant.Status.Namespace.Name,
			},
			Subjects: tenant.Spec.Subjects,
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Role",
				Name:     role.Name,
			},
		}
		desiredRoleBindings = append(desiredRoleBindings, roleBinding)
	}
	return desiredRoles, desiredRoleBindings
}

func (r *CatalogReconciler) buildDesiredProviderRolesAndRoleBindings(
	tenants []*catalogv1alpha1.Account,
	provider *catalogv1alpha1.Account,
	catalogEntry catalogv1alpha1.CatalogEntry,
) ([]rbacv1.Role, []rbacv1.RoleBinding) {
	var desiredRoles []rbacv1.Role
	var desiredRoleBindings []rbacv1.RoleBinding
	providerCRDInfo := catalogEntry.Status.ProviderCRD
	for _, tenant := range tenants {
		role := rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("kubecarrier:provider:%s", catalogEntry.Name),
				Namespace: tenant.Status.Namespace.Name,
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{providerCRDInfo.APIGroup},
					Resources: []string{providerCRDInfo.Plural},
					Verbs:     []string{rbacv1.VerbAll},
				},
			},
		}
		desiredRoles = append(desiredRoles, role)
		roleBinding := rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("kubecarrier:provider:%s", catalogEntry.Name),
				Namespace: tenant.Status.Namespace.Name,
			},
			Subjects: provider.Spec.Subjects,
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Role",
				Name:     role.Name,
			},
		}
		desiredRoleBindings = append(desiredRoleBindings, roleBinding)
	}
	return desiredRoles, desiredRoleBindings
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
		if _, err := multiowner.InsertOwnerReference(catalog, &desiredOffering, r.Scheme); err != nil {
			return fmt.Errorf("inserting OwnerReference: %w", err)
		}

		foundOffering := &catalogv1alpha1.Offering{}
		err := r.Get(ctx, types.NamespacedName{
			Name:      desiredOffering.Name,
			Namespace: desiredOffering.Namespace,
		}, foundOffering)
		if err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("getting Offering: %w", err)
		}
		if errors.IsNotFound(err) {
			// Create the Offering.
			if err := r.Create(ctx, &desiredOffering); err != nil {
				return fmt.Errorf("creating Offering: %w", err)
			}
			foundOffering = &desiredOffering
		}

		ownerChanged, err := multiowner.InsertOwnerReference(catalog, foundOffering, r.Scheme)
		if err != nil {
			return fmt.Errorf("inserting OwnerReference: %w", err)
		}
		if !reflect.DeepEqual(desiredOffering.Spec, foundOffering.Spec) || ownerChanged {
			foundOffering.Spec = desiredOffering.Spec
			if err := r.Update(ctx, foundOffering); err != nil {
				return fmt.Errorf("updaing Offering: %w", err)
			}
		}
	}
	return nil
}

func (r *CatalogReconciler) reconcileProviders(
	ctx context.Context, log logr.Logger,
	catalog *catalogv1alpha1.Catalog,
	desiredProviders []catalogv1alpha1.Provider,
) error {
	if _, err := r.cleanupProviders(ctx, log, catalog, desiredProviders); err != nil {
		return fmt.Errorf("cleanup Provider: %w", err)
	}

	for _, desiredProvider := range desiredProviders {
		if _, err := multiowner.InsertOwnerReference(catalog, &desiredProvider, r.Scheme); err != nil {
			return fmt.Errorf("inserting OwnerReference: %w", err)
		}

		foundProvider := &catalogv1alpha1.Provider{}
		err := r.Get(ctx, types.NamespacedName{
			Name:      desiredProvider.Name,
			Namespace: desiredProvider.Namespace,
		}, foundProvider)
		if err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("getting Provider: %w", err)
		}
		if errors.IsNotFound(err) {
			// Create the Provider.
			if err := r.Create(ctx, &desiredProvider); err != nil {
				return fmt.Errorf("creating Provider: %w", err)
			}
			foundProvider = &desiredProvider
		}

		ownerChanged, err := multiowner.InsertOwnerReference(catalog, foundProvider, r.Scheme)
		if err != nil {
			return fmt.Errorf("inserting OwnerReference: %w", err)
		}
		if !reflect.DeepEqual(desiredProvider.Spec, foundProvider.Spec) || ownerChanged {
			foundProvider.Spec = desiredProvider.Spec
			if err := r.Update(ctx, foundProvider); err != nil {
				return fmt.Errorf("updaing Provider: %w", err)
			}
		}
	}
	return nil
}

func (r *CatalogReconciler) reconcileRegions(
	ctx context.Context, log logr.Logger,
	catalog *catalogv1alpha1.Catalog,
	desiredRegions []catalogv1alpha1.Region,
) error {
	if _, err := r.cleanupRegions(ctx, log, catalog, desiredRegions); err != nil {
		return fmt.Errorf("cleanup Region: %w", err)
	}

	for _, desiredRegion := range desiredRegions {
		if _, err := multiowner.InsertOwnerReference(catalog, &desiredRegion, r.Scheme); err != nil {
			return fmt.Errorf("inserting OwnerReference: %w", err)
		}

		foundRegion := &catalogv1alpha1.Region{}
		err := r.Get(ctx, types.NamespacedName{
			Name:      desiredRegion.Name,
			Namespace: desiredRegion.Namespace,
		}, foundRegion)
		if err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("getting Region: %w", err)
		}
		if errors.IsNotFound(err) {
			// Create the Region.
			if err := r.Create(ctx, &desiredRegion); err != nil {
				return fmt.Errorf("creating Region: %w", err)
			}
			foundRegion = &desiredRegion
		}

		ownerChanged, err := multiowner.InsertOwnerReference(catalog, foundRegion, r.Scheme)
		if err != nil {
			return fmt.Errorf("inserting OwnerReference: %w", err)
		}
		if !reflect.DeepEqual(desiredRegion.Spec, foundRegion.Spec) || ownerChanged {
			foundRegion.Spec = desiredRegion.Spec
			if err := r.Update(ctx, foundRegion); err != nil {
				return fmt.Errorf("updaing Region: %w", err)
			}
		}
	}
	return nil
}

func (r *CatalogReconciler) reconcileServiceClusterAssignments(
	ctx context.Context, log logr.Logger,
	catalog *catalogv1alpha1.Catalog,
	desiredServiceClusterAssignments []corev1alpha1.ServiceClusterAssignment,
) error {
	if _, err := r.cleanupServiceClusterAssignments(ctx, log, catalog, desiredServiceClusterAssignments); err != nil {
		return fmt.Errorf("cleanup ServiceClusterAssignment: %w", err)
	}

	var readyServiceClusterAssignmentsCounter int
	for _, desiredServiceClusterAssignment := range desiredServiceClusterAssignments {
		if _, err := multiowner.InsertOwnerReference(catalog, &desiredServiceClusterAssignment, r.Scheme); err != nil {
			return fmt.Errorf("inserting OwnerReference: %w", err)
		}

		foundServiceClusterAssignment := &corev1alpha1.ServiceClusterAssignment{}
		err := r.Get(ctx, types.NamespacedName{
			Name:      desiredServiceClusterAssignment.Name,
			Namespace: desiredServiceClusterAssignment.Namespace,
		}, foundServiceClusterAssignment)
		if err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("getting ServiceClusterAssignment: %w", err)
		}
		if errors.IsNotFound(err) {
			// Create the ServiceClusterAssignment.
			if err := r.Create(ctx, &desiredServiceClusterAssignment); err != nil {
				return fmt.Errorf("creating ServiceClusterAssignment: %w", err)
			}
			foundServiceClusterAssignment = &desiredServiceClusterAssignment
		}

		ownerChanged, err := multiowner.InsertOwnerReference(catalog, foundServiceClusterAssignment, r.Scheme)
		if err != nil {
			return fmt.Errorf("inserting OwnerReference: %w", err)
		}
		if !reflect.DeepEqual(desiredServiceClusterAssignment.Spec, foundServiceClusterAssignment.Spec) || ownerChanged {
			foundServiceClusterAssignment.Spec = desiredServiceClusterAssignment.Spec
			if err := r.Update(ctx, foundServiceClusterAssignment); err != nil {
				return fmt.Errorf("updaing ServiceClusterAssignment: %w", err)
			}
		}

		if foundServiceClusterAssignment.IsReady() {
			readyServiceClusterAssignmentsCounter++
		}
	}

	// Update AssignmentsReady Status.
	if readyServiceClusterAssignmentsCounter == len(desiredServiceClusterAssignments) {
		if err := r.updateStatus(ctx, catalog, &catalogv1alpha1.CatalogCondition{
			Type:    catalogv1alpha1.ServiceClusterAssignmentReady,
			Status:  catalogv1alpha1.ConditionTrue,
			Reason:  "ServiceClusterAssignmentsReady",
			Message: "All ServiceClusterAssignments are ready.",
		}); err != nil {
			return fmt.Errorf("updating Catalog AssignmentsReady Status: %w", err)
		}
	} else {
		if err := r.updateStatus(ctx, catalog, &catalogv1alpha1.CatalogCondition{
			Type:    catalogv1alpha1.ServiceClusterAssignmentReady,
			Status:  catalogv1alpha1.ConditionFalse,
			Reason:  "ServiceClusterAssignmentsUnready",
			Message: "ServiceClusterAssignments are not ready.",
		}); err != nil {
			return fmt.Errorf("updating Catalog AssignmentsReady Status: %w", err)
		}
	}

	return nil
}

func (r *CatalogReconciler) reconcileRoles(
	ctx context.Context, log logr.Logger,
	catalog *catalogv1alpha1.Catalog,
	desiredRoles []rbacv1.Role,
) error {
	if _, err := r.cleanupRoles(ctx, log, catalog, desiredRoles); err != nil {
		return fmt.Errorf("cleanup Role: %w", err)
	}

	for _, desiredRole := range desiredRoles {
		if _, err := multiowner.InsertOwnerReference(catalog, &desiredRole, r.Scheme); err != nil {
			return fmt.Errorf("inserting OwnerReference: %w", err)
		}

		foundRole := &rbacv1.Role{}
		err := r.Get(ctx, types.NamespacedName{
			Name:      desiredRole.Name,
			Namespace: desiredRole.Namespace,
		}, foundRole)
		if err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("getting Role: %w", err)
		}
		if errors.IsNotFound(err) {
			// Create the Role.
			if err := r.Create(ctx, &desiredRole); err != nil {
				return fmt.Errorf("creating Role: %w", err)
			}
			foundRole = &desiredRole
		}

		ownerChanged, err := multiowner.InsertOwnerReference(catalog, foundRole, r.Scheme)
		if err != nil {
			return fmt.Errorf("inserting OwnerReference: %w", err)
		}
		if !reflect.DeepEqual(desiredRole.Rules, foundRole.Rules) || ownerChanged {
			foundRole.Rules = desiredRole.Rules
			if err := r.Update(ctx, foundRole); err != nil {
				return fmt.Errorf("updaing Role: %w", err)
			}
		}
	}
	return nil
}

func (r *CatalogReconciler) reconcileRoleBindings(
	ctx context.Context, log logr.Logger,
	catalog *catalogv1alpha1.Catalog,
	desiredRoleBindings []rbacv1.RoleBinding,
) error {
	if _, err := r.cleanupRoleBindings(ctx, log, catalog, desiredRoleBindings); err != nil {
		return fmt.Errorf("cleanup RoleBinding: %w", err)
	}

	for _, desiredRoleBinding := range desiredRoleBindings {
		if _, err := multiowner.InsertOwnerReference(catalog, &desiredRoleBinding, r.Scheme); err != nil {
			return fmt.Errorf("inserting OwnerReference: %w", err)
		}

		foundRoleBinding := &rbacv1.RoleBinding{}
		err := r.Get(ctx, types.NamespacedName{
			Name:      desiredRoleBinding.Name,
			Namespace: desiredRoleBinding.Namespace,
		}, foundRoleBinding)
		if err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("getting RoleBinding: %w", err)
		}
		if errors.IsNotFound(err) {
			// Create the RoleBinding.
			if err := r.Create(ctx, &desiredRoleBinding); err != nil {
				return fmt.Errorf("creating RoleBinding: %w", err)
			}
			foundRoleBinding = &desiredRoleBinding
		}

		ownerChanged, err := multiowner.InsertOwnerReference(catalog, foundRoleBinding, r.Scheme)
		if err != nil {
			return fmt.Errorf("inserting OwnerReference: %w", err)
		}
		if !reflect.DeepEqual(desiredRoleBinding.Subjects, foundRoleBinding.Subjects) ||
			!reflect.DeepEqual(desiredRoleBinding.RoleRef, foundRoleBinding.RoleRef) ||
			ownerChanged {
			foundRoleBinding.Subjects = desiredRoleBinding.Subjects
			foundRoleBinding.RoleRef = desiredRoleBinding.RoleRef
			if err := r.Update(ctx, foundRoleBinding); err != nil {
				return fmt.Errorf("updaing RoleBinding: %w", err)
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
	if err := r.List(ctx, foundOfferingList, multiowner.OwnedBy(catalog, r.Scheme)); err != nil {
		return 0, fmt.Errorf("listing Offerings: %w", err)
	}
	return r.cleanupOutdatedReferences(ctx, log,
		catalog,
		offeringsToObjectArray(foundOfferingList.Items),
		offeringsToObjectArray(desiredOfferings))
}

func (r *CatalogReconciler) cleanupProviders(
	ctx context.Context, log logr.Logger,
	catalog *catalogv1alpha1.Catalog,
	desiredProviders []catalogv1alpha1.Provider,
) (deletedProviderCounter int, err error) {
	// Fetch existing Providers.
	foundProviderList := &catalogv1alpha1.ProviderList{}
	if err := r.List(ctx, foundProviderList, multiowner.OwnedBy(catalog, r.Scheme)); err != nil {
		return 0, fmt.Errorf("listing Providers: %w", err)
	}

	return r.cleanupOutdatedReferences(ctx, log,
		catalog,
		providersToObjectArray(foundProviderList.Items),
		providersToObjectArray(desiredProviders))
}

func (r *CatalogReconciler) cleanupRegions(
	ctx context.Context, log logr.Logger,
	catalog *catalogv1alpha1.Catalog,
	desiredRegions []catalogv1alpha1.Region,
) (deletedRegionCounter int, err error) {
	// Fetch existing Regions.
	foundRegionList := &catalogv1alpha1.RegionList{}
	if err := r.List(ctx, foundRegionList, multiowner.OwnedBy(catalog, r.Scheme)); err != nil {
		return 0, fmt.Errorf("listing Regions: %w", err)
	}
	return r.cleanupOutdatedReferences(ctx, log,
		catalog,
		regionsToObjectArray(foundRegionList.Items),
		regionsToObjectArray(desiredRegions))
}

func (r *CatalogReconciler) cleanupServiceClusterAssignments(
	ctx context.Context, log logr.Logger,
	catalog *catalogv1alpha1.Catalog,
	desiredServiceClusterAssignments []corev1alpha1.ServiceClusterAssignment,
) (deletedServiceClusterAssignmentCounter int, err error) {
	// Fetch existing ServiceClusterAssignments.
	foundServiceClusterAssignmentList := &corev1alpha1.ServiceClusterAssignmentList{}
	if err := r.List(ctx, foundServiceClusterAssignmentList, multiowner.OwnedBy(catalog, r.Scheme), client.InNamespace(catalog.Namespace)); err != nil {
		return 0, fmt.Errorf("listing ServiceClusterAssignments: %w", err)
	}
	return r.cleanupOutdatedReferences(ctx, log,
		catalog,
		serviceClusterAssignmentsToObjectArray(foundServiceClusterAssignmentList.Items),
		serviceClusterAssignmentsToObjectArray(desiredServiceClusterAssignments))
}

func (r *CatalogReconciler) cleanupRoles(
	ctx context.Context, log logr.Logger,
	catalog *catalogv1alpha1.Catalog,
	desiredRoles []rbacv1.Role,
) (deletedRoleCounter int, err error) {
	// Fetch existing Roles.
	foundRoleList := &rbacv1.RoleList{}
	if err := r.List(ctx, foundRoleList, multiowner.OwnedBy(catalog, r.Scheme)); err != nil {
		return 0, fmt.Errorf("listing Roles: %w", err)
	}
	return r.cleanupOutdatedReferences(ctx, log,
		catalog,
		rolesToObjectArray(foundRoleList.Items),
		rolesToObjectArray(desiredRoles))
}

func (r *CatalogReconciler) cleanupRoleBindings(
	ctx context.Context, log logr.Logger,
	catalog *catalogv1alpha1.Catalog,
	desiredRoleBindings []rbacv1.RoleBinding,
) (deletedRoleBindingCounter int, err error) {
	// Fetch existing RoleBindings.
	foundRoleBindingList := &rbacv1.RoleBindingList{}
	if err := r.List(ctx, foundRoleBindingList, multiowner.OwnedBy(catalog, r.Scheme)); err != nil {
		return 0, fmt.Errorf("listing RoleBindings: %w", err)
	}
	return r.cleanupOutdatedReferences(ctx, log,
		catalog,
		roleBindingsToObjectArray(foundRoleBindingList.Items),
		roleBindingsToObjectArray(desiredRoleBindings))
}

func offeringsToObjectArray(offerings []catalogv1alpha1.Offering) []object {
	out := make([]object, len(offerings))
	for i := range offerings {
		out[i] = &offerings[i]
	}
	return out
}

func providersToObjectArray(providers []catalogv1alpha1.Provider) []object {
	out := make([]object, len(providers))
	for i := range providers {
		out[i] = &providers[i]
	}
	return out
}

func regionsToObjectArray(regions []catalogv1alpha1.Region) []object {
	out := make([]object, len(regions))
	for i := range regions {
		out[i] = &regions[i]
	}
	return out
}

func serviceClusterAssignmentsToObjectArray(serviceClusterAssignments []corev1alpha1.ServiceClusterAssignment) []object {
	out := make([]object, len(serviceClusterAssignments))
	for i := range serviceClusterAssignments {
		out[i] = &serviceClusterAssignments[i]
	}
	return out
}

func rolesToObjectArray(roles []rbacv1.Role) []object {
	out := make([]object, len(roles))
	for i := range roles {
		out[i] = &roles[i]
	}
	return out
}

func roleBindingsToObjectArray(roleBindings []rbacv1.RoleBinding) []object {
	out := make([]object, len(roleBindings))
	for i := range roleBindings {
		out[i] = &roleBindings[i]
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
		ownerReferenceChanged, err := multiowner.DeleteOwnerReference(catalog, foundObject, r.Scheme)
		if err != nil {
			return 0, fmt.Errorf("deleting OwnerReference: %w", err)
		}

		owned, err := multiowner.IsOwned(foundObject)
		if err != nil {
			return 0, fmt.Errorf("checking object isUnowned: %w", err)
		}

		switch {
		case !owned:
			// The Object object is unowned by any Catalog objects, it can be removed.
			if err := r.Delete(ctx, foundObject); err != nil && !errors.IsNotFound(err) {
				return 0, fmt.Errorf("deleting Object: %w", err)
			}
			log.Info("deleting unowned Object", "Kind", foundObject.GetObjectKind().GroupVersionKind().Kind,
				"ObjectName", foundObject.GetName(),
				"ObjectNamespace", foundObject.GetNamespace())
			deletedObjectCounter++
		case owned && ownerReferenceChanged:
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

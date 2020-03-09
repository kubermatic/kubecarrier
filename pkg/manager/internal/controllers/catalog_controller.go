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
	"github.com/kubermatic/kubecarrier/pkg/internal/multiowner"
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

// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=catalogs,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=catalogs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=catalogentries,verbs=list;watch
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=offerings,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=serviceclusterreferences,verbs=get;list;watch;create;update;delete
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
		desiredProviders                 []runtime.Object
		desiredOfferings                 []runtime.Object
		desiredServiceClusterReferences  []runtime.Object
		desiredServiceClusterAssignments []runtime.Object
		desiredRoles                     []runtime.Object
		desiredRoleBindings              []runtime.Object
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
		desiredServiceClusterReferencesForTenant, desiredServiceClusterAssignmentsForTenant, err := r.buildDesiredServiceClusterReferencesAndAssignments(ctx, log, provider, tenant, readyCatalogEntries)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("building ServiceClusterReferenceAndAssignment: %w", err)
		}
		desiredServiceClusterReferences = append(desiredServiceClusterReferences, desiredServiceClusterReferencesForTenant...)
		desiredServiceClusterAssignments = append(desiredServiceClusterAssignments, desiredServiceClusterAssignmentsForTenant...)
	}

	if err := r.reconcileProviders(ctx, log, catalog, desiredProviders); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconcliing Providers: %w", err)
	}

	if err := r.reconcileOfferings(ctx, log, catalog, desiredOfferings); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconcliing Offerings: %w", err)
	}

	if err := r.reconcileServiceClusterReferences(ctx, log, catalog, desiredServiceClusterReferences); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconcliing ServiceClusterReferences: %w", err)
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
		Watches(&source.Kind{Type: &catalogv1alpha1.ServiceClusterReference{}}, enqueuerForOwner).
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

	cleanedUp, err := multiowner.DeleteOwnedObjects(ctx, log, r.Client, r.Scheme, catalog, []runtime.Object{
		&catalogv1alpha1.Offering{},
		&catalogv1alpha1.ServiceClusterReference{},
		&catalogv1alpha1.Provider{},
		&corev1alpha1.ServiceClusterAssignment{},
		&rbacv1.Role{},
		&rbacv1.RoleBinding{},
	})
	if err != nil {
		return fmt.Errorf("cleanning up owned objects: %w", err)
	}

	if cleanedUp && util.RemoveFinalizer(catalog, catalogControllerFinalizer) {
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
) (desiredOfferings []runtime.Object) {
	for _, catalogEntry := range catalogEntries {
		desiredOfferings = append(desiredOfferings, &catalogv1alpha1.Offering{
			ObjectMeta: metav1.ObjectMeta{
				Name:      catalogEntry.Name,
				Namespace: tenant.Status.Namespace.Name,
			},
			Offering: catalogv1alpha1.OfferingData{
				Metadata: catalogv1alpha1.OfferingMetadata{
					DisplayName: catalogEntry.Spec.Metadata.DisplayName,
					Description: catalogEntry.Spec.Metadata.Description,
				},
				Provider: catalogv1alpha1.ObjectReference{
					Name: provider.Name,
				},
				CRD: *catalogEntry.Status.TenantCRD,
			},
		})
	}
	return
}

func (r *CatalogReconciler) buildDesiredProvider(
	provider *catalogv1alpha1.Account,
	tenant *catalogv1alpha1.Account,
) (desiredProvider runtime.Object) {
	desiredProvider = &catalogv1alpha1.Provider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      provider.Name,
			Namespace: tenant.Status.Namespace.Name,
		},
		Spec: catalogv1alpha1.ProviderSpec{
			Metadata: provider.Spec.Metadata,
		},
	}
	return
}

func (r *CatalogReconciler) buildDesiredServiceClusterReferencesAndAssignments(
	ctx context.Context, log logr.Logger,
	provider *catalogv1alpha1.Account,
	tenant *catalogv1alpha1.Account,
	catalogEntries []catalogv1alpha1.CatalogEntry,
) (desiredServiceClusterReferences []runtime.Object, desiredServiceClusterAssignments []runtime.Object, err error) {
	serviceClusterNames := map[string]struct{}{}
	for _, catalogEntry := range catalogEntries {
		serviceClusterNames[catalogEntry.Status.TenantCRD.ServiceCluster.Name] = struct{}{}
	}
	for serviceClusterName := range serviceClusterNames {
		serviceCluster := &corev1alpha1.ServiceCluster{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      serviceClusterName,
			Namespace: provider.Status.Namespace.Name,
		}, serviceCluster); err != nil {
			return nil, nil, fmt.Errorf("getting ServiceCluster: %w", err)
		}
		desiredServiceClusterReferences = append(desiredServiceClusterReferences, &catalogv1alpha1.ServiceClusterReference{
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceClusterName + "." + provider.Name,
				Namespace: tenant.Status.Namespace.Name,
			},
			Spec: catalogv1alpha1.ServiceClusterReferenceSpec{
				Metadata: serviceCluster.Spec.Metadata,
				Provider: catalogv1alpha1.ObjectReference{
					Name: provider.Name,
				},
			},
		})

		desiredServiceClusterAssignments = append(desiredServiceClusterAssignments, &corev1alpha1.ServiceClusterAssignment{
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
	return
}

func (r *CatalogReconciler) buildDesiredTenantRolesAndRoleBindings(
	tenants []*catalogv1alpha1.Account,
	catalogEntry catalogv1alpha1.CatalogEntry,
) (desiredRoles []runtime.Object, desiredRoleBindings []runtime.Object) {
	tenantCRDInfo := catalogEntry.Status.TenantCRD
	for _, tenant := range tenants {
		role := &rbacv1.Role{
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

		roleBinding := &rbacv1.RoleBinding{
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
) (desiredRoles []runtime.Object, desiredRoleBindings []runtime.Object) {
	providerCRDInfo := catalogEntry.Status.ProviderCRD
	for _, tenant := range tenants {
		role := &rbacv1.Role{
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
		roleBinding := &rbacv1.RoleBinding{
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
	desiredOfferings []runtime.Object,
) error {
	return multiowner.ReconcileOwnedObjects(
		ctx, log,
		r.Client, r.Scheme,
		catalog,
		desiredOfferings, &catalogv1alpha1.Offering{},
		func(actual, desired runtime.Object) error {
			actualOffering := actual.(*catalogv1alpha1.Offering)
			desiredOffering := desired.(*catalogv1alpha1.Offering)
			if !reflect.DeepEqual(actualOffering.Offering, desiredOffering.Offering) {
				actualOffering.Offering = desiredOffering.Offering
			}
			return nil
		})
}

func (r *CatalogReconciler) reconcileProviders(
	ctx context.Context, log logr.Logger,
	catalog *catalogv1alpha1.Catalog,
	desiredProviders []runtime.Object,
) error {
	return multiowner.ReconcileOwnedObjects(
		ctx, log,
		r.Client, r.Scheme,
		catalog,
		desiredProviders, &catalogv1alpha1.Provider{},
		func(actual, desired runtime.Object) error {
			actualProvider := actual.(*catalogv1alpha1.Provider)
			desiredProvider := desired.(*catalogv1alpha1.Provider)
			if !reflect.DeepEqual(actualProvider.Spec, desiredProvider.Spec) {
				actualProvider.Spec = desiredProvider.Spec
			}
			return nil
		})
}

func (r *CatalogReconciler) reconcileServiceClusterReferences(
	ctx context.Context, log logr.Logger,
	catalog *catalogv1alpha1.Catalog,
	desiredServiceClusterReferences []runtime.Object,
) error {
	return multiowner.ReconcileOwnedObjects(
		ctx, log,
		r.Client, r.Scheme,
		catalog,
		desiredServiceClusterReferences, &catalogv1alpha1.ServiceClusterReference{},
		func(actual, desired runtime.Object) error {
			actualServiceClusterReference := actual.(*catalogv1alpha1.ServiceClusterReference)
			desiredServiceClusterReference := desired.(*catalogv1alpha1.ServiceClusterReference)
			if !reflect.DeepEqual(actualServiceClusterReference.Spec, desiredServiceClusterReference.Spec) {
				actualServiceClusterReference.Spec = desiredServiceClusterReference.Spec
			}
			return nil
		})
}

func (r *CatalogReconciler) reconcileServiceClusterAssignments(
	ctx context.Context, log logr.Logger,
	catalog *catalogv1alpha1.Catalog,
	desiredServiceClusterAssignments []runtime.Object,
) error {

	if err := multiowner.ReconcileOwnedObjects(
		ctx, log,
		r.Client, r.Scheme,
		catalog,
		desiredServiceClusterAssignments, &corev1alpha1.ServiceClusterAssignment{},
		func(actual, desired runtime.Object) error {
			actualServiceClusterAssignment := actual.(*corev1alpha1.ServiceClusterAssignment)
			desiredServiceClusterAssignment := desired.(*corev1alpha1.ServiceClusterAssignment)
			if !reflect.DeepEqual(desiredServiceClusterAssignment.Spec, desiredServiceClusterAssignment.Spec) {
				actualServiceClusterAssignment.Spec = desiredServiceClusterAssignment.Spec
			}
			return nil
		}); err != nil {
		return nil
	}

	var readyServiceClusterAssignmentsCounter int
	foundServiceClusterAssignmentList := &corev1alpha1.ServiceClusterAssignmentList{}
	if err := r.List(ctx, foundServiceClusterAssignmentList, multiowner.OwnedBy(catalog, r.Scheme), client.InNamespace(catalog.Namespace)); err != nil {
		return fmt.Errorf("listing ServiceClusterAssignments: %w", err)
	}
	for _, serviceClusterAssignment := range foundServiceClusterAssignmentList.Items {
		if serviceClusterAssignment.IsReady() {
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
	desiredRoles []runtime.Object,
) error {
	return multiowner.ReconcileOwnedObjects(
		ctx, log,
		r.Client, r.Scheme,
		catalog,
		desiredRoles, &rbacv1.Role{},
		func(actual, desired runtime.Object) error {
			actualRole := actual.(*rbacv1.Role)
			desiredRole := desired.(*rbacv1.Role)
			if !reflect.DeepEqual(actualRole.Rules, desiredRole.Rules) {
				actualRole.Rules = desiredRole.Rules
			}
			return nil
		})
}

func (r *CatalogReconciler) reconcileRoleBindings(
	ctx context.Context, log logr.Logger,
	catalog *catalogv1alpha1.Catalog,
	desiredRoleBindings []runtime.Object,
) error {
	return multiowner.ReconcileOwnedObjects(
		ctx, log,
		r.Client, r.Scheme,
		catalog,
		desiredRoleBindings, &rbacv1.RoleBinding{},
		func(actual, desired runtime.Object) error {
			actualRoleBinding := actual.(*rbacv1.RoleBinding)
			desiredRoleBinding := desired.(*rbacv1.RoleBinding)
			if !reflect.DeepEqual(actualRoleBinding.Subjects, desiredRoleBinding.Subjects) ||
				!reflect.DeepEqual(actualRoleBinding.RoleRef, desiredRoleBinding.RoleRef) {
				actualRoleBinding.Subjects = desiredRoleBinding.Subjects
				actualRoleBinding.RoleRef = desiredRoleBinding.RoleRef
			}
			return nil
		})
}

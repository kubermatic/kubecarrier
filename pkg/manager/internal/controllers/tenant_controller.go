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
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
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
	"github.com/kubermatic/kubecarrier/pkg/internal/owner"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

const (
	tenantControllerFinalizer string = "tenant.kubecarrier.io/controller"
)

// TenantReconciler reconciles a Tenant object
type TenantReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=tenants,verbs=get;list;watch;update;
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=tenants/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=tenantreferences,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete

// Reconcile function reconciles the Tenant object which specified by the request. Currently, it does the following:
// 1. Fetch the Tenant object.
// 2. Handle the deletion of the Tenant object (Remove the namespace that the tenant owns, and remove the finalizer).
// 3. Handle the creation/update of the Tenant object (Create/reconcile the namespace and TenantReferences and insert the finalizer).
// 4. Update the status of the Tenant object.
func (r *TenantReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("tenant", req.NamespacedName)

	// 1. Fetch the Tenant object.
	tenant := &catalogv1alpha1.Tenant{}
	if err := r.Get(ctx, req.NamespacedName, tenant); err != nil {
		// If the Tenant object is already gone, we just ignore the NotFound error.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 2. Handle the deletion of the Tenant object.
	if !tenant.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, log, tenant); err != nil {
			return ctrl.Result{}, fmt.Errorf("handling deletion: %w", err)
		}
		return ctrl.Result{}, nil
	}

	// 3. reconcile the Tenant object.
	// check/add the finalizer for the Tenant
	if util.AddFinalizer(tenant, tenantControllerFinalizer) {
		// Update the tenant with the finalizer
		if err := r.Update(ctx, tenant); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating finalizers: %w", err)
		}
	}

	// check/update the NamespaceName
	if tenant.Status.NamespaceName == "" {
		tenant.Status.NamespaceName = strings.Replace(tenant.Name, ".", "-", -1)
		if err := r.Status().Update(ctx, tenant); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating NamespaceName: %w", err)
		}
	}

	// Reconcile the namespace for the tenant
	if err := r.reconcileNamespace(ctx, log, tenant); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconciling namespace: %w", err)
	}

	// Reconcile the TenantReferences for the tenant
	if tenant.IsReady() {
		if err := r.reconcileTenantReferences(ctx, log, tenant); err != nil {
			return ctrl.Result{}, fmt.Errorf("reconciling TenantReferences: %w", err)
		}
	}

	return ctrl.Result{}, nil
}

func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	enqueuer := owner.EnqueueRequestForOwner(&catalogv1alpha1.Tenant{}, r.Scheme)

	return ctrl.NewControllerManagedBy(mgr).
		For(&catalogv1alpha1.Tenant{}).
		Watches(&source.Kind{Type: &corev1.Namespace{}}, enqueuer).
		Watches(&source.Kind{Type: &catalogv1alpha1.TenantReference{}}, enqueuer).
		Watches(&source.Kind{Type: &catalogv1alpha1.Provider{}}, &handler.EnqueueRequestsFromMapFunc{
			ToRequests: handler.ToRequestsFunc(func(mapObject handler.MapObject) (out []reconcile.Request) {
				tenants := &catalogv1alpha1.TenantList{}
				if err := r.List(context.Background(), tenants, client.InNamespace(mapObject.Meta.GetNamespace())); err != nil {
					// This will makes the manager crashes, and it will restart and reconcile all objects again.
					panic(fmt.Errorf("listting tenant: %w", err))
				}
				for _, t := range tenants.Items {
					out = append(out, reconcile.Request{
						NamespacedName: types.NamespacedName{
							Name:      t.Name,
							Namespace: t.Namespace,
						},
					})
				}
				return
			}),
		}).
		Complete(r)
}

// handleDeletion handles the deletion of the Tenant object. Currently, it does:
// 1. Delete the TenantReferences that this tenant owns.
// 2. Delete the Namespace that the tenant owns.
// 3. Remove the finalizer from the tenant object.
func (r *TenantReconciler) handleDeletion(ctx context.Context, log logr.Logger, tenant *catalogv1alpha1.Tenant) error {
	// Update the Tenant Status to Terminating.
	readyCondition, _ := tenant.Status.GetCondition(catalogv1alpha1.TenantReady)
	if readyCondition.Status != catalogv1alpha1.ConditionFalse ||
		readyCondition.Status == catalogv1alpha1.ConditionFalse && readyCondition.Reason != catalogv1alpha1.TenantTerminatingReason {
		tenant.Status.ObservedGeneration = tenant.Generation
		tenant.Status.SetCondition(catalogv1alpha1.TenantCondition{
			Type:    catalogv1alpha1.TenantReady,
			Status:  catalogv1alpha1.ConditionFalse,
			Reason:  catalogv1alpha1.TenantTerminatingReason,
			Message: "Tenant is being terminated",
		})
		if err := r.Status().Update(ctx, tenant); err != nil {
			return fmt.Errorf("updating Tenant status: %w", err)
		}
	}

	// 1. Clean up TenantReferences.
	providerList := &catalogv1alpha1.ProviderList{}
	if err := r.List(ctx, providerList); err != nil {
		return fmt.Errorf("listing Providers: %w", err)
	}

	var tenantReferenceCleanedUp int
	for _, provider := range providerList.Items {

		tenantReference := &catalogv1alpha1.TenantReference{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tenant.Name,
				Namespace: provider.Status.NamespaceName,
			},
		}

		if err := r.Delete(ctx, tenantReference); err != nil {
			if errors.IsNotFound(err) {
				tenantReferenceCleanedUp++
				continue
			}
			return fmt.Errorf("deleting TenantReference: %w", err)
		}
	}

	if tenantReferenceCleanedUp != len(providerList.Items) {
		// Not everything has been cleaned up yet
		// move to the next reconcile call
		return nil
	}

	// 2. Cleanup Namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: tenant.Status.NamespaceName,
		},
	}
	err := r.Get(ctx, types.NamespacedName{
		Name: tenant.Status.NamespaceName,
	}, ns)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("getting Namespace: %w", err)
	}

	if err == nil && ns.DeletionTimestamp.IsZero() {
		if err = r.Delete(ctx, ns); err != nil {
			return fmt.Errorf("deleting Namespace: %w", err)
		}
		// Move to the next reconcile round
		return nil
	}

	if !errors.IsNotFound(err) {
		// Namespace is not yet gone
		return nil
	}

	// 3. The Namespace and TenantReferences are completely removed, then we remove the finalizer here.
	if util.RemoveFinalizer(tenant, tenantControllerFinalizer) {
		if err := r.Update(ctx, tenant); err != nil {
			return fmt.Errorf("updating Tenant Status: %w", err)
		}
	}
	return nil
}

func (r *TenantReconciler) reconcileNamespace(ctx context.Context, log logr.Logger, tenant *catalogv1alpha1.Tenant) error {
	ns := &corev1.Namespace{}
	err := r.Get(ctx, types.NamespacedName{Name: tenant.Status.NamespaceName}, ns)
	if err == nil {
		// No error from the Get, so we update the Tenant Status.
		if readyCondition, _ := tenant.Status.GetCondition(catalogv1alpha1.TenantReady); readyCondition.Status != catalogv1alpha1.ConditionTrue {
			// Update Tenant Status
			tenant.Status.ObservedGeneration = tenant.Generation
			tenant.Status.SetCondition(catalogv1alpha1.TenantCondition{
				Type:    catalogv1alpha1.TenantReady,
				Status:  catalogv1alpha1.ConditionTrue,
				Reason:  "SetupComplete",
				Message: "Tenant setup is complete.",
			})
			if err := r.Status().Update(ctx, tenant); err != nil {
				return fmt.Errorf("updating Tenant status: %w", err)
			}
		}
		return nil

	}
	if !errors.IsNotFound(err) {
		return fmt.Errorf("getting Tenant namepsace: %w", err)
	}

	// When the namespace is not there, we need to make sure to update our Status first,
	// so the rest of the system can act on it.
	// This is especially important for Reconcilations that are prone to errors, or take a long time.
	// In this case it's most likely overkill, but still strictly necessary.
	if readyCondition, _ := tenant.Status.GetCondition(catalogv1alpha1.TenantReady); readyCondition.Status != catalogv1alpha1.ConditionFalse {
		tenant.Status.ObservedGeneration = tenant.Generation
		tenant.Status.SetCondition(catalogv1alpha1.TenantCondition{
			Type:    catalogv1alpha1.TenantReady,
			Status:  catalogv1alpha1.ConditionFalse,
			Reason:  "SetupIncomplete",
			Message: "Tenant setup is incomplete, namespace is missing.",
		})

		if err = r.Status().Update(ctx, tenant); err != nil {
			return fmt.Errorf("updating Tenant status: %w", err)
		}

		// move to the next reconcile round
		return nil
	}

	ns.Name = tenant.Status.NamespaceName
	owner.SetOwnerReference(tenant, ns, r.Scheme)

	// Reconcile the namespace
	if err = r.Create(ctx, ns); err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("creating Tenant namespace: %w", err)
	}
	return nil
}

func (r *TenantReconciler) reconcileTenantReferences(ctx context.Context, log logr.Logger, tenant *catalogv1alpha1.Tenant) error {
	providerList := &catalogv1alpha1.ProviderList{}
	if err := r.List(ctx, providerList); err != nil {
		return fmt.Errorf("listing Providers: %w", err)
	}

	for _, provider := range providerList.Items {
		if condition, _ := provider.Status.GetCondition(catalogv1alpha1.ProviderReady); condition.Status != catalogv1alpha1.ConditionTrue {
			// skip NotReady providers.
			continue
		}

		tenantReference := &catalogv1alpha1.TenantReference{}
		err := r.Get(ctx, types.NamespacedName{
			Name:      tenant.Name,
			Namespace: provider.Status.NamespaceName,
		}, tenantReference)
		if err == nil {
			// No error from the Get, so TenantReference has already been created.
			continue
		}
		if !errors.IsNotFound(err) {
			return fmt.Errorf("getting TenantReference: %w", err)
		}

		// The TenantReference is not found, then we just create it.
		tenantReference.Name = tenant.Name
		tenantReference.Namespace = provider.Status.NamespaceName
		owner.SetOwnerReference(tenant, tenantReference, r.Scheme)

		// Create the TenantReference
		if err = r.Create(ctx, tenantReference); err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("creating TenantReference: %w", err)
		}
	}
	return nil
}

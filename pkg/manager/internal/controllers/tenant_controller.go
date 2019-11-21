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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

const (
	tenantControllerFinalizer string = "tenant.kubecarrier.io/controller"
)

// TenantReconciler reconciles a Tenant object
type TenantReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=tenants,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=tenants/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete

// Reconcile function reconciles the Tenant object which specified by the request. Currently, it does the following:
// 1. Fetch the Tenant object.
// 2. Handle the deletion of the Tenant object (Remove the namespace that the tenant owns, and remove the finalizer).
// 3. Handle the creation/update of the Tenant object (Create/reconcile the namespace and insert the finalizer).
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
		tenant.Status.NamespaceName = fmt.Sprintf("tenant-%s", tenant.Name)
		if err := r.Update(ctx, tenant); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating NamespaceName: %w", err)
		}
	}

	// Reconcile the namespace for the tenant
	if err := r.reconcileNamespace(ctx, log, tenant); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconciling namespace: %w", err)
	}

	return ctrl.Result{}, nil
}

func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	owner := &catalogv1alpha1.Tenant{}
	enqueuer, err := util.EnqueueRequestForOwner(owner, mgr.GetScheme())
	if err != nil {
		return fmt.Errorf("cannot create enqueuer for Tenant: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&catalogv1alpha1.Tenant{}).
		Watches(&source.Kind{Type: &corev1.Namespace{}}, enqueuer).
		Complete(r)
}

// handleDeletion handles the deletion of the Tenant object. Currently, it does:
// 1. Delete the Namespace that the tenant owns.
// 2. Remove the finalizer from the tenant object.
func (r *TenantReconciler) handleDeletion(ctx context.Context, log logr.Logger, tenant *catalogv1alpha1.Tenant) error {

	// 1. Delete the Namespace.
	namespace, namespaceFound, err := r.getTenantNamespace(ctx, tenant)
	if err != nil {
		return fmt.Errorf("getting Tenant namespace: %w", err)
	}

	if namespaceFound {
		// If the namespace is found, then it need to be deleted.
		if err := r.Delete(ctx, namespace); err != nil {
			return fmt.Errorf("deleting Namespace: %w", err)
		}
		// Move to the next reconcile round
		return nil
	}
	// 2. The Namespace is completely removed, then we remove the finalizer here.
	if util.RemoveFinalizer(tenant, tenantControllerFinalizer) {
		if err := r.Update(ctx, tenant); err != nil {
			return fmt.Errorf("updating Tenant Status: %w", err)
		}
	}
	return nil
}

func (r *TenantReconciler) reconcileNamespace(ctx context.Context, log logr.Logger, tenant *catalogv1alpha1.Tenant) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: tenant.Status.NamespaceName,
		},
	}

	if _, err := util.InsertOwnerReference(tenant, ns, r.Scheme); err != nil {
		return fmt.Errorf("setting cross-namespaceed owner reference: %w", err)
	}

	_, namespaceFound, err := r.getTenantNamespace(ctx, tenant)
	if err != nil {
		return fmt.Errorf("getting Tenant namespace: %v", err)
	}

	if !namespaceFound {
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
		// Reconcile the namespace
		if err = r.Create(ctx, ns); err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("creating Tenant namespace: %w", err)
		}
	}

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

// getTenantNamespace is a helper function to perform the Namespace lookup.
func (r *TenantReconciler) getTenantNamespace(ctx context.Context, tenant *catalogv1alpha1.Tenant) (ns *corev1.Namespace, found bool, err error) {
	ns = &corev1.Namespace{}
	if err = r.Get(ctx, types.NamespacedName{Name: tenant.Status.NamespaceName}, ns); err != nil {
		if errors.IsNotFound(err) {
			return ns, false, nil
		}
		return nil, false, err
	}
	return ns, true, nil
}

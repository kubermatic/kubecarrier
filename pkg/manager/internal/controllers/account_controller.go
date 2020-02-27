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
	reconcile2 "github.com/kubermatic/kubecarrier/pkg/internal/reconcile"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

const (
	accountControllerFinalizer string = "account.kubecarrier.io/controller"
)

// AccountReconciler reconciles a Account object
type AccountReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=accounts,verbs=get;list;watch;update;
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=accounts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=tenantreferences,verbs=get;list;watch;create;update;delete

// Reconcile function reconciles the Account object which specified by the request. Currently, it does the following:
// 1. Fetch the Account object.
// 2. Handle the deletion of the Account object (Remove the namespace that the account owns, and remove the finalizer).
// 3. Handle the creation/update of the Account object (Create/reconcile the namespace and insert the finalizer).
// 4. Update the status of the Account object.
func (r *AccountReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("account", req.NamespacedName)

	account := &catalogv1alpha1.Account{}
	if err := r.Get(ctx, req.NamespacedName, account); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !account.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, log, account); err != nil {
			return ctrl.Result{}, fmt.Errorf("handling deletion: %w", err)
		}
		return ctrl.Result{}, nil
	}

	if util.AddFinalizer(account, accountControllerFinalizer) {
		if err := r.Update(ctx, account); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating finalizers: %w", err)
		}
	}
	if err := r.reconcileNamespace(ctx, log, account); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconciling namespace: %w", err)
	}
	if err := r.reconcileTenantReferences(ctx, log, account); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconciling tenant references: %w", err)
	}

	return ctrl.Result{}, nil
}

func (r *AccountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	enqueuer := owner.EnqueueRequestForOwner(&catalogv1alpha1.Account{}, mgr.GetScheme())

	return ctrl.NewControllerManagedBy(mgr).
		For(&catalogv1alpha1.Account{}).
		Watches(&source.Kind{Type: &corev1.Namespace{}}, enqueuer).
		Watches(&source.Kind{Type: &catalogv1alpha1.TenantReference{}}, enqueuer).
		Watches(&source.Kind{Type: &catalogv1alpha1.Account{}}, &handler.EnqueueRequestsFromMapFunc{
			ToRequests: handler.ToRequestsFunc(func(mapObject handler.MapObject) (out []reconcile.Request) {
				provider := mapObject.Object.(*catalogv1alpha1.Account)
				if !provider.HasRole(catalogv1alpha1.ProviderRole) {
					return
				}

				tenants := &catalogv1alpha1.AccountList{}
				if err := r.List(context.Background(), tenants, client.InNamespace(mapObject.Meta.GetNamespace())); err != nil {
					// This will makes the manager crashes, and it will restart and reconcile all objects again.
					panic(fmt.Errorf("listting accounts: %w", err))
				}
				for _, t := range tenants.Items {
					if t.HasRole(catalogv1alpha1.TenantRole) {
						out = append(out, reconcile.Request{
							NamespacedName: types.NamespacedName{
								Name:      t.Name,
								Namespace: t.Namespace,
							},
						})
					}
				}
				return
			}),
		}).
		Complete(r)
}

// handleDeletion handles the deletion of the Account object:
func (r *AccountReconciler) handleDeletion(ctx context.Context, log logr.Logger, account *catalogv1alpha1.Account) error {
	// Update the Account Status to Terminating.
	readyCondition, _ := account.Status.GetCondition(catalogv1alpha1.AccountReady)
	if readyCondition.Status != catalogv1alpha1.ConditionFalse ||
		readyCondition.Status == catalogv1alpha1.ConditionFalse && readyCondition.Reason != catalogv1alpha1.AccountTerminatingReason {
		account.Status.ObservedGeneration = account.Generation
		account.Status.SetCondition(catalogv1alpha1.AccountCondition{
			Type:    catalogv1alpha1.AccountReady,
			Status:  catalogv1alpha1.ConditionFalse,
			Reason:  catalogv1alpha1.AccountTerminatingReason,
			Message: "Account is being terminated",
		})
		if err := r.Status().Update(ctx, account); err != nil {
			return fmt.Errorf("updating Account status: %w", err)
		}
	}

	changed, err := (&reconcile2.OwnedObjectReconciler{
		Scheme:      r.Scheme,
		Log:         log,
		Owner:       account,
		WantedState: nil,
		TypeFilter: []runtime.Object{
			&corev1.Namespace{},
			&catalogv1alpha1.TenantReference{},
		},
	}).ReconcileOwnedObjects(ctx, r.Client)
	if err != nil {
		return fmt.Errorf("cannot reconcile objects: %w", err)
	}

	if !changed && util.RemoveFinalizer(account, accountControllerFinalizer) {
		if err := r.Update(ctx, account); err != nil {
			return fmt.Errorf("updating Account Status: %w", err)
		}
	}
	return nil
}

func (r *AccountReconciler) reconcileNamespace(ctx context.Context, log logr.Logger, account *catalogv1alpha1.Account) error {
	ns := &corev1.Namespace{}
	ns.Name = account.Name
	if _, err := (&reconcile2.OwnedObjectReconciler{
		Scheme: r.Scheme,
		Log:    log,
		Owner:  account,
		TypeFilter: []runtime.Object{
			&corev1.Namespace{},
		},
		WantedState: []util.Object{ns},
	}).ReconcileOwnedObjects(ctx, r.Client); err != nil {
		return fmt.Errorf("cannot reconcile namespace: %w", err)
	}

	if account.Status.Namespace.Name == "" {
		account.Status.Namespace.Name = ns.Name
		if err := r.Status().Update(ctx, account); err != nil {
			return fmt.Errorf("updating NamespaceName: %w", err)
		}
	}
	if readyCondition, _ := account.Status.GetCondition(catalogv1alpha1.AccountReady); readyCondition.Status != catalogv1alpha1.ConditionTrue {
		// Update Account Status
		account.Status.ObservedGeneration = account.Generation
		account.Status.SetCondition(catalogv1alpha1.AccountCondition{
			Type:    catalogv1alpha1.AccountReady,
			Status:  catalogv1alpha1.ConditionTrue,
			Reason:  "SetupComplete",
			Message: "Account setup is complete.",
		})
		if err := r.Status().Update(ctx, account); err != nil {
			return fmt.Errorf("updating Account status: %w", err)
		}
	}
	return nil
}

func (r *AccountReconciler) reconcileTenantReferences(ctx context.Context, log logr.Logger, account *catalogv1alpha1.Account) error {
	accountList := &catalogv1alpha1.AccountList{}
	if err := r.List(ctx, accountList); err != nil {
		return fmt.Errorf("listing Accounts: %w", err)
	}

	wantedRefs := make([]util.Object, 0)
	if account.HasRole(catalogv1alpha1.TenantRole) {
		for _, providerAccount := range accountList.Items {
			if !providerAccount.HasRole(catalogv1alpha1.ProviderRole) {
				continue
			}
			if condition, _ := providerAccount.Status.GetCondition(catalogv1alpha1.AccountReady); condition.Status != catalogv1alpha1.ConditionTrue {
				continue
			}
			tenantReference := &catalogv1alpha1.TenantReference{
				ObjectMeta: metav1.ObjectMeta{
					Name:      account.Name,
					Namespace: providerAccount.Status.Namespace.Name,
				},
			}
			owner.SetOwnerReference(account, tenantReference, r.Scheme)
			wantedRefs = append(wantedRefs, tenantReference)
		}
	}

	_, err := (&reconcile2.OwnedObjectReconciler{
		Scheme:      r.Scheme,
		Log:         log,
		Owner:       account,
		WantedState: wantedRefs,
		TypeFilter: []runtime.Object{
			&catalogv1alpha1.TenantReference{},
		},
	}).ReconcileOwnedObjects(ctx, r.Client)
	if err != nil {
		return fmt.Errorf("cannot reconcile objects: %w", err)
	}
	return err
}

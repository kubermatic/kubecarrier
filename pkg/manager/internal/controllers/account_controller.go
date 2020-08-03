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
	"k8c.io/utils/pkg/util"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	catalogv1alpha1 "k8c.io/kubecarrier/pkg/apis/catalog/v1alpha1"
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
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=tenants,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete;escalate;bind
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete

// Reconcile function reconciles the Account object which specified by the request. Currently, it does the following:
// 1. Fetch the Account object.
// 2. Handle the creation/update of the Account object (Create/reconcile the namespace, tenants, roles, and rolebindings).
// 3. Update the status of the Account object.
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

	if util.AddFinalizer(account, metav1.FinalizerDeleteDependents) {
		if err := r.Client.Update(ctx, account); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating Account finalizers: %w", err)
		}
	}

	if err := r.reconcileNamespace(ctx, log, account); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconciling namespace: %w", err)
	}

	if err := r.reconcileRolesAndRoleBindings(ctx, log, account); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconciling account roles and rolebindings: %w", err)
	}

	if !account.IsReady() {
		// Update Account Status
		account.Status.ObservedGeneration = account.Generation
		account.Status.SetCondition(catalogv1alpha1.AccountCondition{
			Type:    catalogv1alpha1.AccountReady,
			Status:  catalogv1alpha1.ConditionTrue,
			Reason:  "SetupComplete",
			Message: "Account setup is complete.",
		})
		if err := r.Status().Update(ctx, account); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating Account status: %w", err)
		}
	}

	if err := r.reconcileTenants(ctx, log, account); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconciling tenant references: %w", err)
	}

	return ctrl.Result{}, nil
}

func (r *AccountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&catalogv1alpha1.Account{}).
		Owns(&corev1.Namespace{}).
		Owns(&catalogv1alpha1.Tenant{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Watches(&source.Kind{Type: &catalogv1alpha1.Account{}}, &handler.EnqueueRequestsFromMapFunc{
			ToRequests: handler.ToRequestsFunc(func(mapObject handler.MapObject) (out []ctrl.Request) {
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
						out = append(out, ctrl.Request{
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
	return nil
}

func (r *AccountReconciler) reconcileNamespace(ctx context.Context, log logr.Logger, account *catalogv1alpha1.Account) error {
	ns := &corev1.Namespace{}
	ns.Name = account.Name
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, ns, func() error {
		if err := controllerutil.SetControllerReference(account, ns, r.Scheme); err != nil {
			return fmt.Errorf("set controller reference for namespace: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("creating or updating namespace: %w", err)
	}

	if account.Status.Namespace == nil {
		account.Status.Namespace = &catalogv1alpha1.ObjectReference{
			Name: ns.Name,
		}
		if err := r.Status().Update(ctx, account); err != nil {
			return fmt.Errorf("updating NamespaceName: %w", err)
		}
	}
	return nil
}

func (r *AccountReconciler) reconcileTenants(ctx context.Context, log logr.Logger, account *catalogv1alpha1.Account) error {
	accountList := &catalogv1alpha1.AccountList{}
	if err := r.List(ctx, accountList); err != nil {
		return fmt.Errorf("listing Accounts: %w", err)
	}

	wantedRefs := make([]*catalogv1alpha1.Tenant, 0)
	if account.HasRole(catalogv1alpha1.TenantRole) {
		for _, providerAccount := range accountList.Items {
			if !providerAccount.HasRole(catalogv1alpha1.ProviderRole) {
				continue
			}
			if condition, _ := providerAccount.Status.GetCondition(catalogv1alpha1.AccountReady); condition.Status != catalogv1alpha1.ConditionTrue {
				continue
			}
			tenant := &catalogv1alpha1.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name:      account.Name,
					Namespace: providerAccount.Status.Namespace.Name,
				},
			}
			wantedRefs = append(wantedRefs, tenant)
		}
	}
	for _, tenant := range wantedRefs {
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, tenant, func() error {
			if err := controllerutil.SetControllerReference(account, tenant, r.Scheme); err != nil {
				return fmt.Errorf("set controller reference for Tenant: %w", err)
			}
			return nil
		}); err != nil {
			return fmt.Errorf("creating or updating Tenant: %w", err)
		}
	}
	return nil
}

func (r *AccountReconciler) reconcileRolesAndRoleBindings(ctx context.Context, log logr.Logger, account *catalogv1alpha1.Account) error {
	var roles []*rbacv1.Role
	var roleBindings []*rbacv1.RoleBinding
	if account.HasRole(catalogv1alpha1.ProviderRole) {
		desiredProviderRole := &rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kubecarrier:provider",
				Namespace: account.Status.Namespace.Name,
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"secrets"},
					Verbs:     []string{rbacv1.VerbAll},
				},
				{
					APIGroups: []string{"kubecarrier.io"},
					Resources: []string{rbacv1.ResourceAll},
					Verbs:     []string{rbacv1.VerbAll},
				},
				{
					APIGroups: []string{"catalog.kubecarrier.io"},
					Resources: []string{
						"catalogs",
						"catalogentries",
						"catalogentrysets",
						"derivedcustomresources",
					},
					Verbs: []string{rbacv1.VerbAll},
				},
				{
					APIGroups: []string{"catalog.kubecarrier.io"},
					Resources: []string{
						"tenants",
					},
					Verbs: []string{"get", "list", "watch", "update", "patch"},
				},
			},
		}
		roles = append(roles, desiredProviderRole)

		desiredProviderRoleBinding := &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kubecarrier:provider",
				Namespace: account.Status.Namespace.Name,
			},
			Subjects: account.Spec.Subjects,
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Role",
				Name:     desiredProviderRole.Name,
			},
		}
		roleBindings = append(roleBindings, desiredProviderRoleBinding)
	}
	if account.HasRole(catalogv1alpha1.TenantRole) {
		desiredTenantRole := &rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kubecarrier:tenant",
				Namespace: account.Status.Namespace.Name,
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{"catalog.kubecarrier.io"},
					Resources: []string{
						"providers",
						"regions",
						"offerings",
					},
					Verbs: []string{"get", "list", "watch"},
				},
			},
		}
		roles = append(roles, desiredTenantRole)

		desiredTenantRoleBinding := &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kubecarrier:tenant",
				Namespace: account.Status.Namespace.Name,
			},
			Subjects: account.Spec.Subjects,
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Role",
				Name:     desiredTenantRole.Name,
			},
		}
		roleBindings = append(roleBindings, desiredTenantRoleBinding)
	}

	for _, role := range roles {
		desiredRole := role.DeepCopy()
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, role, func() error {
			if err := controllerutil.SetControllerReference(account, role, r.Scheme); err != nil {
				return fmt.Errorf("set controller reference for Role: %w", err)
			}
			role.Rules = desiredRole.Rules
			return nil
		}); err != nil {
			return fmt.Errorf("creating or updating Role: %w", err)
		}
	}

	for _, roleBinding := range roleBindings {
		desiredRoleBinding := roleBinding.DeepCopy()
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, roleBinding, func() error {
			if err := controllerutil.SetControllerReference(account, roleBinding, r.Scheme); err != nil {
				return fmt.Errorf("set controller reference for RoleBinding: %w", err)
			}
			roleBinding.Subjects = desiredRoleBinding.Subjects
			return nil
		}); err != nil {
			return fmt.Errorf("creating or updating RoleBindings: %w", err)
		}
	}

	return nil
}

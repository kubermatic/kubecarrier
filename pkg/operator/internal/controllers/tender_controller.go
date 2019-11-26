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

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	tenderControllerFinalizer string = "tender.kubecarrier.io/controller"
)

// TenderReconciler reconciles a Tender object
type TenderReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Log        logr.Logger
	Kustomize  configBuilder
	reconciler reconciler
}

// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=tenders,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=tenders/status,verbs=get;update;patch

func (r *TenderReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var (
		ctx    = context.Background()
		log    = r.Log.WithValues("tender", req.NamespacedName)
		result ctrl.Result
	)

	tender := &operatorv1alpha1.Tender{}
	if err := r.Get(ctx, req.NamespacedName, tender); err != nil {
		return result, client.IgnoreNotFound(err)
	}

	// handle Deletion
	if !tender.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, log, tender); err != nil {
			return result, fmt.Errorf("handling deletion: %v", err)
		}
		return result, nil
	}

	// Add Finalizer
	if util.AddFinalizer(tender, tenderControllerFinalizer) {
		if err := r.Update(ctx, tender); err != nil {
			return result, fmt.Errorf("updating Tender finalizers: %v", err)
		}
	}

	// Add OwnerReferences
	var objectsToCheck []metav1.Object
	for _, object := range objects {
		if err := addOwner(tender, object, r.Scheme); err != nil {
			return result, err
		}

		obj, err := r.reconciler.Reconcile(ctx, log, object)
		if err != nil {
			return result, fmt.Errorf("reconcile type: %w", err)
		}

		objectsToCheck = append(objectsToCheck, obj)
	}

	// Update Tender Status
	_, unready, allComponentsReady := checkReady(objectsToCheck...)
	readyCondition, _ := tender.Status.GetCondition(operatorv1alpha1.TenderReady)

	var unreadyNames []string
	for _, u := range unready {
		unreadyNames = append(unreadyNames, u.GetName())
	}

	var statusUpdate bool

	// Report Unready Status
	if !allComponentsReady &&
		readyCondition.Status != operatorv1alpha1.ConditionFalse {
		statusUpdate = true
		tender.Status.SetCondition(operatorv1alpha1.TenderCondition{
			Type:    operatorv1alpha1.TenderReady,
			Status:  operatorv1alpha1.ConditionFalse,
			Reason:  "ComponentsUnready",
			Message: "not all components are ready, unready: " + strings.Join(unreadyNames, ", "),
		})
	}

	// Report Ready Status
	if allComponentsReady &&
		readyCondition.Status != operatorv1alpha1.ConditionTrue {
		statusUpdate = true
		tender.Status.SetCondition(operatorv1alpha1.TenderCondition{
			Type:    operatorv1alpha1.TenderReady,
			Status:  operatorv1alpha1.ConditionTrue,
			Reason:  "ComponentsReady",
			Message: "all components report ready",
		})
	}

	if statusUpdate {
		if err = r.Status().Update(ctx, tender); err != nil {
			return result, fmt.Errorf("updating Tender Status: %w", err)
		}
	}

	return result, nil
}

func (r *TenderReconciler) handleDeletion(ctx context.Context, log logr.Logger, tender *operatorv1alpha1.Tender) error {
	if cond, _ := tender.Status.GetCondition(operatorv1alpha1.TenderReady); cond.Reason != operatorv1alpha1.TenderTerminatingReason {
		tender.Status.SetCondition(operatorv1alpha1.TenderCondition{
			Message: "Tender is being deleted",
			Reason:  operatorv1alpha1.TenderTerminatingReason,
			Status:  operatorv1alpha1.ConditionFalse,
			Type:    operatorv1alpha1.TenderReady,
		})
		if err := r.Update(ctx, tender); err != nil {
			return fmt.Errorf("updating Tender: %v", err)
		}
	}
	ownedBy, err := util.OwnedBy(tender, r.Scheme)
	if err != nil {
		return fmt.Errorf("owned by selector: %w", err)
	}

	roleBindingsClean, err := cleanupClusterRoleBindings(ctx, r.Client, ownedBy)
	if err != nil {
		return err
	}

	rolesClean, err := cleanupClusterRoles(ctx, r.Client, ownedBy)
	if err != nil {
		return err
	}

	// Wait for all RoleBindings and Roles to be deleted
	if !roleBindingsClean ||
		!rolesClean {
		return nil
	}

	// Remove Finalizer
	if tenderControllerFinalizer.Delete(tender) {
		if err := r.Update(ctx, tender); err != nil {
			return fmt.Errorf("updating Tender: %v", err)
		}
	}
	return nil
}

func (r *TenderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.reconciler = reconcile.NewUnstructuredReconciler(
		r.Client, r.Scheme,
		reconcile.ServiceAccountReconciler,
		reconcile.RoleReconciler, reconcile.RoleBindingReconciler,
		reconcile.ClusterRoleReconciler, reconcile.ClusterRoleBindingReconciler,
		reconcile.DeploymentReconciler, reconcile.ServiceReconciler,
	)

	owner := &operatorv1alpha1.Tender{}
	ownerHandler, err := util.EnqueueRequestForOwner(owner, r.Scheme)
	if err != nil {
		return fmt.Errorf("create owner handler: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.Tender{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Watches(&source.Kind{Type: &rbacv1.ClusterRole{}}, ownerHandler).
		Watches(&source.Kind{Type: &rbacv1.ClusterRoleBinding{}}, ownerHandler).
		Complete(r)
}

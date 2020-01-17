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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

const tenantAssignmentControllerFinalizer string = "tenantassignment.kubecarrier.io/controller"

// TenantAssignmentReconciler reconciles a TenantAssignment object
type TenantAssignmentReconciler struct {
	Log logr.Logger

	MasterClient       client.Client
	ServiceClient      client.Client
	MasterScheme       *runtime.Scheme
	ServiceClusterName string
}

// +kubebuilder:rbac:groups=kubecarrier.io,resources=tenantassignments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubecarrier.io,resources=tenantassignments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete

func (r *TenantAssignmentReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("tenantassignment", req.NamespacedName)

	tenantAssignment := &corev1alpha1.TenantAssignment{}
	if err := r.MasterClient.Get(ctx, req.NamespacedName, tenantAssignment); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// handle Deletion
	if !tenantAssignment.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, log, tenantAssignment); err != nil {
			return ctrl.Result{}, fmt.Errorf("handling deletion: %w", err)
		}
		return ctrl.Result{}, nil
	}

	if util.AddFinalizer(tenantAssignment, tenantAssignmentControllerFinalizer) {
		if err := r.MasterClient.Update(ctx, tenantAssignment); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating %s finalizers: %w", tenantAssignment.Kind, err)
		}
	}

	ns, err := util.UpsertNamespace(ctx, tenantAssignment, r.ServiceClient, r.MasterScheme)
	if err != nil {
		tenantAssignment.Status.ObservedGeneration = tenantAssignment.Generation
		tenantAssignment.Status.SetCondition(corev1alpha1.TenantAssignmentCondition{
			Type:    corev1alpha1.TenantAssignmentReady,
			Status:  corev1alpha1.ConditionFalse,
			Message: err.Error(),
			Reason:  "CreatingNamespace",
		})
		if err := r.MasterClient.Status().Update(ctx, tenantAssignment); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating TenantAssignment Status: %w", err)
		}
		return ctrl.Result{}, fmt.Errorf("cannot create tenant Assigenment namespace: %w", err)
	}

	tenantAssignment.Status.NamespaceName = ns.Name
	tenantAssignment.Status.ObservedGeneration = tenantAssignment.Generation
	tenantAssignment.Status.SetCondition(corev1alpha1.TenantAssignmentCondition{
		Type:    corev1alpha1.TenantAssignmentReady,
		Status:  corev1alpha1.ConditionTrue,
		Message: "Namespace has been setup.",
		Reason:  "SetupComplete",
	})
	if err = r.MasterClient.Status().Update(ctx, tenantAssignment); err != nil {
		return ctrl.Result{}, fmt.Errorf("updating TenantAssignment Status: %w", err)
	}
	return ctrl.Result{}, nil
}

func (r *TenantAssignmentReconciler) handleDeletion(ctx context.Context, log logr.Logger, tenantAssignment *corev1alpha1.TenantAssignment) error {
	// Delete owned Namespaces
	ownedBy, err := util.OwnedBy(tenantAssignment, r.MasterScheme)
	if err != nil {
		return fmt.Errorf("building owned by selector: %w", err)
	}
	if err := r.ServiceClient.DeleteAllOf(ctx, &corev1.Namespace{}, ownedBy); err != nil {
		return fmt.Errorf("deleting Namespaces: %w", err)
	}

	namespaceList := &corev1.NamespaceList{}
	if err = r.ServiceClient.List(ctx, namespaceList, ownedBy); err != nil {
		return fmt.Errorf("listing Namespaces: %w", err)
	}

	if len(namespaceList.Items) != 0 {
		return nil
	}

	// Remove Finalizer
	if util.RemoveFinalizer(tenantAssignment, tenantAssignmentControllerFinalizer) {
		if err := r.MasterClient.Update(ctx, tenantAssignment); err != nil {
			return fmt.Errorf("updating %s finalizers: %w", tenantAssignment.Kind, err)
		}
	}
	return nil
}

func (r *TenantAssignmentReconciler) SetupWithManagers(serviceMgr, masterMgr ctrl.Manager) error {
	namespaceSource := &source.Kind{Type: &corev1.Namespace{}}
	if err := serviceMgr.SetFields(namespaceSource); err != nil {
		return fmt.Errorf("service cluster namespace source: %w", err)
	}

	enqueuer, err := util.EnqueueRequestForOwner(&corev1alpha1.TenantAssignment{}, r.MasterScheme)
	if err != nil {
		return fmt.Errorf("creating TenantAssignment enqueuer: %w", err)
	}

	return ctrl.NewControllerManagedBy(masterMgr).
		For(&corev1alpha1.TenantAssignment{}).
		Watches(source.Func(namespaceSource.Start), enqueuer).
		WithEventFilter(&util.Predicate{Accept: func(obj runtime.Object) bool {
			if tenantAssignment, ok := obj.(*corev1alpha1.TenantAssignment); ok {
				if tenantAssignment.Spec.ServiceCluster.Name == r.ServiceClusterName {
					return true
				}
			}
			return false
		}}).
		Complete(r)
}

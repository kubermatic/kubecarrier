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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

const serviceClusterAssignmentControllerFinalizer string = "serviceclusterassignment.kubecarrier.io/controller"

// ServiceClusterAssignmentReconciler reconciles a ServiceClusterAssignment object
type ServiceClusterAssignmentReconciler struct {
	Log logr.Logger

	MasterClient       client.Client
	ServiceClient      client.Client
	MasterScheme       *runtime.Scheme
	ServiceClusterName string
}

// +kubebuilder:rbac:groups=kubecarrier.io,resources=serviceclusterassignments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubecarrier.io,resources=serviceclusterassignments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete

func (r *ServiceClusterAssignmentReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("serviceClusterAssignment", req.NamespacedName)

	serviceClusterAssignment := &corev1alpha1.ServiceClusterAssignment{}
	if err := r.MasterClient.Get(ctx, req.NamespacedName, serviceClusterAssignment); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// handle Deletion
	if !serviceClusterAssignment.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, log, serviceClusterAssignment); err != nil {
			return ctrl.Result{}, fmt.Errorf("handling deletion: %w", err)
		}
		return ctrl.Result{}, nil
	}

	if util.AddFinalizer(serviceClusterAssignment, serviceClusterAssignmentControllerFinalizer) {
		if err := r.MasterClient.Update(ctx, serviceClusterAssignment); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating %s finalizers: %w", serviceClusterAssignment.Kind, err)
		}
	}

	ns, err := util.UpsertNamespace(ctx, serviceClusterAssignment, r.ServiceClient, r.MasterScheme)
	if err != nil {
		serviceClusterAssignment.Status.ObservedGeneration = serviceClusterAssignment.Generation
		serviceClusterAssignment.Status.SetCondition(corev1alpha1.ServiceClusterAssignmentCondition{
			Type:    corev1alpha1.ServiceClusterAssignmentReady,
			Status:  corev1alpha1.ConditionFalse,
			Message: err.Error(),
			Reason:  "CreatingNamespace",
		})
		if err := r.MasterClient.Status().Update(ctx, serviceClusterAssignment); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating ServiceClusterAssignment Status: %w", err)
		}
		return ctrl.Result{}, fmt.Errorf("cannot create tenant Assigenment namespace: %w", err)
	}

	serviceClusterAssignment.Status.NamespaceName = ns.Name
	serviceClusterAssignment.Status.ObservedGeneration = serviceClusterAssignment.Generation
	serviceClusterAssignment.Status.SetCondition(corev1alpha1.ServiceClusterAssignmentCondition{
		Type:    corev1alpha1.ServiceClusterAssignmentReady,
		Status:  corev1alpha1.ConditionTrue,
		Message: "Namespace has been setup.",
		Reason:  "SetupComplete",
	})
	if err = r.MasterClient.Status().Update(ctx, serviceClusterAssignment); err != nil {
		return ctrl.Result{}, fmt.Errorf("updating ServiceClusterAssignment Status: %w", err)
	}
	return ctrl.Result{}, nil
}

func (r *ServiceClusterAssignmentReconciler) handleDeletion(ctx context.Context, log logr.Logger, serviceClusterAssignment *corev1alpha1.ServiceClusterAssignment) error {
	if serviceClusterAssignment.Status.NamespaceName != "" {
		ns := &corev1.Namespace{}
		ns.SetName(serviceClusterAssignment.Status.NamespaceName)
		err := r.ServiceClient.Delete(ctx, ns)
		switch {
		case err == nil:
			return nil
		case errors.IsNotFound(err):
			break
		default:
			return fmt.Errorf("cannot delete ns: %w", err)
		}
	}

	// Remove Finalizer
	if util.RemoveFinalizer(serviceClusterAssignment, serviceClusterAssignmentControllerFinalizer) {
		if err := r.MasterClient.Update(ctx, serviceClusterAssignment); err != nil {
			return fmt.Errorf("updating %s finalizers: %w", serviceClusterAssignment.Kind, err)
		}
	}
	return nil
}

func (r *ServiceClusterAssignmentReconciler) SetupWithManagers(serviceMgr, masterMgr ctrl.Manager) error {
	namespaceSource := &source.Kind{Type: &corev1.Namespace{}}
	if err := serviceMgr.SetFields(namespaceSource); err != nil {
		return fmt.Errorf("service cluster namespace source: %w", err)
	}

	enqueuer, err := util.EnqueueRequestForOwner(&corev1alpha1.ServiceClusterAssignment{}, r.MasterScheme)
	if err != nil {
		return fmt.Errorf("creating ServiceClusterAssignment enqueuer: %w", err)
	}

	return ctrl.NewControllerManagedBy(masterMgr).
		For(&corev1alpha1.ServiceClusterAssignment{}).
		Watches(source.Func(namespaceSource.Start), enqueuer).
		WithEventFilter(util.PredicateFn(func(obj runtime.Object) bool {
			if serviceClusterAssignment, ok := obj.(*corev1alpha1.ServiceClusterAssignment); ok {
				if serviceClusterAssignment.Spec.ServiceCluster.Name == r.ServiceClusterName {
					return true
				}
			}

			// for namespace owner reconciliation from the service cluster
			if _, ok := obj.(*corev1.Namespace); ok {
				return true
			}
			return false
		})).
		Complete(r)
}

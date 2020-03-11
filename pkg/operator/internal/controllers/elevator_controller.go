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
	certv1alpha2 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	adminv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/owner"
	"github.com/kubermatic/kubecarrier/pkg/internal/reconcile"
	resourceselevator "github.com/kubermatic/kubecarrier/pkg/internal/resources/elevator"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

const (
	elevatorControllerFinalizer = "elevator.kubecarrier.io/controller"
)

// ElevatorReconciler reconciles an Elevator object
type ElevatorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=elevators,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=elevators/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;create;update;patch;delete;escalate;bind
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete;escalate;bind
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=issuers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete

// Reconcile function reconciles the Elevator object which specified by the request. Currently, it does the following:
// 1. Fetch the Elevator object.
// 2. Handle the deletion of the Elevator object (Remove the objects that the Elevator owns, and remove the finalizer).
// 3. Reconcile the objects that owned by Elevator object.
// 4. Update the status of the Elevator object.
func (r *ElevatorReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("elevator", req.NamespacedName)

	// 1. Fetch the Elevator object.
	elevator := &operatorv1alpha1.Elevator{}
	if err := r.Get(ctx, req.NamespacedName, elevator); err != nil {
		// If the Elevator object is already gone, we just ignore the NotFound error.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 2. Handle the deletion of the Elevator object (Remove the objects that the Elevator owns, and remove the finalizer).
	if !elevator.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, elevator); err != nil {
			return ctrl.Result{}, fmt.Errorf("handle deletion: %w", err)
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer
	if util.AddFinalizer(elevator, elevatorControllerFinalizer) {
		if err := r.Update(ctx, elevator); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating Elevator finalizers: %w", err)
		}
	}

	// 3. Reconcile the objects that owned by Elevator object.
	// Build the manifests of the Elevator controller manager.
	objects, err := resourceselevator.Manifests(
		resourceselevator.Config{
			Name:      elevator.Name,
			Namespace: elevator.Namespace,

			ProviderKind:    elevator.Spec.ProviderCRD.Kind,
			ProviderVersion: elevator.Spec.ProviderCRD.Version,
			ProviderGroup:   elevator.Spec.ProviderCRD.Group,
			ProviderPlural:  elevator.Spec.ProviderCRD.Plural,

			TenantKind:    elevator.Spec.TenantCRD.Kind,
			TenantVersion: elevator.Spec.TenantCRD.Version,
			TenantGroup:   elevator.Spec.TenantCRD.Group,
			TenantPlural:  elevator.Spec.TenantCRD.Plural,

			DerivedCRName: elevator.Spec.DerivedCR.Name,
		})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("creating elevator manifests: %w", err)
	}

	deploymentIsReady, err := r.reconcileOwnedObjects(ctx, log, elevator, objects)
	if err != nil {
		return ctrl.Result{}, err
	}

	// 4. Update the status of the Elevator object.
	if err := r.updateElevatorStatus(ctx, elevator, deploymentIsReady); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ElevatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	enqueuer := owner.EnqueueRequestForOwner(&operatorv1alpha1.Elevator{}, r.Scheme)

	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.Elevator{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&certv1alpha2.Issuer{}).
		Owns(&certv1alpha2.Certificate{}).
		Watches(&source.Kind{Type: &corev1.ServiceAccount{}}, enqueuer).
		Watches(&source.Kind{Type: &rbacv1.ClusterRole{}}, enqueuer).
		Watches(&source.Kind{Type: &rbacv1.ClusterRoleBinding{}}, enqueuer).
		Watches(&source.Kind{Type: &adminv1beta1.MutatingWebhookConfiguration{}}, enqueuer).
		Complete(r)
}

// handleDeletion handles the deletion of the Elevator object. Currently, it does:
// 1. Update the Elevator status to Terminating.
// 2. Delete the objects that the Elevator object owns.
// 3. Remove the finalizer from the Elevator object.
func (r *ElevatorReconciler) handleDeletion(ctx context.Context, elevator *operatorv1alpha1.Elevator) error {

	// 1. Update the Elevator Status to Terminating.
	readyCondition, _ := elevator.Status.GetCondition(operatorv1alpha1.ElevatorReady)
	if readyCondition.Status != operatorv1alpha1.ConditionFalse ||
		readyCondition.Status == operatorv1alpha1.ConditionFalse && readyCondition.Reason != operatorv1alpha1.ElevatorTerminatingReason {
		elevator.Status.ObservedGeneration = elevator.Generation
		elevator.Status.SetCondition(operatorv1alpha1.ElevatorCondition{
			Type:    operatorv1alpha1.ElevatorReady,
			Status:  operatorv1alpha1.ConditionFalse,
			Reason:  operatorv1alpha1.ElevatorTerminatingReason,
			Message: "Elevator is being terminated",
		})
		if err := r.Status().Update(ctx, elevator); err != nil {
			return fmt.Errorf("updating Elevator status: %w", err)
		}
	}

	// 2. Delete Objects.
	ownedByFilter := owner.OwnedBy(elevator, r.Scheme)

	clusterRoleBindingsCleaned, err := cleanupClusterRoleBindings(ctx, r.Client, ownedByFilter)
	if err != nil {
		return fmt.Errorf("cleaning ClusterRoleBinding: %w", err)
	}

	clusterRolesCleaned, err := cleanupClusterRoles(ctx, r.Client, ownedByFilter)
	if err != nil {
		return fmt.Errorf("cleaning ClusterRoles: %w", err)
	}

	customResourceDefinitionsCleaned, err := cleanupCustomResourceDefinitions(ctx, r.Client, ownedByFilter)
	if err != nil {
		return fmt.Errorf("cleaning CustomResourceDefinitions: %w", err)
	}

	mutatingWebhookConfigurationsCleaned, err := cleanupMutatingWebhookConfigurations(ctx, r.Client, ownedByFilter)
	if err != nil {
		return fmt.Errorf("cleaning MutatingWebhookConfigurations: %w", err)
	}

	// Make sure all the ClusterRoleBindings, ClusterRoles and CustomResourceDefinitions are deleted.
	if !clusterRoleBindingsCleaned ||
		!clusterRolesCleaned ||
		!customResourceDefinitionsCleaned ||
		!mutatingWebhookConfigurationsCleaned {
		return nil
	}

	// 3. Remove the Finalizer.
	if util.RemoveFinalizer(elevator, elevatorControllerFinalizer) {
		if err := r.Update(ctx, elevator); err != nil {
			return fmt.Errorf("updating Elevator finalizers: %w", err)
		}
	}
	return nil
}

// reconcileOwnedObjects adds the OwnerReference to the objects that owned by this Elevator object, and reconciles them.
func (r *ElevatorReconciler) reconcileOwnedObjects(ctx context.Context, log logr.Logger, elevator *operatorv1alpha1.Elevator, objects []unstructured.Unstructured) (bool, error) {
	var deploymentIsReady bool
	for _, object := range objects {
		if err := addOwnerReference(elevator, &object, r.Scheme); err != nil {
			return false, err
		}
		curObj, err := reconcile.Unstructured(ctx, log, r.Client, &object)
		if err != nil {
			return false, fmt.Errorf("reconcile kind: %s, err: %w", object.GroupVersionKind().Kind, err)
		}

		switch obj := curObj.(type) {
		case *appsv1.Deployment:
			deploymentIsReady = util.DeploymentIsAvailable(obj)
		}
	}
	return deploymentIsReady, nil
}

// updateElevatorStatus updates the Status of the Elevator object if needed.
func (r *ElevatorReconciler) updateElevatorStatus(ctx context.Context, elevator *operatorv1alpha1.Elevator, deploymentIsReady bool) error {
	var updateStatus bool
	readyCondition, _ := elevator.Status.GetCondition(operatorv1alpha1.ElevatorReady)
	if !deploymentIsReady && readyCondition.Status != operatorv1alpha1.ConditionFalse {
		updateStatus = true
		elevator.Status.ObservedGeneration = elevator.Generation
		elevator.Status.SetCondition(operatorv1alpha1.ElevatorCondition{
			Type:    operatorv1alpha1.ElevatorReady,
			Status:  operatorv1alpha1.ConditionFalse,
			Reason:  "DeploymentUnready",
			Message: "the deployment of the Elevator controller manager is not ready",
		})
	}

	if deploymentIsReady && readyCondition.Status != operatorv1alpha1.ConditionTrue {
		updateStatus = true
		elevator.Status.ObservedGeneration = elevator.Generation
		elevator.Status.SetCondition(operatorv1alpha1.ElevatorCondition{
			Type:    operatorv1alpha1.ElevatorReady,
			Status:  operatorv1alpha1.ConditionTrue,
			Reason:  "DeploymentReady",
			Message: "the deployment of the Elevator controller manager is ready",
		})
	}

	if updateStatus {
		if err := r.Status().Update(ctx, elevator); err != nil {
			return fmt.Errorf("updating Elevator status: %w", err)
		}
	}
	return nil
}

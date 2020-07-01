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
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/owner"
	resourceselevator "github.com/kubermatic/kubecarrier/pkg/internal/resources/elevator"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

const (
	elevatorControllerFinalizer = "elevator.kubecarrier.io/controller"
)

// ElevatorReconciler reconciles a Elevator object
type ElevatorReconciler struct {
	client.Client
	Log        logr.Logger
	Scheme     *runtime.Scheme
	RESTMapper meta.RESTMapper
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

func (r *ElevatorReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("elevator", req.NamespacedName)

	elevator := &operatorv1alpha1.Elevator{}
	if err := r.Get(ctx, req.NamespacedName, elevator); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if elevator.Spec.Paused.IsPaused() {
		if elevator.SetPausedCondition() {
			if err := r.Client.Status().Update(ctx, elevator); err != nil {
				return ctrl.Result{}, fmt.Errorf("updating %s status: %w", elevator.Name, err)
			}
		}
		// reconciliation paused, skip all other handlers
		return ctrl.Result{}, nil
	}

	if !elevator.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, elevator); err != nil {
			return ctrl.Result{}, fmt.Errorf("handle deletion: %w", err)
		}
		return ctrl.Result{}, nil
	}

	if util.AddFinalizer(elevator, elevatorControllerFinalizer) {
		if err := r.Update(ctx, elevator); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating Elevator finalizers: %w", err)
		}
	}

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
		return ctrl.Result{}, fmt.Errorf("creating Elevator manifests: %w", err)
	}

	deploymentIsReady, err := reconcileOwnedObjectsForNamespacedOwner(ctx, log, r.Scheme, r.RESTMapper, r.Client, elevator, objects)
	if err != nil {
		return ctrl.Result{}, err
	}

	// 4. Update the status of the Elevator object.
	if err := r.updateStatus(ctx, elevator, deploymentIsReady); err != nil {
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
		Owns(&corev1.ServiceAccount{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&certv1alpha2.Issuer{}).
		Owns(&certv1alpha2.Certificate{}).
		Watches(&source.Kind{Type: &rbacv1.ClusterRole{}}, enqueuer).
		Watches(&source.Kind{Type: &rbacv1.ClusterRoleBinding{}}, enqueuer).
		Watches(&source.Kind{Type: &adminv1beta1.MutatingWebhookConfiguration{}}, enqueuer).
		Complete(r)
}

func (r *ElevatorReconciler) handleDeletion(ctx context.Context, elevator *operatorv1alpha1.Elevator) error {
	if elevator.SetTerminatingCondition() {
		if err := r.Client.Status().Update(ctx, elevator); err != nil {
			return fmt.Errorf("updating %s status: %w", elevator.Name, err)
		}
	}

	cleanedUp, err := util.DeleteObjects(ctx, r.Client, r.Scheme, []runtime.Object{
		&rbacv1.ClusterRole{},
		&rbacv1.ClusterRoleBinding{},
		&adminv1beta1.MutatingWebhookConfiguration{},
	}, owner.OwnedBy(elevator, r.Scheme))
	if err != nil {
		return fmt.Errorf("DeleteObjects: %w", err)
	}

	if cleanedUp && util.RemoveFinalizer(elevator, elevatorControllerFinalizer) {
		if err := r.Update(ctx, elevator); err != nil {
			return fmt.Errorf("updating Elevator finalizers: %w", err)
		}
	}
	return nil
}

func (r *ElevatorReconciler) updateStatus(ctx context.Context, elevator *operatorv1alpha1.Elevator, deploymentIsReady bool) error {
	var pausedChanged, readyChanged bool

	if !elevator.Spec.Paused.IsPaused() {
		pausedChanged = elevator.SetUnPausedCondition()
	}

	if deploymentIsReady {
		readyChanged = elevator.SetReadyCondition()
	} else {
		readyChanged = elevator.SetUnReadyCondition()
	}

	if readyChanged || pausedChanged {
		if err := r.Client.Status().Update(ctx, elevator); err != nil {
			return fmt.Errorf("updating %s status: %w", elevator.Name, err)
		}
	}
	return nil
}

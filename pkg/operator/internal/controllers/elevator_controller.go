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

	"github.com/go-logr/logr"
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
	resourceselevator "github.com/kubermatic/kubecarrier/pkg/internal/resources/elevator"
)

type elevatorController struct {
	Obj *operatorv1alpha1.Elevator
}

func (c *elevatorController) GetReadyConditionStatus() operatorv1alpha1.ConditionStatus {
	readyCondition, _ := c.Obj.Status.GetCondition(operatorv1alpha1.ElevatorReady)
	return readyCondition.Status
}

func (c *elevatorController) SetReadyCondition() {
	c.Obj.Status.ObservedGeneration = c.Obj.Generation
	c.Obj.Status.SetCondition(operatorv1alpha1.ElevatorCondition{
		Type:    operatorv1alpha1.ElevatorReady,
		Status:  operatorv1alpha1.ConditionTrue,
		Reason:  "DeploymentReady",
		Message: "the deployment of the Elevator controller manager is ready",
	})
}

func (c *elevatorController) SetUnReadyCondition() {
	c.Obj.Status.ObservedGeneration = c.Obj.Generation
	c.Obj.Status.SetCondition(operatorv1alpha1.ElevatorCondition{
		Type:    operatorv1alpha1.ElevatorReady,
		Status:  operatorv1alpha1.ConditionFalse,
		Reason:  "DeploymentUnready",
		Message: "the deployment of the Elevator controller manager is not ready",
	})
}

func (c *elevatorController) SetTerminatingCondition(ctx context.Context) bool {
	readyCondition, _ := c.Obj.Status.GetCondition(operatorv1alpha1.ElevatorReady)
	if readyCondition.Status != operatorv1alpha1.ConditionFalse ||
		readyCondition.Status == operatorv1alpha1.ConditionFalse && readyCondition.Reason != operatorv1alpha1.ElevatorTerminatingReason {
		c.Obj.Status.ObservedGeneration = c.Obj.Generation
		c.Obj.Status.SetCondition(operatorv1alpha1.ElevatorCondition{
			Type:    operatorv1alpha1.ElevatorReady,
			Status:  operatorv1alpha1.ConditionFalse,
			Reason:  operatorv1alpha1.ElevatorTerminatingReason,
			Message: "Elevator is being terminated",
		})
		return true
	}
	return false
}

func (c *elevatorController) GetObj() object {
	return c.Obj
}

func (c *elevatorController) GetManifests(ctx context.Context) ([]unstructured.Unstructured, error) {
	return resourceselevator.Manifests(
		resourceselevator.Config{
			Name:      c.Obj.Name,
			Namespace: c.Obj.Namespace,

			ProviderKind:    c.Obj.Spec.ProviderCRD.Kind,
			ProviderVersion: c.Obj.Spec.ProviderCRD.Version,
			ProviderGroup:   c.Obj.Spec.ProviderCRD.Group,
			ProviderPlural:  c.Obj.Spec.ProviderCRD.Plural,

			TenantKind:    c.Obj.Spec.TenantCRD.Kind,
			TenantVersion: c.Obj.Spec.TenantCRD.Version,
			TenantGroup:   c.Obj.Spec.TenantCRD.Group,
			TenantPlural:  c.Obj.Spec.TenantCRD.Plural,

			DerivedCRName: c.Obj.Spec.DerivedCR.Name,
		})
}

// ElevatorReconciler reconciles an Elevator object
type ElevatorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=elevators,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=elevators/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;create;update;patch;delete;escalate;bind
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete;escalate;bind
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete

// Reconcile function reconciles the Elevator object which specified by the request. Currently, it does the following:
// 1. Fetch the Elevator object.
// 2. Handle the deletion of the Elevator object (Remove the objects that the Elevator owns, and remove the finalizer).
// 3. Reconcile the objects that owned by Elevator object.
// 4. Update the status of the Elevator object.
func (r *ElevatorReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("elevator", req.NamespacedName)
	br := BaseReconciler{
		Client:    r.Client,
		Log:       log,
		Scheme:    r.Scheme,
		Finalizer: "elevator.kubecarrier.io/controller",
		Name:      "Elevator",
	}

	elevator := &operatorv1alpha1.Elevator{}
	ctr := &elevatorController{Obj: elevator}
	return br.Reconcile(ctx, req, ctr)
}

func (r *ElevatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	enqueuer := owner.EnqueueRequestForOwner(&operatorv1alpha1.Elevator{}, r.Scheme)

	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.Elevator{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Watches(&source.Kind{Type: &corev1.ServiceAccount{}}, enqueuer).
		Watches(&source.Kind{Type: &rbacv1.ClusterRole{}}, enqueuer).
		Watches(&source.Kind{Type: &rbacv1.ClusterRoleBinding{}}, enqueuer).
		Complete(r)
}

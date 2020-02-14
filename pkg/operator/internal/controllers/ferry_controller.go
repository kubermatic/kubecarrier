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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/reconcile"
	resourceferry "github.com/kubermatic/kubecarrier/pkg/internal/resources/ferry"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

const (
	ferryControllerFinalizer string = "ferry.kubecarrier.io/controller"
)

// FerryReconciler reconciles a Ferry object
type FerryReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete;escalate;bind
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=ferries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=ferries/status,verbs=get;update;patch

// Reconcile function reconciles the Ferry object which specified by the request. Currently, it does the following:
// * fetches the Ferry object
// * handles object deletion if neccessary
// * create all necessary objects from resource
// * Updates the ferry status
func (r *FerryReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("ferry", req.NamespacedName)

	ferry := &operatorv1alpha1.Ferry{}
	if err := r.Get(ctx, req.NamespacedName, ferry); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	ferry.Status.ObservedGeneration = ferry.Generation

	// handle Deletion
	if !ferry.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, log, ferry); err != nil {
			return ctrl.Result{}, fmt.Errorf("handling deletion: %v", err)
		}
		return ctrl.Result{}, nil
	}

	// Add Finalizer
	if util.AddFinalizer(ferry, ferryControllerFinalizer) {
		if err := r.Update(ctx, ferry); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating Ferry finalizers: %v", err)
		}
	}

	// Reconcile the objects that owned by Ferry object.
	var deploymentReady bool

	// Build the manifests of the Ferry controller manager.
	objects, err := resourceferry.Manifests(ferryConfigurationForObject(ferry))
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("ferry manifests: %w", err)
	}
	for _, object := range objects {
		if err := controllerutil.SetControllerReference(ferry, &object, r.Scheme); err != nil {
			return ctrl.Result{}, fmt.Errorf("setting controller reference: %w", err)
		}
		currObj, err := reconcile.Unstructured(ctx, log, r.Client, &object)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("reconcile type: %w", err)
		}
		switch obj := currObj.(type) {
		case *appsv1.Deployment:
			deploymentReady = util.DeploymentIsAvailable(obj)
		}
	}

	// Update the ferry status
	if deploymentReady {
		ferry.Status.SetCondition(operatorv1alpha1.FerryCondition{
			Type:    operatorv1alpha1.FerryReady,
			Status:  operatorv1alpha1.ConditionTrue,
			Reason:  "ComponentsReady",
			Message: "all components report ready",
		})
	} else {
		ferry.Status.SetCondition(operatorv1alpha1.FerryCondition{
			Type:    operatorv1alpha1.FerryReady,
			Status:  operatorv1alpha1.ConditionFalse,
			Reason:  "ComponentsUnready",
			Message: "Ferry deployment isn't ready",
		})
	}

	if err = r.Status().Update(ctx, ferry); err != nil {
		return ctrl.Result{}, fmt.Errorf("updating Ferry Status: %w", err)
	}
	return ctrl.Result{}, nil
}

func (r *FerryReconciler) handleDeletion(ctx context.Context, log logr.Logger, ferry *operatorv1alpha1.Ferry) error {
	if cond, _ := ferry.Status.GetCondition(operatorv1alpha1.FerryReady); cond.Reason != operatorv1alpha1.FerryTerminatingReason {
		ferry.Status.SetCondition(operatorv1alpha1.FerryCondition{
			Message: "Ferry is being deleted",
			Reason:  operatorv1alpha1.FerryTerminatingReason,
			Status:  operatorv1alpha1.ConditionFalse,
			Type:    operatorv1alpha1.FerryReady,
		})
		if err := r.Status().Update(ctx, ferry); err != nil {
			return fmt.Errorf("updating Ferry: %v", err)
		}
	}

	// Delete Objects.
	allCleared := true
	objects, err := resourceferry.Manifests(ferryConfigurationForObject(ferry))
	if err != nil {
		return fmt.Errorf("deletion: manifests: %w", err)
	}
	for _, obj := range objects {
		err := r.Client.Delete(ctx, &obj)
		if errors.IsNotFound(err) {
			continue
		}
		allCleared = false
		if err != nil {
			return fmt.Errorf("deleting %s: %w", obj, err)
		}
	}

	if allCleared {
		// Remove Finalizer
		if util.RemoveFinalizer(ferry, ferryControllerFinalizer) {
			if err := r.Update(ctx, ferry); err != nil {
				return fmt.Errorf("updating Ferry: %v", err)
			}
		}
	}
	return nil
}

var ferryControllerObjects = []runtime.Object{
	&corev1.Service{},
	&corev1.ServiceAccount{},
	&rbacv1.Role{},
	&rbacv1.RoleBinding{},
	&appsv1.Deployment{},
}

func (r *FerryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	cm := ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.Ferry{})

	for _, obj := range ferryControllerObjects {
		cm = cm.Watches(&source.Kind{Type: obj}, &handler.EnqueueRequestForOwner{OwnerType: &operatorv1alpha1.Ferry{}})
	}

	return cm.Complete(r)
}

func ferryConfigurationForObject(ferry *operatorv1alpha1.Ferry) resourceferry.Config {
	return resourceferry.Config{
		ProviderNamespace:    ferry.Namespace,
		Name:                 ferry.Name,
		KubeconfigSecretName: ferry.Spec.KubeconfigSecret.Name,
	}
}

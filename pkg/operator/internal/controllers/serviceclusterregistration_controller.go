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
	"github.com/kubermatic/kubecarrier/pkg/internal/kustomize"
	"github.com/kubermatic/kubecarrier/pkg/internal/reconcile"
	"github.com/kubermatic/kubecarrier/pkg/internal/resources/tender"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

const (
	serviceclusterregistrationControllerFinalizer string = "serviceclusterregistration.kubecarrier.io/controller"
)

// ServiceClusterRegistrationReconciler reconciles a ServiceClusterRegistration object
type ServiceClusterRegistrationReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Log       logr.Logger
	Kustomize *kustomize.Kustomize
}

// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete;escalate;bind
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=serviceclusterregistrations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=serviceclusterregistrations/status,verbs=get;update;patch

// Reconcile function reconciles the ServiceClusterRegistration object which specified by the request. Currently, it does the following:
// * fetches the ServiceClusterRegistration object
// * handles object deletion if neccessary
// * create all necessary objects from resource
// * Updates the serviceclusterregistration status
func (r *ServiceClusterRegistrationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("serviceclusterregistration", req.NamespacedName)

	serviceclusterregistration := &operatorv1alpha1.ServiceClusterRegistration{}
	if err := r.Get(ctx, req.NamespacedName, serviceclusterregistration); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	serviceclusterregistration.Status.ObservedGeneration = serviceclusterregistration.Generation

	// handle Deletion
	if !serviceclusterregistration.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, log, serviceclusterregistration); err != nil {
			return ctrl.Result{}, fmt.Errorf("handling deletion: %v", err)
		}
		return ctrl.Result{}, nil
	}

	// Add Finalizer
	if util.AddFinalizer(serviceclusterregistration, serviceclusterregistrationControllerFinalizer) {
		if err := r.Update(ctx, serviceclusterregistration); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating ServiceClusterRegistration finalizers: %v", err)
		}
	}

	// Reconcile the objects that owned by ServiceClusterRegistration object.
	var deploymentReady bool

	// Build the manifests of the ServiceClusterRegistration controller manager.
	objects, err := tender.Manifests(r.Kustomize, serviceclusterregistrationConfigurationForObject(serviceclusterregistration))
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("serviceclusterregistration manifests: %w", err)
	}
	for _, object := range objects {
		if err := controllerutil.SetControllerReference(serviceclusterregistration, &object, r.Scheme); err != nil {
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

	// Update the serviceclusterregistration status
	if deploymentReady {
		serviceclusterregistration.Status.SetCondition(operatorv1alpha1.ServiceClusterRegistrationCondition{
			Type:    operatorv1alpha1.ServiceClusterRegistrationReady,
			Status:  operatorv1alpha1.ConditionTrue,
			Reason:  "ComponentsReady",
			Message: "all components report ready",
		})
	} else {
		serviceclusterregistration.Status.SetCondition(operatorv1alpha1.ServiceClusterRegistrationCondition{
			Type:    operatorv1alpha1.ServiceClusterRegistrationReady,
			Status:  operatorv1alpha1.ConditionFalse,
			Reason:  "ComponentsUnready",
			Message: "ServiceClusterRegistration deployment isn't ready",
		})
	}

	if err = r.Status().Update(ctx, serviceclusterregistration); err != nil {
		return ctrl.Result{}, fmt.Errorf("updating ServiceClusterRegistration Status: %w", err)
	}
	return ctrl.Result{}, nil
}

func (r *ServiceClusterRegistrationReconciler) handleDeletion(ctx context.Context, log logr.Logger, serviceclusterregistration *operatorv1alpha1.ServiceClusterRegistration) error {
	if cond, _ := serviceclusterregistration.Status.GetCondition(operatorv1alpha1.ServiceClusterRegistrationReady); cond.Reason != operatorv1alpha1.ServiceClusterRegistrationTerminatingReason {
		serviceclusterregistration.Status.SetCondition(operatorv1alpha1.ServiceClusterRegistrationCondition{
			Message: "ServiceClusterRegistration is being deleted",
			Reason:  operatorv1alpha1.ServiceClusterRegistrationTerminatingReason,
			Status:  operatorv1alpha1.ConditionFalse,
			Type:    operatorv1alpha1.ServiceClusterRegistrationReady,
		})
		if err := r.Status().Update(ctx, serviceclusterregistration); err != nil {
			return fmt.Errorf("updating ServiceClusterRegistration: %v", err)
		}
	}

	// 2. Delete Objects.
	allCleared := true
	objects, err := tender.Manifests(r.Kustomize, serviceclusterregistrationConfigurationForObject(serviceclusterregistration))
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
		if util.RemoveFinalizer(serviceclusterregistration, serviceclusterregistrationControllerFinalizer) {
			if err := r.Update(ctx, serviceclusterregistration); err != nil {
				return fmt.Errorf("updating ServiceClusterRegistration: %v", err)
			}
		}
	}
	return nil
}

var serviceclusterregistrationControllerObjects = []runtime.Object{
	&corev1.Service{},
	&corev1.ServiceAccount{},
	&rbacv1.Role{},
	&rbacv1.RoleBinding{},
	&appsv1.Deployment{},
}

func (r *ServiceClusterRegistrationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	cm := ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.ServiceClusterRegistration{})

	for _, obj := range serviceclusterregistrationControllerObjects {
		cm = cm.Watches(&source.Kind{Type: obj}, &handler.EnqueueRequestForOwner{OwnerType: &operatorv1alpha1.ServiceClusterRegistration{}})
	}

	return cm.Complete(r)
}

func serviceclusterregistrationConfigurationForObject(serviceclusterregistration *operatorv1alpha1.ServiceClusterRegistration) tender.Config {
	return tender.Config{
		ProviderNamespace:    serviceclusterregistration.Namespace,
		Name:                 serviceclusterregistration.Name,
		KubeconfigSecretName: serviceclusterregistration.Spec.KubeconfigSecret.Name,
	}
}

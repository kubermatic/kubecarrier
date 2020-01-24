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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/reconcile"
	"github.com/kubermatic/kubecarrier/pkg/internal/resources/manager"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

const (
	kubeCarrierControllerFinalizer = "kubecarrier.kubecarrier.io/controller"
)

// KubeCarrierReconciler reconciles a KubeCarrier object
type KubeCarrierReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=kubecarriers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=kubecarriers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;create;update;patch;delete;escalate;bind
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete;escalate;bind
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=issuers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete

// Reconcile function reconciles the KubeCarrier object which specified by the request. Currently, it does the following:
// 1. Fetch the KubeCarrier object.
// 2. Handle the deletion of the KubeCarrier object (Remove the objects that the KubeCarrier owns, and remove the finalizer).
// 3. Reconcile the objects that owned by KubeCarrier object.
// 4. Update the status of the KubeCarrier object.
func (r *KubeCarrierReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("kubecarrier", req.NamespacedName)

	// 1. Fetch the KubeCarrier object.
	kubeCarrier := &operatorv1alpha1.KubeCarrier{}
	if err := r.Get(ctx, req.NamespacedName, kubeCarrier); err != nil {
		// If the KubeCarrier object is already gone, we just ignore the NotFound error.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 2. Handle the deletion of the KubeCarrier object (Remove the objects that the KubeCarrier owns, and remove the finalizer).
	if !kubeCarrier.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, kubeCarrier); err != nil {
			return ctrl.Result{}, fmt.Errorf("handle deletion: %w", err)
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer
	if util.AddFinalizer(kubeCarrier, kubeCarrierControllerFinalizer) {
		if err := r.Update(ctx, kubeCarrier); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating KubeCarrier finalizers: %w", err)
		}
	}

	// 3. Reconcile the objects that owned by KubeCarrier object.
	// Build the manifests of the KubeCarrier controller manager.
	objects, err := manager.Manifests(
		manager.Config{
			Namespace: kubeCarrier.Namespace,
		})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("creating manager manifests: %w", err)
	}

	deploymentIsReady, err := r.reconcileOwnedObjects(ctx, log, kubeCarrier, objects)
	if err != nil {
		return ctrl.Result{}, err
	}

	// 4. Update the status of the KubeCarrier object.
	if err := r.updateKubeCarrierStatus(ctx, kubeCarrier, deploymentIsReady); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *KubeCarrierReconciler) SetupWithManager(mgr ctrl.Manager) error {
	owner := &operatorv1alpha1.KubeCarrier{}
	enqueuer, err := util.EnqueueRequestForOwner(owner, mgr.GetScheme())
	if err != nil {
		return fmt.Errorf("cannot create enqueuer for KubeCarrier: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(owner).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Watches(&source.Kind{Type: &rbacv1.ClusterRole{}}, enqueuer).
		Watches(&source.Kind{Type: &rbacv1.ClusterRoleBinding{}}, enqueuer).
		Watches(&source.Kind{Type: &apiextensionsv1.CustomResourceDefinition{}}, enqueuer).
		Complete(r)
}

// handleDeletion handles the deletion of the KubeCarrier object. Currently, it does:
// 1. Update the KubeCarrier status to Terminating.
// 2. Delete the objects that the KubeCarrier object owns.
// 3. Remove the finalizer from the KubeCarrier object.
func (r *KubeCarrierReconciler) handleDeletion(ctx context.Context, kubeCarrier *operatorv1alpha1.KubeCarrier) error {

	// 1. Update the KubeCarrier Status to Terminating.
	readyCondition, _ := kubeCarrier.Status.GetCondition(operatorv1alpha1.KubeCarrierReady)
	if readyCondition.Status != operatorv1alpha1.ConditionFalse ||
		readyCondition.Status == operatorv1alpha1.ConditionFalse && readyCondition.Reason != operatorv1alpha1.KubeCarrierTerminatingReason {
		kubeCarrier.Status.ObservedGeneration = kubeCarrier.Generation
		kubeCarrier.Status.SetCondition(operatorv1alpha1.KubeCarrierCondition{
			Type:    operatorv1alpha1.KubeCarrierReady,
			Status:  operatorv1alpha1.ConditionFalse,
			Reason:  operatorv1alpha1.KubeCarrierTerminatingReason,
			Message: "KubeCarrier is being terminated",
		})
		if err := r.Status().Update(ctx, kubeCarrier); err != nil {
			return fmt.Errorf("updating KubeCarrier status: %w", err)
		}
	}

	// 2. Delete Objects.
	ownedBy, err := util.OwnedBy(kubeCarrier, r.Scheme)
	if err != nil {
		return fmt.Errorf("getting ownedBy list option: %w", err)
	}

	clusterRoleBindingsCleaned, err := cleanupClusterRoleBindings(ctx, r.Client, ownedBy)
	if err != nil {
		return fmt.Errorf("cleaning ClusterRoleBinding: %w", err)
	}

	clusterRolesCleaned, err := cleanupClusterRoles(ctx, r.Client, ownedBy)
	if err != nil {
		return fmt.Errorf("cleaning ClusterRoles: %w", err)
	}

	customResourceDefinitionsCleaned, err := cleanupCustomResourceDefinitions(ctx, r.Client, ownedBy)
	if err != nil {
		return fmt.Errorf("cleaning CustomResourceDefinitions: %w", err)
	}

	mutatingWebhookConfigurationsCleaned, err := cleanupMutatingWebhookConfiguration(ctx, r.Client, ownedBy)
	if err != nil {
		return fmt.Errorf("cleaning MutatingWebhookConfigurations: %w", err)
	}

	validatingWebhookConfigurationsCleaned, err := cleanupValidatingWebhookConfiguration(ctx, r.Client, ownedBy)
	if err != nil {
		return fmt.Errorf("cleaning ValidatingWebhookConfigurations: %w", err)
	}

	// Make sure all the owned objects are deleted.
	if !clusterRoleBindingsCleaned ||
		!clusterRolesCleaned ||
		!customResourceDefinitionsCleaned ||
		!mutatingWebhookConfigurationsCleaned ||
		!validatingWebhookConfigurationsCleaned {
		return nil
	}

	// 3. Remove the Finalizer.
	if util.RemoveFinalizer(kubeCarrier, kubeCarrierControllerFinalizer) {
		if err := r.Update(ctx, kubeCarrier); err != nil {
			return fmt.Errorf("updating KubeCarrier finalizers: %w", err)
		}
	}
	return nil
}

// reconcileOwnedObjects adds the OwnerReference to the objects that owned by this KubeCarrier object, and reconciles them.
func (r *KubeCarrierReconciler) reconcileOwnedObjects(ctx context.Context, log logr.Logger, kubeCarrier *operatorv1alpha1.KubeCarrier, objects []unstructured.Unstructured) (bool, error) {
	var deploymentIsReady bool
	for _, object := range objects {
		if err := addOwnerReference(kubeCarrier, &object, r.Scheme); err != nil {
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

// updateKubeCarrierStatus updates the Status of the KubeCarrier object if needed.
func (r *KubeCarrierReconciler) updateKubeCarrierStatus(ctx context.Context, kubeCarrier *operatorv1alpha1.KubeCarrier, deploymentIsReady bool) error {
	var updateStatus bool
	readyCondition, _ := kubeCarrier.Status.GetCondition(operatorv1alpha1.KubeCarrierReady)
	if !deploymentIsReady && readyCondition.Status != operatorv1alpha1.ConditionFalse {
		updateStatus = true
		kubeCarrier.Status.ObservedGeneration = kubeCarrier.Generation
		kubeCarrier.Status.SetCondition(operatorv1alpha1.KubeCarrierCondition{
			Type:    operatorv1alpha1.KubeCarrierReady,
			Status:  operatorv1alpha1.ConditionFalse,
			Reason:  "DeploymentUnready",
			Message: "the deployment of the KubeCarrier controller manager is not ready",
		})
	}

	if deploymentIsReady && readyCondition.Status != operatorv1alpha1.ConditionTrue {
		updateStatus = true
		kubeCarrier.Status.ObservedGeneration = kubeCarrier.Generation
		kubeCarrier.Status.SetCondition(operatorv1alpha1.KubeCarrierCondition{
			Type:    operatorv1alpha1.KubeCarrierReady,
			Status:  operatorv1alpha1.ConditionTrue,
			Reason:  "DeploymentReady",
			Message: "the deployment of the KubeCarrier controller manager is ready",
		})
	}

	if updateStatus {
		if err := r.Status().Update(ctx, kubeCarrier); err != nil {
			return fmt.Errorf("updating KubeCarrier status: %w", err)
		}
	}
	return nil
}

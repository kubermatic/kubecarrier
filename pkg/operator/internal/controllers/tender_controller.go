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

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/kubermatic/kubecarrier/pkg/internal/reconcile"

	"github.com/kubermatic/kubecarrier/pkg/internal/kustomize"

	tender2 "github.com/kubermatic/kubecarrier/pkg/internal/resources/tender"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	tenderControllerFinalizer string = "tender.kubecarrier.io/controller"
)

var tenderControllerObjects = []runtime.Object{
	&corev1.Service{},
	&corev1.ServiceAccount{},
	&rbacv1.Role{},
	&rbacv1.RoleBinding{},
	&rbacv1.ClusterRole{},
	&rbacv1.ClusterRoleBinding{},
	&appsv1.Deployment{},
}

// TenderReconciler reconciles a Tender object
type TenderReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Log       logr.Logger
	Kustomize *kustomize.Kustomize
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
// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=tenders,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=tenders/status,verbs=get;update;patch

func (r *TenderReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("tender", req.NamespacedName)

	tender := &operatorv1alpha1.Tender{}
	if err := r.Get(ctx, req.NamespacedName, tender); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// handle Deletion
	if !tender.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, log, tender); err != nil {
			return ctrl.Result{}, fmt.Errorf("handling deletion: %v", err)
		}
		return ctrl.Result{}, nil
	}

	// Add Finalizer
	if util.AddFinalizer(tender, tenderControllerFinalizer) {
		if err := r.Update(ctx, tender); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating Tender finalizers: %v", err)
		}
	}

	// 3. Reconcile the objects that owned by Tender object.
	var deploymentReady bool

	// Build the manifests of the Tender controller manager.
	objects, err := tender2.Manifests(r.Kustomize, tender2.Config{ProviderNamespace: req.Namespace})
	for _, object := range objects {
		if _, err := util.InsertOwnerReference(tender, &object, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		currObj, err := reconcile.Unstructured(ctx, log, r.Client, &object)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("reconcile type: %w", err)
		}
		switch currObj.(type) {
		case *appsv1.Deployment:
			deploymentReady = util.DeploymentIsAvailable(currObj.(*appsv1.Deployment))
		}
	}

	if deploymentReady {
		tender.Status.SetCondition(operatorv1alpha1.TenderCondition{
			Type:    operatorv1alpha1.TenderReady,
			Status:  operatorv1alpha1.ConditionTrue,
			Reason:  "ComponentsReady",
			Message: "all components report ready",
		})
	} else {
		tender.Status.SetCondition(operatorv1alpha1.TenderCondition{
			Type:    operatorv1alpha1.TenderReady,
			Status:  operatorv1alpha1.ConditionFalse,
			Reason:  "ComponentsUnready",
			Message: "Tender deployment isn't ready",
		})
	}

	if err = r.Status().Update(ctx, tender); err != nil {
		return ctrl.Result{}, fmt.Errorf("updating Tender Status: %w", err)
	}
	return ctrl.Result{}, nil
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

	// 2. Delete Objects.
	allCleared := true
	ownedBy, err := util.OwnedBy(tender, r.Scheme)
	if err != nil {
		return fmt.Errorf("getting ownedBy list option: %w", err)
	}

	for _, obj := range tenderControllerObjects {
		if err := r.Client.DeleteAllOf(ctx, obj, ownedBy); err != nil {
			return fmt.Errorf("deleteAllOf: %w", err)
		}
		gvk, err := apiutil.GVKForObject(obj, r.Scheme)
		if err != nil {
			return fmt.Errorf("gvk: %w", err)
		}

		lst := &unstructured.UnstructuredList{}
		lst.SetGroupVersionKind(gvk)
		if err := r.Client.List(ctx, lst, ownedBy); err != nil {
			return fmt.Errorf("list: %w", err)
		}
		if len(lst.Items) > 0 {
			allCleared = false
		}
	}

	if allCleared {
		// Remove Finalizer
		if util.RemoveFinalizer(tender, tenderControllerFinalizer) {
			if err := r.Update(ctx, tender); err != nil {
				return fmt.Errorf("updating Tender: %v", err)
			}
		}
	}
	return nil
}

func (r *TenderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	owner := &operatorv1alpha1.Tender{}
	ownerHandler, err := util.EnqueueRequestForOwner(owner, r.Scheme)
	if err != nil {
		return fmt.Errorf("create owner handler: %w", err)
	}

	cm := ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.Tender{})

	for _, obj := range tenderControllerObjects {
		cm = cm.Watches(&source.Kind{Type: obj}, ownerHandler)
	}

	return cm.Complete(r)
}

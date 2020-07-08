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
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	resourceferry "github.com/kubermatic/kubecarrier/pkg/internal/resources/ferry"
)

// FerryReconciler reconciles a Ferry object
type FerryReconciler struct {
	client.Client
	Log        logr.Logger
	Scheme     *runtime.Scheme
	RESTMapper meta.RESTMapper
}

// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete;escalate;bind
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=ferries,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=ferries/status,verbs=get;update;patch

func (r *FerryReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("ferry", req.NamespacedName)

	ferry := &operatorv1alpha1.Ferry{}
	if err := r.Get(ctx, req.NamespacedName, ferry); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if ferry.Spec.Paused.IsPaused() {
		if ferry.SetPausedCondition() {
			if err := r.Client.Status().Update(ctx, ferry); err != nil {
				return ctrl.Result{}, fmt.Errorf("updating %s status: %w", ferry.Name, err)
			}
		}
		// reconciliation paused, skip all other handlers
		return ctrl.Result{}, nil
	}

	if !ferry.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, ferry); err != nil {
			return ctrl.Result{}, fmt.Errorf("handle deletion: %w", err)
		}
		return ctrl.Result{}, nil
	}

	objects, err := resourceferry.Manifests(
		resourceferry.Config{
			ProviderNamespace:    ferry.Namespace,
			Name:                 ferry.Name,
			KubeconfigSecretName: ferry.Spec.KubeconfigSecret.Name,
		})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("creating Ferry manifests: %w", err)
	}

	deploymentIsReady, err := reconcileOwnedObjectsForNamespacedOwner(ctx, log, r.Scheme, r.RESTMapper, r.Client, ferry, objects)
	if err != nil {
		return ctrl.Result{}, err
	}

	// 4. Update the status of the Ferry object.
	if err := r.updateStatus(ctx, ferry, deploymentIsReady); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *FerryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.Ferry{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&corev1.ServiceAccount{}).
		Complete(r)
}

func (r *FerryReconciler) handleDeletion(ctx context.Context, ferry *operatorv1alpha1.Ferry) error {
	if ferry.SetTerminatingCondition() {
		if err := r.Client.Status().Update(ctx, ferry); err != nil {
			return fmt.Errorf("updating %s status: %w", ferry.Name, err)
		}
	}
	return nil
}

func (r *FerryReconciler) updateStatus(ctx context.Context, ferry *operatorv1alpha1.Ferry, deploymentIsReady bool) error {
	var pausedChanged, readyChanged bool

	if !ferry.Spec.Paused.IsPaused() {
		pausedChanged = ferry.SetUnPausedCondition()
	}

	if deploymentIsReady {
		readyChanged = ferry.SetReadyCondition()
	} else {
		readyChanged = ferry.SetUnReadyCondition()
	}

	if readyChanged || pausedChanged {
		if err := r.Client.Status().Update(ctx, ferry); err != nil {
			return fmt.Errorf("updating %s status: %w", ferry.Name, err)
		}
	}
	return nil
}

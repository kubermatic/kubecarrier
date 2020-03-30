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
	"sigs.k8s.io/controller-runtime/pkg/source"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/owner"
	resourcetower "github.com/kubermatic/kubecarrier/pkg/internal/resources/tower"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

const (
	towerControllerFinalizer = "tower.kubecarrier.io/controller"
)

// TowerReconciler reconciles a Tower object
type TowerReconciler struct {
	client.Client
	Log        logr.Logger
	Scheme     *runtime.Scheme
	RESTMapper meta.RESTMapper
}

// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=towers,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=towers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;create;update;patch;delete;escalate;bind
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete;escalate;bind
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete

func (r *TowerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("tower", req.NamespacedName)

	tower := &operatorv1alpha1.Tower{}
	if err := r.Get(ctx, req.NamespacedName, tower); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !tower.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, tower); err != nil {
			return ctrl.Result{}, fmt.Errorf("handle deletion: %w", err)
		}
		return ctrl.Result{}, nil
	}

	if util.AddFinalizer(tower, towerControllerFinalizer) {
		if err := r.Update(ctx, tower); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating Tower finalizers: %w", err)
		}
	}

	objects, err := resourcetower.Manifests(
		resourcetower.Config{
			Name:      tower.Name,
			Namespace: tower.Namespace,
		})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("creating Tower manifests: %w", err)
	}

	deploymentIsReady, err := reconcileOwnedObjectsForNamespacedOwner(ctx, log, r.Scheme, r.RESTMapper, r.Client, tower, objects)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := r.updateStatus(ctx, tower, deploymentIsReady); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *TowerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	enqueuer := owner.EnqueueRequestForOwner(&operatorv1alpha1.Tower{}, r.Scheme)

	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.Tower{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Watches(&source.Kind{Type: &rbacv1.ClusterRole{}}, enqueuer).
		Watches(&source.Kind{Type: &rbacv1.ClusterRoleBinding{}}, enqueuer).
		Complete(r)
}

func (r *TowerReconciler) handleDeletion(ctx context.Context, tower *operatorv1alpha1.Tower) error {
	if tower.SetTerminatingCondition() {
		if err := r.Client.Status().Update(ctx, tower); err != nil {
			return fmt.Errorf("updating %s status: %w", tower.Name, err)
		}
	}

	cleanedUp, err := util.DeleteObjects(ctx, r.Client, r.Scheme, []runtime.Object{
		&rbacv1.ClusterRole{},
		&rbacv1.ClusterRoleBinding{},
	}, owner.OwnedBy(tower, r.Scheme))
	if err != nil {
		return fmt.Errorf("DeleteObjects: %w", err)
	}

	if cleanedUp && util.RemoveFinalizer(tower, towerControllerFinalizer) {
		if err := r.Update(ctx, tower); err != nil {
			return fmt.Errorf("updating Tower finalizers: %w", err)
		}
	}
	return nil
}

func (r *TowerReconciler) updateStatus(ctx context.Context, tower *operatorv1alpha1.Tower, deploymentIsReady bool) error {
	var statusChanged bool

	if deploymentIsReady {
		statusChanged = tower.SetReadyCondition()
	} else {
		statusChanged = tower.SetUnReadyCondition()
	}

	if statusChanged {
		if err := r.Client.Status().Update(ctx, tower); err != nil {
			return fmt.Errorf("updating %s status: %w", tower.Name, err)
		}
	}
	return nil
}

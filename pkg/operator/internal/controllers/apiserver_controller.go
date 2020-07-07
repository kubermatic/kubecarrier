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
	"github.com/kubermatic/kubecarrier/pkg/internal/resources/apiserver"
	"github.com/kubermatic/utils/pkg/owner"
	"github.com/kubermatic/utils/pkg/util"
)

const (
	apiServerControllerFinalizer = "apiserver.kubecarrier.io/controller"
)

type APIServerReconciler struct {
	client.Client
	Log        logr.Logger
	Scheme     *runtime.Scheme
	RESTMapper meta.RESTMapper
}

// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=apiservers,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=apiservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;create;update;patch;delete;escalate;bind
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete;escalate;bind
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=issuers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete

func (r *APIServerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("apiserver", req.NamespacedName)

	apiServer := &operatorv1alpha1.APIServer{}
	if err := r.Get(ctx, req.NamespacedName, apiServer); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !apiServer.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, apiServer); err != nil {
			return ctrl.Result{}, fmt.Errorf("handle deletion: %w", err)
		}
		return ctrl.Result{}, nil
	}

	if util.AddFinalizer(apiServer, apiServerControllerFinalizer) {
		if err := r.Update(ctx, apiServer); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating APIServer finalizers: %w", err)
		}
	}

	objects, err := apiserver.Manifests(
		apiserver.Config{
			Name:      apiServer.Name,
			Namespace: apiServer.Namespace,
			Spec:      apiServer.Spec,
		},
	)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("creating APIServer manifests: %w", err)
	}

	deploymentIsReady, err := reconcileOwnedObjectsForNamespacedOwner(ctx, log, r.Scheme, r.RESTMapper, r.Client, apiServer, objects)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := r.updateStatus(ctx, apiServer, deploymentIsReady); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil

}

func (r *APIServerReconciler) updateStatus(ctx context.Context, apiServer *operatorv1alpha1.APIServer, deploymentIsReady bool) error {
	var statusChanged bool

	if deploymentIsReady {
		statusChanged = apiServer.SetReadyCondition()
	} else {
		statusChanged = apiServer.SetUnReadyCondition()
	}

	if statusChanged {
		if err := r.Client.Status().Update(ctx, apiServer); err != nil {
			return fmt.Errorf("updating %s status: %w", apiServer.Name, err)
		}
	}
	return nil
}

func (r *APIServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	enqueuer := owner.EnqueueRequestForOwner(&operatorv1alpha1.APIServer{}, r.Scheme)
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.APIServer{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Watches(&source.Kind{Type: &rbacv1.ClusterRole{}}, enqueuer).
		Watches(&source.Kind{Type: &rbacv1.ClusterRoleBinding{}}, enqueuer).
		Complete(r)
}

func (r *APIServerReconciler) handleDeletion(ctx context.Context, apiServer *operatorv1alpha1.APIServer) error {
	if apiServer.SetTerminatingCondition() {
		if err := r.Client.Status().Update(ctx, apiServer); err != nil {
			return fmt.Errorf("updating %s status: %w", apiServer.Name, err)
		}
	}

	cleanedUp, err := util.DeleteObjects(ctx, r.Client, r.Scheme, []runtime.Object{
		&rbacv1.ClusterRole{},
		&rbacv1.ClusterRoleBinding{},
	}, owner.OwnedBy(apiServer, r.Scheme))
	if err != nil {
		return fmt.Errorf("DeleteObjects: %w", err)
	}

	if cleanedUp && util.RemoveFinalizer(apiServer, apiServerControllerFinalizer) {
		if err := r.Update(ctx, apiServer); err != nil {
			return fmt.Errorf("updating APIServer finalizers: %w", err)
		}
	}
	return nil
}

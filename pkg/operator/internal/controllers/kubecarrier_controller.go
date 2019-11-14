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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubermatic/kubecarrier/pkg/internal/kustomize"
	"github.com/kubermatic/kubecarrier/pkg/internal/reconcile"
	"github.com/kubermatic/kubecarrier/pkg/internal/resources/manager"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
)

// KubeCarrierReconciler reconciles a KubeCarrier object
type KubeCarrierReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=kubecarriers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=kubecarriers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;create;update;patch;delete;escalate;bind
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete;escalate;bind
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete

func (r *KubeCarrierReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("kubecarrier", req.NamespacedName)

	kubeCarrier := &operatorv1alpha1.KubeCarrier{}
	if err := r.Get(ctx, req.NamespacedName, kubeCarrier); err != nil {
		// If the KubeCarrier object is already gone, we just ignore the NotFound error.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !kubeCarrier.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	objects, err := manager.Manifests(
		kustomize.NewDefaultKustomize(),
		manager.Config{
			Namespace: kubeCarrier.Namespace,
		})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("creating manager manifests: %w", err)
	}

	for _, object := range objects {
		_, err := reconcile.Unstructured(ctx, log, r.Client, &object)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("reconcile kind: %s, err: %w", object.GroupVersionKind().Kind, err)
		}
	}

	return ctrl.Result{}, nil
}

func (r *KubeCarrierReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.KubeCarrier{}).
		Complete(r)
}

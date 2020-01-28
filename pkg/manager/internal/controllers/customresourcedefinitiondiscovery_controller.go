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

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

const crdDiscoveryControllerFinalizer string = "custormresourcedefinitiondiscovery.kubecarrier.io/manager-controller"

// CustomResourceDefinitionDiscoveryReconciler reconciles a CustomResourceDefinitionDiscovery object
type CustomResourceDefinitionDiscoveryReconciler struct {
	Log logr.Logger

	Client client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=kubecarrier.io,resources=customresourcedefinitiondiscoveries,verbs=get;list;watch
// +kubebuilder:rbac:groups=kubecarrier.io,resources=customresourcedefinitiondiscoveries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch;update;create;patch;delete

func (r *CustomResourceDefinitionDiscoveryReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("crddiscovery", req.NamespacedName)

	crdDiscovery := &corev1alpha1.CustomResourceDefinitionDiscovery{}
	if err := r.Client.Get(ctx, req.NamespacedName, crdDiscovery); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	crdDiscovery.Status.ObservedGeneration = crdDiscovery.Generation

	if !crdDiscovery.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, log, crdDiscovery); err != nil {
			return ctrl.Result{}, fmt.Errorf("handling deletion: %w", err)
		}
		return ctrl.Result{}, nil
	}

	if util.AddFinalizer(crdDiscovery, crdDiscoveryControllerFinalizer) {
		if err := r.Client.Update(ctx, crdDiscovery); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating CustomResourceDefinitionDiscovery finalizers: %w", err)
		}
	}

	return ctrl.Result{}, nil
}

func (r *CustomResourceDefinitionDiscoveryReconciler) handleDeletion(ctx context.Context, log logr.Logger, crdDiscovery *corev1alpha1.CustomResourceDefinitionDiscovery) error {
	if util.RemoveFinalizer(crdDiscovery, crdDiscoveryControllerFinalizer) {
		if err := r.Client.Update(ctx, crdDiscovery); err != nil {
			return fmt.Errorf("updating CustomResourceDefinitionDiscovery finalizers: %w", err)
		}
	}
	return nil
}

func (r *CustomResourceDefinitionDiscoveryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.CustomResourceDefinitionDiscovery{}).
		Complete(r)
}

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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/owner"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

const crDiscoveryControllerFinalizer string = "crdiscovery.kubecarrier.io/ferry"

var (
	CRDNotFound = fmt.Errorf("CRDNotFound")
)

// CustomResourceDiscoveryReconciler reconciles a CustomResourceDiscovery object
type CustomResourceDiscoveryReconciler struct {
	Log logr.Logger

	ManagementClient client.Client
	ManagementScheme *runtime.Scheme
	ServiceClient    client.Client
	ServiceCache     cache.Cache

	ServiceClusterName string
}

// +kubebuilder:rbac:groups=kubecarrier.io,resources=customresourcediscoveries,verbs=get;list;watch;patch;update
// +kubebuilder:rbac:groups=kubecarrier.io,resources=customresourcediscoveries/status,verbs=get;update;patch
// Service cluster permission for this controller
// https://github.com/kubermatic/kubecarrier/issues/143
// +servicecluster:kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch;update

func (r *CustomResourceDiscoveryReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("crdiscovery", req.NamespacedName)

	crDiscovery := &corev1alpha1.CustomResourceDiscovery{}
	if err := r.ManagementClient.Get(ctx, req.NamespacedName, crDiscovery); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	crDiscovery.Status.ObservedGeneration = crDiscovery.Generation

	if !crDiscovery.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, log, crDiscovery); err != nil {
			return ctrl.Result{}, fmt.Errorf("handling deletion: %w", err)
		}
		return ctrl.Result{}, nil
	}

	if util.AddFinalizer(crDiscovery, crDiscoveryControllerFinalizer) {
		if err := r.ManagementClient.Update(ctx, crDiscovery); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating CustomResourceDiscovery finalizers: %w", err)
		}
	}

	// Lookup CRD
	crd := &apiextensionsv1.CustomResourceDefinition{}
	err := r.ServiceClient.Get(ctx, types.NamespacedName{
		Name: crDiscovery.Spec.CRD.Name,
	}, crd)

	switch {
	case errors.IsNotFound(err):
		crDiscovery.Status.CRD = nil
		crDiscovery.Status.SetCondition(corev1alpha1.CustomResourceDiscoveryCondition{
			Type:    corev1alpha1.CustomResourceDiscoveryDiscovered,
			Status:  corev1alpha1.ConditionFalse,
			Message: err.Error(),
			Reason:  CRDNotFound.Error(),
		})
		if err = r.ManagementClient.Status().Update(ctx, crDiscovery); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating CustomResourceDiscovery Status - notFound: %w", err)
		}
		// requeue until the CRD is found
		return ctrl.Result{Requeue: true}, nil
	case err == nil:
		// Add owner ref on CRD in the service cluster
		changed, err := owner.SetOwnerReference(crDiscovery, crd, r.ManagementScheme)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("setting owner reference: %w", err)
		}
		if changed {
			if err := r.ServiceClient.Update(ctx, crd); err != nil {
				return ctrl.Result{}, fmt.Errorf("updating CRD: %w", err)
			}
		}

		crDiscovery.Status.CRD = crd
		crDiscovery.Status.SetCondition(corev1alpha1.CustomResourceDiscoveryCondition{
			Type:    corev1alpha1.CustomResourceDiscoveryDiscovered,
			Status:  corev1alpha1.ConditionTrue,
			Message: "CRD was found on the cluster.",
			Reason:  "CRDFound",
		})
		if err = r.ManagementClient.Status().Update(ctx, crDiscovery); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating CustomResourceDiscovery Status -- ready: %w", err)
		}
		return ctrl.Result{}, nil
	default:
		return ctrl.Result{}, fmt.Errorf("getting CRD: %w", err)
	}
}

func (r *CustomResourceDiscoveryReconciler) handleDeletion(ctx context.Context, log logr.Logger, crDiscovery *corev1alpha1.CustomResourceDiscovery) error {
	crd := &apiextensionsv1.CustomResourceDefinition{}
	err := r.ServiceClient.Get(ctx, types.NamespacedName{
		Name: crDiscovery.Spec.CRD.Name,
	}, crd)
	switch {
	case errors.IsNotFound(err):
		if util.RemoveFinalizer(crDiscovery, crDiscoveryControllerFinalizer) {
			if err := r.ManagementClient.Update(ctx, crDiscovery); err != nil {
				return fmt.Errorf("updating CustomResourceDiscovery finalizers: %w", err)
			}
		}
		return nil
	case err == nil:
		// CRD still exists, ensure we're not owning it anymore
		if owner.RemoveOwnerReference(crDiscovery, crd) {
			if err = r.ServiceClient.Update(ctx, crd); err != nil {
				return fmt.Errorf("updating CRD: %w", err)
			}
		}
		if util.RemoveFinalizer(crDiscovery, crDiscoveryControllerFinalizer) {
			if err := r.ManagementClient.Update(ctx, crDiscovery); err != nil {
				return fmt.Errorf("updating CustomResourceDiscovery finalizers: %w", err)
			}
		}
		return nil
	default:
		return fmt.Errorf("getting CRD: %w", err)
	}
}

func (r *CustomResourceDiscoveryReconciler) SetupWithManager(managementMgr ctrl.Manager) error {
	crdSource := &source.Kind{Type: &apiextensionsv1.CustomResourceDefinition{}}
	if err := crdSource.InjectCache(r.ServiceCache); err != nil {
		return fmt.Errorf("injecting cache: %w", err)
	}

	return ctrl.NewControllerManagedBy(managementMgr).
		For(&corev1alpha1.CustomResourceDiscovery{}).
		Watches(source.Func(crdSource.Start), owner.EnqueueRequestForOwner(&corev1alpha1.CustomResourceDiscovery{}, r.ManagementScheme)).
		WithEventFilter(util.PredicateFn(func(obj runtime.Object) bool {
			if crDiscovery, ok := obj.(*corev1alpha1.CustomResourceDiscovery); ok {
				// We are just interested in CustomResourceDiscovery objects assigned to this ServiceCluster.
				if crDiscovery.Spec.ServiceCluster.Name != r.ServiceClusterName {
					return false
				}
			}

			// We don't want to filter CustomResourceDefinition events
			return true
		})).
		Complete(r)
}

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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

const crdReferenceControllerFinalizer string = "crdreference.kubecarrier.io/controller"

// CustomResourceDefinitionDiscoveryReconciler reconciles a CustomResourceDefinitionDiscovery object
type CustomResourceDefinitionDiscoveryReconciler struct {
	Log logr.Logger

	MasterClient  client.Client
	ServiceClient client.Client

	MasterScheme       *runtime.Scheme
	ServiceClusterName string
}

// +kubebuilder:rbac:groups=kubecarrier.io,resources=customresourcedefinitiondiscoveries,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=kubecarrier.io,resources=customresourcedefinitiondiscoveries/status,verbs=get;update;patch
// TODO: Figure out what to do for service cluster permissions?
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch;update

func (r *CustomResourceDefinitionDiscoveryReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("crdreference", req.NamespacedName)

	crdDiscovery := &corev1alpha1.CustomResourceDefinitionDiscovery{}
	if err := r.MasterClient.Get(ctx, req.NamespacedName, crdDiscovery); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	crdDiscovery.Status.ObservedGeneration = crdDiscovery.Generation

	if !crdDiscovery.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, log, crdDiscovery); err != nil {
			return ctrl.Result{}, fmt.Errorf("handling deletion: %w", err)
		}
		return ctrl.Result{}, nil
	}

	if util.AddFinalizer(crdDiscovery, crdReferenceControllerFinalizer) {
		if err := r.MasterClient.Update(ctx, crdDiscovery); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating CustomResourceDefinitionDiscovery finalizers: %w", err)
		}
	}

	// Lookup CRD
	crd := &apiextensionsv1.CustomResourceDefinition{}
	err := r.ServiceClient.Get(ctx, types.NamespacedName{
		Name: crdDiscovery.Spec.CRD.Name,
	}, crd)

	switch {
	case errors.IsNotFound(err):
		crdDiscovery.Status.CRD = nil
		crdDiscovery.Status.SetCondition(corev1alpha1.CustomResourceDefinitionDiscoveryCondition{
			Type:    corev1alpha1.CustomResourceDefinitionDiscoveryReady,
			Status:  corev1alpha1.ConditionFalse,
			Message: err.Error(),
			Reason:  util.ErrorReason(err),
		})
		if err = r.MasterClient.Status().Update(ctx, crdDiscovery); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating CustomResourceDefinitionDiscovery Status - notFound: %w", err)
		}
		// requeue until the CRD is found
		return ctrl.Result{Requeue: true}, nil
	case err == nil:
		// Add owner ref on CRD in the service cluster
		changed, err := util.InsertOwnerReference(crd, crdDiscovery, r.MasterScheme)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("inserting OwnerReference: %w", err)
		}
		if changed {
			if err := r.ServiceClient.Update(ctx, crd); err != nil {
				return ctrl.Result{}, fmt.Errorf("updating CRD: %w", err)
			}
		}

		crdDiscovery.Status.CRD = crd
		crdDiscovery.Status.SetCondition(corev1alpha1.CustomResourceDefinitionDiscoveryCondition{
			Type:    corev1alpha1.CustomResourceDefinitionDiscoveryReady,
			Status:  corev1alpha1.ConditionTrue,
			Message: "CRD was found on the cluster.",
			Reason:  "CRDFound",
		})
		if err = r.MasterClient.Status().Update(ctx, crdDiscovery); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating CustomResourceDefinitionDiscovery Status -- ready: %w", err)
		}
		return ctrl.Result{}, nil
	default:
		return ctrl.Result{}, fmt.Errorf("getting CRD: %w", err)
	}
}

func (r *CustomResourceDefinitionDiscoveryReconciler) handleDeletion(ctx context.Context, log logr.Logger, crdReference *corev1alpha1.CustomResourceDefinitionDiscovery) error {
	crd := &apiextensionsv1.CustomResourceDefinition{}
	err := r.ServiceClient.Get(ctx, types.NamespacedName{
		Name: crdReference.Spec.CRD.Name,
	}, crd)
	switch {
	case errors.IsNotFound(err):
		if util.RemoveFinalizer(crdReference, crdReferenceControllerFinalizer) {
			if err := r.MasterClient.Update(ctx, crdReference); err != nil {
				return fmt.Errorf("updating CustomResourceDefinitionDiscovery finalizers: %w", err)
			}
		}
		return nil
	case err == nil:
		// CRD still exists, ensure we're not owning it anymore
		changed, err := util.DeleteOwnerReference(crd, crdReference, r.MasterScheme)
		if err != nil {
			return fmt.Errorf("deleting OwnerReference: %w", err)
		}
		if changed {
			if err = r.ServiceClient.Update(ctx, crd); err != nil {
				return fmt.Errorf("updating CRD: %w", err)
			}
		}
		if util.RemoveFinalizer(crdReference, crdReferenceControllerFinalizer) {
			if err := r.MasterClient.Update(ctx, crdReference); err != nil {
				return fmt.Errorf("updating CustomResourceDefinitionDiscovery finalizers: %w", err)
			}
		}
		return nil
	default:
		return fmt.Errorf("getting CRD: %w", err)
	}
}

func (r *CustomResourceDefinitionDiscoveryReconciler) SetupWithManagers(serviceMgr, masterMgr ctrl.Manager) error {
	crdSource := &source.Kind{Type: &apiextensionsv1.CustomResourceDefinition{}}
	if err := serviceMgr.SetFields(crdSource); err != nil {
		return fmt.Errorf("setFields: %w", err)
	}

	enqueuer, err := util.EnqueueRequestForOwner(&corev1alpha1.CustomResourceDefinitionDiscovery{}, r.MasterScheme)
	if err != nil {
		return fmt.Errorf("creating CustomResourceDefinitionDiscovery enqueuer: %w", err)
	}

	return ctrl.NewControllerManagedBy(masterMgr).
		For(&corev1alpha1.CustomResourceDefinitionDiscovery{}).
		Watches(source.Func(crdSource.Start), enqueuer).
		WithEventFilter(util.PredicateFn(func(obj runtime.Object) bool {
			if crdReference, ok := obj.(*corev1alpha1.CustomResourceDefinitionDiscovery); ok {
				if crdReference.Spec.ServiceCluster.Name == r.ServiceClusterName {
					return true
				}
			}
			return false
		})).
		Complete(r)
}

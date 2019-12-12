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
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
)

// CatalogEntryReconciler reconciles a CatalogEntry object
type CatalogEntryReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=catalogentries,verbs=get;list;watch;update;
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=catalogentries/status,verbs=get;update;patch

// Reconcile function reconciles the CatalogEntry object which specified by the request. Currently, it does the following:
func (r *CatalogEntryReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("catalogEntry", req.NamespacedName)

	// 1. Fetch the CatalogEntry object.
	catalogEntry := &catalogv1alpha1.CatalogEntry{}
	if err := r.Get(ctx, req.NamespacedName, catalogEntry); err != nil {
		// If the CatalogEntry object is already gone, we just ignore the NotFound error.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 2. Handle the deletion of the CatalogEntry object.
	if !catalogEntry.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, log, catalogEntry); err != nil {
			return ctrl.Result{}, fmt.Errorf("handling deletion: %w", err)
		}
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *CatalogEntryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&catalogv1alpha1.CatalogEntry{}).
		Complete(r)
}

func (r *CatalogEntryReconciler) handleDeletion(ctx context.Context, log logr.Logger, catalogEntry *catalogv1alpha1.CatalogEntry) error {
	// Update the CatalogEntry Status to Terminating.
	readyCondition, _ := catalogEntry.Status.GetCondition(catalogv1alpha1.CatalogEntryReady)
	if readyCondition.Status != catalogv1alpha1.ConditionFalse ||
		readyCondition.Status == catalogv1alpha1.ConditionFalse && readyCondition.Reason != catalogv1alpha1.CatalogEntryTerminatingReason {
		catalogEntry.Status.ObservedGeneration = catalogEntry.Generation
		catalogEntry.Status.SetCondition(catalogv1alpha1.CatalogEntryCondition{
			Type:    catalogv1alpha1.CatalogEntryReady,
			Status:  catalogv1alpha1.ConditionFalse,
			Reason:  catalogv1alpha1.CatalogEntryTerminatingReason,
			Message: "CatalogEntry is being terminated",
		})
		if err := r.Status().Update(ctx, catalogEntry); err != nil {
			return fmt.Errorf("updating CatalogEntry status: %w", err)
		}
	}
	return nil
}

func (r *CatalogEntryReconciler) reconcile(ctx context.Context, log logr.Logger, catalogEntry *catalogv1alpha1.CatalogEntry) error {
	crdSelector, err := metav1.LabelSelectorAsSelector(catalogEntry.Spec.CRDSelector)
	if err != nil {
		return fmt.Errorf("crd selector: %w", err)
	}
	crdList := &apiextensionsv1beta1.CustomResourceDefinitionList{}
	err := r.Client.List(ctx, crdList, client.MatchingLabelsSelector{Selector: crdSelector})
	if err == nil {
		for _, crd := range crdList.Items {
			crdInfo := catalogv1alpha1.CRDInformation{}
			catalogEntry.Status.CRDs = append(catalogEntry.Status.CRDs)
		}
	}
	return nil
}

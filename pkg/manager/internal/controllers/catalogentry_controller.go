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
	"strings"

	"github.com/go-logr/logr"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
)

const (
	// This annotation is used to make sure a CRD can only be referenced by a single CatalogEntry object.
	catalogEntryReferenceAnnotation = "kubecarrier.io/catalog-entry"

	serviceClusterAnnotation = "kubecarrier.io/service-cluster"
)

// CatalogEntryReconciler reconciles a CatalogEntry object
type CatalogEntryReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=catalogentries,verbs=get;list;watch;update;
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=catalogentries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch

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

	namespacedName := req.NamespacedName.String()
	if err := r.manipulateCRDInfo(ctx, log, catalogEntry, namespacedName); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconciling CRDInfo: %w", err)
	}

	return ctrl.Result{}, nil
}

func (r *CatalogEntryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&catalogv1alpha1.CatalogEntry{}).
		Watches(&source.Kind{Type: &apiextensionsv1beta1.CustomResourceDefinition{}}, &handler.EnqueueRequestsFromMapFunc{
			ToRequests: handler.ToRequestsFunc(func(o handler.MapObject) (requests []reconcile.Request) {
				annotations := o.Meta.GetAnnotations()
				if annotations == nil {
					return
				}
				if val, present := annotations[catalogEntryReferenceAnnotation]; !present {
					namespacedName := strings.SplitN(val, string(types.Separator), 2)
					requests = append(requests, reconcile.Request{
						NamespacedName: types.NamespacedName{
							Name:      namespacedName[1],
							Namespace: namespacedName[0],
						},
					})
				}
				return
			}),
		}).
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

func (r *CatalogEntryReconciler) manipulateCRDInfo(ctx context.Context, log logr.Logger, catalogEntry *catalogv1alpha1.CatalogEntry, namespacedName string) error {
	crdSelector, err := metav1.LabelSelectorAsSelector(catalogEntry.Spec.CRDSelector)
	if err != nil {
		return fmt.Errorf("crd selector: %w", err)
	}

	var (
		crdReferencedOther int
		crdReferenced      int

		crdInfos []catalogv1alpha1.CRDInformation
	)

	crdList := &apiextensionsv1beta1.CustomResourceDefinitionList{}
	if err := r.Client.List(ctx, crdList, client.MatchingLabelsSelector{Selector: crdSelector}); err != nil {
		return fmt.Errorf("listing CRD: %w", err)
	}
	for _, crd := range crdList.Items {

		// check the annotation of the CRD
		catalogEntryName, present := crd.Annotations[catalogEntryReferenceAnnotation]
		if !present {
			// The reference annotation is not set yet, we just set it and update the CRD
			crd.Annotations[catalogEntryReferenceAnnotation] = namespacedName
			if err := r.Update(ctx, &crd); err != nil {
				return fmt.Errorf("updating CRD annotation: %w", err)
			}
			continue
		}

		// Since a CRD only can be referenced by one CatalogEntry, if this CRD is referenced by
		// another CatalogEntry, we just skip it.
		if catalogEntryName != namespacedName {
			crdReferencedOther++
			continue
		}
		crdReferenced++

		crdInfo := catalogv1alpha1.CRDInformation{
			Name:     crd.Name,
			APIGroup: crd.Spec.Group,
			Kind:     crd.Spec.Names.Kind,
		}

		for _, ver := range crd.Spec.Versions {
			schema := ver.Schema
			if schema == nil {
				schema = crd.Spec.Validation
			}
			crdInfo.Versions = append(crdInfo.Versions, catalogv1alpha1.CRDVersion{
				Name:   ver.Name,
				Schema: schema,
			})
		}

		// legacy Schema handling
		if crd.Spec.Version != "" && len(crd.Spec.Versions) == 0 {
			crdInfo.Versions = append(crdInfo.Versions, catalogv1alpha1.CRDVersion{
				Name:   crd.Spec.Version,
				Schema: crd.Spec.Validation,
			})
		}

		// Service Cluster
		serviceCluster, present := crd.Annotations[serviceClusterAnnotation]
		if !present {
			return fmt.Errorf("getting ServiceCluster of the CRD error: CRD should have an annotation to indicate the ServiceCluster")
		}

		crdInfo.ServiceCluster = catalogv1alpha1.ObjectReference{
			Name: serviceCluster,
		}

		crdInfos = append(crdInfos, crdInfo)
	}

	if crdReferencedOther+crdReferenced == len(crdList.Items) {
		catalogEntry.Status.CRDs = crdInfos
		catalogEntry.Status.ObservedGeneration = catalogEntry.Generation
		catalogEntry.Status.SetCondition(catalogv1alpha1.CatalogEntryCondition{
			Type:    catalogv1alpha1.CatalogEntryReady,
			Status:  catalogv1alpha1.ConditionTrue,
			Reason:  "CatalogEntryReady",
			Message: "CatalogEntry is Ready.",
		})
	} else {
		catalogEntry.Status.ObservedGeneration = catalogEntry.Generation
		catalogEntry.Status.SetCondition(catalogv1alpha1.CatalogEntryCondition{
			Type:    catalogv1alpha1.CatalogEntryReady,
			Status:  catalogv1alpha1.ConditionFalse,
			Reason:  "CatalogEntryNotReady",
			Message: "CatalogEntry is not Ready.",
		})
	}
	if err := r.Status().Update(ctx, catalogEntry); err != nil {
		return fmt.Errorf("updating Catalog Status: %w", err)
	}
	return nil
}

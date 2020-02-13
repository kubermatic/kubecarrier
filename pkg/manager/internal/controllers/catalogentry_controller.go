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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

const (
	// This annotation is used to make sure a ReferencedCRD can only be referenced by a single CatalogEntry object.
	catalogEntryReferenceAnnotation = "kubecarrier.io/catalog-entry"
	catalogEntryControllerFinalizer = "catalogentry.kubecarrier.io/controller"
)

// CatalogEntryReconciler reconciles a CatalogEntry object
type CatalogEntryReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=catalogentries,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=catalogentries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch;update

// Reconcile function reconciles the CatalogEntry object which specified by the request. Currently, it does the following:
// - Fetch the CatalogEntry object.
// - Handle the deletion of the CatalogEntry object (Remove the annotations from the ReferencedCRD, and remove the finalizer).
// - Manipulate/Update the CRDInformation in the CatalogEntry status.
// - Update the status of the CatalogEntry object.
func (r *CatalogEntryReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("catalogEntry", req.NamespacedName)
	namespacedName := req.NamespacedName.String()

	// Fetch the CatalogEntry object.
	catalogEntry := &catalogv1alpha1.CatalogEntry{}
	if err := r.Get(ctx, req.NamespacedName, catalogEntry); err != nil {
		// If the CatalogEntry object is already gone, we just ignore the NotFound error.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle the deletion of the CatalogEntry object.
	if !catalogEntry.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, log, catalogEntry, namespacedName); err != nil {
			return ctrl.Result{}, fmt.Errorf("handling deletion: %w", err)
		}
		return ctrl.Result{}, nil
	}

	if util.AddFinalizer(catalogEntry, catalogEntryControllerFinalizer) {
		// Update the CatalogEntry with the finalizer
		if err := r.Update(ctx, catalogEntry); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating finalizers: %w", err)
		}
	}

	crd := &apiextensionsv1.CustomResourceDefinition{}
	if err := r.Client.Get(ctx, types.NamespacedName{
		Name: catalogEntry.Spec.ReferencedCRD.Name,
	}, crd); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, r.updateStatus(ctx, catalogEntry, &catalogv1alpha1.CatalogEntryCondition{
				Type:    catalogv1alpha1.CatalogEntryReady,
				Status:  catalogv1alpha1.ConditionFalse,
				Reason:  "NotFound",
				Message: "The referenced ReferencedCRD was not found.",
			})
		}
		return ctrl.Result{}, fmt.Errorf("getting ReferencedCRD: %w", err)
	}
	if crd.Spec.Scope != apiextensionsv1.NamespaceScoped {
		return ctrl.Result{}, r.updateStatus(ctx, catalogEntry, &catalogv1alpha1.CatalogEntryCondition{
			Type:    catalogv1alpha1.CatalogEntryReady,
			Status:  catalogv1alpha1.ConditionFalse,
			Reason:  "NotNamespaced",
			Message: "The referenced CRD needs to Namespace scoped.",
		})
	}

	// check the annotation of the ReferencedCRD
	annotations := crd.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	if catalogEntryNamespacedName, ok := annotations[catalogEntryReferenceAnnotation]; ok && catalogEntryNamespacedName != namespacedName {
		// referenced by another instance
		return ctrl.Result{}, r.updateStatus(ctx, catalogEntry, &catalogv1alpha1.CatalogEntryCondition{
			Type:    catalogv1alpha1.CatalogEntryReady,
			Status:  catalogv1alpha1.ConditionFalse,
			Reason:  "AlreadyInUse",
			Message: fmt.Sprintf("The referenced ReferencedCRD is already referenced by %q.", catalogEntryNamespacedName),
		})
	} else if !ok {
		// not yet referenced
		annotations[catalogEntryReferenceAnnotation] = namespacedName
		crd.SetAnnotations(annotations)
		if err := r.Update(ctx, crd); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating ReferencedCRD annotation: %w", err)
		}
	}

	// lookup Provider
	provider, err := catalogv1alpha1.GetProviderByProviderNamespace(ctx, r.Client, catalogEntry.Namespace)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("getting the Provider by Provider Namespace: %w", err)
	}

	// check if Provider is allowed to use the CRD
	if crd.Labels == nil ||
		crd.Labels[ProviderLabel] != provider.Name {
		return ctrl.Result{}, r.updateStatus(ctx, catalogEntry, &catalogv1alpha1.CatalogEntryCondition{
			Type:    catalogv1alpha1.CatalogEntryReady,
			Status:  catalogv1alpha1.ConditionFalse,
			Reason:  "NotAssignedToProvider",
			Message: fmt.Sprintf("The referenced CRD not assigned to this Provider or is missing a %s label.", ProviderLabel),
		})
	}

	// lookup ServiceCluster
	if crd.Labels == nil ||
		crd.Labels[serviceClusterLabel] == "" {
		return ctrl.Result{}, r.updateStatus(ctx, catalogEntry, &catalogv1alpha1.CatalogEntryCondition{
			Type:    catalogv1alpha1.CatalogEntryReady,
			Status:  catalogv1alpha1.ConditionFalse,
			Reason:  "MissingServiceClusterLabel",
			Message: fmt.Sprintf("The referenced CRD is missing a %s label.", serviceClusterLabel),
		})
	}
	crdInfo, err := getCRDInformation(crd)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("getting CRD Info: %w", err)
	}

	catalogEntry.Status.CRD = crdInfo
	return ctrl.Result{}, r.updateStatus(ctx, catalogEntry, &catalogv1alpha1.CatalogEntryCondition{
		Type:    catalogv1alpha1.CatalogEntryReady,
		Status:  catalogv1alpha1.ConditionTrue,
		Reason:  "CatalogEntryReady",
		Message: "CatalogEntry is Ready.",
	})
}

func (r *CatalogEntryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&catalogv1alpha1.CatalogEntry{}).
		Watches(&source.Kind{Type: &apiextensionsv1.CustomResourceDefinition{}}, &handler.EnqueueRequestsFromMapFunc{
			ToRequests: handler.ToRequestsFunc(func(o handler.MapObject) (requests []reconcile.Request) {
				annotations := o.Meta.GetAnnotations()
				if annotations == nil {
					return
				}
				if val, present := annotations[catalogEntryReferenceAnnotation]; present {
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

func (r *CatalogEntryReconciler) handleDeletion(ctx context.Context, log logr.Logger, catalogEntry *catalogv1alpha1.CatalogEntry, desiredNamespacedName string) error {
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

	// Clean up annotation for the ReferencedCRD
	crd := &apiextensionsv1.CustomResourceDefinition{}
	err := r.Get(ctx, types.NamespacedName{
		Name: catalogEntry.Spec.ReferencedCRD.Name,
	}, crd)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("getting ReferencedCRD: %w", err)
	}
	if err == nil && crd.Annotations != nil {
		delete(crd.Annotations, catalogEntryReferenceAnnotation)
		if err := r.Update(ctx, crd); err != nil {
			return fmt.Errorf("updating ReferencedCRD: %w", err)
		}
	}

	if util.RemoveFinalizer(catalogEntry, catalogEntryControllerFinalizer) {
		if err := r.Update(ctx, catalogEntry); err != nil {
			return fmt.Errorf("updating CatalogEntry: %w", err)
		}
	}

	return nil
}

func getCRDInformation(crd *apiextensionsv1.CustomResourceDefinition) (catalogv1alpha1.CRDInformation, error) {
	crdInfo := catalogv1alpha1.CRDInformation{
		Name:     crd.Name,
		APIGroup: crd.Spec.Group,
		Kind:     crd.Spec.Names.Kind,
	}

	for _, ver := range crd.Spec.Versions {
		crdInfo.Versions = append(crdInfo.Versions, catalogv1alpha1.CRDVersion{
			Name:   ver.Name,
			Schema: ver.Schema,
		})
	}

	// Service Cluster
	serviceCluster, present := crd.Labels[serviceClusterLabel]
	if !present {
		return catalogv1alpha1.CRDInformation{}, fmt.Errorf("getting ServiceCluster of the ReferencedCRD error: ReferencedCRD should have an annotation to indicate the ServiceCluster")
	}

	crdInfo.ServiceCluster = catalogv1alpha1.ObjectReference{
		Name: serviceCluster,
	}

	return crdInfo, nil
}

func (r *CatalogEntryReconciler) updateStatus(
	ctx context.Context,
	catalogEntry *catalogv1alpha1.CatalogEntry,
	condition *catalogv1alpha1.CatalogEntryCondition,
) error {
	catalogEntry.Status.ObservedGeneration = catalogEntry.Generation
	if condition != nil {
		catalogEntry.Status.SetCondition(*condition)
	}
	if err := r.Status().Update(ctx, catalogEntry); err != nil {
		return fmt.Errorf("updating CatalogEntry status: %w", err)
	}
	return nil
}

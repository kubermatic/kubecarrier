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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	// This annotation is used to make sure a CRD can only be referenced by a single CatalogEntry object.
	catalogEntryReferenceAnnotation = "kubecarrier.io/catalog-entry"
	catalogEntryControllerFinalizer = "catalogentry.kubecarrier.io/controller"
	providerLabel                   = "kubecarrier.io/provider"
	serviceClusterAnnotation        = "kubecarrier.io/service-cluster"
)

// CatalogEntryReconciler reconciles a CatalogEntry object
type CatalogEntryReconciler struct {
	client.Client
	Log                        logr.Logger
	Scheme                     *runtime.Scheme
	KubeCarrierSystemNamespace string
}

// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=catalogentries,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=catalogentries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch;update

// Reconcile function reconciles the CatalogEntry object which specified by the request. Currently, it does the following:
// - Fetch the CatalogEntry object.
// - Handle the deletion of the CatalogEntry object (Remove the annotations from the CRDs, and remove the finalizer).
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

	provider, err := getProviderByProviderNamespace(ctx, r.Client, r.KubeCarrierSystemNamespace, catalogEntry.Namespace)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("getting the Provider by Provider Namespace: %w", err)
	}

	// Defaulting the `kubecarrier.io/provider` matchlabel
	if catalogEntry.Spec.CRDSelector == nil {
		catalogEntry.Spec.CRDSelector = &metav1.LabelSelector{}
	}
	if catalogEntry.Spec.CRDSelector.MatchLabels == nil {
		catalogEntry.Spec.CRDSelector.MatchLabels = map[string]string{}
	}
	if catalogEntry.Spec.CRDSelector.MatchLabels[providerLabel] != provider.Name {
		catalogEntry.Spec.CRDSelector.MatchLabels[providerLabel] = provider.Name
		if err := r.Update(ctx, catalogEntry); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating the provider matchLabel: %w", err)
		}
	}

	// Manipulate the CRD information to CatalogEntry status, and update the status of the CatalogEntry.
	if err := r.manipulateCRDInfo(ctx, log, catalogEntry, namespacedName); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconciling CRDInfo: %w", err)
	}

	return ctrl.Result{}, nil
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

	// Clean up annotations for the CRD
	crdSelector, err := metav1.LabelSelectorAsSelector(catalogEntry.Spec.CRDSelector)
	if err != nil {
		return fmt.Errorf("crd selector: %w", err)
	}

	var (
		crdReferencedOther   int
		crdReferencedCleanUp int
	)

	crdList := &apiextensionsv1.CustomResourceDefinitionList{}
	if err := r.Client.List(ctx, crdList, client.MatchingLabelsSelector{Selector: crdSelector}); err != nil {
		return fmt.Errorf("listing CRD: %w", err)
	}
	for _, crd := range crdList.Items {
		annotations := crd.GetAnnotations()
		if catalogEntryNamespacedName, present := annotations[catalogEntryReferenceAnnotation]; present {
			if catalogEntryNamespacedName == desiredNamespacedName {
				delete(annotations, catalogEntryReferenceAnnotation)
				crd.SetAnnotations(annotations)
				if err := r.Update(ctx, &crd); err != nil {
					return fmt.Errorf("updating CRD: %w", err)
				}
			} else {
				crdReferencedOther++
			}
		} else {
			crdReferencedCleanUp++
		}

	}

	if crdReferencedOther+crdReferencedCleanUp != len(crdList.Items) {
		return nil
	}

	if util.RemoveFinalizer(catalogEntry, catalogEntryControllerFinalizer) {
		if err := r.Update(ctx, catalogEntry); err != nil {
			return fmt.Errorf("updating CatalogEntry: %w", err)
		}
	}

	return nil
}

func (r *CatalogEntryReconciler) manipulateCRDInfo(ctx context.Context, log logr.Logger, catalogEntry *catalogv1alpha1.CatalogEntry, desiredNamespacedName string) error {
	crdSelector, err := metav1.LabelSelectorAsSelector(catalogEntry.Spec.CRDSelector)
	if err != nil {
		return fmt.Errorf("crd selector: %w", err)
	}

	var (
		crdReferencedOther int
		crdReferenced      int

		crdInfos []catalogv1alpha1.CRDInformation
	)

	crdList := &apiextensionsv1.CustomResourceDefinitionList{}
	if err := r.Client.List(ctx, crdList, client.MatchingLabelsSelector{Selector: crdSelector}); err != nil {
		return fmt.Errorf("listing CRD: %w", err)
	}
	for _, crd := range crdList.Items {

		// check the annotation of the CRD
		annotations := crd.GetAnnotations()
		if catalogEntryNamespacedName, present := annotations[catalogEntryReferenceAnnotation]; present {
			if catalogEntryNamespacedName == desiredNamespacedName {
				crdReferenced++
				crdInfo, err := getCRDInformation(crd)
				if err != nil {
					return err
				}
				crdInfos = append(crdInfos, crdInfo)

			} else {
				// Since a CRD only can be referenced by one CatalogEntry, if this CRD is referenced by
				// another CatalogEntry, we just skip it.
				crdReferencedOther++

			}
		} else {
			// The reference annotation is not set yet, we just set it and update the CRD
			annotations[catalogEntryReferenceAnnotation] = desiredNamespacedName
			crd.SetAnnotations(annotations)
			if err := r.Update(ctx, &crd); err != nil {
				return fmt.Errorf("updating CRD annotation: %w", err)
			}
		}
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
		return fmt.Errorf("updating CatalogEntry Status: %w", err)
	}
	return nil
}

func getCRDInformation(crd apiextensionsv1.CustomResourceDefinition) (catalogv1alpha1.CRDInformation, error) {
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
	serviceCluster, present := crd.Annotations[serviceClusterAnnotation]
	if !present {
		return catalogv1alpha1.CRDInformation{}, fmt.Errorf("getting ServiceCluster of the CRD error: CRD should have an annotation to indicate the ServiceCluster")
	}

	crdInfo.ServiceCluster = catalogv1alpha1.ObjectReference{
		Name: serviceCluster,
	}

	return crdInfo, nil
}

func getProviderByProviderNamespace(ctx context.Context, c client.Client, kubecarrierNamespace, providerNamespace string) (*catalogv1alpha1.Provider, error) {
	providerList := &catalogv1alpha1.ProviderList{}
	if err := c.List(ctx, providerList,
		client.InNamespace(kubecarrierNamespace),
		client.MatchingFields{
			catalogv1alpha1.ProviderNamespaceFieldIndex: providerNamespace,
		},
	); err != nil {
		return nil, err
	}
	switch len(providerList.Items) {
	case 0:
		// not found
		return nil, fmt.Errorf("providers.catalog.kubecarrier.io with index %q not found", catalogv1alpha1.ProviderNamespaceFieldIndex)
	case 1:
		// found!
		return &providerList.Items[0], nil
	default:
		// found too many
		return nil, fmt.Errorf("multiple providers.catalog.kubecarrier.io with index %q found", catalogv1alpha1.ProviderNamespaceFieldIndex)
	}
}

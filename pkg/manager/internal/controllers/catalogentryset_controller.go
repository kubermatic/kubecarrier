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
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/owner"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

const catalogEntriesLabel = "catalogentries.kubecarrier.io/controlled-by"

// CatalogEntrySetReconciler reconciles a CatalogEntrySet object
type CatalogEntrySetReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=catalogentrysets,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=catalogentrysets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=catalogentries,verbs=get;list;watch;update;create;delete
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=catalogentries/status,verbs=get
// +kubebuilder:rbac:groups=kubecarrier.io,resources=customresourcediscoverysets,verbs=get;watch;update;create;delete
// +kubebuilder:rbac:groups=kubecarrier.io,resources=customresourcediscoveries,verbs=list

func (r *CatalogEntrySetReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	catalogEntrySet := &catalogv1alpha1.CatalogEntrySet{}
	if err := r.Get(ctx, req.NamespacedName, catalogEntrySet); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if util.AddFinalizer(catalogEntrySet, metav1.FinalizerDeleteDependents) {
		if err := r.Client.Update(ctx, catalogEntrySet); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating CatalogEntrySet finalizers: %w", err)
		}
	}
	if !catalogEntrySet.DeletionTimestamp.IsZero() {
		// nothing to do, let kube controller-manager foregroundDeletion wait until every created object is deleted
		return ctrl.Result{}, nil
	}

	// Reconcile CustomResourceDiscoverySet object
	currentCustomResourceDiscoverySet, err := r.reconcileCustomResourceDiscoverySet(ctx, catalogEntrySet)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("reconciling CustomResourceDiscoverySet: %w", err)
	}

	if !currentCustomResourceDiscoverySet.IsReady() {
		if err := r.updateStatus(ctx, catalogEntrySet, catalogv1alpha1.CatalogEntrySetCondition{
			Type:    catalogv1alpha1.CustomResourceDiscoverySetReady,
			Status:  catalogv1alpha1.ConditionFalse,
			Reason:  "CustomResourceDiscoverySetUnready",
			Message: "CustomResourceDiscoverySet is unready.",
		}); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if err := r.updateStatus(ctx, catalogEntrySet, catalogv1alpha1.CatalogEntrySetCondition{
		Type:    catalogv1alpha1.CustomResourceDiscoverySetReady,
		Status:  catalogv1alpha1.ConditionTrue,
		Reason:  "CustomResourceDiscoverySetReady",
		Message: "CustomResourceDiscoverySet is ready.",
	}); err != nil {
		return ctrl.Result{}, err
	}

	// Reconcile CatalogEntry objects
	customResourceDiscoveryList := &corev1alpha1.CustomResourceDiscoveryList{}
	if err := r.List(ctx,
		customResourceDiscoveryList,
		owner.OwnedBy(currentCustomResourceDiscoverySet, r.Scheme),
	); err != nil {
		return ctrl.Result{}, fmt.Errorf("listing CustomResourceDiscovery: %w", err)
	}
	var unreadyCatalogEntryNames []string
	existingCatalogEntryNames := map[string]struct{}{}
	for _, customResourceDiscovery := range customResourceDiscoveryList.Items {
		currentCatalogEntry, err := r.reconcileCatalogEntry(ctx, &customResourceDiscovery, catalogEntrySet)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf(
				"reconciling CatalogEntry for CRD %s: %w", customResourceDiscovery.Status.CRD.Name, err)
		}
		existingCatalogEntryNames[currentCatalogEntry.Name] = struct{}{}

		if !currentCatalogEntry.IsReady() {
			unreadyCatalogEntryNames = append(unreadyCatalogEntryNames, currentCatalogEntry.Name)
		}
	}

	// Cleanup uncontrolled CatalogEntry objects
	catalogEntryList := &catalogv1alpha1.CatalogEntryList{}
	if err := r.List(ctx, catalogEntryList, client.MatchingLabels{
		catalogEntriesLabel: catalogEntrySet.Name,
	}); err != nil {
		return ctrl.Result{}, fmt.Errorf(
			"listing all CatalogEntry for this CatalogEntrySet: %w", err)
	}
	for _, catalogEntry := range catalogEntryList.Items {
		if _, ok := existingCatalogEntryNames[catalogEntry.Name]; ok {
			continue
		}

		// delete CatalogEntry objects that should no longer exist
		if err := r.Delete(ctx, &catalogEntry); err != nil {
			return ctrl.Result{}, fmt.Errorf("deleting CatalogEntry: %w", err)
		}
	}

	// Report status
	if len(unreadyCatalogEntryNames) > 0 {
		if err := r.updateStatus(ctx, catalogEntrySet, catalogv1alpha1.CatalogEntrySetCondition{
			Type:   catalogv1alpha1.CatalogEntriesReady,
			Status: catalogv1alpha1.ConditionFalse,
			Reason: "CatalogEntriesUnready",
			Message: fmt.Sprintf(
				"Some CatalogEntry objects are unready [%s]", strings.Join(unreadyCatalogEntryNames, ",")),
		}); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		if err := r.updateStatus(ctx, catalogEntrySet, catalogv1alpha1.CatalogEntrySetCondition{
			Type:    catalogv1alpha1.CatalogEntriesReady,
			Status:  catalogv1alpha1.ConditionTrue,
			Reason:  "CatalogEntriesReady",
			Message: "All CatalogEntry objects are ready.",
		}); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *CatalogEntrySetReconciler) reconcileCustomResourceDiscoverySet(
	ctx context.Context, catalogEntrySet *catalogv1alpha1.CatalogEntrySet,
) (*corev1alpha1.CustomResourceDiscoverySet, error) {
	desiredCustomResourceDiscoverySet := &corev1alpha1.CustomResourceDiscoverySet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      catalogEntrySet.Name,
			Namespace: catalogEntrySet.Namespace,
			Labels: map[string]string{
				catalogEntriesLabel: catalogEntrySet.Name,
			},
		},
		Spec: corev1alpha1.CustomResourceDiscoverySetSpec{
			CRD: corev1alpha1.ObjectReference{
				Name: catalogEntrySet.Spec.DiscoverySet.CRD.Name,
			},
			ServiceClusterSelector: catalogEntrySet.Spec.DiscoverySet.ServiceClusterSelector,
			KindOverride:           catalogEntrySet.Spec.DiscoverySet.KindOverride,
		},
	}
	err := controllerutil.SetControllerReference(catalogEntrySet, desiredCustomResourceDiscoverySet, r.Scheme)
	if err != nil {
		return nil, fmt.Errorf("set controller reference: %w", err)
	}

	currentCustomResourceDiscoverySet := &corev1alpha1.CustomResourceDiscoverySet{}
	err = r.Get(ctx, types.NamespacedName{
		Name:      desiredCustomResourceDiscoverySet.Name,
		Namespace: desiredCustomResourceDiscoverySet.Namespace,
	}, currentCustomResourceDiscoverySet)
	if err != nil && !errors.IsNotFound(err) {
		return nil, fmt.Errorf("getting CustromResourceDiscoverySet: %w", err)
	}

	if errors.IsNotFound(err) {
		// create CustomResourceDiscovery object
		if err := r.Create(ctx, desiredCustomResourceDiscoverySet); err != nil {
			return nil, fmt.Errorf("creating CustomResourceDiscoverySet: %w", err)
		}
		return desiredCustomResourceDiscoverySet, nil
	}

	// update existing
	if !reflect.DeepEqual(currentCustomResourceDiscoverySet.Spec, desiredCustomResourceDiscoverySet.Spec) {
		currentCustomResourceDiscoverySet.Spec = desiredCustomResourceDiscoverySet.Spec
		if err := r.Update(ctx, currentCustomResourceDiscoverySet); err != nil {
			return nil, fmt.Errorf("updating CustomResourceDiscoverySet: %w", err)
		}
	}
	return currentCustomResourceDiscoverySet, nil
}

func (r *CatalogEntrySetReconciler) reconcileCatalogEntry(
	ctx context.Context, customResourceDiscovery *corev1alpha1.CustomResourceDiscovery,
	catalogEntrySet *catalogv1alpha1.CatalogEntrySet,
) (*catalogv1alpha1.CatalogEntry, error) {
	crd, err := r.getCRDByCustomResourceDiscovery(ctx, customResourceDiscovery)
	if err != nil {
		return nil, fmt.Errorf("getting CRD that owned by CustomResourceDiscovery: %w", err)
	}
	desiredCatalogEntry := &catalogv1alpha1.CatalogEntry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      crd.Name,
			Namespace: catalogEntrySet.Namespace,
			Labels: map[string]string{
				catalogEntriesLabel: catalogEntrySet.Name,
			},
		},
		Spec: catalogv1alpha1.CatalogEntrySpec{
			Metadata: catalogv1alpha1.CatalogEntryMetadata{
				DisplayName: catalogEntrySet.Spec.Metadata.DisplayName,
				Description: catalogEntrySet.Spec.Metadata.Description,
			},
			BaseCRD: catalogv1alpha1.ObjectReference{
				Name: crd.Name,
			},
			Derive: catalogEntrySet.Spec.Derive,
		},
	}
	err = controllerutil.SetControllerReference(catalogEntrySet, desiredCatalogEntry, r.Scheme)
	if err != nil {
		return nil, fmt.Errorf("set controller reference: %w", err)
	}

	currentCatalogEntry := &catalogv1alpha1.CatalogEntry{}
	err = r.Get(ctx, types.NamespacedName{
		Name:      desiredCatalogEntry.Name,
		Namespace: desiredCatalogEntry.Namespace,
	}, currentCatalogEntry)
	if err != nil && !errors.IsNotFound(err) {
		return nil, fmt.Errorf("getting CatalogEntry: %w", err)
	}

	if errors.IsNotFound(err) {
		// create CatalogEntry object
		if err := r.Create(ctx, desiredCatalogEntry); err != nil {
			return nil, fmt.Errorf("creating CatalogEntry: %w", err)
		}
		return desiredCatalogEntry, nil
	}

	// update existing
	if !reflect.DeepEqual(currentCatalogEntry.Spec, desiredCatalogEntry.Spec) {
		currentCatalogEntry.Spec = desiredCatalogEntry.Spec
		if err := r.Update(ctx, currentCatalogEntry); err != nil {
			return nil, fmt.Errorf("updating CatalogEntry: %w", err)
		}
	}
	return currentCatalogEntry, nil
}

func (r *CatalogEntrySetReconciler) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(&catalogv1alpha1.CatalogEntrySet{}).
		Owns(&catalogv1alpha1.CatalogEntry{}).
		Owns(&corev1alpha1.CustomResourceDiscoverySet{}).
		Complete(r)
}

func (r *CatalogEntrySetReconciler) updateStatus(
	ctx context.Context,
	catalogEntrySet *catalogv1alpha1.CatalogEntrySet,
	condition catalogv1alpha1.CatalogEntrySetCondition,
) error {
	catalogEntrySet.Status.ObservedGeneration = catalogEntrySet.Generation
	catalogEntrySet.Status.SetCondition(condition)

	customResourceDiscoverySetReady, _ := catalogEntrySet.Status.GetCondition(
		catalogv1alpha1.CustomResourceDiscoverySetReady)
	catalogEntriesReady, _ := catalogEntrySet.Status.GetCondition(
		catalogv1alpha1.CatalogEntriesReady)

	if customResourceDiscoverySetReady.True() && catalogEntriesReady.True() {
		// Everything is ready
		catalogEntrySet.Status.SetCondition(catalogv1alpha1.CatalogEntrySetCondition{
			Type:    catalogv1alpha1.CatalogEntrySetReady,
			Status:  catalogv1alpha1.ConditionTrue,
			Reason:  "ComponentsReady",
			Message: "CustomResourceDiscoverySet and All CatalogEntry objects are ready.",
		})
	} else if !customResourceDiscoverySetReady.True() {
		catalogEntrySet.Status.SetCondition(catalogv1alpha1.CatalogEntrySetCondition{
			Type:    catalogv1alpha1.CatalogEntrySetReady,
			Status:  catalogv1alpha1.ConditionFalse,
			Reason:  "CustomResourceDiscoverySetUnready",
			Message: "CustomResourceDiscoverySet is unready.",
		})
	} else if !catalogEntriesReady.True() {
		catalogEntrySet.Status.SetCondition(catalogv1alpha1.CatalogEntrySetCondition{
			Type:    catalogv1alpha1.CatalogEntrySetReady,
			Status:  catalogv1alpha1.ConditionFalse,
			Reason:  "CatalogEntriesUnready",
			Message: "Not all CatalogEntries are unready.",
		})
	}
	if err := r.Status().Update(ctx, catalogEntrySet); err != nil {
		return fmt.Errorf("updating CatalogEntrySet status: %w", err)
	}
	return nil
}

func (r *CatalogEntrySetReconciler) getCRDByCustomResourceDiscovery(ctx context.Context, customResourceDiscovery *corev1alpha1.CustomResourceDiscovery) (*apiextensionsv1.CustomResourceDefinition, error) {
	crdList := &apiextensionsv1.CustomResourceDefinitionList{}
	if err := r.List(ctx,
		crdList,
		owner.OwnedBy(customResourceDiscovery, r.Scheme),
	); err != nil {
		return nil, err
	}
	switch len(crdList.Items) {
	case 0:
		// not found
		return nil, fmt.Errorf("crd that owned by CustomResourceDiscovery object %s not found", customResourceDiscovery.Name)
	case 1:
		// found!
		return &crdList.Items[0], nil
	default:
		// found too many
		return nil, fmt.Errorf("multiple crds that owned by CustomResourceDiscovery object %s found", customResourceDiscovery.Name)
	}
}

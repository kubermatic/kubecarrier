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
	"github.com/gobuffalo/flect"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/source"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

const crddControllerFinalizer string = "crdd.kubecarrier.io/controller"

// CustomResourceDefinitionDiscoveryReconciler reconciles a CustomResourceDefinitionDiscovery object
type CustomResourceDefinitionDiscoveryReconciler struct {
	client.Client
	Log                        logr.Logger
	Scheme                     *runtime.Scheme
	KubeCarrierSystemNamespace string
}

// +kubebuilder:rbac:groups=kubecarrier.io,resources=customresourcedefinitiondiscoveries,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=kubecarrier.io,resources=customresourcedefinitiondiscoveries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch;update;create;patch;delete

func (r *CustomResourceDefinitionDiscoveryReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var (
		ctx    = context.Background()
		log    = r.Log.WithValues("crddiscovery", req.NamespacedName)
		result ctrl.Result
	)

	crdDiscovery := &corev1alpha1.CustomResourceDefinitionDiscovery{}
	if err := r.Get(ctx, req.NamespacedName, crdDiscovery); err != nil {
		return result, client.IgnoreNotFound(err)
	}

	if !crdDiscovery.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, log, crdDiscovery); err != nil {
			return result, fmt.Errorf("handling deletion: %w", err)
		}
		return result, nil
	}

	if crdDiscovery.Status.CRD == nil {
		// Just-in-case sanity check - so the manager does not panic,
		// if the EventFilter further down is not enought.
		log.Info("skipping, missing discovered CRD in status")
		return result, nil
	}

	if util.AddFinalizer(crdDiscovery, crddControllerFinalizer) {
		if err := r.Update(ctx, crdDiscovery); err != nil {
			return result, fmt.Errorf("updating CustomResourceDefinitionDiscovery finalizers: %w", err)
		}
	}

	provider, err := catalogv1alpha1.GetProviderByProviderNamespace(ctx, r, r.KubeCarrierSystemNamespace, req.Namespace)
	if err != nil {
		return result, fmt.Errorf("getting Provider: %w", err)
	}

	currentCRD, err := r.reconcileCRD(ctx, crdDiscovery, provider)
	if err != nil {
		return result, fmt.Errorf("reconciling CRD: %w", err)
	}

	// Report Status
	if !isCRDReady(currentCRD) {
		if err = r.updateStatus(ctx, crdDiscovery, corev1alpha1.CustomResourceDefinitionDiscoveryCondition{
			Type:    corev1alpha1.CustomResourceDefinitionDiscoveryEstablished,
			Status:  corev1alpha1.ConditionFalse,
			Reason:  "Establishing",
			Message: "CRD is not yet established with the kubernetes apiserver.",
		}); err != nil {
			return result, fmt.Errorf("updating CRDD Status: %w", err)
		}
		return result, nil
	}

	if err = r.updateStatus(ctx, crdDiscovery, corev1alpha1.CustomResourceDefinitionDiscoveryCondition{
		Type:    corev1alpha1.CustomResourceDefinitionDiscoveryEstablished,
		Status:  corev1alpha1.ConditionTrue,
		Reason:  "Established",
		Message: "CRD is established with the kubernetes apiserver.",
	}); err != nil {
		return result, fmt.Errorf("updating CRDD Status: %w", err)
	}

	currentCatapult, err := r.reconcileCatapult(ctx, crdDiscovery, currentCRD)
	if err != nil {
		return result, fmt.Errorf("reconciling Catapult: %w", err)
	}

	if readyCondition, _ := currentCatapult.Status.GetCondition(operatorv1alpha1.CatapultReady); readyCondition.Status != operatorv1alpha1.ConditionTrue {
		if err = r.updateStatus(ctx, crdDiscovery, corev1alpha1.CustomResourceDefinitionDiscoveryCondition{
			Type:    corev1alpha1.CustomResourceDefinitionDiscoveryControllerReady,
			Status:  corev1alpha1.ConditionFalse,
			Reason:  "Unready",
			Message: "The controller is unready.",
		}); err != nil {
			return result, fmt.Errorf("updating status: %w", err)
		}
		return result, nil
	}

	if err = r.updateStatus(ctx, crdDiscovery, corev1alpha1.CustomResourceDefinitionDiscoveryCondition{
		Type:    corev1alpha1.CustomResourceDefinitionDiscoveryControllerReady,
		Status:  corev1alpha1.ConditionTrue,
		Reason:  "Ready",
		Message: "The controller is ready.",
	}); err != nil {
		return result, fmt.Errorf("updating status: %w", err)
	}
	return result, nil
}

func (r *CustomResourceDefinitionDiscoveryReconciler) reconcileCRD(
	ctx context.Context, crdDiscovery *corev1alpha1.CustomResourceDefinitionDiscovery,
	provider *catalogv1alpha1.Provider,
) (*apiextensionsv1.CustomResourceDefinition, error) {
	// Build desired CRD
	kind := crdDiscovery.Spec.KindOverride
	if kind == "" {
		kind = crdDiscovery.Status.CRD.Spec.Names.Kind
	}

	desiredCRD := &apiextensionsv1.CustomResourceDefinition{
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: crdDiscovery.Spec.ServiceCluster.Name + "." + provider.Name,
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Plural:   flect.Pluralize(strings.ToLower(kind)),
				Singular: strings.ToLower(kind),
				Kind:     kind,
				ListKind: kind + "List",
			},
			Scope:    apiextensionsv1.NamespaceScoped,
			Versions: crdDiscovery.Status.CRD.Spec.Versions,
		},
		Status: apiextensionsv1.CustomResourceDefinitionStatus{},
	}
	desiredCRD.Name = desiredCRD.Spec.Names.Plural + "." + desiredCRD.Spec.Group
	if _, err := util.InsertOwnerReference(crdDiscovery, desiredCRD, r.Scheme); err != nil {
		return nil, fmt.Errorf("insert object reference: %w", err)
	}

	currentCRD := &apiextensionsv1.CustomResourceDefinition{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      desiredCRD.Name,
		Namespace: desiredCRD.Namespace,
	}, currentCRD)
	if err != nil && !errors.IsNotFound(err) {
		return nil, fmt.Errorf("getting CustomResourceDefinition: %w", err)
	}
	if errors.IsNotFound(err) {
		// Create CRD
		if err = r.Create(ctx, desiredCRD); err != nil {
			return nil, fmt.Errorf("creating CustomResourceDefinition: %w", err)
		}
		return desiredCRD, nil
	}

	// Update CRD
	currentCRD.Spec.PreserveUnknownFields = desiredCRD.Spec.PreserveUnknownFields
	currentCRD.Spec.Versions = desiredCRD.Spec.Versions
	if err = r.Update(ctx, currentCRD); err != nil {
		return nil, fmt.Errorf("updating CustomResourceDefinition: %w", err)
	}

	return currentCRD, nil
}

func (r *CustomResourceDefinitionDiscoveryReconciler) reconcileCatapult(
	ctx context.Context, crdDiscovery *corev1alpha1.CustomResourceDefinitionDiscovery,
	currentCRD *apiextensionsv1.CustomResourceDefinition,
) (*operatorv1alpha1.Catapult, error) {
	// Reconcile Catapult
	storageVersion := getStorageVersion(currentCRD)
	desiredCatapult := &operatorv1alpha1.Catapult{
		ObjectMeta: metav1.ObjectMeta{
			Name:      crdDiscovery.Name,
			Namespace: crdDiscovery.Namespace,
		},
		Spec: operatorv1alpha1.CatapultSpec{
			MasterClusterCRD: operatorv1alpha1.CRDReference{
				Kind:    currentCRD.Status.AcceptedNames.Kind,
				Group:   currentCRD.Spec.Group,
				Version: storageVersion,
				Plural:  currentCRD.Status.AcceptedNames.Plural,
			},
			ServiceClusterCRD: operatorv1alpha1.CRDReference{
				Kind:    crdDiscovery.Status.CRD.Status.AcceptedNames.Kind,
				Group:   crdDiscovery.Status.CRD.Spec.Group,
				Version: storageVersion,
				Plural:  crdDiscovery.Status.CRD.Status.AcceptedNames.Plural,
			},
			ServiceCluster: operatorv1alpha1.ObjectReference{
				Name: crdDiscovery.Spec.ServiceCluster.Name,
			},
		},
	}
	if err := controllerutil.SetControllerReference(
		crdDiscovery, desiredCatapult, r.Scheme); err != nil {
		return nil, fmt.Errorf("set controller reference: %w", err)
	}

	currentCatapult := &operatorv1alpha1.Catapult{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      desiredCatapult.Name,
		Namespace: desiredCatapult.Namespace,
	}, currentCatapult)
	if err != nil && !errors.IsNotFound(err) {
		return nil, fmt.Errorf("getting Catapult: %w", err)
	}

	if errors.IsNotFound(err) {
		// Create Catapult
		if err = r.Create(ctx, desiredCatapult); err != nil {
			return nil, fmt.Errorf("creating Catapult: %w", err)
		}
		return desiredCatapult, nil
	}

	// Update Catapult
	currentCatapult.Spec = desiredCatapult.Spec
	if err = r.Update(ctx, currentCatapult); err != nil {
		return nil, fmt.Errorf("updating Catapult: %w", err)
	}

	return currentCatapult, nil
}

func (r *CustomResourceDefinitionDiscoveryReconciler) updateStatus(
	ctx context.Context, crdd *corev1alpha1.CustomResourceDefinitionDiscovery,
	condition corev1alpha1.CustomResourceDefinitionDiscoveryCondition,
) error {
	crdd.Status.ObservedGeneration = crdd.Generation
	crdd.Status.SetCondition(condition)

	established, _ := crdd.Status.GetCondition(
		corev1alpha1.CustomResourceDefinitionDiscoveryEstablished)
	controllerReady, _ := crdd.Status.GetCondition(
		corev1alpha1.CustomResourceDefinitionDiscoveryControllerReady)

	if established.True() && controllerReady.True() {
		// Everything is ready
		crdd.Status.SetCondition(corev1alpha1.CustomResourceDefinitionDiscoveryCondition{
			Type:    corev1alpha1.CustomResourceDefinitionDiscoveryReady,
			Status:  corev1alpha1.ConditionTrue,
			Reason:  "ComponentsReady",
			Message: "The CRD is established and the controller is ready.",
		})
	} else if !established.True() {
		// CRD is not yet established
		crdd.Status.SetCondition(corev1alpha1.CustomResourceDefinitionDiscoveryCondition{
			Type:    corev1alpha1.CustomResourceDefinitionDiscoveryReady,
			Status:  corev1alpha1.ConditionFalse,
			Reason:  "CRDNotEstablished",
			Message: "The CRD is not yet established.",
		})
	} else if !controllerReady.True() {
		// Controller not ready
		crdd.Status.SetCondition(corev1alpha1.CustomResourceDefinitionDiscoveryCondition{
			Type:    corev1alpha1.CustomResourceDefinitionDiscoveryReady,
			Status:  corev1alpha1.ConditionFalse,
			Reason:  "ControllerUnready",
			Message: "The controller is unready.",
		})
	}

	if err := r.Status().Update(ctx, crdd); err != nil {
		return fmt.Errorf("updating status: %w", err)
	}
	return nil
}

func (r *CustomResourceDefinitionDiscoveryReconciler) handleDeletion(ctx context.Context, log logr.Logger, crdDiscovery *corev1alpha1.CustomResourceDefinitionDiscovery) error {
	cond, ok := crdDiscovery.Status.GetCondition(corev1alpha1.CustomResourceDefinitionDiscoveryDiscovered)
	if !ok || cond.Status != corev1alpha1.ConditionFalse || cond.Reason != corev1alpha1.TerminatingReason {
		crdDiscovery.Status.SetCondition(corev1alpha1.CustomResourceDefinitionDiscoveryCondition{
			Message: "custom resource definition discovery is being teminated",
			Reason:  corev1alpha1.TerminatingReason,
			Status:  corev1alpha1.ConditionFalse,
			Type:    corev1alpha1.CustomResourceDefinitionDiscoveryDiscovered,
		})
		if err := r.Status().Update(ctx, crdDiscovery); err != nil {
			return fmt.Errorf("update discovered state: %w", err)
		}
	}

	crds := &apiextensionsv1.CustomResourceDefinitionList{}
	if err := r.List(ctx, crds, util.OwnedBy(crdDiscovery, r.Scheme)); err != nil {
		return fmt.Errorf("cannot list crds: %w", err)
	}
	if len(crds.Items) > 0 {
		for _, crd := range crds.Items {
			if err := r.Delete(ctx, &crd); err != nil {
				return fmt.Errorf("cannot delete: %w", err)
			}
		}
		return nil
	}

	if util.RemoveFinalizer(crdDiscovery, crddControllerFinalizer) {
		if err := r.Update(ctx, crdDiscovery); err != nil {
			return fmt.Errorf("updating CustomResourceDefinitionDiscovery finalizers: %w", err)
		}
	}
	return nil
}

func (r *CustomResourceDefinitionDiscoveryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.CustomResourceDefinitionDiscovery{}).
		Owns(&operatorv1alpha1.Catapult{}).
		Watches(
			&source.Kind{Type: &apiextensionsv1.CustomResourceDefinition{}},
			util.EnqueueRequestForOwner(&corev1alpha1.CustomResourceDefinitionDiscovery{}, r.Scheme)).
		WithEventFilter(util.PredicateFn(func(obj runtime.Object) bool {
			crdd, ok := obj.(*corev1alpha1.CustomResourceDefinitionDiscovery)
			if !ok {
				// we only want to filter CustomResourceDefinitionDiscovery objects
				return true
			}

			// CustomResourceDefinitionDiscoveryDiscovered that are not yet discovered, should be skipped.
			discoveredCondition, _ := crdd.Status.GetCondition(corev1alpha1.CustomResourceDefinitionDiscoveryDiscovered)
			return discoveredCondition.Status == corev1alpha1.ConditionTrue
		})).
		Complete(r)
}

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

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/owner"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

const crDiscoveryControllerFinalizer string = "crdiscovery.kubecarrier.io/controller"

// CustomResourceDiscoveryReconciler reconciles a CustomResourceDiscovery object
type CustomResourceDiscoveryReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=kubecarrier.io,resources=customresourcediscoveries,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=kubecarrier.io,resources=customresourcediscoveries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch;update;create;patch;delete

func (r *CustomResourceDiscoveryReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var (
		ctx    = context.Background()
		log    = r.Log.WithValues("crdiscovery", req.NamespacedName)
		result ctrl.Result
	)

	crDiscovery := &corev1alpha1.CustomResourceDiscovery{}
	if err := r.Get(ctx, req.NamespacedName, crDiscovery); err != nil {
		return result, client.IgnoreNotFound(err)
	}

	if !crDiscovery.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, log, crDiscovery); err != nil {
			return result, fmt.Errorf("handling deletion: %w", err)
		}
		return result, nil
	}

	if crDiscovery.Status.CRD == nil {
		// Just-in-case sanity check - so the manager does not panic,
		// if the EventFilter further down is not enought.
		log.Info("skipping, missing discovered CRD in status")
		return result, nil
	}

	if util.AddFinalizer(crDiscovery, crDiscoveryControllerFinalizer) {
		if err := r.Update(ctx, crDiscovery); err != nil {
			return result, fmt.Errorf("updating CustomResourceDiscovery finalizers: %w", err)
		}
	}

	currentCRD, err := r.reconcileCRD(ctx, crDiscovery)
	if err != nil {
		return result, fmt.Errorf("reconciling CRD: %w", err)
	}

	// Report Status
	if !isCRDReady(currentCRD) {
		if err = r.updateStatus(ctx, crDiscovery, corev1alpha1.CustomResourceDiscoveryCondition{
			Type:    corev1alpha1.CustomResourceDiscoveryEstablished,
			Status:  corev1alpha1.ConditionFalse,
			Reason:  "Establishing",
			Message: "CRD is not yet established with the kubernetes apiserver.",
		}); err != nil {
			return result, fmt.Errorf("updating CRDiscovery Status: %w", err)
		}
		return result, nil
	}

	if err = r.updateStatus(ctx, crDiscovery, corev1alpha1.CustomResourceDiscoveryCondition{
		Type:    corev1alpha1.CustomResourceDiscoveryEstablished,
		Status:  corev1alpha1.ConditionTrue,
		Reason:  "Established",
		Message: "CRD is established with the kubernetes apiserver.",
	}); err != nil {
		return result, fmt.Errorf("updating CRDiscovery Status: %w", err)
	}

	currentCatapult, err := r.reconcileCatapult(ctx, crDiscovery, currentCRD)
	if err != nil {
		return result, fmt.Errorf("reconciling Catapult: %w", err)
	}

	if readyCondition, _ := currentCatapult.Status.GetCondition(operatorv1alpha1.CatapultReady); readyCondition.Status != operatorv1alpha1.ConditionTrue {
		if err = r.updateStatus(ctx, crDiscovery, corev1alpha1.CustomResourceDiscoveryCondition{
			Type:    corev1alpha1.CustomResourceDiscoveryControllerReady,
			Status:  corev1alpha1.ConditionFalse,
			Reason:  "Unready",
			Message: "The controller is unready.",
		}); err != nil {
			return result, fmt.Errorf("updating status: %w", err)
		}
		return result, nil
	}

	if err = r.updateStatus(ctx, crDiscovery, corev1alpha1.CustomResourceDiscoveryCondition{
		Type:    corev1alpha1.CustomResourceDiscoveryControllerReady,
		Status:  corev1alpha1.ConditionTrue,
		Reason:  "Ready",
		Message: "The controller is ready.",
	}); err != nil {
		return result, fmt.Errorf("updating status: %w", err)
	}
	return result, nil
}

func (r *CustomResourceDiscoveryReconciler) reconcileCRD(
	ctx context.Context, crDiscovery *corev1alpha1.CustomResourceDiscovery,
) (*apiextensionsv1.CustomResourceDefinition, error) {
	// Build desired CRD
	kind := crDiscovery.Spec.KindOverride
	if kind == "" {
		kind = crDiscovery.Status.CRD.Spec.Names.Kind
	}
	plural := flect.Pluralize(strings.ToLower(kind))
	group := crDiscovery.Spec.ServiceCluster.Name + "." + crDiscovery.Namespace

	desiredCRD := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: plural + "." + group,
			Labels: map[string]string{
				OriginNamespaceLabel: crDiscovery.Namespace,
				ServiceClusterLabel:  crDiscovery.Spec.ServiceCluster.Name,
			},
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: group,
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Plural:   plural,
				Singular: strings.ToLower(kind),
				Kind:     kind,
				ListKind: kind + "List",
			},
			Scope:    apiextensionsv1.NamespaceScoped,
			Versions: crDiscovery.Status.CRD.Spec.Versions,
		},
		Status: apiextensionsv1.CustomResourceDefinitionStatus{},
	}
	owner.SetOwnerReference(crDiscovery, desiredCRD, r.Scheme)

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

func (r *CustomResourceDiscoveryReconciler) reconcileCatapult(
	ctx context.Context, crDiscovery *corev1alpha1.CustomResourceDiscovery,
	currentCRD *apiextensionsv1.CustomResourceDefinition,
) (*operatorv1alpha1.Catapult, error) {
	// Reconcile Catapult
	storageVersion := getStorageVersion(currentCRD)
	desiredCatapult := &operatorv1alpha1.Catapult{
		ObjectMeta: metav1.ObjectMeta{
			Name:      crDiscovery.Name,
			Namespace: crDiscovery.Namespace,
		},
		Spec: operatorv1alpha1.CatapultSpec{
			ManagementClusterCRD: operatorv1alpha1.CRDReference{
				Kind:    currentCRD.Status.AcceptedNames.Kind,
				Group:   currentCRD.Spec.Group,
				Version: storageVersion,
				Plural:  currentCRD.Status.AcceptedNames.Plural,
			},
			ServiceClusterCRD: operatorv1alpha1.CRDReference{
				Kind:    crDiscovery.Status.CRD.Status.AcceptedNames.Kind,
				Group:   crDiscovery.Status.CRD.Spec.Group,
				Version: storageVersion,
				Plural:  crDiscovery.Status.CRD.Status.AcceptedNames.Plural,
			},
			ServiceCluster: operatorv1alpha1.ObjectReference{
				Name: crDiscovery.Spec.ServiceCluster.Name,
			},
			WebhookStrategy: crDiscovery.Spec.WebhookStrategy,
		},
	}
	if err := controllerutil.SetControllerReference(
		crDiscovery, desiredCatapult, r.Scheme); err != nil {
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

func (r *CustomResourceDiscoveryReconciler) updateStatus(
	ctx context.Context, crDiscovery *corev1alpha1.CustomResourceDiscovery,
	condition corev1alpha1.CustomResourceDiscoveryCondition,
) error {
	crDiscovery.Status.ObservedGeneration = crDiscovery.Generation
	crDiscovery.Status.SetCondition(condition)

	established, _ := crDiscovery.Status.GetCondition(
		corev1alpha1.CustomResourceDiscoveryEstablished)
	controllerReady, _ := crDiscovery.Status.GetCondition(
		corev1alpha1.CustomResourceDiscoveryControllerReady)

	if established.True() && controllerReady.True() {
		// Everything is ready
		crDiscovery.Status.SetCondition(corev1alpha1.CustomResourceDiscoveryCondition{
			Type:    corev1alpha1.CustomResourceDiscoveryReady,
			Status:  corev1alpha1.ConditionTrue,
			Reason:  "ComponentsReady",
			Message: "The CRD is established and the controller is ready.",
		})
	} else if !established.True() {
		// CRD is not yet established
		crDiscovery.Status.SetCondition(corev1alpha1.CustomResourceDiscoveryCondition{
			Type:    corev1alpha1.CustomResourceDiscoveryReady,
			Status:  corev1alpha1.ConditionFalse,
			Reason:  "CRDNotEstablished",
			Message: "The CRD is not yet established.",
		})
	} else if !controllerReady.True() {
		// Controller not ready
		crDiscovery.Status.SetCondition(corev1alpha1.CustomResourceDiscoveryCondition{
			Type:    corev1alpha1.CustomResourceDiscoveryReady,
			Status:  corev1alpha1.ConditionFalse,
			Reason:  "ControllerUnready",
			Message: "The controller is unready.",
		})
	}

	if err := r.Status().Update(ctx, crDiscovery); err != nil {
		return fmt.Errorf("updating status: %w", err)
	}
	return nil
}

func (r *CustomResourceDiscoveryReconciler) handleDeletion(ctx context.Context, log logr.Logger, crDiscovery *corev1alpha1.CustomResourceDiscovery) error {
	cond, ok := crDiscovery.Status.GetCondition(corev1alpha1.CustomResourceDiscoveryDiscovered)
	if !ok || cond.Status != corev1alpha1.ConditionFalse || cond.Reason != corev1alpha1.TerminatingReason {
		crDiscovery.Status.SetCondition(corev1alpha1.CustomResourceDiscoveryCondition{
			Message: "custom resource definition discovery is being teminated",
			Reason:  corev1alpha1.TerminatingReason,
			Status:  corev1alpha1.ConditionFalse,
			Type:    corev1alpha1.CustomResourceDiscoveryDiscovered,
		})
		if err := r.Status().Update(ctx, crDiscovery); err != nil {
			return fmt.Errorf("update discovered state: %w", err)
		}
	}

	crds := &apiextensionsv1.CustomResourceDefinitionList{}
	if err := r.List(ctx, crds, owner.OwnedBy(crDiscovery, r.Scheme)); err != nil {
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

	if util.RemoveFinalizer(crDiscovery, crDiscoveryControllerFinalizer) {
		if err := r.Update(ctx, crDiscovery); err != nil {
			return fmt.Errorf("updating CustomResourceDiscovery finalizers: %w", err)
		}
	}
	return nil
}

func (r *CustomResourceDiscoveryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.CustomResourceDiscovery{}).
		Owns(&operatorv1alpha1.Catapult{}).
		Watches(
			&source.Kind{Type: &apiextensionsv1.CustomResourceDefinition{}},
			owner.EnqueueRequestForOwner(&corev1alpha1.CustomResourceDiscovery{}, r.Scheme)).
		WithEventFilter(util.PredicateFn(func(obj runtime.Object) bool {
			crDiscovery, ok := obj.(*corev1alpha1.CustomResourceDiscovery)
			if !ok {
				// we only want to filter CustomResourceDiscovery objects
				return true
			}

			// CustomResourceDiscoveryDiscovered that are not yet discovered, should be skipped.
			discoveredCondition, _ := crDiscovery.Status.GetCondition(corev1alpha1.CustomResourceDiscoveryDiscovered)
			return discoveredCondition.Status == corev1alpha1.ConditionTrue
		})).
		Complete(r)
}

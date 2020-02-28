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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/owner"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

const (
	crDiscoveriesLabel                = "crdiscoveries.kubecarrier.io/controlled-by"
	crDiscoverySetControllerFinalizer = "crdiscoveryset.kubecarrier.io/controller"
)

// CustomResourceDiscoverySetReconciler reconciles a CustomResourceDiscovery object
type CustomResourceDiscoverySetReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=kubecarrier.io,resources=customresourcediscoverysets,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=kubecarrier.io,resources=customresourcediscoverysets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kubecarrier.io,resources=customresourcediscoveries,verbs=get;list;watch;update;create;delete
// +kubebuilder:rbac:groups=kubecarrier.io,resources=customresourcediscoveries/status,verbs=get

func (r *CustomResourceDiscoverySetReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var (
		ctx    = context.Background()
		result ctrl.Result
	)

	crDiscoverySet := &corev1alpha1.CustomResourceDiscoverySet{}
	if err := r.Get(ctx, req.NamespacedName, crDiscoverySet); err != nil {
		return result, client.IgnoreNotFound(err)
	}

	if util.AddFinalizer(crDiscoverySet, crDiscoverySetControllerFinalizer) {
		if err := r.Client.Update(ctx, crDiscoverySet); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating CustomResourceDiscoverySet finalizers: %w", err)
		}
	}
	if !crDiscoverySet.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, r.Log, crDiscoverySet); err != nil {
			return result, fmt.Errorf("handling deletion: %w", err)
		}
		return ctrl.Result{}, nil
	}

	// List ServiceClusters
	serviceClusterSelector, err := metav1.LabelSelectorAsSelector(&crDiscoverySet.Spec.ServiceClusterSelector)
	if err != nil {
		return result, fmt.Errorf("parsing ServiceCluster selector: %w", err)
	}
	serviceClusterList := &corev1alpha1.ServiceClusterList{}
	if err := r.List(ctx, serviceClusterList,
		client.InNamespace(crDiscoverySet.Namespace),
		client.MatchingLabelsSelector{Selector: serviceClusterSelector},
	); err != nil {
		return result, fmt.Errorf("listing ServiceClusters: %w", err)
	}

	// Reconcile CRDiscoveries
	var unreadyCRDiscoveryNames []string
	existingCRDiscoveryNames := map[string]struct{}{}
	for _, serviceCluster := range serviceClusterList.Items {
		currentCRDiscovery, err := r.reconcileCRDiscovery(ctx, &serviceCluster, crDiscoverySet)
		if err != nil {
			return result, fmt.Errorf(
				"reconciling CustomResourceDiscovery for ServiceCluster %s: %w", serviceCluster.Name, err)
		}
		existingCRDiscoveryNames[currentCRDiscovery.Name] = struct{}{}

		ready, _ := currentCRDiscovery.Status.GetCondition(corev1alpha1.CustomResourceDiscoveryReady)
		if ready.Status != corev1alpha1.ConditionTrue {
			unreadyCRDiscoveryNames = append(unreadyCRDiscoveryNames, currentCRDiscovery.Name)
		}
	}

	// Cleanup uncontrolled CRDiscoveries
	crDiscoveryList := &corev1alpha1.CustomResourceDiscoveryList{}
	if err := r.List(ctx, crDiscoveryList, client.MatchingLabels{
		crDiscoveriesLabel: crDiscoverySet.Name,
	}); err != nil {
		return result, fmt.Errorf(
			"listing all CustomResourceDiscovery for this Set: %w", err)
	}
	for _, crDiscovery := range crDiscoveryList.Items {
		_, ok := existingCRDiscoveryNames[crDiscovery.Name]
		if ok {
			continue
		}

		// delete crDiscovery that should no longer exist
		if err := r.Delete(ctx, &crDiscovery); err != nil {
			return result, fmt.Errorf("deleting CustomResourceDiscovery: %w", err)
		}
	}

	// Report status
	crDiscoverySet.Status.ObservedGeneration = crDiscoverySet.Generation
	if len(unreadyCRDiscoveryNames) > 0 {
		// Unready
		crDiscoverySet.Status.SetCondition(corev1alpha1.CustomResourceDiscoverySetCondition{
			Type:   corev1alpha1.CustomResourceDiscoverySetReady,
			Status: corev1alpha1.ConditionFalse,
			Reason: "ComponentsUnready",
			Message: fmt.Sprintf(
				"Some CustomResourceDiscovery objects are unready [%s]", strings.Join(unreadyCRDiscoveryNames, ",")),
		})
	} else {
		// Ready
		crDiscoverySet.Status.SetCondition(corev1alpha1.CustomResourceDiscoverySetCondition{
			Type:    corev1alpha1.CustomResourceDiscoverySetReady,
			Status:  corev1alpha1.ConditionTrue,
			Reason:  "ComponentsReady",
			Message: "All CustomResourceDiscovery objects are ready.",
		})
	}

	if err := r.Status().Update(ctx, crDiscoverySet); err != nil {
		return result, fmt.Errorf("updating Status: %w", err)
	}
	return result, nil
}

func (r *CustomResourceDiscoverySetReconciler) reconcileCRDiscovery(
	ctx context.Context, serviceCluster *corev1alpha1.ServiceCluster,
	crDiscoverySet *corev1alpha1.CustomResourceDiscoverySet,
) (*corev1alpha1.CustomResourceDiscovery, error) {
	desiredCRDiscovery := &corev1alpha1.CustomResourceDiscovery{
		ObjectMeta: metav1.ObjectMeta{
			Name:      crDiscoverySet.Name + "." + serviceCluster.Name,
			Namespace: crDiscoverySet.Namespace,
			Labels: map[string]string{
				crDiscoveriesLabel: crDiscoverySet.Name,
			},
		},
		Spec: corev1alpha1.CustomResourceDiscoverySpec{
			CRD: crDiscoverySet.Spec.CRD,
			ServiceCluster: corev1alpha1.ObjectReference{
				Name: serviceCluster.Name,
			},
			KindOverride: crDiscoverySet.Spec.KindOverride,
		},
	}
	owner.SetOwnerReference(crDiscoverySet, desiredCRDiscovery, r.Scheme)

	currentCRDiscovery := &corev1alpha1.CustomResourceDiscovery{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      desiredCRDiscovery.Name,
		Namespace: desiredCRDiscovery.Namespace,
	}, currentCRDiscovery)
	if err != nil && !errors.IsNotFound(err) {
		return nil, fmt.Errorf("getting CustomResourceDiscovery: %w", err)
	}

	if errors.IsNotFound(err) {
		// create CRDiscovery instance
		if err := r.Create(ctx, desiredCRDiscovery); err != nil {
			return nil, fmt.Errorf("creating CustomResourceDiscovery: %w", err)
		}
		return desiredCRDiscovery, nil
	}

	// update existing
	currentCRDiscovery.Spec = desiredCRDiscovery.Spec
	if err := r.Update(ctx, currentCRDiscovery); err != nil {
		return nil, fmt.Errorf("updating CustomResourceDiscovery: %w", err)
	}
	return currentCRDiscovery, nil
}

func (r *CustomResourceDiscoverySetReconciler) handleDeletion(ctx context.Context, log logr.Logger, crDiscoverySet *corev1alpha1.CustomResourceDiscoverySet) error {
	cond, ok := crDiscoverySet.Status.GetCondition(corev1alpha1.CustomResourceDiscoverySetReady)
	if !ok || cond.Status != corev1alpha1.ConditionFalse || cond.Reason != corev1alpha1.TerminatingReason {
		crDiscoverySet.Status.SetCondition(corev1alpha1.CustomResourceDiscoverySetCondition{
			Message: "CustomResourceDiscoverySet is being terminated",
			Reason:  corev1alpha1.TerminatingReason,
			Status:  corev1alpha1.ConditionFalse,
			Type:    corev1alpha1.CustomResourceDiscoverySetReady,
		})
		if err := r.Status().Update(ctx, crDiscoverySet); err != nil {
			return fmt.Errorf("update CustomResourceDiscoverySet state: %w", err)
		}
	}

	customResourceDiscoveries := &corev1alpha1.CustomResourceDiscoveryList{}
	if err := r.List(ctx, customResourceDiscoveries, owner.OwnedBy(crDiscoverySet, r.Scheme)); err != nil {
		return fmt.Errorf("cannot list CustomResourceDiscoveries: %w", err)
	}
	if len(customResourceDiscoveries.Items) > 0 {
		for _, crd := range customResourceDiscoveries.Items {
			if err := r.Delete(ctx, &crd); err != nil {
				return fmt.Errorf("cannot delete: %w", err)
			}
		}
		return nil
	}

	if util.RemoveFinalizer(crDiscoverySet, crDiscoverySetControllerFinalizer) {
		if err := r.Update(ctx, crDiscoverySet); err != nil {
			return fmt.Errorf("updating CustomResourceDiscovery finalizers: %w", err)
		}
	}
	return nil
}

func (r *CustomResourceDiscoverySetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	enqueueAllCRDS := &handler.EnqueueRequestsFromMapFunc{
		ToRequests: handler.ToRequestsFunc(func(mapObject handler.MapObject) (out []reconcile.Request) {
			crdsList := &corev1alpha1.CustomResourceDiscoverySetList{}
			if err := r.List(context.Background(), crdsList, client.InNamespace(mapObject.Meta.GetNamespace())); err != nil {
				// This will makes the manager crashes, and it will restart and reconcile all objects again.
				panic(fmt.Errorf("listting CustomResourceDiscovery: %w", err))
			}
			for _, crds := range crdsList.Items {
				out = append(out, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      crds.Name,
						Namespace: crds.Namespace,
					},
				})
			}
			return
		}),
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.CustomResourceDiscoverySet{}).
		Owns(&corev1alpha1.CustomResourceDiscovery{}).
		Watches(&source.Kind{Type: &corev1alpha1.ServiceCluster{}}, enqueueAllCRDS).
		Complete(r)
}

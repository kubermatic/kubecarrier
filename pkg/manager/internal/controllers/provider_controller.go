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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

const (
	providerControllerFinalizer string = "provider.kubecarrier.io/controller"
)

// ProviderReconciler reconciles a Provider object
type ProviderReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=providers,verbs=get;list;watch;update;
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=providers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete

// Reconcile function reconciles the Provider object which specified by the request. Currently, it does the following:
// 1. Fetch the Provider object.
// 2. Handle the deletion of the Provider object (Remove the namespace that the provider owns, and remove the finalizer).
// 3. Handle the creation/update of the Provider object (Create/reconcile the namespace and insert the finalizer).
// 4. Update the status of the Provider object.
func (r *ProviderReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("provider", req.NamespacedName)

	// 1. Fetch the Provider object.
	provider := &catalogv1alpha1.Provider{}
	if err := r.Get(ctx, req.NamespacedName, provider); err != nil {
		// If the Provider object is already gone, we just ignore the NotFound error.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 2. Handle the deletion of the Provider object.
	if !provider.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, log, provider); err != nil {
			return ctrl.Result{}, fmt.Errorf("handling deletion: %w", err)
		}
		return ctrl.Result{}, nil
	}

	// 3. reconcile the Provider object.
	// check/add the finalizer for the Provider
	if util.AddFinalizer(provider, providerControllerFinalizer) {
		// Update the provider with the finalizer
		if err := r.Update(ctx, provider); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating finalizers: %w", err)
		}
	}

	// check/update the NamespaceName
	if provider.Status.NamespaceName == "" {
		provider.Status.NamespaceName = fmt.Sprintf("provider-%s", strings.Replace(provider.Name, ".", "-", -1))
		if err := r.Status().Update(ctx, provider); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating NamespaceName: %w", err)
		}
	}

	// Reconcile the namespace for the provider
	if err := r.reconcileNamespace(ctx, log, provider); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconciling namespace: %w", err)
	}

	return ctrl.Result{}, nil
}

func (r *ProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	owner := &catalogv1alpha1.Provider{}
	enqueuer := util.EnqueueRequestForOwner(owner, mgr.GetScheme())

	return ctrl.NewControllerManagedBy(mgr).
		For(&catalogv1alpha1.Provider{}).
		Watches(&source.Kind{Type: &corev1.Namespace{}}, enqueuer).
		Complete(r)
}

// handleDeletion handles the deletion of the Provider object. Currently, it does:
// 1. Delete the Namespace that the provider owns.
// 2. Remove the finalizer from the provider object.
func (r *ProviderReconciler) handleDeletion(ctx context.Context, log logr.Logger, provider *catalogv1alpha1.Provider) error {
	// Update the Provider Status to Terminating.
	readyCondition, _ := provider.Status.GetCondition(catalogv1alpha1.ProviderReady)
	if readyCondition.Status != catalogv1alpha1.ConditionFalse ||
		readyCondition.Status == catalogv1alpha1.ConditionFalse && readyCondition.Reason != catalogv1alpha1.ProviderTerminatingReason {
		provider.Status.ObservedGeneration = provider.Generation
		provider.Status.SetCondition(catalogv1alpha1.ProviderCondition{
			Type:    catalogv1alpha1.ProviderReady,
			Status:  catalogv1alpha1.ConditionFalse,
			Reason:  catalogv1alpha1.ProviderTerminatingReason,
			Message: "Provider is being terminated",
		})
		if err := r.Status().Update(ctx, provider); err != nil {
			return fmt.Errorf("updating Provider status: %w", err)
		}
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: provider.Status.NamespaceName,
		},
	}

	if err := r.Delete(ctx, ns); err == nil {
		// Move to the next reconcile round
		return nil
	} else if !errors.IsNotFound(err) {
		return fmt.Errorf("deleting namespace: %w", err)
	}
	// else the error is IsNotFound, so the namespace is gone, and then we remove the finalizer.

	// 2. The Namespace is completely removed, then we remove the finalizer here.
	if util.RemoveFinalizer(provider, providerControllerFinalizer) {
		if err := r.Update(ctx, provider); err != nil {
			return fmt.Errorf("updating Provider Status: %w", err)
		}
	}
	return nil
}

func (r *ProviderReconciler) reconcileNamespace(ctx context.Context, log logr.Logger, provider *catalogv1alpha1.Provider) error {
	ns := &corev1.Namespace{}
	err := r.Get(ctx, types.NamespacedName{Name: provider.Status.NamespaceName}, ns)
	if err == nil {
		// No error from the Get, so we update the Provider Status.
		if readyCondition, _ := provider.Status.GetCondition(catalogv1alpha1.ProviderReady); readyCondition.Status != catalogv1alpha1.ConditionTrue {
			// Update Provider Status
			provider.Status.ObservedGeneration = provider.Generation
			provider.Status.SetCondition(catalogv1alpha1.ProviderCondition{
				Type:    catalogv1alpha1.ProviderReady,
				Status:  catalogv1alpha1.ConditionTrue,
				Reason:  "SetupComplete",
				Message: "Provider setup is complete.",
			})
			if err := r.Status().Update(ctx, provider); err != nil {
				return fmt.Errorf("updating Provider status: %w", err)
			}
		}
		return nil

	}
	if !errors.IsNotFound(err) {
		return fmt.Errorf("getting Provider namepsace: %w", err)
	}

	// When the namespace is not there, we need to make sure to update our Status first,
	// so the rest of the system can act on it.
	// This is especially important for Reconcilations that are prone to errors, or take a long time.
	// In this case it's most likely overkill, but still strictly necessary.
	if readyCondition, _ := provider.Status.GetCondition(catalogv1alpha1.ProviderReady); readyCondition.Status != catalogv1alpha1.ConditionFalse {
		provider.Status.ObservedGeneration = provider.Generation
		provider.Status.SetCondition(catalogv1alpha1.ProviderCondition{
			Type:    catalogv1alpha1.ProviderReady,
			Status:  catalogv1alpha1.ConditionFalse,
			Reason:  "SetupIncomplete",
			Message: "Provider setup is incomplete, namespace is missing.",
		})

		if err = r.Status().Update(ctx, provider); err != nil {
			return fmt.Errorf("updating Provider status: %w", err)
		}

		// move to the next reconcile round
		return nil
	}

	ns.Name = provider.Status.NamespaceName
	if _, err := util.InsertOwnerReference(provider, ns, r.Scheme); err != nil {
		return fmt.Errorf("setting cross-namespaceed owner reference: %w", err)
	}
	// Reconcile the namespace
	if err = r.Create(ctx, ns); err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("creating Provider namespace: %w", err)
	}
	return nil
}

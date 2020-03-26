/*
Copyright 2020 The KubeCarrier Authors.

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
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
)

type ServiceClusterReconciler struct {
	client.Client
	Log                  logr.Logger
	Scheme               *runtime.Scheme
	MonitorGraceDuration time.Duration
}

// +kubebuilder:rbac:groups=kubecarrier.io,resources=serviceclusters,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=kubecarrier.io,resources=serviceclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=ferries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=ferries/status,verbs=get;update;patch

func (r *ServiceClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {

	ctx := context.Background()
	log := r.Log.WithValues("servicecluster", req.NamespacedName)

	serviceCluster := &corev1alpha1.ServiceCluster{}
	if err := r.Get(ctx, req.NamespacedName, serviceCluster); err != nil {
		// If the ServiceCluster object is already gone, we just ignore the NotFound error.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 2. Handle the deletion of the ServiceCluster object.
	if !serviceCluster.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, log, serviceCluster); err != nil {
			return ctrl.Result{}, fmt.Errorf("handling deletion: %w", err)
		}
		return ctrl.Result{}, nil
	}

	currentFerry, err := r.reconcileFerry(ctx, serviceCluster)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("reconciling Ferry: %w", err)
	}

	if !currentFerry.IsReady() {
		if err = r.updateStatus(ctx, serviceCluster, corev1alpha1.ServiceClusterCondition{
			Type:    corev1alpha1.ServiceClusterControllerReady,
			Status:  corev1alpha1.ConditionFalse,
			Reason:  "Unready",
			Message: "The controller is unready.",
		}); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating status: %w", err)
		}
		return ctrl.Result{}, nil
	}

	if err = r.updateStatus(ctx, serviceCluster, corev1alpha1.ServiceClusterCondition{
		Type:    corev1alpha1.ServiceClusterControllerReady,
		Status:  corev1alpha1.ConditionTrue,
		Reason:  "Ready",
		Message: "The controller is ready.",
	}); err != nil {
		return ctrl.Result{}, fmt.Errorf("updating status: %w", err)
	}

	// added extra seconds to requeue to ensure previous MonitorGraceDuration
	// expires if no new heartbeats arrive
	return ctrl.Result{RequeueAfter: r.MonitorGraceDuration + time.Second}, nil
}

func (r *ServiceClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.ServiceCluster{}).
		Owns(&operatorv1alpha1.Ferry{}).
		Complete(r)
}

func (r *ServiceClusterReconciler) reconcileFerry(
	ctx context.Context,
	serviceCluster *corev1alpha1.ServiceCluster) (*operatorv1alpha1.Ferry, error) {
	desiredFerry := &operatorv1alpha1.Ferry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceCluster.Name,
			Namespace: serviceCluster.Namespace,
		},
		Spec: operatorv1alpha1.FerrySpec{
			KubeconfigSecret: operatorv1alpha1.ObjectReference{
				Name: serviceCluster.Spec.KubeconfigSecret.Name,
			},
		},
	}

	if err := controllerutil.SetControllerReference(
		serviceCluster, desiredFerry, r.Scheme); err != nil {
		return nil, fmt.Errorf("set controller reference: %w", err)
	}

	currentFerry := &operatorv1alpha1.Ferry{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      desiredFerry.Name,
		Namespace: desiredFerry.Namespace,
	}, currentFerry)
	if err != nil && !errors.IsNotFound(err) {
		return nil, fmt.Errorf("getting Ferry: %w", err)
	}

	if errors.IsNotFound(err) {
		// Create Ferry
		if err = r.Create(ctx, desiredFerry); err != nil {
			return nil, fmt.Errorf("creating Ferry: %w", err)
		}
		return desiredFerry, nil
	}

	// Update Ferry
	currentFerry.Spec = desiredFerry.Spec
	if err = r.Update(ctx, currentFerry); err != nil {
		return nil, fmt.Errorf("updating Ferry: %w", err)
	}

	return currentFerry, nil
}

func (r *ServiceClusterReconciler) handleDeletion(ctx context.Context, log logr.Logger, serviceCluster *corev1alpha1.ServiceCluster) error {
	// Update the Catalog Status to Terminating.
	readyCondition, _ := serviceCluster.Status.GetCondition(corev1alpha1.ServiceClusterReady)
	if readyCondition.Status != corev1alpha1.ConditionFalse ||
		readyCondition.Status == corev1alpha1.ConditionFalse && readyCondition.Reason != corev1alpha1.ServiceClusterTerminatingReason {
		if err := r.updateStatus(ctx, serviceCluster, corev1alpha1.ServiceClusterCondition{
			Type:    corev1alpha1.ServiceClusterReady,
			Status:  corev1alpha1.ConditionFalse,
			Reason:  corev1alpha1.ServiceClusterTerminatingReason,
			Message: "ServiceCluster is being terminated",
		}); err != nil {
			return fmt.Errorf("updating ServiceCluster Status: %w", err)
		}
	}
	return nil
}

func (r *ServiceClusterReconciler) updateStatus(
	ctx context.Context, serviceCluster *corev1alpha1.ServiceCluster,
	condition corev1alpha1.ServiceClusterCondition,
) error {
	serviceCluster.Status.ObservedGeneration = serviceCluster.Generation
	serviceCluster.Status.SetCondition(condition)

	controllerReady, _ := serviceCluster.Status.GetCondition(
		corev1alpha1.ServiceClusterControllerReady)
	serviceClusterReachable, _ := serviceCluster.Status.GetCondition(
		corev1alpha1.ServiceClusterReachable)
	timenow := metav1.Now()
	if timenow.Sub(serviceClusterReachable.LastHeartbeatTime.Time) > r.MonitorGraceDuration {
		serviceClusterReachable.Type = corev1alpha1.ServiceClusterReachable
		serviceClusterReachable.LastTransitionTime = timenow
		serviceClusterReachable.Status = corev1alpha1.ConditionUnknown
		serviceClusterReachable.Message = "cluster reachable heartbeat hasn't been updaed within monitor grace duration"
		serviceClusterReachable.Reason = "GracePeriodTimeout"
		serviceCluster.Status.SetCondition(serviceClusterReachable)
	}

	if controllerReady.True() && serviceClusterReachable.True() {
		serviceCluster.Status.SetCondition(corev1alpha1.ServiceClusterCondition{
			Type:    corev1alpha1.ServiceClusterReady,
			Status:  corev1alpha1.ConditionTrue,
			Reason:  "ComponentsReady",
			Message: "The Ferry controller is ready and service cluster is reachable.",
		})
	} else if !controllerReady.True() {
		serviceCluster.Status.SetCondition(corev1alpha1.ServiceClusterCondition{
			Type:    corev1alpha1.ServiceClusterReady,
			Status:  corev1alpha1.ConditionFalse,
			Reason:  "FerryControllerUnready",
			Message: "The Ferry controller is not ready.",
		})
	} else if !serviceClusterReachable.True() {
		serviceCluster.Status.SetCondition(corev1alpha1.ServiceClusterCondition{
			Type:    corev1alpha1.ServiceClusterReady,
			Status:  corev1alpha1.ConditionFalse,
			Reason:  "ServiceClusterUnreachable",
			Message: "The service cluster is unreachable.",
		})
	}

	if err := r.Status().Update(ctx, serviceCluster); err != nil {
		return fmt.Errorf("updating status: %w", err)
	}
	return nil
}

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
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
)

type ServerVersionInfo interface {
	ServerVersion() (*version.Info, error)
}

// ServiceClusterReconciler sends a heartbeat to KubeCarrier to signal its readyness.
type ServiceClusterReconciler struct {
	Log logr.Logger

	ManagementClient          client.Client
	ServiceClusterVersionInfo ServerVersionInfo
	ProviderNamespace         string
	ServiceClusterName        string
	StatusUpdatePeriod        time.Duration
}

// +kubebuilder:rbac:groups=kubecarrier.io,resources=serviceclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubecarrier.io,resources=serviceclusters/status,verbs=get;update;patch

func (r *ServiceClusterReconciler) Reconcile(req ctrl.Request) (res ctrl.Result, err error) {
	ctx := context.Background()

	serviceCluster := &corev1alpha1.ServiceCluster{}
	if err := r.ManagementClient.Get(ctx, req.NamespacedName, serviceCluster); err != nil {
		// If the ServiceCluster object is already gone, we just ignore the NotFound error.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	serverVersion, svcErr := r.ServiceClusterVersionInfo.ServerVersion()
	serviceCluster.Status.KubernetesVersion = serverVersion

	if svcErr != nil {
		reason := "ClusterUnreachable"
		statusErr, ok := svcErr.(*errors.StatusError)
		if ok {
			reason = string(statusErr.Status().Reason)
		}
		if err = r.updateStatus(ctx, serviceCluster, corev1alpha1.ServiceClusterCondition{
			Type:    corev1alpha1.ServiceClusterReachable,
			Status:  corev1alpha1.ConditionFalse,
			Reason:  reason,
			Message: svcErr.Error(),
		}); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating status: %w", err)
		}
		return ctrl.Result{}, svcErr
	}

	if err = r.updateStatus(ctx, serviceCluster, corev1alpha1.ServiceClusterCondition{
		Type:    corev1alpha1.ServiceClusterReachable,
		Status:  corev1alpha1.ConditionTrue,
		Reason:  "ServiceClusterReachable",
		Message: "service cluster is posting ready status",
	}); err != nil {
		return ctrl.Result{}, fmt.Errorf("updateing status: %w", err)
	}

	return ctrl.Result{RequeueAfter: r.StatusUpdatePeriod}, nil
}

func (r *ServiceClusterReconciler) SetupWithManager(managementMgr ctrl.Manager) error {
	c, err := controller.New("servicecluster-controller", managementMgr, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		return fmt.Errorf("manager-setup: %w", err)
	}

	// Bootstrap ServiceCluster enqueuement
	if err = c.Watch(
		source.Func(r.enqueueOwnCluster),
		&handler.EnqueueRequestForObject{},
	); err != nil {
		return fmt.Errorf("initial watch: %w", err)
	}
	return nil
}

func (r *ServiceClusterReconciler) enqueueOwnCluster(h handler.EventHandler, q workqueue.RateLimitingInterface, p ...predicate.Predicate) error {
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Name:      r.ServiceClusterName,
		Namespace: r.ProviderNamespace,
	}})
	return nil
}

func (r *ServiceClusterReconciler) updateStatus(
	ctx context.Context, serviceCluster *corev1alpha1.ServiceCluster,
	condition corev1alpha1.ServiceClusterCondition,
) error {
	serviceCluster.Status.ObservedGeneration = serviceCluster.Generation
	serviceCluster.Status.SetCondition(condition)

	if err := r.ManagementClient.Status().Update(ctx, serviceCluster); err != nil {
		return fmt.Errorf("updating status: %w", err)
	}
	return nil
}

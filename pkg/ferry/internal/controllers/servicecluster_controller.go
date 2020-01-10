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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
)

// ServiceClusterReconciler sends a heartbeat to KubeCarrier to signal its readyness.
type ServiceClusterReconciler struct {
	Log logr.Logger

	MasterClient       client.Client
	ServiceClient      client.Client
	ProviderNamespace  string
	ServiceClusterName string
	StatusUpdatePeriod time.Duration
}

// +kubebuilder:rbac:groups=kubecarrier.io,resources=serviceclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubecarrier.io,resources=serviceclusters/status,verbs=get;update;patch

func (r *ServiceClusterReconciler) Reconcile(req ctrl.Request) (res ctrl.Result, err error) {
	ctx := context.Background()
	log := r.Log.WithValues("servicecluster", req.NamespacedName)

	cm := &corev1.ConfigMap{}
	svcErr := r.ServiceClient.Get(ctx, types.NamespacedName{
		Namespace: "kube-public",
		Name:      "cluster-info",
	}, cm)

	var cond corev1alpha1.ServiceClusterCondition

	if svcErr != nil {
		reason := ""
		statusErr, ok := svcErr.(*errors.StatusError)
		if ok {
			reason = string(statusErr.Status().Reason)
		}
		cond = corev1alpha1.ServiceClusterCondition{
			Message: svcErr.Error(),
			Reason:  reason,
			Status:  corev1alpha1.ConditionFalse,
			Type:    corev1alpha1.ServiceClusterReady,
		}
	} else {
		cond = corev1alpha1.ServiceClusterCondition{
			Message: "service cluster is posting ready status",
			Reason:  "ServiceClusterReady",
			Status:  corev1alpha1.ConditionTrue,
			Type:    corev1alpha1.ServiceClusterReady,
		}
	}

	serviceCluster := &corev1alpha1.ServiceCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.ServiceClusterName,
			Namespace: r.ProviderNamespace,
		},
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r.MasterClient, serviceCluster, func() error {
		return nil
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("cannot create or update ServiceCluster: %w", err)
	}
	log.Info(fmt.Sprintf("ServiceCluster: %s", string(op)))

	serviceCluster.Status.ObservedGeneration = serviceCluster.Generation
	serviceCluster.Status.SetCondition(cond)

	if err := r.MasterClient.Status().Update(ctx, serviceCluster); err != nil {
		return ctrl.Result{}, fmt.Errorf("status update: %w", err)
	}

	if svcErr != nil {
		return ctrl.Result{}, svcErr
	}
	return ctrl.Result{RequeueAfter: r.StatusUpdatePeriod}, nil
}

func (r *ServiceClusterReconciler) SetupWithManagers(serviceMgr, masterMgr ctrl.Manager) error {
	c, err := controller.New("servicecluster-controller", masterMgr, controller.Options{
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

	err = c.Watch(
		&source.Kind{Type: &corev1alpha1.ServiceCluster{}},
		&handler.EnqueueRequestsFromMapFunc{
			ToRequests: handler.ToRequestsFunc(func(obj handler.MapObject) []reconcile.Request {
				if obj.Meta.GetName() == r.ServiceClusterName {
					return []reconcile.Request{{
						NamespacedName: types.NamespacedName{
							Namespace: r.ProviderNamespace,
							Name:      r.ServiceClusterName,
						},
					}}
				}
				return nil
			}),
		},
	)
	if err != nil {
		return fmt.Errorf("serviceCluster watch: %w", err)
	}
	return masterMgr.Add(c)
}

func (r *ServiceClusterReconciler) enqueueOwnCluster(h handler.EventHandler, q workqueue.RateLimitingInterface, p ...predicate.Predicate) error {
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Name:      r.ServiceClusterName,
		Namespace: r.ProviderNamespace,
	}})
	return nil
}

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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	masterv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/master/v1alpha1"
)

// ManagementClusterReconciler reconciles a ManagementCluster object
type ManagementClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=master.kubecarrier.io,resources=managementclusters,verbs=get;list;watch;update;
// +kubebuilder:rbac:groups=master.kubecarrier.io,resources=managementclusters/status,verbs=get;update;patch

func (r ManagementClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	managementCluster := &masterv1alpha1.ManagementCluster{}
	if err := r.Get(ctx, req.NamespacedName, managementCluster); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if managementCluster.Name == "local" {
		if err := r.updateStatus(ctx, managementCluster, &masterv1alpha1.ManagementClusterCondition{
			Type:    masterv1alpha1.ManagementClusterReady,
			Status:  masterv1alpha1.ConditionTrue,
			Reason:  "MasterManagementClusterIsReady",
			Message: "Master ManagementCluster is ready",
		}); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *ManagementClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&masterv1alpha1.ManagementCluster{}).
		Complete(r)
}

func (r ManagementClusterReconciler) updateStatus(
	ctx context.Context,
	managementCluster *masterv1alpha1.ManagementCluster,
	condition *masterv1alpha1.ManagementClusterCondition,
) error {
	managementCluster.Status.ObservedGeneration = managementCluster.Generation
	if condition != nil {
		managementCluster.Status.SetCondition(*condition)
	}
	if err := r.Status().Update(ctx, managementCluster); err != nil {
		return fmt.Errorf("updating ManagementCluster status: %w", err)
	}
	return nil
}

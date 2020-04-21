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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	fakev1 "github.com/kubermatic/kubecarrier/pkg/apis/fake/v1"
)

// SnapshotReconciler reconciles a Snapshot object
type SnapshotReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=fake.kubecarrier.io,resources=snapshots,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=fake.kubecarrier.io,resources=snapshots/status,verbs=get;update;patch

func (r *SnapshotReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	s := &fakev1.Snapshot{}
	if err := r.Get(ctx, req.NamespacedName, s); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("cannot fetch database: %w", err)
	}

	if !s.GetDeletionTimestamp().IsZero() {
		if s.SetTerminatingCondition() {
			if err := r.Client.Status().Update(ctx, s); err != nil {
				return ctrl.Result{}, fmt.Errorf("updating %s status: %w", "FakeSnapshot controller", err)
			}
		}
		return ctrl.Result{}, nil
	}

	if s.SetReadyCondition() {
		s.Status.Date = metav1.Now()
		if err := r.Status().Update(ctx, s); err != nil {
			return ctrl.Result{}, fmt.Errorf("cannot update s: %w", err)
		}
	}
	return ctrl.Result{}, nil

}

func (r *SnapshotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fakev1.Snapshot{}).
		Complete(r)
}

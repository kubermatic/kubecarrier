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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubermatic/utils/pkg/util"

	fakev1 "k8c.io/kubecarrier/pkg/apis/fake/v1"
)

const finalizer = "fake.kubecarrier.io/controller"

// DBReconciler reconciles a DB object
type DBReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=fake.kubecarrier.io,resources=dbs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=fake.kubecarrier.io,resources=dbs/status,verbs=get;update;patch

func (r *DBReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	db := &fakev1.DB{}
	if err := r.Get(ctx, req.NamespacedName, db); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("cannot fetch database: %w", err)
	}

	if !db.GetDeletionTimestamp().IsZero() {
		if db.SetTerminatingCondition() {
			if err := r.Client.Status().Update(ctx, db); err != nil {
				return ctrl.Result{}, fmt.Errorf("updating %s status: %w", "FakeDB controller", err)
			}
		}
		cond, _ := db.Status.GetCondition(fakev1.DBReady)
		if time.Since(cond.LastTransitionTime.Time) < time.Duration(db.Spec.Config.DeletionAfterSeconds)*time.Second {
			return ctrl.Result{RequeueAfter: time.Second * time.Duration(db.Spec.Config.DeletionAfterSeconds)}, nil
		}
		if util.RemoveFinalizer(db, finalizer) {
			if err := r.Client.Update(ctx, db); err != nil {
				return ctrl.Result{}, fmt.Errorf("updating %s Status: %w", "FakeDB controller", err)
			}
		}

		return ctrl.Result{}, nil
	}

	if util.AddFinalizer(db, finalizer) {
		if err := r.Client.Update(ctx, db); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating %s finalizers: %w", "FakeDB controller", err)
		}
	}

	readyTime := db.GetCreationTimestamp().Time.Add(time.Duration(db.Spec.Config.ReadyAfterSeconds) * time.Second)
	now := time.Now()
	if readyTime.Before(now) {
		if db.SetReadyCondition() {
			db.Status.Connection = fmt.Sprintf("fake endpoint:%s:%s", db.Spec.DatabaseName, db.Spec.DatabaseUser)
			if err := r.Status().Update(ctx, db); err != nil {
				return ctrl.Result{}, fmt.Errorf("cannot update db: %w", err)
			}
		}
		return ctrl.Result{}, nil
	}

	return ctrl.Result{
		// requeue with a bit of wiggle room (10ms)
		RequeueAfter: readyTime.Sub(now) + 10*time.Millisecond,
	}, nil
}

func (r *DBReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fakev1.DB{}).
		Complete(r)
}

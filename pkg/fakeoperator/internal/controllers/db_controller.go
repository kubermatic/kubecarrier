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
	"math/rand"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	fakev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/fake/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

const finalizer = "fake.kubecarrier.io/controller"

// DBReconciler reconciles a Joke object
type DBReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=fake.kubecarrier.io,resources=dbs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=fake.kubecarrier.io,resources=dbs/status,verbs=get;update;patch
func (r *DBReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	db := &fakev1alpha1.DB{}
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
		cond, _ := db.Status.GetCondition(fakev1alpha1.DBReady)
		if time.Now().UTC().Sub(cond.LastTransitionTime.Time).Seconds() < float64(db.Spec.Config.DeletionAfterSeconds) {
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

	cond, ok := db.Status.GetCondition(fakev1alpha1.DBReady)
	// mark as unready
	if !ok {
		if db.SetUnReadyCondition() {
			if err := r.Status().Update(ctx, db); err != nil {
				return ctrl.Result{}, fmt.Errorf("cannot update db: %w", err)
			}
		}
		return ctrl.Result{}, nil
	}

	// DB is in ready state
	if db.IsReady() {
		if db.SetReadyCondition() {
			db.Status.Connection = &fakev1alpha1.Connection{
				Endpoint: "fake endpoint",
				Name:     db.Spec.DatabaseName,
				Username: db.Spec.DatabaseUser,
			}
			if err := r.Status().Update(ctx, db); err != nil {
				return ctrl.Result{}, fmt.Errorf("cannot update db: %w", err)
			}
			return ctrl.Result{}, nil
		}
	}

	// DB is not ready and waiting for timeout
	if time.Now().UTC().Sub(cond.LastTransitionTime.Time).Seconds() < float64(db.Spec.Config.ReadyAfterSeconds) {
		return ctrl.Result{RequeueAfter: time.Second * time.Duration(db.Spec.Config.ReadyAfterSeconds)}, nil
	}

	// mark as ready after delay
	if db.SetReadyCondition() {
		db.Status.Connection = &fakev1alpha1.Connection{
			Endpoint: "fake endpoint",
			Name:     db.Spec.DatabaseName,
			Username: db.Spec.DatabaseUser,
		}
		if err := r.Status().Update(ctx, db); err != nil {
			return ctrl.Result{}, fmt.Errorf("cannot update db: %w", err)
		}
	}
	return ctrl.Result{}, nil
}

func (r *DBReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fakev1alpha1.DB{}).
		Complete(r)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

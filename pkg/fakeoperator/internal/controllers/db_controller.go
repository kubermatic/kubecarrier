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
)

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

	db.Status.ObservedGeneration = db.Generation

	db.Status.SetCondition(fakev1alpha1.DBCondition{
		Message: "No dbs of the appropriate type were found in the dbs database",
		Reason:  "NoValidJokes",
		Status:  fakev1alpha1.ConditionTrue,
		Type:    fakev1alpha1.DBReady,
	})

	if err := r.Status().Update(ctx, db); err != nil {
		return ctrl.Result{}, fmt.Errorf("cannot update db: %w", err)
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

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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubermatic/kubecarrier/pkg/apis/e2e/v1alpha2"
)

// JokeReconciler reconciles a Joke object
type JokeReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=e2e.kubecarrier.io,resources=jokes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=e2e.kubecarrier.io,resources=jokes/status,verbs=get;update;patch

func (r *JokeReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("joke", req.NamespacedName)

	joke := &v1alpha2.Joke{}
	if err := r.Get(ctx, req.NamespacedName, joke); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("cannot fetch joke: %w", err)
	}
	log.Info("doing some actual work...TODO")
	// TODO --> finish this tomorrow
	return ctrl.Result{}, nil
}

func (r *JokeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha2.Joke{}).
		Complete(r)
}

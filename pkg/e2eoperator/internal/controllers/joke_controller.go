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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	e2ev1alpha2 "github.com/kubermatic/kubecarrier/pkg/apis/e2e/v1alpha2"
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
	joke := &e2ev1alpha2.Joke{}
	if err := r.Get(ctx, req.NamespacedName, joke); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("cannot fetch joke: %w", err)
	}

	joke.Status.ObservedGeneration = joke.Generation
	if joke.Spec.Disabled {
		joke.Status.SelectedJoke = nil
		joke.Status.SetCondition(e2ev1alpha2.JokeCondition{
			Message: "Joke object is disabled",
			Reason:  "Disabled",
			Status:  e2ev1alpha2.ConditionFalse,
			Type:    e2ev1alpha2.JokeReady,
		})
		if err := r.Status().Update(ctx, joke); err != nil {
			return ctrl.Result{}, fmt.Errorf("cannot update joke: %w", err)
		}
		return ctrl.Result{}, nil

	}

	// This should be handled by the validation webhook
	if len(joke.Spec.Jokes) == 0 {
		joke.Status.SelectedJoke = nil
		joke.Status.SetCondition(e2ev1alpha2.JokeCondition{
			Message: "No jokes were defined in the database",
			Reason:  "EmptyDatabase",
			Status:  e2ev1alpha2.ConditionFalse,
			Type:    e2ev1alpha2.JokeReady,
		})
		if err := r.Status().Update(ctx, joke); err != nil {
			return ctrl.Result{}, fmt.Errorf("cannot update joke: %w", err)
		}
		return ctrl.Result{}, nil
	}

	joke.Status.SelectedJoke = &joke.Spec.Jokes[rand.Intn(len(joke.Spec.Jokes))]
	joke.Status.SetCondition(e2ev1alpha2.JokeCondition{
		Message: "Joke has been found and setup",
		Reason:  "JokeSetup",
		Status:  e2ev1alpha2.ConditionTrue,
		Type:    e2ev1alpha2.JokeReady,
	})
	if err := r.Status().Update(ctx, joke); err != nil {
		return ctrl.Result{}, fmt.Errorf("cannot update joke: %w", err)
	}
	return ctrl.Result{}, nil
}

func (r *JokeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&e2ev1alpha2.Joke{}).
		Complete(r)
}

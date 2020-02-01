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
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

type AdoptionReconciler struct {
	client.Client
	Log                  logr.Logger
	ServiceClusterClient client.Client
	ServiceClusterCache  cache.Cache

	// Dynamic types we work with
	MasterClusterGVK, ServiceClusterGVK   schema.GroupVersionKind
	MasterClusterType, ServiceClusterType *unstructured.Unstructured

	ProviderNamespace string
}

func (r *AdoptionReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var (
		ctx    = context.Background()
		result ctrl.Result
	)

	serviceClusterObj := r.ServiceClusterType.DeepCopy()
	if err := r.ServiceClusterClient.Get(ctx, req.NamespacedName, serviceClusterObj); err != nil {
		return result, client.IgnoreNotFound(err)
	}

	// Build desired master cluster object
	desiredMasterClusterObj := serviceClusterObj.DeepCopy()
	desiredMasterClusterObj.SetGroupVersionKind(r.MasterClusterGVK)
	if err := unstructured.SetNestedField(
		desiredMasterClusterObj.Object, map[string]interface{}{}, "metadata"); err != nil {
		return result, fmt.Errorf(
			"deleting %s .metadata: %w", r.MasterClusterGVK.Kind, err)
	}
	desiredMasterClusterObj.SetName(serviceClusterObj.GetName())
	desiredMasterClusterObj.SetNamespace(r.ProviderNamespace)

	// Reconcile
	currentMasterClusterObj := r.MasterClusterType.DeepCopy()
	err := r.Get(ctx, types.NamespacedName{
		Name:      desiredMasterClusterObj.GetName(),
		Namespace: desiredMasterClusterObj.GetNamespace(),
	}, currentMasterClusterObj)
	if err != nil && !errors.IsNotFound(err) {
		return result, fmt.Errorf(
			"getting %s: %w", r.MasterClusterGVK.Kind, err)
	}

	if errors.IsNotFound(err) {
		// Create the master cluster object
		if err = r.Create(ctx, desiredMasterClusterObj); err != nil {
			return result, fmt.Errorf(
				"creating %s: %w", r.MasterClusterGVK.Kind, err)
		}
	}

	return result, nil
}

func (r *AdoptionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	serviceClusterSource := &source.Kind{Type: r.ServiceClusterType}
	if err := serviceClusterSource.InjectCache(r.ServiceClusterCache); err != nil {
		return fmt.Errorf("injecting cache: %w", err)
	}

	c, err := controller.New(
		strings.ToLower(r.ServiceClusterGVK.Kind),
		mgr, controller.Options{
			Reconciler: r,
		})
	if err != nil {
		return fmt.Errorf("creating controller: %w", err)
	}

	return c.Watch(
		source.Func(serviceClusterSource.Start),
		&handler.EnqueueRequestForObject{},
		util.PredicateFn(func(obj runtime.Object) bool {
			// we are only interested in unowned objects
			meta, ok := obj.(metav1.Object)
			if !ok {
				return false
			}

			return len(meta.GetOwnerReferences()) == 0
		}))
}

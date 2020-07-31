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

	"github.com/kubermatic/utils/pkg/owner"
	"github.com/kubermatic/utils/pkg/util"

	corev1alpha1 "k8c.io/kubecarrier/pkg/apis/core/v1alpha1"
)

type AdoptionReconciler struct {
	client.Client
	Log                  logr.Logger
	NamespacedClient     client.Client
	ServiceClusterClient client.Client
	ServiceClusterCache  cache.Cache

	// Dynamic types we work with
	ManagementClusterGVK, ServiceClusterGVK schema.GroupVersionKind

	ProviderNamespace string
}

func (r *AdoptionReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var (
		ctx    = context.Background()
		result ctrl.Result
	)

	serviceClusterObj := r.newServiceObject()
	if err := r.ServiceClusterClient.Get(ctx, req.NamespacedName, serviceClusterObj); err != nil {
		return result, client.IgnoreNotFound(err)
	}
	if owner.IsOwned(serviceClusterObj) ||
		!serviceClusterObj.GetDeletionTimestamp().IsZero() {
		// the object is already owned or was deleted, so we don't want to do anything.
		// otherwise we might recreate the owning object preventing the deletion of this instance.
		return result, nil
	}

	// Lookup SCA to see where we need to put this in the management cluster.
	sca, err := corev1alpha1.GetServiceClusterAssignmentByServiceClusterNamespace(ctx, r.NamespacedClient, serviceClusterObj.GetNamespace())
	if err != nil {
		return result, fmt.Errorf("getting ServiceClusterAssignment: %w", err)
	}

	// Build desired management cluster object
	desiredManagementClusterObj := serviceClusterObj.DeepCopy()
	desiredManagementClusterObj.SetGroupVersionKind(r.ManagementClusterGVK)
	if err := unstructured.SetNestedField(
		desiredManagementClusterObj.Object, map[string]interface{}{}, "metadata"); err != nil {
		return result, fmt.Errorf("deleting %s .metadata: %w", r.ManagementClusterGVK.Kind, err)
	}
	desiredManagementClusterObj.SetName(serviceClusterObj.GetName())
	desiredManagementClusterObj.SetNamespace(sca.Spec.ManagementClusterNamespace.Name)

	// Reconcile
	currentManagementClusterObj := r.newManagementObject()
	err = r.Get(ctx, types.NamespacedName{
		Name:      desiredManagementClusterObj.GetName(),
		Namespace: desiredManagementClusterObj.GetNamespace(),
	}, currentManagementClusterObj)
	if err != nil && !errors.IsNotFound(err) {
		return result, fmt.Errorf("getting %s: %w", r.ManagementClusterGVK.Kind, err)
	}
	if errors.IsNotFound(err) {
		// Create the management cluster object
		if err = r.Create(ctx, desiredManagementClusterObj); err != nil {
			return result, fmt.Errorf("creating %s: %w", r.ManagementClusterGVK.Kind, err)
		}
	}

	return result, nil
}

func (r *AdoptionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := controller.New(
		strings.ToLower(r.ServiceClusterGVK.Kind),
		mgr, controller.Options{
			Reconciler: r,
		})
	if err != nil {
		return fmt.Errorf("creating controller: %w", err)
	}

	return c.Watch(
		source.NewKindWithCache(r.newServiceObject(), r.ServiceClusterCache),
		&handler.EnqueueRequestForObject{},
		util.PredicateFn(func(obj runtime.Object) bool {
			// we are only interested in unowned objects
			meta, ok := obj.(metav1.Object)
			if !ok {
				return false
			}
			return !owner.IsOwned(meta)
		}))
}

func (r *AdoptionReconciler) newServiceObject() *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(r.ServiceClusterGVK)
	return obj
}

func (r *AdoptionReconciler) newManagementObject() *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(r.ManagementClusterGVK)
	return obj
}

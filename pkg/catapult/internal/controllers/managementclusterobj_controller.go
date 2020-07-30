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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/kubermatic/utils/pkg/owner"
	"github.com/kubermatic/utils/pkg/util"

	corev1alpha1 "k8c.io/kubecarrier/pkg/apis/core/v1alpha1"
)

const catapultControllerFinalizer string = "catapult.kubecarrier.io/controller"

// ManagementClusterObjReconciler reconciles CRD instances in the management cluster,
// by creating a matching instance in the service cluster and syncing it's status back.
type ManagementClusterObjReconciler struct {
	client.Client
	Log              logr.Logger
	NamespacedClient client.Client
	Scheme           *runtime.Scheme

	ServiceClusterClient              client.Client
	ServiceClusterCache               cache.Cache
	ProviderNamespace, ServiceCluster string

	// Dynamic types we work with
	ManagementClusterGVK, ServiceClusterGVK schema.GroupVersionKind
}

// +kubebuilder:rbac:groups=kubecarrier.io,resources=serviceclusterassignments,verbs=get;list;watch
// +kubebuilder:rbac:groups=kubecarrier.io,resources=serviceclusterassignments/status,verbs=get

func (r *ManagementClusterObjReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var (
		result ctrl.Result
		ctx    = context.Background()
	)

	managementClusterObj := r.newManagementObject()
	if err := r.Get(ctx, req.NamespacedName, managementClusterObj); err != nil {
		return result, client.IgnoreNotFound(err)
	}

	if !managementClusterObj.GetDeletionTimestamp().IsZero() {
		if err := r.handleDeletion(ctx, managementClusterObj); err != nil {
			return result, fmt.Errorf("handling deletion: %w", err)
		}
		return result, nil
	}

	if util.AddFinalizer(managementClusterObj, catapultControllerFinalizer) {
		if err := r.Update(ctx, managementClusterObj); err != nil {
			return result, fmt.Errorf("updating %s finalizers: %w", r.ManagementClusterGVK.Kind, err)
		}
	}

	// There needs to be a ServiceClusterAssignment Object
	// so we know where to put this object on the ServiceCluster.
	sca := &corev1alpha1.ServiceClusterAssignment{}
	if err := r.NamespacedClient.Get(ctx, types.NamespacedName{
		Name:      managementClusterObj.GetNamespace() + "." + r.ServiceCluster,
		Namespace: r.ProviderNamespace,
	}, sca); err != nil {
		return result, fmt.Errorf("getting ServiceClusterAssignment: %w", err)
	}
	if !sca.IsReady() {
		// SCA not yet ready
		result.Requeue = true
		return result, nil
	}

	// Build desired service cluster object
	desiredServiceClusterObj := managementClusterObj.DeepCopy()
	if err := unstructured.SetNestedField(
		desiredServiceClusterObj.Object, map[string]interface{}{}, "metadata"); err != nil {
		return result, fmt.Errorf(
			"deleting %s .metadata: %w", r.ServiceClusterGVK.Kind, err)
	}
	delete(desiredServiceClusterObj.Object, "status")
	desiredServiceClusterObj.SetGroupVersionKind(r.ServiceClusterGVK)
	desiredServiceClusterObj.SetName(managementClusterObj.GetName())
	desiredServiceClusterObj.SetNamespace(sca.Status.ServiceClusterNamespace.Name)
	if _, err := owner.SetOwnerReference(
		managementClusterObj, desiredServiceClusterObj, r.Scheme); err != nil {
		return result, fmt.Errorf("setting owner reference: %w", err)
	}

	// Reconcile
	currentServiceClusterObj := r.newServiceObject()
	err := r.ServiceClusterClient.Get(ctx, types.NamespacedName{
		Name:      desiredServiceClusterObj.GetName(),
		Namespace: desiredServiceClusterObj.GetNamespace(),
	}, currentServiceClusterObj)
	if err != nil && !errors.IsNotFound(err) {
		return result, fmt.Errorf(
			"getting %s: %w", r.ServiceClusterGVK.Kind, err)
	}

	if errors.IsNotFound(err) {
		// Create the service cluster object
		if err = r.ServiceClusterClient.Create(ctx, desiredServiceClusterObj); err != nil {
			return result, fmt.Errorf(
				"creating %s: %w", r.ServiceClusterGVK.Kind, err)
		}
		return result, nil
	}

	// Make sure we take ownership of the service cluster instance,
	// if the OwnerReference is not yet set.
	if _, err := owner.SetOwnerReference(
		managementClusterObj, currentServiceClusterObj, r.Scheme); err != nil {
		return result, fmt.Errorf("setting owner reference: %w", err)
	}

	// Update existing service cluster instance
	// This is a bit complicated, because we want to support arbitrary fields and not only .spec.
	// Thats why we are updating everything, with the exception of .status and .metadata
	updatedServiceClusterObj := desiredServiceClusterObj.DeepCopy()
	if err := unstructured.SetNestedField(
		updatedServiceClusterObj.Object,
		currentServiceClusterObj.Object["metadata"], "metadata"); err != nil {
		return result, fmt.Errorf(
			"updating %s .metadata: %w", r.ServiceClusterGVK.Kind, err)
	}
	if err := unstructured.SetNestedField(
		updatedServiceClusterObj.Object,
		currentServiceClusterObj.Object["status"], "status"); err != nil {
		return result, fmt.Errorf("updating %s .status: %w", r.ServiceClusterGVK.Kind, err)
	}
	if err := r.ServiceClusterClient.Update(ctx, updatedServiceClusterObj); err != nil {
		return result, fmt.Errorf(
			"updating %s: %w", r.ServiceClusterGVK.Kind, err)
	}

	// Sync Status from service cluster to management cluster
	if err := unstructured.SetNestedField(
		managementClusterObj.Object,
		currentServiceClusterObj.Object["status"], "status"); err != nil {
		return result, fmt.Errorf(
			"updating %s .status: %w", r.ManagementClusterGVK.Kind, err)
	}
	if err = util.UpdateObservedGeneration(
		currentServiceClusterObj, managementClusterObj); err != nil {
		return result, fmt.Errorf(
			"update observedGeneration, by comparing %s to %s: %w",
			r.ServiceClusterGVK.Kind, r.ManagementClusterGVK.Kind, err)
	}
	if err = r.Status().Update(ctx, managementClusterObj); err != nil {
		return result, fmt.Errorf(
			"updating %s status: %w", r.ManagementClusterGVK.Kind, err)
	}

	return result, nil
}

func (r *ManagementClusterObjReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(r.newManagementObject()).
		Watches(
			source.NewKindWithCache(r.newServiceObject(), r.ServiceClusterCache),
			owner.EnqueueRequestForOwner(r.newManagementObject(), mgr.GetScheme()),
		).
		Complete(r)
}

func (r *ManagementClusterObjReconciler) newServiceObject() *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(r.ServiceClusterGVK)
	return obj
}

func (r *ManagementClusterObjReconciler) newManagementObject() *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(r.ManagementClusterGVK)
	return obj
}

func (r *ManagementClusterObjReconciler) handleDeletion(
	ctx context.Context, managementClusterObj *unstructured.Unstructured,
) error {
	sca := &corev1alpha1.ServiceClusterAssignment{}
	err := r.NamespacedClient.Get(ctx, types.NamespacedName{
		Name:      managementClusterObj.GetNamespace() + "." + r.ServiceCluster,
		Namespace: r.ProviderNamespace,
	}, sca)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("getting ServiceClusterAssignment: %w", err)
	}

	if err == nil {
		// if the ServiceClusterAssignment is not found,
		// we can skip deleting the instance on the ServiceCluster.
		serviceClusterObj := &unstructured.Unstructured{}
		serviceClusterObj.SetGroupVersionKind(r.ServiceClusterGVK)
		err := r.ServiceClusterClient.Get(ctx, types.NamespacedName{
			Name:      managementClusterObj.GetName(),
			Namespace: sca.Status.ServiceClusterNamespace.Name,
		}, serviceClusterObj)
		if err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("getting %s: %w", r.ServiceClusterGVK.Kind, err)
		}

		if err == nil && serviceClusterObj.GetDeletionTimestamp().IsZero() {
			if err = r.ServiceClusterClient.Delete(ctx, serviceClusterObj); err != nil {
				return fmt.Errorf("deleting %s: %w", r.ServiceClusterGVK.Kind, err)
			}
			return nil
		}

		if !errors.IsNotFound(err) {
			// wait until object is realy gone
			return nil
		}
	}

	if util.RemoveFinalizer(managementClusterObj, catapultControllerFinalizer) {
		if err := r.Update(ctx, managementClusterObj); err != nil {
			return fmt.Errorf("updating %s finalizers: %w", r.ManagementClusterGVK.Kind, err)
		}
	}
	return nil
}

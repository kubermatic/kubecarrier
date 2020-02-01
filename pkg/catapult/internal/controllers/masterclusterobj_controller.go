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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

// MasterClusterObjReconciler reconciles CRD instances in the master cluster,
// by creating a matching instance in the service cluster and syncing it's status back.
type MasterClusterObjReconciler struct {
	client.Client
	Log                  logr.Logger
	Scheme               *runtime.Scheme
	ServiceClusterClient client.Client
	ServiceClusterCache  cache.Cache

	// Dynamic types we work with
	MasterClusterGVK, ServiceClusterGVK   schema.GroupVersionKind
	MasterClusterType, ServiceClusterType *unstructured.Unstructured

	ProviderNamespace string
	ServiceCluster    string
}

func (r *MasterClusterObjReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var (
		result ctrl.Result
		ctx    = context.Background()
	)

	masterClusterObj := r.MasterClusterType.DeepCopy()
	if err := r.Get(ctx, req.NamespacedName, masterClusterObj); err != nil {
		return result, client.IgnoreNotFound(err)
	}

	// There needs to be a ServiceClusterAssignment Object
	// so we know where to put this object on the ServiceCluster.
	sca, err := r.reconcileServiceClusterAssignment(ctx, masterClusterObj)
	if err != nil {
		return result, fmt.Errorf("reconciling ServiceClusterAssignment: %w", err)
	}
	if readyCondition, _ := sca.Status.GetCondition(
		corev1alpha1.ServiceClusterAssignmentReady,
	); readyCondition.Status != corev1alpha1.ConditionTrue {
		// SCA not yet ready
		result.Requeue = true
		return result, nil
	}

	// Build desired service cluster object
	desiredServiceClusterObj := masterClusterObj.DeepCopy()
	if err := unstructured.SetNestedField(
		desiredServiceClusterObj.Object, map[string]interface{}{}, "metadata"); err != nil {
		return result, fmt.Errorf(
			"deleting %s .metadata: %w", r.ServiceClusterGVK.Kind, err)
	}
	delete(desiredServiceClusterObj.Object, "status")
	desiredServiceClusterObj.SetGroupVersionKind(r.ServiceClusterGVK)
	desiredServiceClusterObj.SetName(masterClusterObj.GetName())
	desiredServiceClusterObj.SetNamespace(sca.Status.ServiceClusterNamespace.Name)

	if _, err := util.InsertOwnerReference(
		masterClusterObj, desiredServiceClusterObj, r.Scheme); err != nil {
		return result, fmt.Errorf("inserting owner reference: %w", err)
	}

	// Reconcile
	currentServiceClusterObj := r.ServiceClusterType.DeepCopy()
	err = r.ServiceClusterClient.Get(ctx, types.NamespacedName{
		Name:      desiredServiceClusterObj.GetName(),
		Namespace: desiredServiceClusterObj.GetNamespace(),
	}, currentServiceClusterObj)
	if err != nil && !errors.IsNotFound(err) {
		return result, fmt.Errorf(
			"getting %s: %w", r.ServiceClusterGVK.Kind, err)
	}

	if errors.IsNotFound(err) {
		// Create the service cluster object
		fmt.Println(desiredServiceClusterObj)
		if err = r.ServiceClusterClient.Create(ctx, desiredServiceClusterObj); err != nil {
			return result, fmt.Errorf(
				"creating %s: %w", r.ServiceClusterGVK.Kind, err)
		}
		return result, nil
	}

	// Make sure we take ownership of the service cluster instance,
	// if the OwnerReference is not yet set.
	if _, err := util.InsertOwnerReference(
		masterClusterObj, currentServiceClusterObj, r.Scheme); err != nil {
		return result, fmt.Errorf("inserting owner reference: %w", err)
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

	// Sync Status from service cluster to master cluster
	if err := unstructured.SetNestedField(
		masterClusterObj.Object,
		currentServiceClusterObj.Object["status"], "status"); err != nil {
		return result, fmt.Errorf(
			"updating %s .status: %w", r.MasterClusterGVK.Kind, err)
	}
	if err = util.UpdateObservedGeneration(
		currentServiceClusterObj, masterClusterObj); err != nil {
		return result, fmt.Errorf(
			"update observedGeneration, by comparing %s to %s: %w",
			r.ServiceClusterGVK.Kind, r.MasterClusterGVK.Kind, err)
	}
	if err = r.Status().Update(ctx, masterClusterObj); err != nil {
		return result, fmt.Errorf(
			"updating %s status: %w", r.MasterClusterGVK.Kind, err)
	}

	return result, nil
}

func (r *MasterClusterObjReconciler) SetupWithManager(mgr ctrl.Manager) error {
	serviceClusterSource := &source.Kind{Type: r.ServiceClusterType}
	if err := serviceClusterSource.InjectCache(r.ServiceClusterCache); err != nil {
		return fmt.Errorf("injecting cache: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(r.MasterClusterType).
		Watches(
			source.Func(serviceClusterSource.Start),
			util.EnqueueRequestForOwner(r.MasterClusterType, r.Scheme),
		).
		Complete(r)
}

func (r *MasterClusterObjReconciler) reconcileServiceClusterAssignment(
	ctx context.Context, masterClusterObj *unstructured.Unstructured,
) (*corev1alpha1.ServiceClusterAssignment, error) {
	scaList := &corev1alpha1.ServiceClusterAssignmentList{}
	if err := r.List(
		ctx, scaList,
		client.InNamespace(r.ProviderNamespace),
		client.MatchingFields{
			corev1alpha1.
				ServiceClusterAssignmentMasterClusterNamespaceFieldIndex: masterClusterObj.GetNamespace(),
			corev1alpha1.
				ServiceClusterAssignmentServiceClusterFieldIndex: r.ServiceCluster,
		}); err != nil {
		return nil, fmt.Errorf(
			"listing ServiceClusterAssignments for ServiceCluster and MasterCluster Namespace: %w", err)
	}
	if len(scaList.Items) > 1 {
		return nil, fmt.Errorf(
			"multiple ServiceClusterAssignments for same ServiceCluster/MasterCluster Namespace found")
	}

	var serviceClusterAssignment *corev1alpha1.ServiceClusterAssignment
	if len(scaList.Items) == 0 {
		// We have to create a ServiceClusterAssignment.
		serviceClusterAssignment = &corev1alpha1.ServiceClusterAssignment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      masterClusterObj.GetNamespace() + "." + r.ServiceCluster,
				Namespace: r.ProviderNamespace,
			},
			Spec: corev1alpha1.ServiceClusterAssignmentSpec{
				MasterClusterNamespace: corev1alpha1.ObjectReference{
					Name: masterClusterObj.GetNamespace(),
				},
				ServiceCluster: corev1alpha1.ObjectReference{
					Name: r.ServiceCluster,
				},
			},
		}

		if err := r.Create(ctx, serviceClusterAssignment); err != nil {
			return nil, fmt.Errorf("creating ServiceClusterAssignment: %w", err)
		}
	}

	if len(scaList.Items) == 1 {
		serviceClusterAssignment = &scaList.Items[0]
	}

	return serviceClusterAssignment, nil
}

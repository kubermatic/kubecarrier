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
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/reconcile"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

type Controller interface {
	// GetObj - return origin controller object
	GetObj() object
	GetManifests(context.Context) ([]unstructured.Unstructured, error)
	GetReadyConditionStatus() operatorv1alpha1.ConditionStatus
	SetTerminatingCondition(context.Context) bool
	SetUnReadyCondition()
	SetReadyCondition()
}

type BaseReconciler struct {
	Client    client.Client
	Scheme    *runtime.Scheme
	Log       logr.Logger
	Name      string
	Finalizer string
}

// FetchObject - fetch the object
func (r *BaseReconciler) FetchObject(ctx context.Context, req ctrl.Request, ctr Controller) error {
	if err := r.Client.Get(ctx, req.NamespacedName, ctr.GetObj()); err != nil {
		// If the object is already gone, we just ignore the NotFound error.
		return client.IgnoreNotFound(err)
	}
	return nil
}

// Reconcile - base reconcile logic
func (r *BaseReconciler) Reconcile(ctx context.Context, req ctrl.Request, ctr Controller) (ctrl.Result, error) {

	if err := r.FetchObject(ctx, req, ctr); err != nil {
		return ctrl.Result{}, err
	}

	if !ctr.GetObj().GetDeletionTimestamp().IsZero() {
		if err := r.HandleDeletion(ctx, ctr); err != nil {
			return ctrl.Result{}, fmt.Errorf("handle deletion: %w", err)
		}
		return ctrl.Result{}, nil
	}

	if util.AddFinalizer(ctr.GetObj(), r.Finalizer) {
		if err := r.Client.Update(ctx, ctr.GetObj()); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating %s finalizers: %w", r.Name, err)
		}
	}

	objects, err := ctr.GetManifests(ctx)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("creating %s manifests: %w", r.Name, err)
	}

	deploymentIsReady, err := r.ReconcileOwnedObjects(ctx, ctr.GetObj(), objects)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := r.UpdateStatus(ctx, ctr, deploymentIsReady); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// HandleDeletion - handle the deletion of the object (Remove the objects that the object owns, and remove the finalizer).
func (r *BaseReconciler) HandleDeletion(ctx context.Context, ctr Controller) error {

	// Update the object Status to Terminating.
	if ctr.SetTerminatingCondition(ctx) {
		if err := r.Client.Status().Update(ctx, ctr.GetObj()); err != nil {
			return fmt.Errorf("updating %s status: %w", r.Name, err)
		}
	}
	// Delete Objects.
	allCleared := true
	objects, err := ctr.GetManifests(ctx)
	if err != nil {
		return fmt.Errorf("deletion: manifests: %w", err)
	}
	for _, obj := range objects {
		err := r.Client.Delete(ctx, &obj)
		if errors.IsNotFound(err) {
			continue
		}
		allCleared = false
		if err != nil {
			return fmt.Errorf("deleting %s: %w", obj, err)
		}
	}

	if allCleared {
		// Remove Finalizer
		if util.RemoveFinalizer(ctr.GetObj(), r.Finalizer) {
			if err := r.Client.Update(ctx, ctr.GetObj()); err != nil {
				return fmt.Errorf("updating %s finalizers: %w", r.Name, err)
			}
		}
	}
	return nil
}

// UpdateStatus - update the status of the object
func (r *BaseReconciler) UpdateStatus(ctx context.Context, ctr Controller, deploymentIsReady bool) error {
	var updateStatus bool
	readyConditionStatus := ctr.GetReadyConditionStatus()

	if !deploymentIsReady && readyConditionStatus != operatorv1alpha1.ConditionFalse {
		updateStatus = true
		ctr.SetUnReadyCondition()
	}

	if deploymentIsReady && readyConditionStatus != operatorv1alpha1.ConditionTrue {
		updateStatus = true
		ctr.SetReadyCondition()
	}

	if updateStatus {
		if err := r.Client.Status().Update(ctx, ctr.GetObj()); err != nil {
			return fmt.Errorf("updating %s status: %w", r.Name, err)
		}
	}
	return nil
}

// ReconcileOwnedObjects - reconcile the objects that owned by obj
func (r *BaseReconciler) ReconcileOwnedObjects(ctx context.Context, obj object, objects []unstructured.Unstructured) (bool, error) {
	var deploymentIsReady bool
	for _, object := range objects {
		if err := addOwnerReference(obj, &object, r.Scheme); err != nil {
			return false, err
		}
		curObj, err := reconcile.Unstructured(ctx, r.Log, r.Client, &object)
		if err != nil {
			return false, fmt.Errorf("reconcile kind: %s, err: %w", object.GroupVersionKind().Kind, err)
		}

		switch ctr := curObj.(type) {
		case *appsv1.Deployment:
			deploymentIsReady = util.DeploymentIsAvailable(ctr)
		}
	}
	return deploymentIsReady, nil
}

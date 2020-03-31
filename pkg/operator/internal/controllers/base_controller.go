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
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/kubermatic/kubecarrier/pkg/internal/owner"
	"github.com/kubermatic/kubecarrier/pkg/internal/reconcile"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

type Component interface {
	runtime.Object
	metav1.Object
	SetTerminatingCondition() bool
	SetUnReadyCondition() bool
	SetReadyCondition() bool
}

type ControllerStrategy interface {
	// GetObj - return origin controller object
	GetObj() Component
	GetManifests(context.Context, Component) ([]unstructured.Unstructured, error)
	GetDeletionObjectTypes() []runtime.Object
	AddWatches(builder *builder.Builder, scheme *runtime.Scheme) *builder.Builder
}

type BaseReconciler struct {
	Controller ControllerStrategy
	Client     client.Client
	Scheme     *runtime.Scheme
	RESTMapper meta.RESTMapper
	Log        logr.Logger
	Name       string
	Finalizer  string
}

func NewBaseReconciler(ctr ControllerStrategy, c client.Client, scheme *runtime.Scheme, RESTMapper meta.RESTMapper, log logr.Logger, name, finalizer string) *BaseReconciler {
	return &BaseReconciler{
		Controller: ctr,
		Client:     c,
		Scheme:     scheme,
		RESTMapper: RESTMapper,
		Log:        log,
		Name:       name,
		Finalizer:  finalizer,
	}
}

func (r *BaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	controllerBuilder := ctrl.NewControllerManagedBy(mgr)
	controllerBuilder = controllerBuilder.For(r.Controller.GetObj())
	enqueuer := owner.EnqueueRequestForOwner(r.Controller.GetObj(), mgr.GetScheme())
	for _, obj := range r.Controller.GetDeletionObjectTypes() {
		controllerBuilder = controllerBuilder.Watches(&source.Kind{Type: obj}, enqueuer)
	}
	controllerBuilder = r.Controller.AddWatches(controllerBuilder, mgr.GetScheme())
	return controllerBuilder.Complete(r)
}

// Reconcile function reconciles the object which specified by the request. Currently, it does the following:
// 1. Fetch the Catapult object.
// 2. Handle the deletion of the object (Remove the objects that the it owns, and remove the finalizer).
// 3. Reconcile the objects that owned by object.
// 4. Update the status of the object.
func (r *BaseReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	obj := r.Controller.GetObj()

	if err := r.Client.Get(ctx, req.NamespacedName, obj); err != nil {
		// If the object is already gone, we just ignore the NotFound error.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !obj.GetDeletionTimestamp().IsZero() {
		if err := r.HandleDeletion(ctx, obj, r.Controller.GetDeletionObjectTypes()); err != nil {
			return ctrl.Result{}, fmt.Errorf("handle deletion: %w", err)
		}
		return ctrl.Result{}, nil
	}

	if len(r.Finalizer) > 0 {
		if util.AddFinalizer(obj, r.Finalizer) {
			if err := r.Client.Update(ctx, obj); err != nil {
				return ctrl.Result{}, fmt.Errorf("updating %s finalizers: %w", r.Name, err)
			}
		}
	}

	objects, err := r.Controller.GetManifests(ctx, obj)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("creating %s manifests: %w", r.Name, err)
	}

	deploymentIsReady, err := r.ReconcileOwnedObjects(ctx, obj, objects)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := r.UpdateStatus(ctx, obj, deploymentIsReady); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// HandleDeletion - handle the deletion of the object (Remove the objects that the object owns, and remove the finalizer).
func (r *BaseReconciler) HandleDeletion(ctx context.Context, c Component, ownObjectsTypes []runtime.Object) error {
	// Update the object Status to Terminating.
	if c.SetTerminatingCondition() {
		if err := r.Client.Status().Update(ctx, c); err != nil {
			return fmt.Errorf("updating %s status: %w", r.Name, err)
		}
	}

	cleanedUp, err := util.DeleteObjects(ctx, r.Client, r.Scheme, ownObjectsTypes, owner.OwnedBy(c, r.Scheme))
	if err != nil {
		return fmt.Errorf("DeleteObjects: %w", err)
	}
	if cleanedUp && len(r.Finalizer) > 0 && util.RemoveFinalizer(c, r.Finalizer) {
		if err := r.Client.Update(ctx, c); err != nil {
			return fmt.Errorf("updating %s Status: %w", r.Name, err)
		}
	}
	return nil
}

// UpdateStatus - update the status of the object
func (r *BaseReconciler) UpdateStatus(ctx context.Context, c Component, deploymentIsReady bool) error {
	var statusChanged bool

	if deploymentIsReady {
		statusChanged = c.SetReadyCondition()
	} else {
		statusChanged = c.SetUnReadyCondition()
	}

	if statusChanged {
		if err := r.Client.Status().Update(ctx, c); err != nil {
			return fmt.Errorf("updating %s status: %w", r.Name, err)
		}
	}
	return nil
}

// ReconcileOwnedObjects - reconcile the objects that owned by obj
func (r *BaseReconciler) ReconcileOwnedObjects(ctx context.Context, obj Component, objects []unstructured.Unstructured) (bool, error) {
	var deploymentIsReady bool
	for _, object := range objects {
		gvk, err := apiutil.GVKForObject(&object, r.Scheme)
		if err != nil {
			return false, err
		}
		restMapping, err := r.RESTMapper.RESTMapping(schema.GroupKind{
			Group: gvk.Group,
			Kind:  gvk.Kind,
		}, gvk.Version)
		if err != nil {
			return false, err
		}

		switch restMapping.Scope.Name() {
		case meta.RESTScopeNameNamespace:
			if err := controllerutil.SetControllerReference(obj, &object, r.Scheme); err != nil {
				return false, err
			}
		case meta.RESTScopeNameRoot:
			owner.SetOwnerReference(obj, &object, r.Scheme)
		default:
			return false, fmt.Errorf("unknown REST scope: %s", restMapping.Scope.Name())
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

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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

const (
	internalcrdControllerFinalizer string = "internalcrd.kubecarrier.io/controller"
)

type InternalObjectReconciler struct {
	MasterClient         client.Client
	MasterScheme         *runtime.Scheme
	ServiceClient        client.Client
	Log                  logr.Logger
	InternalGVK          schema.GroupVersionKind
	ServiceClusterGVK    schema.GroupVersionKind
	ProviderNamespace    string
	KubecarrierNamespace string
}

// +kubebuilder:rbac:groups=kubecarrier.io,resources=serviceclusterassignments,verbs=get;list;watch
// +kubebuilder:rbac:groups=kubecarrier.io,resources=serviceclusterassignments/status,verbs=get
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=tenants,verbs=get;list;watch
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=tenants/status,verbs=get

func (r *InternalObjectReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var (
		ctx = context.Background()
		log = r.Log.WithValues(r.InternalGVK.String(), req.NamespacedName)
	)

	internalObj := &unstructured.Unstructured{}
	internalObj.SetGroupVersionKind(r.InternalGVK)
	if err := r.MasterClient.Get(ctx, req.NamespacedName, internalObj); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !internalObj.GetDeletionTimestamp().IsZero() {
		if err := r.handleDeletion(ctx, log, internalObj); err != nil {
			return ctrl.Result{}, fmt.Errorf("handling deletion: %w", err)
		}
		return ctrl.Result{}, nil
	}

	if util.AddFinalizer(internalObj, internalcrdControllerFinalizer) {
		if err := r.MasterClient.Update(ctx, internalObj); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating Internal CRD finalizers: %w", err)
		}
	}
	serviceClusterObj := &unstructured.Unstructured{}
	targetNamespace, err := r.targetNamespace(ctx, internalObj)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("target-namespace: %w", err)
	}

	configureDesiredObj := func() error {
		internalObjCpy := internalObj.DeepCopy()
		serviceClusterObj.SetGroupVersionKind(r.ServiceClusterGVK)
		serviceClusterObj.SetNamespace(targetNamespace)
		if _, err := util.InsertOwnerReference(internalObj, serviceClusterObj, r.MasterScheme); err != nil {
			return fmt.Errorf("cannot insert owner reference: %w", err)
		}
		for k := range internalObj.Object {
			if k != "status" && k != "metadata" && k != "kind" && k != "apiVersion" {
				serviceClusterObj.Object[k] = internalObjCpy.Object[k]
			}
		}
		return nil
	}
	if err := configureDesiredObj(); err != nil {
		return ctrl.Result{}, err
	}

	op, err := ctrl.CreateOrUpdate(ctx, r.ServiceClient, serviceClusterObj, configureDesiredObj)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("cannot create of update desired service cluster's obj: %w", err)
	}
	log.Info("serviceCluster's CRD: %s", op)
	return ctrl.Result{}, nil
}

func (r *InternalObjectReconciler) handleDeletion(ctx context.Context, log logr.Logger, internalObj *unstructured.Unstructured) error {
	targetNamespace, err := r.targetNamespace(ctx, internalObj)
	if err != nil {
		return fmt.Errorf("target-namespace: %w", err)
	}
	serviceClusterObj := &unstructured.Unstructured{}
	serviceClusterObj.SetGroupVersionKind(r.ServiceClusterGVK)
	serviceClusterObj.SetNamespace(targetNamespace)
	serviceClusterObj.SetName(internalObj.GetName())

	err = r.ServiceClient.Delete(ctx, serviceClusterObj)
	switch {
	case err == nil:
		return nil
	case errors.IsNotFound(err):
		if util.RemoveFinalizer(internalObj, internalcrdControllerFinalizer) {
			if err := r.MasterClient.Update(ctx, internalObj); err != nil {
				return fmt.Errorf("updating master cluster CRD: %w", err)
			}
		}
		return nil
	default:
		return fmt.Errorf("deleting service's cluster CRD: %w", err)
	}
}

func (r *InternalObjectReconciler) targetNamespace(ctx context.Context, obj *unstructured.Unstructured) (namespace string, err error) {
	tenantList := &catalogv1alpha1.TenantList{}
	if err := r.MasterClient.List(ctx, tenantList,
		client.InNamespace(r.KubecarrierNamespace),
		client.MatchingFields{
			catalogv1alpha1.TenantNamespaceFieldIndex: obj.GetNamespace(),
		},
	); err != nil {
		return "", err
	}
	if len(tenantList.Items) != 1 {
		return "", fmt.Errorf("expected single tenant matching tenant namespace, found %d", len(tenantList.Items))
	}
	tenant := tenantList.Items[0]
	if !tenant.IsReady() {
		return "", fmt.Errorf("tenant not yet ready")
	}
	serviceClusterAssignment := &corev1alpha1.ServiceClusterAssignment{}
	if err := r.MasterClient.Get(ctx, types.NamespacedName{
		Namespace: r.ProviderNamespace,
		Name:      tenant.GetName(),
	}, serviceClusterAssignment); err != nil {
		return "", fmt.Errorf("getting serviceClusterAssignment: %w", err)
	}
	if cond, _ := serviceClusterAssignment.Status.GetCondition(corev1alpha1.ServiceClusterAssignmentReady); cond.Status != corev1alpha1.ConditionTrue {
		return "", fmt.Errorf("serviceClusterAssignment not yet ready")
	}
	return serviceClusterAssignment.Status.NamespaceName, nil
}

func (r *InternalObjectReconciler) SetupWithManagers(serviceMgr, masterMgr ctrl.Manager) error {
	serviceClusterType := &unstructured.Unstructured{}
	serviceClusterType.SetGroupVersionKind(r.ServiceClusterGVK)

	serviceClusterSource := &source.Kind{Type: serviceClusterType}
	if err := serviceMgr.SetFields(serviceClusterSource); err != nil {
		return err
	}

	internalType := &unstructured.Unstructured{}
	internalType.SetGroupVersionKind(r.InternalGVK)

	enqueuer, err := util.EnqueueRequestForOwner(internalType, r.MasterScheme)
	if err != nil {
		return fmt.Errorf("creating internal enqueuer: %w", err)
	}

	return ctrl.NewControllerManagedBy(masterMgr).
		Named(strings.ToLower(r.InternalGVK.Kind)+"."+r.InternalGVK.Group).
		For(internalType).
		Watches(source.Func(serviceClusterSource.Start), enqueuer).
		Complete(r)
}

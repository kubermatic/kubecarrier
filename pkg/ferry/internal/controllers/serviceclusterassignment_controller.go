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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/kubermatic/utils/pkg/owner"
	"github.com/kubermatic/utils/pkg/util"

	corev1alpha1 "k8c.io/kubecarrier/pkg/apis/core/v1alpha1"
)

const serviceClusterAssignmentControllerFinalizer string = "serviceclusterassignment.kubecarrier.io/controller"

// ServiceClusterAssignmentReconciler reconciles a ServiceClusterAssignment object
type ServiceClusterAssignmentReconciler struct {
	Log logr.Logger

	ManagementClient   client.Client
	ServiceClient      client.Client
	ServiceCache       cache.Cache
	ManagementScheme   *runtime.Scheme
	ServiceClusterName string
}

// +kubebuilder:rbac:groups=kubecarrier.io,resources=serviceclusterassignments,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=kubecarrier.io,resources=serviceclusterassignments/status,verbs=get;update;patch
// https://github.com/kubermatic/kubecarrier/issues/143
// +servicecluster:kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete

func (r *ServiceClusterAssignmentReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("serviceClusterAssignment", req.NamespacedName)

	serviceClusterAssignment := &corev1alpha1.ServiceClusterAssignment{}
	if err := r.ManagementClient.Get(ctx, req.NamespacedName, serviceClusterAssignment); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// handle Deletion
	if !serviceClusterAssignment.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, log, serviceClusterAssignment); err != nil {
			return ctrl.Result{}, fmt.Errorf("handling deletion: %w", err)
		}
		return ctrl.Result{}, nil
	}

	if util.AddFinalizer(serviceClusterAssignment, serviceClusterAssignmentControllerFinalizer) {
		if err := r.ManagementClient.Update(ctx, serviceClusterAssignment); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating %s finalizers: %w", serviceClusterAssignment.Kind, err)
		}
	}

	ns, err := ensureUniqueNamespace(
		ctx, r.ServiceClient,
		serviceClusterAssignment,
		serviceClusterAssignment.Spec.ManagementClusterNamespace.Name,
		r.ManagementScheme)
	if err != nil {
		serviceClusterAssignment.Status.ObservedGeneration = serviceClusterAssignment.Generation
		serviceClusterAssignment.Status.SetCondition(corev1alpha1.ServiceClusterAssignmentCondition{
			Type:    corev1alpha1.ServiceClusterAssignmentReady,
			Status:  corev1alpha1.ConditionFalse,
			Message: err.Error(),
			Reason:  "CreatingNamespace",
		})
		if err := r.ManagementClient.Status().Update(ctx, serviceClusterAssignment); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating ServiceClusterAssignment Status: %w", err)
		}
		return ctrl.Result{}, fmt.Errorf("cannot create TenantAssignment namespace: %w", err)
	}

	serviceClusterAssignment.Status.ServiceClusterNamespace = &corev1alpha1.ObjectReference{
		Name: ns.Name,
	}
	serviceClusterAssignment.Status.ObservedGeneration = serviceClusterAssignment.Generation
	serviceClusterAssignment.Status.SetCondition(corev1alpha1.ServiceClusterAssignmentCondition{
		Type:    corev1alpha1.ServiceClusterAssignmentReady,
		Status:  corev1alpha1.ConditionTrue,
		Message: "Namespace has been setup.",
		Reason:  "SetupComplete",
	})
	if err = r.ManagementClient.Status().Update(ctx, serviceClusterAssignment); err != nil {
		return ctrl.Result{}, fmt.Errorf("updating ServiceClusterAssignment Status: %w", err)
	}
	return ctrl.Result{}, nil
}

func (r *ServiceClusterAssignmentReconciler) handleDeletion(ctx context.Context, log logr.Logger, serviceClusterAssignment *corev1alpha1.ServiceClusterAssignment) error {
	// Update the Provider Status to Terminating.
	readyCondition, _ := serviceClusterAssignment.Status.GetCondition(corev1alpha1.ServiceClusterAssignmentReady)
	if readyCondition.Status != corev1alpha1.ConditionFalse ||
		readyCondition.Status == corev1alpha1.ConditionFalse && readyCondition.Reason != corev1alpha1.TerminatingReason {
		serviceClusterAssignment.Status.ObservedGeneration = serviceClusterAssignment.Generation
		serviceClusterAssignment.Status.SetCondition(corev1alpha1.ServiceClusterAssignmentCondition{
			Type:    corev1alpha1.ServiceClusterAssignmentReady,
			Status:  corev1alpha1.ConditionFalse,
			Reason:  corev1alpha1.TerminatingReason,
			Message: "ServiceClusterAssignment is being terminated",
		})
		if err := r.ManagementClient.Status().Update(ctx, serviceClusterAssignment); err != nil {
			return fmt.Errorf("updating ServiceClusterAssignment status: %w", err)
		}
	}

	nsList := &corev1.NamespaceList{}
	if err := r.ServiceClient.List(ctx, nsList, owner.OwnedBy(serviceClusterAssignment, r.ManagementScheme)); err != nil {
		return fmt.Errorf("listing Namespaces: %w", err)
	}

	if len(nsList.Items) > 0 {
		for _, ns := range nsList.Items {
			if err := r.ServiceClient.Delete(ctx, &ns); err != nil && !errors.IsNotFound(err) {
				return fmt.Errorf("cannot delete ns %s: %w", ns.Name, err)
			}
		}
		return nil
	}

	// Remove Finalizer
	if util.RemoveFinalizer(serviceClusterAssignment, serviceClusterAssignmentControllerFinalizer) {
		if err := r.ManagementClient.Update(ctx, serviceClusterAssignment); err != nil {
			return fmt.Errorf("updating %s finalizers: %w", serviceClusterAssignment.Kind, err)
		}
	}
	return nil
}

func (r *ServiceClusterAssignmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.ServiceClusterAssignment{}).
		Watches(source.NewKindWithCache(&corev1.Namespace{}, r.ServiceCache), owner.EnqueueRequestForOwner(&corev1alpha1.ServiceClusterAssignment{}, r.ManagementScheme)).
		WithEventFilter(util.PredicateFn(func(obj runtime.Object) bool {
			if serviceClusterAssignment, ok := obj.(*corev1alpha1.ServiceClusterAssignment); ok {
				if serviceClusterAssignment.Spec.ServiceCluster.Name == r.ServiceClusterName {
					return true
				}
			}

			// for namespace owner reconciliation from the service cluster
			if _, ok := obj.(*corev1.Namespace); ok {
				return true
			}
			return false
		})).
		Complete(r)
}

// ensureUniqueNamespace generates unique namespace for obj if one already doesn't exists
//
// It's required that OwnerReverseFieldIndex exists for corev1.Namespace
func ensureUniqueNamespace(ctx context.Context, c client.Client, ownerObj runtime.Object, prefix string, scheme *runtime.Scheme) (*corev1.Namespace, error) {
	namespaceList := &corev1.NamespaceList{}
	if err := c.List(ctx, namespaceList, owner.OwnedBy(ownerObj, scheme)); err != nil {
		return nil, fmt.Errorf("listing Namespaces: %w", err)
	}

	switch len(namespaceList.Items) {
	case 0:
		// Create Namespace
		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: prefix + "-",
			},
		}
		if _, err := owner.SetOwnerReference(ownerObj, namespace, scheme); err != nil {
			return nil, fmt.Errorf("setting owner reference: %w", err)
		}
		if err := c.Create(ctx, namespace); err != nil {
			return nil, fmt.Errorf("creating Namespace: %w", err)
		}
		return namespace, nil
	case 1:
		return namespaceList.Items[0].DeepCopy(), nil
	default:
		nss := make([]string, len(namespaceList.Items))
		for i, ns := range namespaceList.Items {
			nss[i] = ns.Name
		}
		return nil, fmt.Errorf("multiple owned namespaces found: %s", strings.Join(nss, ","))
	}
}

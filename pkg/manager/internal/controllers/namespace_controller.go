/*
Copyright 2020 The KubeCarrier Authors.

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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

const (
	namespaceControllerFinalizer string = "namespace.kubecarrier.io/controller"
)

// NamespaceReconciler is creating ServiceClusterAssignment objects when Namespaces have the right label set.
type NamespaceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=kubecarrier.io,resources=serviceclusterassignments,verbs=get;list;watch;create;delete

func (r *NamespaceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var (
		ctx    = context.Background()
		result ctrl.Result
	)

	// list all the Namespaces
	nsList := &corev1.NamespaceList{}
	if err := r.List(ctx, nsList); err != nil {
		return result, fmt.Errorf("listing Namespaces: %w", err)
	}

	var (
		desiredServiceClusterAssignments []corev1alpha1.ServiceClusterAssignment
		deletedNamespaces                []corev1.Namespace
	)
	for i, ns := range nsList.Items {
		if !ns.DeletionTimestamp.IsZero() {
			deletedNamespaces = append(deletedNamespaces, nsList.Items[i])
			continue
		}

		scas, err := r.scasFromNamespaceLabels(&ns)
		if err != nil {
			return result, fmt.Errorf("building ServiceClusterAssignments for Namespace %s: %w", ns.Name, err)
		}

		// add finalizer
		if len(scas) > 0 &&
			util.AddFinalizer(&ns, namespaceControllerFinalizer) {
			if err := r.Update(ctx, &ns); err != nil {
				return result, fmt.Errorf("adding finalizers on Namespace %s: %w", ns.Name, err)
			}
		}

		desiredServiceClusterAssignments = append(desiredServiceClusterAssignments, scas...)
	}

	currentServiceClusterAssignmentList := &corev1alpha1.ServiceClusterAssignmentList{}
	if err := r.List(ctx, currentServiceClusterAssignmentList); err != nil {
		return result, fmt.Errorf("listing ServiceClusterAssignments: %w", err)
	}

	if err := r.reconcileServiceClusterAssignments(
		ctx, desiredServiceClusterAssignments, currentServiceClusterAssignmentList.Items); err != nil {
		return result, err
	}

	// remove our finalizers on deleted Namespaces
	for _, ns := range deletedNamespaces {
		if util.RemoveFinalizer(&ns, namespaceControllerFinalizer) {
			if err := r.Update(ctx, &ns); err != nil {
				return result, fmt.Errorf("removing finalizer on Namespace %s: %w", ns.Name, err)
			}
		}
	}

	return result, nil
}

func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	enqueuerForOwner := util.EnqueueRequestForOwner(&corev1.Namespace{}, mgr.GetScheme())
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Watches(&source.Kind{Type: &corev1alpha1.ServiceClusterAssignment{}}, enqueuerForOwner).
		Complete(r)
}

func (r *NamespaceReconciler) reconcileServiceClusterAssignments(
	ctx context.Context,
	desiredServiceClusterAssignments []corev1alpha1.ServiceClusterAssignment,
	currentServiceClusterAssignments []corev1alpha1.ServiceClusterAssignment,
) error {
	existing := map[string]struct{}{}

	// reconcile desired ServiceClusterAssignments
	for _, desiredSCA := range desiredServiceClusterAssignments {
		nn := types.NamespacedName{
			Name:      desiredSCA.Name,
			Namespace: desiredSCA.Namespace,
		}.String()
		existing[nn] = struct{}{}

		if err := r.reconcileServiceClusterAssignment(ctx, &desiredSCA); err != nil {
			return fmt.Errorf("reconciling ServiceClusterAssignment %s: %w", nn, err)
		}
	}

	// cleanup old ServiceClusterAssignments
	for _, sca := range currentServiceClusterAssignments {
		nn := types.NamespacedName{
			Name:      sca.Name,
			Namespace: sca.Namespace,
		}.String()
		if _, ok := existing[nn]; ok {
			// skip items that should exist
			continue
		}

		if err := r.Delete(ctx, &sca); err != nil {
			return fmt.Errorf("deleting ServiceClusterAssignment %s: %w", nn, err)
		}
	}

	return nil
}

func (r *NamespaceReconciler) reconcileServiceClusterAssignment(
	ctx context.Context, desiredSCA *corev1alpha1.ServiceClusterAssignment,
) error {
	if err := r.Create(ctx, desiredSCA); err != nil && errors.IsAlreadyExists(err) {
		return fmt.Errorf("creating ServiceClusterAssignment: %w", err)
	}
	return nil
}

const assignmentLabelPrefix = "assignment.kubecarrier.io/"

func (r *NamespaceReconciler) scasFromNamespaceLabels(ns *corev1.Namespace) ([]corev1alpha1.ServiceClusterAssignment, error) {
	var scas []corev1alpha1.ServiceClusterAssignment

	for label := range ns.Labels {
		if !strings.HasPrefix(label, assignmentLabelPrefix) {
			continue
		}

		// remove prefix
		label = label[len(assignmentLabelPrefix):]
		parts := strings.SplitN(label, ".", 2)
		if len(parts) != 2 {
			continue
		}
		serviceCluster := parts[0]
		managementNamespace := parts[1]

		sca := corev1alpha1.ServiceClusterAssignment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ns.Name + "." + serviceCluster,
				Namespace: managementNamespace,
			},
			Spec: corev1alpha1.ServiceClusterAssignmentSpec{
				ServiceCluster: corev1alpha1.ObjectReference{
					Name: serviceCluster,
				},
				MasterClusterNamespace: corev1alpha1.ObjectReference{
					Name: ns.Name,
				},
			},
		}
		if _, err := util.InsertOwnerReference(ns, &sca, r.Scheme); err != nil {
			return nil, fmt.Errorf("setting cross-namespaced owner reference: %w", err)
		}
		scas = append(scas, sca)
	}

	return scas, nil
}

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
	"github.com/gobuffalo/flect"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

const crdDiscoveryControllerFinalizer string = "custormresourcedefinitiondiscovery.kubecarrier.io/manager-controller"

// CustomResourceDefinitionDiscoveryReconciler reconciles a CustomResourceDefinitionDiscovery object
type CustomResourceDefinitionDiscoveryReconciler struct {
	Log            logr.Logger
	Client         client.Client
	Scheme         *runtime.Scheme
	ProviderGetter ProviderGetterByProviderNamespace
}

// +kubebuilder:rbac:groups=kubecarrier.io,resources=customresourcedefinitiondiscoveries,verbs=get;list;watch
// +kubebuilder:rbac:groups=kubecarrier.io,resources=customresourcedefinitiondiscoveries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch;update;create;patch;delete

func (r *CustomResourceDefinitionDiscoveryReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("crddiscovery", req.NamespacedName)

	crdDiscovery := &corev1alpha1.CustomResourceDefinitionDiscovery{}
	if err := r.Client.Get(ctx, req.NamespacedName, crdDiscovery); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	crdDiscovery.Status.ObservedGeneration = crdDiscovery.Generation

	if !crdDiscovery.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, log, crdDiscovery); err != nil {
			return ctrl.Result{}, fmt.Errorf("handling deletion: %w", err)
		}
		return ctrl.Result{}, nil
	}

	if util.AddFinalizer(crdDiscovery, crdDiscoveryControllerFinalizer) {
		if err := r.Client.Status().Update(ctx, crdDiscovery); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating CustomResourceDefinitionDiscovery finalizers: %w", err)
		}
	}

	cond, ok := crdDiscovery.Status.GetCondition(corev1alpha1.CustomResourceDefinitionDiscoveryReady)
	if !ok || cond.Status != corev1alpha1.ConditionTrue {
		crdDiscovery.Status.SetCondition(corev1alpha1.CustomResourceDefinitionDiscoveryCondition{
			Message: "CustomResourceDefinitionDiscovery isn't ready",
			Reason:  "CustomResourceDefinitionDiscoveryUnready",
			Status:  corev1alpha1.ConditionFalse,
			Type:    corev1alpha1.CustomResourceDefinitionDiscoveryDiscovered,
		})
		if err := r.Client.Status().Update(ctx, crdDiscovery); err != nil {
			return ctrl.Result{}, fmt.Errorf("update discovered state: %w", err)
		}
		return ctrl.Result{}, nil
	}

	provider, err := r.ProviderGetter.GetProviderByProviderNamespace(ctx, r.Client, req.Namespace)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("getting Provider: %w", err)
	}

	kind := crdDiscovery.Spec.KindOverride
	if kind == "" {
		kind = crdDiscovery.Status.CRD.Spec.Names.Kind
	}

	crd := &apiextensionsv1.CustomResourceDefinition{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: crdDiscovery.Spec.ServiceCluster.Name + "." + provider.Name,
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Plural:   flect.Pluralize(strings.ToLower(kind)),
				Singular: strings.ToLower(kind),
				Kind:     kind,
				ListKind: kind + "List",
			},
			Scope:                 apiextensionsv1.NamespaceScoped,
			Versions:              crdDiscovery.Status.CRD.Spec.Versions,
			Conversion:            nil, // TODO: implement via webhooks
			PreserveUnknownFields: crdDiscovery.Status.CRD.Spec.PreserveUnknownFields,
		},
		Status: apiextensionsv1.CustomResourceDefinitionStatus{},
	}
	crd.Name = crd.Spec.Names.Plural + "." + crd.Spec.Group
	if _, err := util.InsertOwnerReference(crdDiscovery, crd, r.Scheme); err != nil {
		return ctrl.Result{}, fmt.Errorf("insert object reference: %w", err)
	}

	op, err := ctrl.CreateOrUpdate(ctx, r.Client, crd, func() error {
		crd.Spec.Versions = crdDiscovery.Status.CRD.Spec.Versions
		_, err := util.InsertOwnerReference(crdDiscovery, crd, r.Scheme)
		return err
	})

	if err != nil {
		return ctrl.Result{}, fmt.Errorf("cannot create or update CRD: %w", err)
	}
	log.Info(fmt.Sprintf("CRD %s: %s", crd.Name, op))

	crdDiscovery.Status.SetCondition(corev1alpha1.CustomResourceDefinitionDiscoveryCondition{
		Message: "CustomResourceDefinitionDiscovery ready",
		Reason:  "CustomResourceDefinitionDiscoveryReady",
		Status:  corev1alpha1.ConditionTrue,
		Type:    corev1alpha1.CustomResourceDefinitionDiscoveryDiscovered,
	})
	if err := r.Client.Status().Update(ctx, crdDiscovery); err != nil {
		return ctrl.Result{}, fmt.Errorf("update discovered state: %w", err)
	}
	return ctrl.Result{}, nil
}

func (r *CustomResourceDefinitionDiscoveryReconciler) handleDeletion(ctx context.Context, log logr.Logger, crdDiscovery *corev1alpha1.CustomResourceDefinitionDiscovery) error {
	cond, ok := crdDiscovery.Status.GetCondition(corev1alpha1.CustomResourceDefinitionDiscoveryDiscovered)
	if !ok || cond.Status != corev1alpha1.ConditionFalse || cond.Reason != "Deleting" {
		crdDiscovery.Status.SetCondition(corev1alpha1.CustomResourceDefinitionDiscoveryCondition{
			Message: "custom resource definition discovery is being teminated",
			Reason:  "Deleting", // TODO replace with constant from PR https://github.com/kubermatic/kubecarrier/pull/136
			Status:  corev1alpha1.ConditionFalse,
			Type:    corev1alpha1.CustomResourceDefinitionDiscoveryDiscovered,
		})
		if err := r.Client.Status().Update(ctx, crdDiscovery); err != nil {
			return fmt.Errorf("update discovered state: %w", err)
		}
	}

	crds := &apiextensionsv1.CustomResourceDefinitionList{}
	ownedBy, err := util.OwnedBy(crdDiscovery, r.Scheme)
	if err != nil {
		return fmt.Errorf("cannot created owned by: %w", err)
	}
	if err := r.Client.List(ctx, crds, ownedBy); err != nil {
		return fmt.Errorf("cannot list crds: %w", err)
	}
	if len(crds.Items) > 0 {
		for _, crd := range crds.Items {
			if err := r.Client.Delete(ctx, &crd); err != nil {
				return fmt.Errorf("cannot delete: %w", err)
			}
		}
		return nil
	}

	if util.RemoveFinalizer(crdDiscovery, crdDiscoveryControllerFinalizer) {
		if err := r.Client.Update(ctx, crdDiscovery); err != nil {
			return fmt.Errorf("updating CustomResourceDefinitionDiscovery finalizers: %w", err)
		}
	}
	return nil
}

func (r *CustomResourceDefinitionDiscoveryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	enqueuer, err := util.EnqueueRequestForOwner(&corev1alpha1.CustomResourceDefinitionDiscovery{}, r.Scheme)
	if err != nil {
		return fmt.Errorf("enqueuer: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.CustomResourceDefinitionDiscovery{}).
		Watches(&source.Kind{Type: &apiextensionsv1.CustomResourceDefinition{}}, enqueuer).
		Complete(r)
}

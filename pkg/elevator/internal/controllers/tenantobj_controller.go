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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

// TenantObjReconciler reconciles a tenant-side CRD by converting it into a provider-side object and syncing the status back:
// Tenant obj (spec)   -> Provider obj (spec)
// Tenant obj (status) <- Provider obj (status)
type TenantObjReconciler struct {
	client.Client
	Log              logr.Logger
	Scheme           *runtime.Scheme
	NamespacedClient client.Client

	// Dynamic types we work with
	ProviderGVK, TenantGVK   schema.GroupVersionKind
	ProviderType, TenantType *unstructured.Unstructured

	DerivedCRDName, ProviderNamespace string
}

// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=derivedcustomresourcedefinitions,verbs=get;list;watch
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=derivedcustomresourcedefinitions/status,verbs=get

func (r *TenantObjReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var (
		ctx    = context.Background()
		result ctrl.Result
	)

	tenantObj := r.TenantType.DeepCopy()
	if err := r.Get(ctx, req.NamespacedName, tenantObj); err != nil {
		return result, client.IgnoreNotFound(err)
	}

	// Get DerivedCustomResourceDefinition field configs
	derivedCRD := &catalogv1alpha1.DerivedCustomResourceDefinition{}
	if err := r.NamespacedClient.Get(ctx, types.NamespacedName{
		Name:      r.DerivedCRDName,
		Namespace: r.ProviderNamespace,
	}, derivedCRD); err != nil {
		return result, fmt.Errorf("getting DerivedCustomResourceDefinition: %w", err)
	}
	version := r.ProviderGVK.Version
	exposeConfig, ok := versionExposeConfigForVersion(derivedCRD.Spec.Expose, version)
	if !ok {
		return result, fmt.Errorf("missing version expose config for version %q", version)
	}

	// Reconcile TenantCRD
	err := r.reconcileTenantObj(
		ctx, tenantObj, exposeConfig)
	if err != nil {
		return result, fmt.Errorf("reconciling %s: %w", r.ProviderGVK.Kind, err)
	}

	return result, nil
}

func (r *TenantObjReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(r.TenantType).
		Owns(r.ProviderType).
		Complete(r)
}

func (r *TenantObjReconciler) reconcileTenantObj(
	ctx context.Context, tenantObj *unstructured.Unstructured,
	config catalogv1alpha1.VersionExposeConfig,
) error {
	desiredProviderObj := r.ProviderType.DeepCopy()
	desiredProviderObj.SetName(tenantObj.GetName())
	desiredProviderObj.SetNamespace(tenantObj.GetNamespace())

	err := controllerutil.SetControllerReference(
		tenantObj, desiredProviderObj, r.Scheme)
	if err != nil {
		return fmt.Errorf("set controller reference: %w", err)
	}

	// prepare config
	statusFields, otherFields := splitStatusFields(config.Fields)

	// Lookup current instance
	currentProviderObj := r.ProviderType.DeepCopy()
	err = r.Get(ctx, types.NamespacedName{
		Name:      desiredProviderObj.GetName(),
		Namespace: desiredProviderObj.GetNamespace(),
	}, currentProviderObj)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("getting %s: %w", r.ProviderGVK.Kind, err)
	}

	if errors.IsNotFound(err) {
		// Create the Provider Obj
		if err = copyFields(tenantObj, desiredProviderObj, otherFields); err != nil {
			return fmt.Errorf("copy fields: %w", err)
		}

		if err = r.Create(ctx, desiredProviderObj); err != nil {
			return fmt.Errorf("creating %s: %w", r.ProviderGVK.Kind, err)
		}
		return nil
	}

	// Make sure we take ownership of the provider instance,
	// if the OwnerReference is not yet set.
	// Note:
	// this will raise an error,
	// if the provider object is already owned by someone else
	err = controllerutil.SetControllerReference(
		tenantObj, currentProviderObj, r.Scheme)
	if err != nil {
		return fmt.Errorf("set controller reference: %w", err)
	}

	// Update existing provider instance
	if err = copyFields(tenantObj, currentProviderObj, otherFields); err != nil {
		return fmt.Errorf(
			"copy fields from %s to %s: %w",
			r.TenantGVK.Kind, r.ProviderGVK.Kind, err)
	}
	if err = r.Update(ctx, currentProviderObj); err != nil {
		return fmt.Errorf("updating %s: %w", r.ProviderGVK.Kind, err)
	}

	// Sync status from provider to tenant instance
	if err = copyFields(currentProviderObj, tenantObj, statusFields); err != nil {
		return fmt.Errorf(
			"copy status fields from %s to %s: %w",
			r.ProviderGVK.Kind, r.TenantGVK.Kind, err)
	}
	if err = util.UpdateObservedGeneration(currentProviderObj, tenantObj); err != nil {
		return fmt.Errorf(
			"update observedGeneration, by comparing %s to %s: %w",
			r.ProviderGVK.Kind, r.TenantGVK.Kind, err)
	}
	if err = r.Status().Update(ctx, tenantObj); err != nil {
		return fmt.Errorf("updating %s: %w", r.TenantGVK.Kind, err)
	}

	return nil
}

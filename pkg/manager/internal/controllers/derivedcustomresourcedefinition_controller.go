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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	unstructuredv1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	internalreconcile "github.com/kubermatic/kubecarrier/pkg/internal/reconcile"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

const (
	dcrdFinalizer  = "dcrd.kubecarrier.io/controller"
	dcrdAnnotation = "dcrd.kubecarrier.io/referenced-by"
)

type DerivedCustomResourceDefinitionReconciler struct {
	client.Client
	Log                        logr.Logger
	Scheme                     *runtime.Scheme
	KubeCarrierSystemNamespace string
}

// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=derivedcustomresourcedefinitions,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=derivedcustomresourcedefinitions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch;update;create;delete

func (r *DerivedCustomResourceDefinitionReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	result := ctrl.Result{}
	log := r.Log.WithValues("dcrd", req.NamespacedName)

	dcrd := &catalogv1alpha1.DerivedCustomResourceDefinition{}
	if err := r.Get(ctx, req.NamespacedName, dcrd); err != nil {
		// If the Provider object is already gone, we just ignore the NotFound error.
		return result, client.IgnoreNotFound(err)
	}

	// deletion
	if !dcrd.DeletionTimestamp.IsZero() {
		// remove finalizer
		if err := r.handleDeletion(ctx, dcrd); err != nil {
			return result, fmt.Errorf("handling deletion: %w", err)
		}
		return result, nil
	}

	// add finalizer
	if util.AddFinalizer(dcrd, dcrdFinalizer) {
		// Update the DCRD with the finalizer
		if err := r.Update(ctx, dcrd); err != nil {
			return result, fmt.Errorf("updating finalizers: %w", err)
		}
	}

	// lookup base CRD
	baseCRD := &apiextensionsv1.CustomResourceDefinition{}
	if err := r.Get(ctx, types.NamespacedName{
		Name: dcrd.Spec.BaseCRD.Name,
	}, baseCRD); err != nil {
		if errors.IsNotFound(err) {
			return result, r.updateStatus(ctx, dcrd, catalogv1alpha1.DerivedCustomResourceDefinitionCondition{
				Type:    catalogv1alpha1.DerivedCustomResourceDefinitionReady,
				Status:  catalogv1alpha1.ConditionFalse,
				Reason:  "NotFound",
				Message: "The referenced CRD was not found.",
			})
		}

		return result, fmt.Errorf("get base CRD: %w", err)
	}
	if baseCRD.Spec.Scope != apiextensionsv1.NamespaceScoped {
		return result, r.updateStatus(ctx, dcrd, catalogv1alpha1.DerivedCustomResourceDefinitionCondition{
			Type:    catalogv1alpha1.DerivedCustomResourceDefinitionReady,
			Status:  catalogv1alpha1.ConditionFalse,
			Reason:  "NotNamespaced",
			Message: "The referenced CRD needs to Namespace scoped.",
		})
	}
	if baseCRD.Annotations == nil {
		baseCRD.Annotations = map[string]string{}
	}
	if ref, ok := baseCRD.Annotations[dcrdAnnotation]; ok && ref != req.NamespacedName.String() {
		// referenced by another instance
		return result, r.updateStatus(ctx, dcrd, catalogv1alpha1.DerivedCustomResourceDefinitionCondition{
			Type:    catalogv1alpha1.DerivedCustomResourceDefinitionReady,
			Status:  catalogv1alpha1.ConditionFalse,
			Reason:  "AlreadyInUse",
			Message: fmt.Sprintf("The referenced CRD is already referenced by %q.", ref),
		})
	} else if !ok {
		// not yet referenced
		baseCRD.Annotations[dcrdAnnotation] = req.NamespacedName.String()
		if err := r.Update(ctx, baseCRD); err != nil {
			return result, fmt.Errorf("updating base CRD: %w", err)
		}
	}

	// lookup Provider
	provider, err := GetProviderByProviderNamespace(ctx, r.Client, r.KubeCarrierSystemNamespace, dcrd.Namespace)
	if err != nil {
		return result, fmt.Errorf("getting the Provider by Provider Namespace: %w", err)
	}

	// check if Provider is allowed to use the CRD
	if baseCRD.Labels == nil ||
		baseCRD.Labels[ProviderLabel] != provider.Name {
		return result, r.updateStatus(ctx, dcrd, catalogv1alpha1.DerivedCustomResourceDefinitionCondition{
			Type:    catalogv1alpha1.DerivedCustomResourceDefinitionReady,
			Status:  catalogv1alpha1.ConditionFalse,
			Reason:  "NotAssignedToProvider",
			Message: fmt.Sprintf("The referenced CRD not assigned to this Provider or is missing a %s label.", ProviderLabel),
		})
	}

	// lookup ServiceCluster
	if baseCRD.Labels == nil ||
		baseCRD.Labels[serviceClusterLabel] == "" {
		return result, r.updateStatus(ctx, dcrd, catalogv1alpha1.DerivedCustomResourceDefinitionCondition{
			Type:    catalogv1alpha1.DerivedCustomResourceDefinitionReady,
			Status:  catalogv1alpha1.ConditionFalse,
			Reason:  "MissingServiceClusterLabel",
			Message: fmt.Sprintf("The referenced CRD is missing a %s label.", serviceClusterLabel),
		})
	}
	serviceClusterName := baseCRD.Labels[serviceClusterLabel]

	// kindOverride
	names := *baseCRD.Spec.Names.DeepCopy()
	if dcrd.Spec.KindOverride != "" {
		// Analog to controller-gen:
		// https://github.com/kubernetes-sigs/controller-tools/blob/v0.2.4/pkg/crd/spec.go#L58-L77
		names.Kind = dcrd.Spec.KindOverride
		names.ListKind = names.Kind + "List"
		names.Plural = flect.Pluralize(strings.ToLower(names.Kind))
		names.Singular = strings.ToLower(names.Kind)
	}
	group := serviceClusterName + "." + provider.Name

	derivedCRD := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: names.Plural + "." + group,
			Labels: map[string]string{
				serviceClusterLabel: serviceClusterName,
				ProviderLabel:       provider.Name,
			},
		},
		Spec: *baseCRD.Spec.DeepCopy(),
	}
	derivedCRD.Spec.Group = group
	derivedCRD.Spec.Names = names

	if _, err = util.InsertOwnerReference(dcrd, derivedCRD, r.Scheme); err != nil {
		return result, fmt.Errorf("setting owner reference: %w", err)
	}
	if err = r.applyExposeConfig(dcrd, derivedCRD); err != nil {
		return result, fmt.Errorf("apply expose config: %w", err)
	}
	currentDerivedCRD, err := internalreconcile.CustomResourceDefinition(
		ctx, log, r.Client, derivedCRD,
	)
	if err != nil {
		return result, fmt.Errorf("reconciling CRD: %w", err)
	}

	// check derived CRD conditions
	if !isCRDReady(currentDerivedCRD) {
		// waiting for CRD to be ready
		if err = r.updateStatus(ctx, dcrd, catalogv1alpha1.DerivedCustomResourceDefinitionCondition{
			Type:    catalogv1alpha1.DerivedCustomResourceDefinitionReady,
			Status:  catalogv1alpha1.ConditionFalse,
			Reason:  "Establishing",
			Message: "The derived CRD is not yet established.",
		}); err != nil {
			return result, fmt.Errorf("updating status: %w", err)
		}
		return result, nil
	}

	dcrd.Status.DerivedCRD = &catalogv1alpha1.DerivedCustomResourceDefinitionReference{
		Name:     currentDerivedCRD.Name,
		Group:    currentDerivedCRD.Spec.Group,
		Kind:     currentDerivedCRD.Status.AcceptedNames.Kind,
		Plural:   currentDerivedCRD.Status.AcceptedNames.Plural,
		Singular: currentDerivedCRD.Status.AcceptedNames.Singular,
	}
	if err = r.updateStatus(ctx, dcrd, catalogv1alpha1.DerivedCustomResourceDefinitionCondition{
		Type:    catalogv1alpha1.DerivedCustomResourceDefinitionReady,
		Status:  catalogv1alpha1.ConditionTrue,
		Reason:  "Established",
		Message: "The derived CRD is established.",
	}); err != nil {
		return result, fmt.Errorf("updating status: %w", err)
	}

	return result, nil
}

func isCRDReady(crd *apiextensionsv1.CustomResourceDefinition) bool {
	for _, condition := range crd.Status.Conditions {
		if condition.Type == apiextensionsv1.Established &&
			condition.Status == apiextensionsv1.ConditionTrue {
			return true
		}
	}
	return false
}

func (r *DerivedCustomResourceDefinitionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	owner := &catalogv1alpha1.DerivedCustomResourceDefinition{}
	enqueuer, err := util.EnqueueRequestForOwner(owner, mgr.GetScheme())
	if err != nil {
		return fmt.Errorf("cannot create enqueuer for DerivedCustomResourceDefinition: %w", err)
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&catalogv1alpha1.DerivedCustomResourceDefinition{}).
		Watches(&source.Kind{Type: &apiextensionsv1.CustomResourceDefinition{}}, enqueuer).
		Watches(&source.Kind{Type: &apiextensionsv1.CustomResourceDefinition{}}, &handler.EnqueueRequestsFromMapFunc{
			ToRequests: handler.ToRequestsFunc(func(mapObject handler.MapObject) (out []reconcile.Request) {
				annotations := mapObject.Meta.GetAnnotations()
				if annotations == nil {
					return
				}

				ref, ok := annotations[dcrdAnnotation]
				if !ok {
					return
				}

				parts := strings.SplitN(ref, string(types.Separator), 2)
				if len(parts) != 2 {
					return
				}
				out = append(out, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      parts[1],
						Namespace: parts[0],
					},
				})
				return
			}),
		}).
		Complete(r)
}

func (r *DerivedCustomResourceDefinitionReconciler) handleDeletion(ctx context.Context, dcrd *catalogv1alpha1.DerivedCustomResourceDefinition) error {
	ownedBySelector, err := util.OwnedBy(dcrd, r.Scheme)
	if err != nil {
		return fmt.Errorf("creating owned by selector: %w", err)
	}

	crdList := &apiextensionsv1.CustomResourceDefinitionList{}
	if err = r.List(ctx, crdList, ownedBySelector); err != nil {
		return fmt.Errorf("listing owned CRDs: %w", err)
	}

	for _, crd := range crdList.Items {
		if err = r.Delete(ctx, &crd); err != nil {
			return fmt.Errorf("deleting CRD: %w", err)
		}
	}

	if len(crdList.Items) != 0 {
		// wait for requeue
		return nil
	}

	// remove referenced-by annotation
	baseCRD := &apiextensionsv1.CustomResourceDefinition{}
	err = r.Get(ctx, types.NamespacedName{
		Name: dcrd.Spec.BaseCRD.Name,
	}, baseCRD)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("getting baseCRD: %w", err)
	}
	if err == nil && baseCRD.Annotations != nil {
		delete(baseCRD.Annotations, dcrdAnnotation)
		if err := r.Update(ctx, baseCRD); err != nil {
			return fmt.Errorf("updating base CRD: %w", err)
		}
	}

	// remove finalizer
	if util.RemoveFinalizer(dcrd, dcrdFinalizer) {
		if err := r.Update(ctx, dcrd); err != nil {
			return fmt.Errorf("updating DerivedCRD finalizer: %w", err)
		}
	}
	return nil
}

func (r *DerivedCustomResourceDefinitionReconciler) updateStatus(
	ctx context.Context, dcrd *catalogv1alpha1.DerivedCustomResourceDefinition,
	condition catalogv1alpha1.DerivedCustomResourceDefinitionCondition,
) error {
	dcrd.Status.ObservedGeneration = dcrd.Generation
	dcrd.Status.SetCondition(condition)
	if err := r.Status().Update(ctx, dcrd); err != nil {
		return fmt.Errorf("updating status: %w", err)
	}
	return nil
}

func (r *DerivedCustomResourceDefinitionReconciler) applyExposeConfig(
	dcrd *catalogv1alpha1.DerivedCustomResourceDefinition,
	crd *apiextensionsv1.CustomResourceDefinition,
) error {
	versionExposeMap := map[string]catalogv1alpha1.VersionExposeConfig{}
	for _, exposeConfig := range dcrd.Spec.Expose {
		for _, version := range exposeConfig.Versions {
			versionExposeMap[version] = exposeConfig
		}
	}

	// filter each Schema for fields that are not exposed and
	// remove versions that have no ExposeConfig assigned
	var filteredVersions []apiextensionsv1.CustomResourceDefinitionVersion
	for _, crdVersion := range crd.Spec.Versions {
		exposeConfig, exposeConfigExists := versionExposeMap[crdVersion.Name]
		if !exposeConfigExists {
			continue
		}

		schema := crdVersion.Schema.OpenAPIV3Schema
		if schema == nil {
			// this version has no schema to check
			continue
		}

		filteredSchema, err := filterSchema(*schema, exposeConfig)
		if err != nil {
			return fmt.Errorf("filtering schema: %w", err)
		}
		crdVersion.Schema.OpenAPIV3Schema = &filteredSchema
		filteredVersions = append(filteredVersions, crdVersion)
	}
	crd.Spec.Versions = filteredVersions

	return nil
}

// filterSchema removes fields from the schema that are not exposed
func filterSchema(
	jsonSchema apiextensionsv1.JSONSchemaProps,
	exposeConfig catalogv1alpha1.VersionExposeConfig,
) (apiextensionsv1.JSONSchemaProps, error) {
	obj := dummyObject{}
	walkDummyObject(jsonSchema, obj)

	obj, err := markDummyObject(exposeConfig, obj)
	if err != nil {
		return apiextensionsv1.JSONSchemaProps{}, fmt.Errorf("filter: %w", err)
	}

	return walkFilterSchema(jsonSchema, obj)
}

type dummyObject map[string]dummyObject

const arrayKey string = "__ARRAY__"

func (d dummyObject) IsArray() bool {
	_, ok := d[arrayKey]
	return ok
}

func dummyObjectToInterface(d dummyObject) map[string]interface{} {
	m := map[string]interface{}{}
	for key, subd := range d {
		m[key] = dummyObjectToInterface(subd)
	}
	return m
}

func interfaceToDummyObject(m map[string]interface{}) dummyObject {
	d := dummyObject{}
	for key, subd := range m {
		if subd == nil {
			d[key] = nil
			continue
		}
		d[key] = interfaceToDummyObject(subd.(map[string]interface{}))
	}
	return d
}

// walkDummyObject builds an example object from a JSONSchema.
func walkDummyObject(in apiextensionsv1.JSONSchemaProps, obj dummyObject) {
	if in.Properties == nil && in.Items == nil {
		// no sub-fields
		return
	}

	// key value handling:
	if in.Properties != nil {
		for field, props := range in.Properties {
			obj[field] = dummyObject{}
			walkDummyObject(props, obj[field])
		}
		return
	}

	// array handling
	obj[arrayKey] = dummyObject{}
	walkDummyObject(*in.Items.Schema, obj[arrayKey])
}

// markDummyObject sets all fields to <nil> that are targeted by a field selector,
// so walkFilterSchema can remove unselected fields from the schema.
func markDummyObject(
	exposeConfig catalogv1alpha1.VersionExposeConfig,
	d dummyObject,
) (dummyObject, error) {
	obj := dummyObjectToInterface(d)
	for _, field := range exposeConfig.Fields {
		path := strings.Replace(field.JSONPath, "[]", "."+arrayKey, -1)
		path = strings.Trim(path, ".") // trim trailing and leading dots

		err := unstructuredv1.SetNestedField(obj, nil, strings.Split(path, ".")...)
		if err != nil {
			return dummyObject{}, fmt.Errorf("filtering object fields by json path %s: %w", field.JSONPath, err)
		}
	}
	// always allow metadata
	err := unstructuredv1.SetNestedField(obj, nil, "metadata")
	if err != nil {
		return dummyObject{}, fmt.Errorf("adding default fields: %w", err)
	}
	err = unstructuredv1.SetNestedField(obj, nil, "apiVersion")
	if err != nil {
		return dummyObject{}, fmt.Errorf("adding default fields: %w", err)
	}
	err = unstructuredv1.SetNestedField(obj, nil, "kind")
	if err != nil {
		return dummyObject{}, fmt.Errorf("adding default fields: %w", err)
	}

	return interfaceToDummyObject(obj), nil
}

// walkFilterSchema recurses over the apiextensionsv1.JSONSchemaProps and only retains properties that are marked with <nil> in the filterObj.
func walkFilterSchema(in apiextensionsv1.JSONSchemaProps, filterObj dummyObject) (apiextensionsv1.JSONSchemaProps, error) {
	out := in.DeepCopy()

	var (
		inProperties = in.Properties
		obj          = filterObj
	)

	// array
	if filterObj.IsArray() {
		if in.Items == nil || in.Items.Schema == nil {
			return *out, fmt.Errorf("path is for type array, but schema is not")
		}
		inProperties = in.Items.Schema.Properties
		obj = filterObj[arrayKey]
	}

	// properties
	updatedProperties := map[string]apiextensionsv1.JSONSchemaProps{}
	for field, properties := range inProperties {
		if obj[field] == nil {
			// this field is explicitly included/targeted by a selector
			updatedProperties[field] = properties
			continue
		}

		// check if sub-properties are included by a selector
		subProperties, err := walkFilterSchema(properties, obj[field])
		if err != nil {
			return *out, err
		}
		if len(subProperties.Properties) == 0 &&
			subProperties.Items == nil {
			// this no sub-field or array sub-field is targeted by a selector
			// so we can omit this whole key
			continue
		}
		updatedProperties[field] = subProperties
	}

	if filterObj.IsArray() {
		out.Items.Schema.Properties = updatedProperties
	} else {
		out.Properties = updatedProperties
	}
	return *out, nil
}

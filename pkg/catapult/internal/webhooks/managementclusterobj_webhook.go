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

package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
)

// ManagementClusterObjWebhookHandler handles validating of ManagementClusterObjs.
type ManagementClusterObjWebhookHandler struct {
	Log     logr.Logger
	Scheme  *runtime.Scheme
	decoder *admission.Decoder

	ManagementClusterClient, ServiceClusterClient client.Client

	ManagementClusterGVK, ServiceClusterGVK schema.GroupVersionKind

	ProviderNamespace, ServiceCluster string

	WebhookStrategy corev1alpha1.WebhookStrategyType
}

var _ admission.Handler = (*ManagementClusterObjWebhookHandler)(nil)

// Handle is the function to handle validating requests of ManagementClusterObjs.
func (r *ManagementClusterObjWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	if r.WebhookStrategy != corev1alpha1.WebhookStrategyTypeNone &&
		r.WebhookStrategy != corev1alpha1.WebhookStrategyTypeServiceCluster {
		return admission.Errored(http.StatusInternalServerError,
			fmt.Errorf("invalid WebhookStrategy, WebhookStrategy should be {None(by default), ServiceCluster}, found: %s", r.WebhookStrategy))
	}
	obj := &unstructured.Unstructured{}
	if err := r.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// If the obj is being deleted, just skip the webhook.
	if !obj.GetDeletionTimestamp().IsZero() {
		return admission.Allowed("Allow to delete object")
	}

	// Check if the GVK from request is as same as the GVK from configuration
	objGVK := obj.GroupVersionKind()
	if !reflect.DeepEqual(objGVK, r.ManagementClusterGVK) {
		return admission.Errored(http.StatusBadRequest,
			fmt.Errorf("the GVK (group, version and kind) is wrong with the requestd object, expected: %s, got: %s", r.ManagementClusterGVK, objGVK))
	}

	// Fetch the ServiceClusterAssignment obj to get the namespace to perform dry-run in the service cluster.
	serviceClusterAssignment := &corev1alpha1.ServiceClusterAssignment{}
	if err := r.ManagementClusterClient.Get(ctx, types.NamespacedName{
		Name:      obj.GetNamespace() + "." + r.ServiceCluster,
		Namespace: r.ProviderNamespace,
	}, serviceClusterAssignment); err != nil {
		return admission.Errored(http.StatusInternalServerError,
			fmt.Errorf("getting the ServiceClusterAssignment: %w", err))
	}

	// Check if the ServiceClusterAssignment is ready.
	if !serviceClusterAssignment.IsReady() {
		return admission.Errored(http.StatusInternalServerError,
			fmt.Errorf("ServiceClusterAssignment is not ready"))
	}

	if r.WebhookStrategy == corev1alpha1.WebhookStrategyTypeNone {
		return admission.Allowed("Allow to commit this admission request.")
	}

	serviceClusterObj := obj.DeepCopy()
	if err := unstructured.SetNestedField(
		serviceClusterObj.Object, map[string]interface{}{}, "metadata"); err != nil {
		return admission.Errored(http.StatusInternalServerError,
			fmt.Errorf("deleting %s .metadata: %w", r.ServiceClusterGVK.Kind, err))
	}
	serviceClusterObj.SetGroupVersionKind(r.ServiceClusterGVK)
	serviceClusterObj.SetName(obj.GetName())
	serviceClusterObj.SetNamespace(serviceClusterAssignment.Status.ServiceClusterNamespace.Name)

	// Check if the ServiceClusterObj has already been created, if it is created, then regards this request
	// as a UPDATE, if it is not crated, regards this request as a CREATE.
	err := r.ServiceClusterClient.Get(ctx, types.NamespacedName{
		Name:      serviceClusterObj.GetName(),
		Namespace: serviceClusterObj.GetNamespace(),
	}, serviceClusterObj)
	if err != nil && !errors.IsNotFound(err) {
		return admission.Errored(http.StatusInternalServerError, fmt.Errorf("getting serviceClusterObj: %w", err))
	}

	if errors.IsNotFound(err) {
		r.Log.Info("validate create", "name", obj.GetName())
		if err := r.ServiceClusterClient.Create(ctx, serviceClusterObj, client.DryRunAll); err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
	} else {
		r.Log.Info("validate update", "name", obj.GetName())
		if err := unstructured.SetNestedField(
			serviceClusterObj.Object, obj.Object["spec"], "spec"); err != nil {
			return admission.Errored(http.StatusInternalServerError,
				fmt.Errorf("changing %s .spec: %w", r.ServiceClusterGVK.Kind, err))
		}

		if err := r.ServiceClusterClient.Update(ctx, serviceClusterObj, client.DryRunAll); err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
	}

	newObj := obj.DeepCopy()
	if err := unstructured.SetNestedField(
		newObj.Object, serviceClusterObj.Object["spec"], "spec"); err != nil {
		return admission.Errored(http.StatusInternalServerError,
			fmt.Errorf("changing %s .spec: %w", r.ManagementClusterGVK.Kind, err))
	}

	marshalledObj, err := json.Marshal(newObj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	// Create the patch
	return admission.PatchResponseFromRaw(req.Object.Raw, marshalledObj)
}

// ManagementClusterObjWebhookHandler implements admission.DecoderInjector.
// A decoder will be automatically injected.
// InjectDecoder injects the decoder.
func (r *ManagementClusterObjWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	r.decoder = d
	return nil
}

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

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	elevatorutil "github.com/kubermatic/kubecarrier/pkg/elevator/internal/util"
)

// TenantObjWebhookHandler handles TenantObjs validation.
type TenantObjWebhookHandler struct {
	Log     logr.Logger
	Scheme  *runtime.Scheme
	decoder *admission.Decoder

	// Client has a global cache, and is used to perform Create/Update request with dry-run flag to against the Catapult webhook.
	client.Client
	// NamespacedClient has a namespace-only cache, and is only allowed to access the provider namespace,
	// this is used to fetch the DerivedCustomResource object.
	NamespacedClient client.Client

	TenantGVK, ProviderGVK schema.GroupVersionKind

	DerivedCRName, ProviderNamespace string
}

var _ admission.Handler = (*TenantObjWebhookHandler)(nil)

// Handle is the function to handle TenantObjs validating requests.
func (r *TenantObjWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
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
	if !reflect.DeepEqual(objGVK, r.TenantGVK) {
		return admission.Errored(http.StatusBadRequest,
			fmt.Errorf("the GVK (group, version and kind) is wrong with the requestd object, expected: %s, got: %s", r.TenantGVK, objGVK))
	}

	// Get DerivedCustomResource field configs
	derivedCustomResource := &catalogv1alpha1.DerivedCustomResource{}
	if err := r.NamespacedClient.Get(ctx, types.NamespacedName{
		Name:      r.DerivedCRName,
		Namespace: r.ProviderNamespace,
	}, derivedCustomResource); err != nil {
		return admission.Errored(http.StatusInternalServerError,
			fmt.Errorf("getting the DerivedCustomResource: %w", err))
	}

	// Check if the DerivedCustomResource object is ready
	if !derivedCustomResource.IsReady() {
		return admission.Errored(http.StatusInternalServerError,
			fmt.Errorf("DerivedCustomResource object is not ready"))
	}

	// Get the exposed config and version
	version := r.ProviderGVK.Version
	exposeConfig, ok := elevatorutil.VersionExposeConfigForVersion(derivedCustomResource.Spec.Expose, version)
	if !ok {
		return admission.Errored(http.StatusInternalServerError,
			fmt.Errorf("DerivedCustomResource object is missing version expose config for version %q", version))
	}
	// prepare config
	_, nonStatusExposedFields := elevatorutil.SplitStatusFields(exposeConfig.Fields)

	// Check if the ProviderObj has already been created, if it is created, then regards this request
	// as a UPDATE, if it is not crated, regards this request as a CREATE.
	// Using `req.admission.Request` to determine if the request is CREATE or UPDATE was the first attempt and finally,
	// we decided not to use that because it just doesn't work with the `dry-run` requests, here is an example in the
	// following and you will see the problem:
	//
	// If we use the following for our elevator webhook:
	// ```
	// switch req.Operation {
	// case adminv1beta1.Create:
	// 	if err := r.Create(ctx, providerObj, client.DryRunAll); err != nil {
	// 		return admission.Errored(http.StatusInternalServerError, err)
	// 	}
	// case adminv1beta1.Update:
	// 	if err := r.Update(ctx, providerObj, client.DryRunAll); err != nil {
	// 		return admission.Errored(http.StatusInternalServerError, err)
	// 	}
	// }
	// ```
	// Then go through the request flow when the tenant tries to create the TenantCRD object:
	// 1. Tenant sends a `CREATE` request.
	// 2. The above webhook regards this as a `CREATE` request, the `DryRun` request can pass (NO problem with this step).
	// 3. TenantObj controller of `Elevator` tries to update the finalizer for the TenantCRD object, and send an `UPDATE` request.
	// Then the problem happens: the above webhook regards this as an `UPDATE` request, and the `DryRun` request will fail
	// with `IsNotFound` error, since the ProviderObj has not been created.
	// The similar problem also happens when the TenantObj is removed, and update the finalizer later. Also, there are
	// also some other corner cases that can not be handled by this above approach.
	// There is a workaround that we can remove the `DryRun` flag and do an actual `Create`/`Update` call, in the webhook,
	// but we feels like creating objects and introducing some side effects is not the right way to go.
	// That's why we decided to not use the `req.Operation` but to check if the `ProviderObj` is created or not.
	// Also, if you think about our approach, it also makes sense, i.e., if the `ProviderObj` is not there,
	// of course it is a `CREATE` request, and it also works fine.

	tenantObj := obj.DeepCopy()
	providerObj := &unstructured.Unstructured{}
	providerObj.SetGroupVersionKind(r.ProviderGVK)
	defaults, err := elevatorutil.FormDefaults(exposeConfig.Default)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, fmt.Errorf("forming defaults: %w", err))
	}
	err = r.Get(ctx, types.NamespacedName{
		Name:      tenantObj.GetName(),
		Namespace: tenantObj.GetNamespace(),
	}, providerObj)
	if err != nil && !errors.IsNotFound(err) {
		return admission.Errored(http.StatusInternalServerError, fmt.Errorf("getting providerObj: %w", err))
	}
	if err := elevatorutil.BuildProviderObj(tenantObj, providerObj, r.Scheme, nonStatusExposedFields, defaults); err != nil {
		return admission.Errored(http.StatusInternalServerError, fmt.Errorf("build and elevate: %w", err))
	}
	if errors.IsNotFound(err) {
		r.Log.Info("validate create", "name", obj.GetName())
		if err := r.Create(ctx, providerObj, client.DryRunAll); err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
	} else {
		r.Log.Info("validate update", "name", obj.GetName())
		if err := r.Update(ctx, providerObj, client.DryRunAll); err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
	}

	newObj := obj.DeepCopy()
	if err := elevatorutil.CopyFields(providerObj, newObj, nonStatusExposedFields); err != nil {
		return admission.Errored(http.StatusInternalServerError,
			fmt.Errorf("changing %s .fields back: %w", r.ProviderGVK.Kind, err))
	}

	marshalledObj, err := json.Marshal(newObj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	// Create the defaults
	return admission.PatchResponseFromRaw(req.Object.Raw, marshalledObj)
}

// TenantObjWebhookHandler implements admission.DecoderInjector.
// A decoder will be automatically injected.
// InjectDecoder injects the decoder.
func (r *TenantObjWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	r.decoder = d
	return nil
}

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
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	adminv1beta1 "k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util/webhook"
)

// ProviderReferenceWebhookHandler handles mutating/validating of ProviderReferences.
type ProviderReferenceWebhookHandler struct {
	decoder *admission.Decoder
	Log     logr.Logger
}

var _ admission.Handler = (*ProviderReferenceWebhookHandler)(nil)

// +kubebuilder:webhook:path=/validate-catalog-kubecarrier-io-v1alpha1-providerreference,mutating=false,failurePolicy=fail,groups=catalog.kubecarrier.io,resources=providerreferences,verbs=create;update,versions=v1alpha1,name=vproviderreference.kubecarrier.io

// Handle is the function to handle create/update requests of ProviderReferences.
func (r *ProviderReferenceWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &catalogv1alpha1.ProviderReference{}
	if err := r.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	switch req.Operation {
	case adminv1beta1.Create:
		if err := r.validateCreate(obj); err != nil {
			return admission.Denied(err.Error())
		}
	case adminv1beta1.Update:
		oldObj := obj.DeepCopyObject()
		if err := r.decoder.DecodeRaw(req.OldObject, oldObj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if err := r.validateUpdate(obj, oldObj); err != nil {
			return admission.Denied(err.Error())
		}
	}
	return admission.Allowed("allowed to commit the request")

}

// ProviderReferenceWebhookHandler implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (r *ProviderReferenceWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	r.decoder = d
	return nil
}

func (r *ProviderReferenceWebhookHandler) validateCreate(providerReference *catalogv1alpha1.ProviderReference) error {
	r.Log.Info("validate create", "name", providerReference.Name)
	if !webhook.IsDNS1123Label(providerReference.Name) {
		return fmt.Errorf("providerReference name: %s is not a valid DNS 1123 Label, %s", providerReference.Name, webhook.DNS1123LabelDescription)
	}
	return r.validateMetadata(providerReference)
}

func (r *ProviderReferenceWebhookHandler) validateUpdate(obj *catalogv1alpha1.ProviderReference, oldObj runtime.Object) error {
	r.Log.Info("validate create", "name", obj.Name)
	oldProviderReference, ok := oldObj.(*catalogv1alpha1.ProviderReference)
	if !ok {
		return fmt.Errorf("expect old object to be a %T instead of %T\n", oldProviderReference, oldObj)
	}
	return r.validateMetadata(obj)
}

func (r *ProviderReferenceWebhookHandler) validateMetadata(providerReference *catalogv1alpha1.ProviderReference) error {
	if providerReference.Spec.Metadata.Description == "" || providerReference.Spec.Metadata.DisplayName == "" {
		return fmt.Errorf("the description or the display name of the ProviderReference: %s cannot be empty", providerReference.Name)
	}
	return nil
}

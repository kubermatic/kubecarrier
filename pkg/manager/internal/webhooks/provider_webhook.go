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
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
)

// ProviderWebhookHandler handles mutating/validating of Providers.
type ProviderWebhookHandler struct {
	decoder *admission.Decoder
	Log     logr.Logger
}

var _ admission.Handler = (*ProviderWebhookHandler)(nil)

// +kubebuilder:webhook:path=/validate-catalog-kubecarrier-io-v1alpha1-provider,mutating=false,failurePolicy=fail,groups=catalog.kubecarrier.io,resources=providers,verbs=create;update,versions=v1alpha1,name=vprovider.kb.io

// Handle is the function to handle create/update requests of Providers.
func (r *ProviderWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &catalogv1alpha1.Provider{}
	if err := r.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if req.Operation == adminv1beta1.Create || req.Operation == adminv1beta1.Update {
		r.Log.Info("validate create/update", "name", obj.Name)
		if err := r.validateMetadata(obj); err != nil {
			return admission.Denied(err.Error())
		}
	}
	return admission.Allowed("allowed to commit the request")

}

// ProviderWebhookHandler implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (r *ProviderWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	r.decoder = d
	return nil
}

func (r *ProviderWebhookHandler) validateMetadata(provider *catalogv1alpha1.Provider) error {
	if provider.Spec.Metadata.Description == "" || provider.Spec.Metadata.DisplayName == "" {
		return fmt.Errorf("the description or the display name of the Provider: %s cannot be empty", provider.Name)
	}
	return nil
}

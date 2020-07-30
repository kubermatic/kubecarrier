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

	catalogv1alpha1 "k8c.io/kubecarrier/pkg/apis/catalog/v1alpha1"
)

// RegionWebhookHandler handles mutating/validating of Regions.
type RegionWebhookHandler struct {
	decoder *admission.Decoder
	Log     logr.Logger
}

var _ admission.Handler = (*RegionWebhookHandler)(nil)

// +kubebuilder:webhook:path=/validate-catalog-kubecarrier-io-v1alpha1-region,mutating=false,failurePolicy=fail,groups=catalog.kubecarrier.io,resources=regions,verbs=update,versions=v1alpha1,name=vregion.kubecarrier.io

// Handle is the function to handle create/update requests of Regions.
func (r *RegionWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &catalogv1alpha1.Region{}
	if err := r.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	switch req.Operation {
	case adminv1beta1.Update:
		oldObj := &catalogv1alpha1.Region{}
		if err := r.decoder.DecodeRaw(req.OldObject, oldObj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if err := r.validateUpdate(oldObj, obj); err != nil {
			return admission.Denied(err.Error())
		}
	}
	return admission.Allowed("allowed to commit the request")

}

// RegionWebhookHandler implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (r *RegionWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	r.decoder = d
	return nil
}

func (r *RegionWebhookHandler) validateUpdate(oldObj, newObj *catalogv1alpha1.Region) error {
	r.Log.Info("validate update", "name", newObj.Name)
	if newObj.Spec.Provider.Name != oldObj.Spec.Provider.Name {
		return fmt.Errorf("the Provider of Region is immutable")
	}
	return nil
}

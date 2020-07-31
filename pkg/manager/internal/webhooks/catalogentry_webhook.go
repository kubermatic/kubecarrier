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

// CatalogEntryWebhookHandler handles validating of CatalogEntries.
type CatalogEntryWebhookHandler struct {
	decoder *admission.Decoder
	Log     logr.Logger
}

var _ admission.Handler = (*CatalogEntryWebhookHandler)(nil)

// +kubebuilder:webhook:path=/validate-catalog-kubecarrier-io-v1alpha1-catalogentry,mutating=false,failurePolicy=fail,groups=catalog.kubecarrier.io,resources=catalogentries,verbs=update,versions=v1alpha1,name=vcatalogentry.kubecarrier.io

// Handle is the function to handle validating requests of CatalogEntries.
func (r *CatalogEntryWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &catalogv1alpha1.CatalogEntry{}
	if err := r.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	switch req.Operation {
	case adminv1beta1.Update:
		oldObj := &catalogv1alpha1.CatalogEntry{}
		if err := r.decoder.DecodeRaw(req.OldObject, oldObj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if err := r.validateUpdate(oldObj, obj); err != nil {
			return admission.Denied(err.Error())
		}
	}
	return admission.Allowed("allowed to commit the request")
}

// CatalogEntryWebhookHandler implements admission.DecoderInjector.
// A decoder will be automatically injected.
// InjectDecoder injects the decoder.
func (r *CatalogEntryWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	r.decoder = d
	return nil
}

func (r *CatalogEntryWebhookHandler) validateUpdate(oldObj, newObj *catalogv1alpha1.CatalogEntry) error {
	r.Log.Info("validate update", "name", newObj.Name)
	if newObj.Spec.BaseCRD.Name != oldObj.Spec.BaseCRD.Name {
		return fmt.Errorf("the Referenced CRD of CatalogEntry is immutable")
	}
	return nil
}

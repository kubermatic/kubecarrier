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

	"github.com/go-logr/logr"
	adminv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/manager/internal/controllers"
)

// CatalogEntryWebhookHandler handles mutating/validating of CatalogEntries.
type CatalogEntryWebhookHandler struct {
	client  client.Client
	decoder *admission.Decoder
	Log     logr.Logger

	KubeCarrierNamespace string
	ProviderLabel        string
}

var _ admission.Handler = (*CatalogEntryWebhookHandler)(nil)

// +kubebuilder:webhook:path=/mutate-catalog-kubecarrier-io-v1alpha1-catalogentry,mutating=true,failurePolicy=fail,groups=catalog.kubecarrier.io,resources=catalogentries,verbs=create;update,versions=v1alpha1,name=mcatalogentry.kb.io

// Handle is the function to handle defaulting requests of CatalogEntries.
func (r *CatalogEntryWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &catalogv1alpha1.CatalogEntry{}
	if err := r.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	// Default the object
	r.Log.Info("default", "name", obj.Name)
	provider, err := controllers.GetProviderByProviderNamespace(context.Background(), r.client, r.KubeCarrierNamespace, obj.Namespace)
	if err != nil {
		return admission.Denied(err.Error())
	}
	if err := r.defaultMatchLabels(obj, provider); err != nil {
		return admission.Denied(err.Error())
	}

	// Validate the object
	if req.Operation == adminv1beta1.Create || req.Operation == adminv1beta1.Update {
		r.Log.Info("validate create/update", "name", obj.Name)
		if err := r.validateMetadata(obj); err != nil {
			return admission.Denied(err.Error())
		}
	}

	marshalled, err := json.Marshal(obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	// Create the patch
	return admission.PatchResponseFromRaw(req.Object.Raw, marshalled)

}

// CatalogEntryWebhookHandler implements inject.Client.
// A client will be automatically injected.
// InjectClient injects the client.
func (r *CatalogEntryWebhookHandler) InjectClient(c client.Client) error {
	r.client = c
	return nil
}

// CatalogEntryWebhookHandler implements admission.DecoderInjector.
// A decoder will be automatically injected.
// InjectDecoder injects the decoder.
func (r *CatalogEntryWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	r.decoder = d
	return nil
}

func (r *CatalogEntryWebhookHandler) defaultMatchLabels(catalogEntry *catalogv1alpha1.CatalogEntry, provider *catalogv1alpha1.Provider) error {
	// Defaulting the `kubecarrier.io/provider` matchlabel
	if catalogEntry.Spec.CRDSelector == nil {
		catalogEntry.Spec.CRDSelector = &metav1.LabelSelector{}
	}
	if catalogEntry.Spec.CRDSelector.MatchLabels == nil {
		catalogEntry.Spec.CRDSelector.MatchLabels = map[string]string{}
	}
	if catalogEntry.Spec.CRDSelector.MatchLabels[r.ProviderLabel] != provider.Name {
		catalogEntry.Spec.CRDSelector.MatchLabels[r.ProviderLabel] = provider.Name
	}

	return nil
}

func (r *CatalogEntryWebhookHandler) validateMetadata(catalogEntry *catalogv1alpha1.CatalogEntry) error {
	if catalogEntry.Spec.Metadata.Description == "" || catalogEntry.Spec.Metadata.DisplayName == "" {
		return fmt.Errorf("the description or the display name of the CatalogEntry: %s cannot be empty", catalogEntry.Name)
	}
	return nil
}

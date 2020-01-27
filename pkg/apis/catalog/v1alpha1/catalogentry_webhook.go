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

package v1alpha1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	adminv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var catalogEntryLog = ctrl.Log.WithName("CatalogEntry")

const (
	KubeCarrierNamespace = "kubecarrier-system"
	providerLabel        = "kubecarrier.io/provider"
)

// +kubebuilder:object:generate=false
// CatalogEntryValidator validates CatalogEntries
type CatalogEntryValidator struct {
	client  client.Client
	decoder *admission.Decoder
}

// +kubebuilder:webhook:path=/validate-catalog-kubecarrier-io-v1alpha1-catalogentry,mutating=false,failurePolicy=fail,groups=catalog.kubecarrier.io,resources=catalogentries,verbs=create;update,versions=v1alpha1,name=vcatalogentry.kb.io

// Handle is the function to handle validating requests of CatalogEntries.
func (r *CatalogEntryValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &CatalogEntry{}
	if err := r.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	switch req.Operation {
	case adminv1beta1.Create:
		catalogEntryLog.Info("validate create", "name", obj.Name)
		if err := r.validateCreate(); err != nil {
			return admission.Denied(err.Error())
		}
	case adminv1beta1.Update:
		catalogEntryLog.Info("validate update", "name", obj.Name)
		oldObj := obj.DeepCopyObject()
		if err := r.decoder.DecodeRaw(req.OldObject, oldObj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if err := r.validateUpdate(); err != nil {
			return admission.Denied(err.Error())
		}

	}
	return admission.Allowed("allowed to be admitted")
}

func (r *CatalogEntryValidator) validateCreate() error {
	return nil
}

func (r *CatalogEntryValidator) validateUpdate() error {
	return nil
}

// CatalogEntryValidator implements inject.Client.
// A client will be automatically injected.

// InjectClient injects the client.
func (r *CatalogEntryValidator) InjectClient(c client.Client) error {
	r.client = c
	return nil
}

// CatalogEntryValidator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (r *CatalogEntryValidator) InjectDecoder(d *admission.Decoder) error {
	r.decoder = d
	return nil
}

// +kubebuilder:object:generate=false
// CatalogEntryDefaulter defaults CatalogEntries
type CatalogEntryDefaulter struct {
	client  client.Client
	decoder *admission.Decoder
}

// +kubebuilder:webhook:path=/mutate-catalog-kubecarrier-io-v1alpha1-catalogentry,mutating=true,failurePolicy=fail,groups=catalog.kubecarrier.io,resources=catalogentries,verbs=create;update,versions=v1alpha1,name=mcatalogentry.kb.io

// Handle is the function to handle defaulting requests of CatalogEntries.
func (r *CatalogEntryDefaulter) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &CatalogEntry{}
	if err := r.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	catalogEntryLog.Info("default", "name", obj.Name)
	// Default the object
	if err := r.defaultFn(obj); err != nil {
		return admission.Denied(err.Error())
	}
	marshalled, err := json.Marshal(obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	// Create the patch
	return admission.PatchResponseFromRaw(req.Object.Raw, marshalled)

}

func (r *CatalogEntryDefaulter) defaultFn(catalogEntry *CatalogEntry) error {
	provider, err := getProviderByProviderNamespace(context.Background(), r.client, "kubecarrier-system", catalogEntry.Namespace)
	if err != nil {
		return fmt.Errorf("getting the Provider by Provider Namespace: %w", err)
	}

	// Defaulting the `kubecarrier.io/provider` matchlabel
	if catalogEntry.Spec.CRDSelector == nil {
		catalogEntry.Spec.CRDSelector = &metav1.LabelSelector{}
	}
	if catalogEntry.Spec.CRDSelector.MatchLabels == nil {
		catalogEntry.Spec.CRDSelector.MatchLabels = map[string]string{}
	}
	if catalogEntry.Spec.CRDSelector.MatchLabels[providerLabel] != provider.Name {
		catalogEntry.Spec.CRDSelector.MatchLabels[providerLabel] = provider.Name
	}

	return nil
}

// CatalogEntryDefaulter implements inject.Client.
// A client will be automatically injected.

// InjectClient injects the client.
func (r *CatalogEntryDefaulter) InjectClient(c client.Client) error {
	r.client = c
	return nil
}

// CatalogEntryDefaulter implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (r *CatalogEntryDefaulter) InjectDecoder(d *admission.Decoder) error {
	r.decoder = d
	return nil
}

func getProviderByProviderNamespace(ctx context.Context, c client.Client, kubecarrierNamespace, providerNamespace string) (*Provider, error) {
	providerList := &ProviderList{}
	if err := c.List(ctx, providerList,
		client.InNamespace(kubecarrierNamespace),
		client.MatchingFields{
			ProviderNamespaceFieldIndex: providerNamespace,
		},
	); err != nil {
		return nil, err
	}
	switch len(providerList.Items) {
	case 0:
		// not found
		return nil, fmt.Errorf("providers.catalog.kubecarrier.io with index %q not found", ProviderNamespaceFieldIndex)
	case 1:
		// found!
		return &providerList.Items[0], nil
	default:
		// found too many
		return nil, fmt.Errorf("multiple providers.catalog.kubecarrier.io with index %q found", ProviderNamespaceFieldIndex)
	}
}

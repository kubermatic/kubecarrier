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
	"errors"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	adminv1beta1 "k8s.io/api/admission/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/constants"
)

// KubeCarrierWebhookHandler handles validating of KubeCarrier objects.
type KubeCarrierWebhookHandler struct {
	decoder *admission.Decoder
	Log     logr.Logger
}

var _ admission.Handler = (*KubeCarrierWebhookHandler)(nil)

// +kubebuilder:webhook:path=/validate-operator-kubecarrier-io-v1alpha1-kubecarrier,mutating=false,failurePolicy=fail,groups=operator.kubecarrier.io,resources=kubecarriers,verbs=create,versions=v1alpha1,name=vkubecarrier.kubecarrier.io

// Handle is the function to handle create requests of KubeCarriers.
func (r *KubeCarrierWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &operatorv1alpha1.KubeCarrier{}
	if err := r.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	switch req.Operation {
	case adminv1beta1.Create:
		if err := r.validateCreate(obj); err != nil {
			return admission.Denied(err.Error())
		}
	}
	return admission.Allowed("allowed to commit the request")

}

// KubeCarrierWebhookHandler implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (r *KubeCarrierWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	r.decoder = d
	return nil
}

func (r *KubeCarrierWebhookHandler) validateCreate(kubeCarrier *operatorv1alpha1.KubeCarrier) error {
	r.Log.Info("validate create", "name", kubeCarrier.Name)
	if kubeCarrier.Name != constants.KubeCarrierDefaultName {
		return fmt.Errorf("KubeCarrier object name should be 'kubecarrier', found: %s", kubeCarrier.Name)
	}
	auth := map[string]bool{}
	for _, a := range kubeCarrier.Spec.API.Authentication {
		var enabled int
		if a.OIDC != nil {
			enabled++
			if auth["OIDC"] {
				return errors.New("Duplicate OIDC configuration")
			}
			auth["OIDC"] = true
		}
		if a.StaticUsers != nil {
			enabled++
			if auth["Htpasswd"] {
				return errors.New("Duplicate StaticUsers configuration")
			}
			auth["Htpasswd"] = true
		}
		if a.ServiceAccount != nil {
			enabled++
			if auth["Token"] {
				return errors.New("Duplicate ServiceAccount configuration")
			}
			auth["Token"] = true
		}
		if a.Anonymous != nil {
			enabled++
			if auth["Anonymous"] {
				return errors.New("Duplicate Anonymous configuration")
			}
			auth["Anonymous"] = true
		}
		if enabled != 1 {
			return errors.New("Authentication item should have one and only one configuration")
		}
	}
	return nil
}

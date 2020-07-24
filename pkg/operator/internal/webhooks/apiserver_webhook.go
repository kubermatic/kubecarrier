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
	"net/http"

	"github.com/go-logr/logr"
	adminv1beta1 "k8s.io/api/admission/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util/webhook"
)

// APIServerWebhookHandler handles validating of APIServer objects.
type APIServerWebhookHandler struct {
	decoder *admission.Decoder
	Log     logr.Logger
	Client  client.Client
}

var _ admission.Handler = (*APIServerWebhookHandler)(nil)

// +kubebuilder:webhook:path=/mutate-operator-kubecarrier-io-v1alpha1-apiserver,mutating=true,failurePolicy=fail,groups=operator.kubecarrier.io,resources=apiservers,verbs=create,versions=v1alpha1,name=mapiserver.kubecarrier.io

// Handle is the function to handle create requests of APIServer.
func (r *APIServerWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &operatorv1alpha1.APIServer{}
	if err := r.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	switch req.Operation {
	case adminv1beta1.Create:
		var changed bool
		if obj.Spec.LogLevel == nil {
			changed = true
			if err := webhook.SetDefaultLogLevel(ctx, r.Client, &obj.Spec); err != nil {
				return admission.Errored(http.StatusInternalServerError, err)
			}
		}
		if err := r.validateCreate(obj); err != nil {
			return admission.Denied(err.Error())
		}
		if changed {
			marshalledObj, err := json.Marshal(obj)
			if err != nil {
				return admission.Errored(http.StatusInternalServerError, err)
			}
			// Create the patch
			return admission.PatchResponseFromRaw(req.Object.Raw, marshalledObj)
		}
	}
	return admission.Allowed("allowed to commit the request")

}

// APIServerWebhookHandler implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (r *APIServerWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	r.decoder = d
	return nil
}

func (r *APIServerWebhookHandler) validateCreate(apiServer *operatorv1alpha1.APIServer) error {
	r.Log.Info("validate create", "name", apiServer.Name)
	if err := apiServer.Spec.Validate(); err != nil {
		return err
	}
	return nil
}

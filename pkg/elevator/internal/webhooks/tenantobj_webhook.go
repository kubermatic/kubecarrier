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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// TenantObjWebhookHandler handles validating of TenantObjs.
type TenantObjWebhookHandler struct {
	Log     logr.Logger
	Scheme  *runtime.Scheme
	decoder *admission.Decoder

	ManagementClusterClient client.Client

	TenantGVK, ProviderGVK schema.GroupVersionKind

	ProviderNamespace, ServiceCluster string
}

var _ admission.Handler = (*TenantObjWebhookHandler)(nil)

// Handle is the function to handle validating requests of ManagementClusterObjs.
func (r *TenantObjWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	return admission.Allowed("allowed")
}

// TenantObjWebhookHandler implements admission.DecoderInjector.
// A decoder will be automatically injected.
// InjectDecoder injects the decoder.
func (r *TenantObjWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	r.decoder = d
	return nil
}

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

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util/webhook"
)

// ServiceClusterWebhookHandler handles mutating/validating of ServiceClusters.
type ServiceClusterWebhookHandler struct {
	decoder *admission.Decoder
	Log     logr.Logger
}

var _ admission.Handler = (*ServiceClusterWebhookHandler)(nil)

// +kubebuilder:webhook:path=/validate-kubecarrier-io-v1alpha1-servicecluster,mutating=false,failurePolicy=fail,groups=kubecarrier.io,resources=serviceclusters,verbs=create,versions=v1alpha1,name=vservicecluster.kubecarrier.io

// Handle is the function to handle create/update requests of ServiceClusters.
func (r *ServiceClusterWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &corev1alpha1.ServiceCluster{}
	if err := r.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if req.Operation == adminv1beta1.Create {
		if err := r.validateCreate(obj); err != nil {
			return admission.Denied(err.Error())
		}
	}
	return admission.Allowed("allowed to commit the request")
}

// ServiceClusterWebhookHandler implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (r *ServiceClusterWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	r.decoder = d
	return nil
}

func (r *ServiceClusterWebhookHandler) validateCreate(serviceCluster *corev1alpha1.ServiceCluster) error {
	r.Log.Info("validate create", "name", serviceCluster.Name)
	if !webhook.IsDNS1123Label(serviceCluster.Name) {
		return fmt.Errorf("serviceCluster name: %s is not a valid DNS 1123 Label, %s", serviceCluster.Name, webhook.DNS1123LabelDescription)
	}
	return nil
}

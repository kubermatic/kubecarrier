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
)

// CustomResourceDefinitionDiscoveryWebhookHandler handles mutating/validating of CustomResourceDefinitionDiscoveries.
type CustomResourceDefinitionDiscoveryWebhookHandler struct {
	decoder *admission.Decoder
	Log     logr.Logger
}

var _ admission.Handler = (*CustomResourceDefinitionDiscoveryWebhookHandler)(nil)

// +kubebuilder:webhook:path=/validate-kubecarrier-io-v1alpha1-customresourcedefinitiondiscoveries,mutating=false,failurePolicy=fail,groups=kubecarrier.io,resources=customresourcedefinitiondiscoveries,verbs=create;update,versions=v1alpha1,name=vcustomresourcedefinitiondiscovery.kubecarrier.io

// Handle is the function to handle create/update requests of CustomResourceDefinitionDiscoverys.
func (r *CustomResourceDefinitionDiscoveryWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &corev1alpha1.CustomResourceDefinitionDiscovery{}
	if err := r.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	switch req.Operation {
	case adminv1beta1.Create:
		if err := r.validateCreate(obj); err != nil {
			return admission.Denied(err.Error())
		}
	case adminv1beta1.Update:
		oldObj := &corev1alpha1.CustomResourceDefinitionDiscovery{}
		if err := r.decoder.DecodeRaw(req.OldObject, oldObj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if err := r.validateUpdate(oldObj, obj); err != nil {
			return admission.Denied(err.Error())
		}
	}
	return admission.Allowed("allowed to commit the request")

}

// CustomResourceDefinitionDiscoveryWebhookHandler implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (r *CustomResourceDefinitionDiscoveryWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	r.decoder = d
	return nil
}

func (r *CustomResourceDefinitionDiscoveryWebhookHandler) validateCreate(crdDiscovery *corev1alpha1.CustomResourceDefinitionDiscovery) error {
	r.Log.Info("validate create", "name", crdDiscovery.Name)
	if crdDiscovery.Spec.ServiceCluster.Name == "" ||
		crdDiscovery.Spec.CRD.Name == "" ||
		crdDiscovery.Spec.KindOverride == "" {
		return fmt.Errorf("the ServiceCluster, CRD, or KindOverride of CustomResourceDefinitionDiscovery is not specifed")
	}
	return nil
}

func (r *CustomResourceDefinitionDiscoveryWebhookHandler) validateUpdate(oldObj, newObj *corev1alpha1.CustomResourceDefinitionDiscovery) error {
	r.Log.Info("validate update", "name", newObj.Name)
	if newObj.Spec.ServiceCluster.Name != oldObj.Spec.ServiceCluster.Name ||
		newObj.Spec.CRD.Name != oldObj.Spec.CRD.Name ||
		newObj.Spec.KindOverride != oldObj.Spec.KindOverride {
		return fmt.Errorf("the Spec (ServiceCluster, CRD, and KindOverride) of CustomResourceDefinitionDiscovery is immutable")
	}
	return nil
}

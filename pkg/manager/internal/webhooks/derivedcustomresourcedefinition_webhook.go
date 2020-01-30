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
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
)

// DerivedCustomResourceDefinitionWebhookHandler handles mutating/validating of DerivedCustomResourceDefinitions.
type DerivedCustomResourceDefinitionWebhookHandler struct {
	decoder *admission.Decoder
	Log     logr.Logger
}

var _ admission.Handler = (*DerivedCustomResourceDefinitionWebhookHandler)(nil)

// +kubebuilder:webhook:path=/validate-catalog-kubecarrier-io-v1alpha1-derivedcustomresourcedefinition,mutating=false,failurePolicy=fail,groups=catalog.kubecarrier.io,resources=derivedcustomresourcedefinitions,verbs=create;update,versions=v1alpha1,name=vderivedcustomresourcedefinition.kubecarrier.io

// Handle is the function to handle create/update requests of DerivedCustomResourceDefinitions.
func (r *DerivedCustomResourceDefinitionWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &catalogv1alpha1.DerivedCustomResourceDefinition{}
	if err := r.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	switch req.Operation {
	case adminv1beta1.Create:
		if err := r.validateCreate(obj); err != nil {
			return admission.Denied(err.Error())
		}
	case adminv1beta1.Update:
		oldObj := obj.DeepCopyObject()
		if err := r.decoder.DecodeRaw(req.OldObject, oldObj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if err := r.validateUpdate(obj, oldObj); err != nil {
			return admission.Denied(err.Error())
		}
	}
	return admission.Allowed("allowed to commit the request")

}

// DerivedCustomResourceDefinitionWebhookHandler implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (r *DerivedCustomResourceDefinitionWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	r.decoder = d
	return nil
}

func (r *DerivedCustomResourceDefinitionWebhookHandler) validateCreate(derivedCustomResourceDefinition *catalogv1alpha1.DerivedCustomResourceDefinition) error {
	r.Log.Info("validate create", "name", derivedCustomResourceDefinition.Name)
	if derivedCustomResourceDefinition.Spec.BaseCRD.Name == "" {
		return fmt.Errorf("the BaseCRD of DerivedCustomResourceDefinition is not specifed")
	}
	return r.validateExpose(derivedCustomResourceDefinition)
}

func (r *DerivedCustomResourceDefinitionWebhookHandler) validateUpdate(obj *catalogv1alpha1.DerivedCustomResourceDefinition, oldObj runtime.Object) error {
	r.Log.Info("validate update", "name", obj.Name)
	oldDerivedCRD, ok := oldObj.(*catalogv1alpha1.DerivedCustomResourceDefinition)
	if !ok {
		return fmt.Errorf("expect old object to be a %T instead of %T\n", oldDerivedCRD, oldObj)
	}
	if obj.Spec.BaseCRD.Name != oldDerivedCRD.Spec.BaseCRD.Name {
		return fmt.Errorf("the BaseCRD of DerivedCustomResourceDefinition is immutable")
	}
	return r.validateExpose(obj)
}

func (r *DerivedCustomResourceDefinitionWebhookHandler) validateExpose(derivedCustomResourceDefinition *catalogv1alpha1.DerivedCustomResourceDefinition) error {
	if derivedCustomResourceDefinition.Spec.Expose == nil || len(derivedCustomResourceDefinition.Spec.Expose) == 0 {
		return fmt.Errorf("the DerivedCustomResourceDefinition should expose at least one version config")
	}

	for _, versionExposeConfig := range derivedCustomResourceDefinition.Spec.Expose {
		if versionExposeConfig.Versions == nil || len(versionExposeConfig.Versions) == 0 {
			return fmt.Errorf("the VersionExposeConfig should contain at least one exposed version")
		}
		if versionExposeConfig.Fields == nil || len(versionExposeConfig.Fields) == 0 {
			return fmt.Errorf("the VersionExposeConfig should contain at least one exposed Field")
		}
	}

	return nil
}

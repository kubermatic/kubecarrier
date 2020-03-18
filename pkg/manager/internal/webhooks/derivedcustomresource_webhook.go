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
	"strings"

	"github.com/go-logr/logr"
	adminv1beta1 "k8s.io/api/admission/v1beta1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

// DerivedCustomResourceWebhookHandler handles mutating/validating of DerivedCustomResources.
type DerivedCustomResourceWebhookHandler struct {
	decoder *admission.Decoder
	Log     logr.Logger
	client.Client
	Scheme *runtime.Scheme
}

var _ admission.Handler = (*DerivedCustomResourceWebhookHandler)(nil)

// +kubebuilder:webhook:path=/validate-catalog-kubecarrier-io-v1alpha1-derivedcustomresource,mutating=false,failurePolicy=fail,groups=catalog.kubecarrier.io,resources=derivedcustomresources,verbs=update;delete,versions=v1alpha1,name=vderivedcustomresource.kubecarrier.io

// Handle is the function to handle create/update requests of DerivedCustomResources.
func (r *DerivedCustomResourceWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	switch req.Operation {
	case adminv1beta1.Update:
		obj := &catalogv1alpha1.DerivedCustomResource{}
		if err := r.decoder.Decode(req, obj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		oldObj := &catalogv1alpha1.DerivedCustomResource{}
		if err := r.decoder.DecodeRaw(req.OldObject, oldObj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if err := r.validateUpdate(oldObj, obj); err != nil {
			return admission.Denied(err.Error())
		}
	case adminv1beta1.Delete:
		oldObj := &catalogv1alpha1.DerivedCustomResource{}
		if err := r.decoder.DecodeRaw(req.OldObject, oldObj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if err := r.validateDelete(ctx, oldObj); err != nil {
			return admission.Denied(err.Error())
		}
	}
	return admission.Allowed("allowed to commit the request")

}

// DerivedCustomResourceWebhookHandler implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (r *DerivedCustomResourceWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	r.decoder = d
	return nil
}

func (r *DerivedCustomResourceWebhookHandler) validateUpdate(oldObj, newObj *catalogv1alpha1.DerivedCustomResource) error {
	r.Log.Info("validate update", "name", newObj.Name)
	if newObj.Spec.BaseCRD.Name != oldObj.Spec.BaseCRD.Name {
		return fmt.Errorf("the BaseCRD of DerivedCustomResource is immutable")
	}
	return nil
}

func (r *DerivedCustomResourceWebhookHandler) validateDelete(ctx context.Context, obj *catalogv1alpha1.DerivedCustomResource) error {
	if obj.Status.DerivedCR == nil {
		return nil
	}

	crd := &apiextensionsv1.CustomResourceDefinition{}
	if err := r.Get(ctx, types.NamespacedName{
		Name: obj.Status.DerivedCR.Plural + "." + obj.Status.DerivedCR.Group,
	}, crd); err != nil {
		return err
	}
	u := &unstructured.UnstructuredList{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   crd.Spec.Group,
		Version: crd.Spec.Versions[0].Name,
		Kind:    crd.Spec.Names.ListKind,
	})
	if err := r.List(ctx, u); err != nil {
		return err
	}
	if len(u.Items) == 0 {
		return nil
	}

	errorMsg := new(strings.Builder)
	errorMsg.WriteString("derived CRD instances are still present in the cluster\n")
	for _, it := range u.Items {
		errorMsg.WriteString(util.MustLogLine(&it, r.Scheme) + " still present\n")
	}
	return fmt.Errorf("%s", errorMsg)
}

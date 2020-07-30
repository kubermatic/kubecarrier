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

	corev1alpha1 "k8c.io/kubecarrier/pkg/apis/core/v1alpha1"
)

// ServiceClusterAssignmentWebhookHandler handles mutating/validating of ServiceClusterAssignments.
type ServiceClusterAssignmentWebhookHandler struct {
	decoder *admission.Decoder
	Log     logr.Logger
}

var _ admission.Handler = (*ServiceClusterAssignmentWebhookHandler)(nil)

// +kubebuilder:webhook:path=/validate-kubecarrier-io-v1alpha1-serviceclusterassignment,mutating=false,failurePolicy=fail,groups=kubecarrier.io,resources=serviceclusterassignments,verbs=create;update,versions=v1alpha1,name=vserviceclusterassignment.kubecarrier.io

// Handle is the function to handle create/update requests of ServiceClusterAssignments.
func (r *ServiceClusterAssignmentWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &corev1alpha1.ServiceClusterAssignment{}
	if err := r.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Validate the object
	switch req.Operation {
	case adminv1beta1.Create:
		if err := r.validateCreate(obj); err != nil {
			return admission.Denied(err.Error())
		}
	case adminv1beta1.Update:
		oldObj := &corev1alpha1.ServiceClusterAssignment{}
		if err := r.decoder.DecodeRaw(req.OldObject, oldObj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if err := r.validateUpdate(oldObj, obj); err != nil {
			return admission.Denied(err.Error())
		}
	}
	return admission.Allowed("allowed to commit the request")

}

// ServiceClusterAssignmentWebhookHandler implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (r *ServiceClusterAssignmentWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	r.decoder = d
	return nil
}

func (r *ServiceClusterAssignmentWebhookHandler) validateCreate(serviceClusterAssignment *corev1alpha1.ServiceClusterAssignment) error {
	r.Log.Info("validate create", "name", serviceClusterAssignment.Name)
	return r.validateName(serviceClusterAssignment)
}

func (r *ServiceClusterAssignmentWebhookHandler) validateUpdate(oldObj, newObj *corev1alpha1.ServiceClusterAssignment) error {
	r.Log.Info("validate update", "name", newObj.Name)
	if newObj.Spec.ServiceCluster.Name != oldObj.Spec.ServiceCluster.Name ||
		newObj.Spec.ManagementClusterNamespace.Name != oldObj.Spec.ManagementClusterNamespace.Name {
		return fmt.Errorf("the ServiceCluster and ManagementClusterNamespace of ServiceClusterAssignment are immutable")
	}
	return r.validateName(newObj)
}

func (r *ServiceClusterAssignmentWebhookHandler) validateName(serviceClusterAssignment *corev1alpha1.ServiceClusterAssignment) error {
	desiredName := fmt.Sprintf("%s.%s",
		serviceClusterAssignment.Spec.ManagementClusterNamespace.Name,
		serviceClusterAssignment.Spec.ServiceCluster.Name)
	if serviceClusterAssignment.Name != desiredName {
		return fmt.Errorf("the Name of the ServiceClusterAssignment should be the compound of <management cluster namespace>.<service cluster>, found: %s, management cluster namespace: %s, service cluster: %s",
			serviceClusterAssignment.Name,
			serviceClusterAssignment.Spec.ManagementClusterNamespace.Name,
			serviceClusterAssignment.Spec.ServiceCluster.Name)
	}
	return nil
}

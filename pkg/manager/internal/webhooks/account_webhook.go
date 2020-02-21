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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
	"github.com/kubermatic/kubecarrier/pkg/internal/util/webhook"
)

// AccountWebhookHandler handles mutating/validating of Accounts.
type AccountWebhookHandler struct {
	decoder *admission.Decoder
	Log     logr.Logger
	Scheme  *runtime.Scheme
	Client  client.Client
}

var _ admission.Handler = (*AccountWebhookHandler)(nil)

// +kubebuilder:webhook:path=/validate-catalog-kubecarrier-io-v1alpha1-account,mutating=false,failurePolicy=fail,groups=catalog.kubecarrier.io,resources=accounts,verbs=create;update;delete,versions=v1alpha1,name=vaccount.kubecarrier.io

// Handle is the function to handle create/update requests of Accounts.
func (r *AccountWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	switch req.Operation {
	case adminv1beta1.Create:
		obj := &catalogv1alpha1.Account{}
		if err := r.decoder.Decode(req, obj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if err := r.validateCreate(ctx, obj); err != nil {
			return admission.Denied(err.Error())
		}
	case adminv1beta1.Update:
		obj := &catalogv1alpha1.Account{}
		if err := r.decoder.Decode(req, obj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		oldObj := &catalogv1alpha1.Account{}
		if err := r.decoder.DecodeRaw(req.OldObject, oldObj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if err := r.validateUpdate(oldObj, obj); err != nil {
			return admission.Denied(err.Error())
		}
	case adminv1beta1.Delete:
		oldObj := &catalogv1alpha1.Account{}
		if err := r.decoder.DecodeRaw(req.OldObject, oldObj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if err := r.validateDelete(ctx, oldObj); err != nil {
			return admission.Denied(err.Error())
		}
	}
	return admission.Allowed("allowed to commit the request")
}

// AccountWebhookHandler implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (r *AccountWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	r.decoder = d
	return nil
}

func (r *AccountWebhookHandler) validateCreate(ctx context.Context, account *catalogv1alpha1.Account) error {
	r.Log.Info("validate create", "name", account.Name)
	if !webhook.IsDNS1123Label(account.Name) {
		return fmt.Errorf("account name: %s is not a valid DNS 1123 Label, %s", account.Name, webhook.DNS1123LabelDescription)
	}
	err := r.Client.Get(ctx, types.NamespacedName{
		Name: account.Name,
	}, &corev1.Namespace{})
	switch {
	case errors.IsNotFound(err):
		break
	case err == nil:
		return fmt.Errorf("namespace %s already exists: %w", account.Name, err)
	default:
		return fmt.Errorf("getting namespace: %w", err)
	}
	return r.validateMetadataAndRoles(account)
}

func (r *AccountWebhookHandler) validateUpdate(oldObj, newObj *catalogv1alpha1.Account) error {
	r.Log.Info("validate update", "name", newObj.Name)
	return r.validateMetadataAndRoles(newObj)
}

func (r *AccountWebhookHandler) validateMetadataAndRoles(account *catalogv1alpha1.Account) error {
	if account.Spec.Metadata.Description == "" || account.Spec.Metadata.DisplayName == "" {
		return fmt.Errorf("the description or the display name of the Account: %s cannot be empty", account.Name)
	}

	roles := make(map[catalogv1alpha1.AccountRole]struct{})
	for _, role := range account.Spec.Roles {
		_, duplicate := roles[role]
		if duplicate {
			return fmt.Errorf("role %v is duplicated", role)
		}
		roles[role] = struct{}{}
	}
	if len(account.Spec.Roles) == 0 {
		return fmt.Errorf("no roles assigned")
	}
	return nil
}

// +kubebuilder:rbac:groups=kubecarrier.io,resources=serviceclusterassignments,verbs=get;list;watch

func (r *AccountWebhookHandler) validateDelete(ctx context.Context, obj *catalogv1alpha1.Account) error {
	// if no namespace was created for the object, we are safe to delete it
	// there's unlikely race condition here if the namespace was created, but not propagated to the account and
	// deletion blocking objects were created in the namespace
	if obj.Status.NamespaceName == "" {
		return nil
	}

	deletionBlockingObjects, err := util.ListObjects(ctx, r.Client, r.Scheme, []runtime.Object{
		&catalogv1alpha1.DerivedCustomResource{},
		&corev1alpha1.CustomResourceDiscovery{},
		&corev1alpha1.CustomResourceDiscoverySet{},
		&corev1alpha1.ServiceClusterAssignment{},
	}, client.InNamespace(obj.Status.NamespaceName))
	if err != nil {
		return fmt.Errorf("listingOwnedObjects: %w", err)
	}

	if len(deletionBlockingObjects) == 0 {
		return nil
	}

	buff := new(strings.Builder)
	buff.WriteString("deletion blocking objects found:\n")
	for _, obj := range deletionBlockingObjects {
		u := &unstructured.Unstructured{}
		if err := r.Scheme.Convert(obj, u, nil); err != nil {
			return fmt.Errorf("cannot convert %T: %w", obj, err)
		}
		_, _ = fmt.Fprintf(buff, "%s.%s: %s\n", u.GetKind(), u.GetAPIVersion(), u.GetName())
	}
	return fmt.Errorf("%s", buff.String())
}

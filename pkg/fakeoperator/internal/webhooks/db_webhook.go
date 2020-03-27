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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	fakev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/fake/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util/webhook"
)

// DBWebhookHandler handles mutating/validating of DBs.
type DBWebhookHandler struct {
	decoder *admission.Decoder
	Log     logr.Logger
	Scheme  *runtime.Scheme
	Client  client.Client
}

var _ admission.Handler = (*DBWebhookHandler)(nil)

// +kubebuilder:webhook:path=/validate-fake-kubecarrier-io-v1alpha1-db,mutating=false,failurePolicy=fail,groups=fake.kubecarrier.io,resources=dbs,verbs=create;update;delete,versions=v1alpha1,name=vdb.kubecarrier.io

// Handle is the function to handle create/update requests of DBs.
func (r *DBWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	switch req.Operation {
	case adminv1beta1.Create:
		obj := &fakev1alpha1.DB{}
		if err := r.decoder.Decode(req, obj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if err := r.validateCreate(ctx, obj); err != nil {
			return admission.Denied(err.Error())
		}
	case adminv1beta1.Update:
		obj := &fakev1alpha1.DB{}
		if err := r.decoder.Decode(req, obj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		oldObj := &fakev1alpha1.DB{}
		if err := r.decoder.DecodeRaw(req.OldObject, oldObj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if err := r.validateUpdate(oldObj, obj); err != nil {
			return admission.Denied(err.Error())
		}
	case adminv1beta1.Delete:
		oldObj := &fakev1alpha1.DB{}
		if err := r.decoder.DecodeRaw(req.OldObject, oldObj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if err := r.validateDelete(ctx, oldObj); err != nil {
			return admission.Denied(err.Error())
		}
	}
	return admission.Allowed("allowed to commit the request")
}

// DBWebhookHandler implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (r *DBWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	r.decoder = d
	return nil
}

func (r *DBWebhookHandler) validateCreate(ctx context.Context, db *fakev1alpha1.DB) error {
	r.Log.Info("validate create", "name", db.Name)
	if !webhook.IsDNS1123Label(db.Name) {
		return fmt.Errorf("db name: %s is not a valid DNS 1123 Label, %s", db.Name, webhook.DNS1123LabelDescription)
	}
	err := r.Client.Get(ctx, types.NamespacedName{
		Name: db.Name,
	}, &corev1.Namespace{})
	switch {
	case errors.IsNotFound(err):
		break
	case err == nil:
		return fmt.Errorf("namespace %s already exists", db.Name)
	default:
		return fmt.Errorf("getting namespace: %w", err)
	}
	return nil
}

func (r *DBWebhookHandler) validateUpdate(oldObj, newObj *fakev1alpha1.DB) error {
	r.Log.Info("validate update", "name", newObj.Name)
	return nil
}

// +kubebuilder:rbac:groups=fake.kubecarrier.io,resources=dbs,verbs=get;list;watch
func (r *DBWebhookHandler) validateDelete(ctx context.Context, db *fakev1alpha1.DB) error {
	return nil
}

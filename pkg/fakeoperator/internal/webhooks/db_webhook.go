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
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	adminv1beta1 "k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	fakev1 "k8c.io/kubecarrier/pkg/apis/fake/v1"
	"k8c.io/kubecarrier/pkg/internal/util/webhook"
)

const defaultPassword = "password"

// DBWebhookHandler handles mutating/validating of DBs.
type DBWebhookHandler struct {
	decoder *admission.Decoder
	Log     logr.Logger
	Scheme  *runtime.Scheme
	Client  client.Client
}

var _ admission.Handler = (*DBWebhookHandler)(nil)

func (r *DBWebhookHandler) defaultObject(obj *fakev1.DB) (*fakev1.DB, bool) {
	newObj := obj.DeepCopy()
	changed := false
	if obj.Spec.DatabasePassword == "" {
		newObj.Spec.DatabasePassword = defaultPassword
		changed = true
	}
	if obj.Spec.RootPassword == "" {
		newObj.Spec.RootPassword = defaultPassword
		changed = true
	}
	return newObj, changed
}

// +kubebuilder:webhook:path=/mutate-fake-kubecarrier-io-v1-db,mutating=true,failurePolicy=fail,groups=fake.kubecarrier.io,resources=dbs,verbs=create;update;delete,versions=v1,name=mdb.kubecarrier.io,matchPolicy=Equivalent,sideEffects=NoneOnDryRun

// Handle is the function to handle create/update requests of DBs.
func (r *DBWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	switch req.Operation {
	case adminv1beta1.Create:
		obj := &fakev1.DB{}
		if err := r.decoder.Decode(req, obj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if err := r.validateCreate(ctx, obj); err != nil {
			return admission.Denied(err.Error())
		}
		if newObj, changed := r.defaultObject(obj); changed {
			marshalledObj, err := json.Marshal(newObj)
			if err != nil {
				return admission.Errored(http.StatusInternalServerError, err)
			}
			// Create the patch
			return admission.PatchResponseFromRaw(req.Object.Raw, marshalledObj)
		}
	case adminv1beta1.Update:
		obj := &fakev1.DB{}
		if err := r.decoder.Decode(req, obj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		oldObj := &fakev1.DB{}
		if err := r.decoder.DecodeRaw(req.OldObject, oldObj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if err := r.validateUpdate(oldObj, obj); err != nil {
			return admission.Denied(err.Error())
		}
	case adminv1beta1.Delete:
		oldObj := &fakev1.DB{}
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

func (r *DBWebhookHandler) validateCreate(ctx context.Context, db *fakev1.DB) error {
	r.Log.Info("mutate create", "name", db.Name)
	if !db.Spec.Config.Create.Enabled() {
		return fmt.Errorf("create operation disabled for %s", db.Name)
	}
	if !webhook.IsDNS1123Label(db.Name) {
		return fmt.Errorf("DB name: %s is not a valid DNS 1123 Label, %s", db.Name, webhook.DNS1123LabelDescription)
	}
	return nil
}

func (r *DBWebhookHandler) validateUpdate(oldObj, newObj *fakev1.DB) error {
	r.Log.Info("mutate update", "name", newObj.Name)
	if !oldObj.Spec.Config.Update.Enabled() && !newObj.Spec.Config.Update.Enabled() {
		return fmt.Errorf("update operation disabled for %s", oldObj.Name)
	}
	if newObj.Spec.DatabaseName != oldObj.Spec.DatabaseName {
		return fmt.Errorf("the Database name is immutable")
	}
	return nil
}

// +kubebuilder:rbac:groups=fake.kubecarrier.io,resources=dbs,verbs=get;list;watch
func (r *DBWebhookHandler) validateDelete(ctx context.Context, db *fakev1.DB) error {
	if !db.Spec.Config.Delete.Enabled() {
		return fmt.Errorf("delete operation disabled for %s", db.Name)
	}
	return nil
}

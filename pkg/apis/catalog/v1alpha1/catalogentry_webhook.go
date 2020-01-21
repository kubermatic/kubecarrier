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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var catalogEntryLog = logf.Log.WithName("CatalogEntry")

func (r *CatalogEntry) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-catalog-kubecarrier-io-v1alpha1-catalogentry,mutating=true,failurePolicy=fail,groups=catalog.kubecarrier.io,resources=catalogentries,verbs=create;update,versions=v1alpha1,name=mcatalogentry.kb.io
// +kubebuilder:webhook:path=/validate-catalog-kubecarrier-io-v1alpha1-catalogentry,mutating=false,failurePolicy=fail,groups=catalog.kubecarrier.io,resources=catalogentries,verbs=create;update,versions=v1alpha1,name=vcatalogentry.kb.io

var (
	_ webhook.Defaulter = &CatalogEntry{}
	_ webhook.Validator = &CatalogEntry{}
)

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *CatalogEntry) Default() {
	catalogEntryLog.Info("default", "name", r.Name)
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CatalogEntry) ValidateCreate() error {
	catalogEntryLog.Info("validate create", "name", r.Name)
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *CatalogEntry) ValidateUpdate(old runtime.Object) error {
	catalogEntryLog.Info("validate update", "name", r.Name)
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *CatalogEntry) ValidateDelete() error {
	catalogEntryLog.Info("validate delete", "name", r.Name)
	return nil
}

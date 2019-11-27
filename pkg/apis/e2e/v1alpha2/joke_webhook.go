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

package v1alpha2

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var _ webhook.Validator = &Joke{}

var jokelog = controllerruntime.Log.WithName("joke-resource")

func (in *Joke) ValidateCreate() error {
	jokelog.Info("validate create", "name", in.Name)
	if len(in.Spec.Jokes) == 0 {
		return fmt.Errorf("joke database cannot be empty")
	}
	return nil
}

func (in *Joke) ValidateUpdate(old runtime.Object) error {
	jokelog.Info("validate update", "name", in.Name)
	if len(in.Spec.Jokes) == 0 {
		return fmt.Errorf("joke database cannot be empty")
	}
	return nil
}

func (in *Joke) ValidateDelete() error {
	jokelog.Info("validate delete", "name", in.Name)
	return nil
}

// Hub marks this type as a conversion hub.
func (*Joke) Hub() {}

// +kubebuilder:webhook:path=/joke-e2e-kubebuilder-io-webhook,mutating=false,failurePolicy=fail,groups=e2e.kubebuilder.io,resources=joke,verbs=create;update,versions=v1alpha2,name=mcronjob.kb.io
// +kubebuilder:webhook:verbs=create;update,path=/validate-batch-tutorial-kubebuilder-io-v1-cronjob,mutating=false,failurePolicy=fail,groups=batch.tutorial.kubebuilder.io,resources=cronjobs,versions=v1,name=vcronjob.kb.io

func (r *Joke) SetupWebhookWithManager(mgr controllerruntime.Manager) error {
	return controllerruntime.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

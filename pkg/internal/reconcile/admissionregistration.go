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

package reconcile

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	adminv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MutatingWebhookConfiguration reconciles a admissionregistration.k8s.io/v1beta1, Kind=MutatingWebhookConfiguration.
func MutatingWebhookConfiguration(
	ctx context.Context,
	log logr.Logger,
	c client.Client,
	desiredMutatingWebhookConfiguration *adminv1beta1.MutatingWebhookConfiguration,
) (currentMutatingWebhookConfiguration *adminv1beta1.MutatingWebhookConfiguration, err error) {
	nn := types.NamespacedName{
		Name:      desiredMutatingWebhookConfiguration.Name,
		Namespace: desiredMutatingWebhookConfiguration.Namespace,
	}
	currentMutatingWebhookConfiguration = &adminv1beta1.MutatingWebhookConfiguration{}
	err = c.Get(ctx, nn, currentMutatingWebhookConfiguration)
	if err != nil && !errors.IsNotFound(err) {
		return currentMutatingWebhookConfiguration, fmt.Errorf("getting MutatingWebhookConfiguration: %w", err)
	}
	if errors.IsNotFound(err) {
		// Create missing ValidatingWebhookConfiguration
		log.Info("creating", "MutatingWebhookConfiguration", nn.String())
		if err = c.Create(ctx, desiredMutatingWebhookConfiguration); err != nil {
			return currentMutatingWebhookConfiguration, fmt.Errorf("creating MutatingWebhookConfiguration: %w", err)
		}
	}
	return currentMutatingWebhookConfiguration, nil
}

// ValidatingWebhookConfiguration reconciles a admissionregistration.k8s.io/v1beta1, Kind=ValidatingWebhookConfiguration.
func ValidatingWebhookConfiguration(
	ctx context.Context,
	log logr.Logger,
	c client.Client,
	desiredValidatingWebhookConfiguration *adminv1beta1.ValidatingWebhookConfiguration,
) (currentValidatingWebhookConfiguration *adminv1beta1.ValidatingWebhookConfiguration, err error) {
	nn := types.NamespacedName{
		Name:      desiredValidatingWebhookConfiguration.Name,
		Namespace: desiredValidatingWebhookConfiguration.Namespace,
	}
	currentValidatingWebhookConfiguration = &adminv1beta1.ValidatingWebhookConfiguration{}
	err = c.Get(ctx, nn, currentValidatingWebhookConfiguration)
	if err != nil && !errors.IsNotFound(err) {
		return currentValidatingWebhookConfiguration, fmt.Errorf("getting ValidatingWebhookConfiguration: %w", err)
	}
	if errors.IsNotFound(err) {
		// Create missing ValidatingWebhookConfiguration
		log.Info("creating", "ValidatingWebhookConfiguration", nn.String())
		if err = c.Create(ctx, desiredValidatingWebhookConfiguration); err != nil {
			return currentValidatingWebhookConfiguration, fmt.Errorf("creating ValidatingWebhookConfiguration: %w", err)
		}
	}
	return currentValidatingWebhookConfiguration, nil
}

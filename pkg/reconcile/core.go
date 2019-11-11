/*
Copyright 2019 The Kubecarrier Authors.

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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceAccount reconciles a /v1, Kind=ServiceAccount.
func ServiceAccount(
	ctx context.Context,
	log logr.Logger,
	c client.Client,
	desiredServiceAccount *corev1.ServiceAccount,
) (currentServiceAccount *corev1.ServiceAccount, err error) {
	name := types.NamespacedName{
		Name:      desiredServiceAccount.Name,
		Namespace: desiredServiceAccount.Namespace,
	}

	// Lookup current version of the object
	currentServiceAccount = &corev1.ServiceAccount{}
	err = c.Get(ctx, name, currentServiceAccount)
	if err != nil && !errors.IsNotFound(err) {
		return nil, fmt.Errorf("getting ServiceAccount: %w", err)
	}

	if errors.IsNotFound(err) {
		// ServiceAccount needs to be created
		log.V(1).Info("creating", "ServiceAccount", name.String())
		if err = c.Create(ctx, desiredServiceAccount); err != nil {
			return nil, fmt.Errorf("creating ServiceAccount: %w", err)
		}
		// no need to check for updates, object was created just now
		return desiredServiceAccount, nil
	}

	return currentServiceAccount, nil
}

// Service reconciles a /v1, Kind=Service.
func Service(
	ctx context.Context,
	log logr.Logger,
	c client.Client,
	desiredService *corev1.Service,
) (currentService *corev1.Service, err error) {
	name := types.NamespacedName{
		Name:      desiredService.Name,
		Namespace: desiredService.Namespace,
	}

	// Lookup current version of the object
	currentService = &corev1.Service{}
	err = c.Get(ctx, name, currentService)
	if err != nil && !errors.IsNotFound(err) {
		// unexpected error
		return nil, fmt.Errorf("getting Service: %w", err)
	}

	if errors.IsNotFound(err) {
		// Service needs to be created
		log.Info("creating", "Service", name.String())
		if err = c.Create(ctx, desiredService); err != nil {
			return nil, fmt.Errorf("creating Service: %w", err)
		}
		// no need to check for updates, object was created just now
		return desiredService, nil
	}

	if !equality.Semantic.DeepEqual(desiredService.Spec.Selector, currentService.Spec.Selector) &&
		!equality.Semantic.DeepEqual(desiredService.Spec.Ports, currentService.Spec.Ports) {
		// desired and current Service .Spec are not equal -> trigger an update
		currentService.Spec.Selector = desiredService.Spec.Selector
		currentService.Spec.Ports = desiredService.Spec.Ports
		if err = c.Update(ctx, currentService); err != nil {
			return nil, fmt.Errorf("updating Service: %w", err)
		}
	}

	return currentService, nil
}

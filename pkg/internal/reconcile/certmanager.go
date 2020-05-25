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
	certv1alpha2 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Issuer reconciles a cert-manager.io/v1alpha2, Kind=Issuer.
func Issuer(
	ctx context.Context,
	log logr.Logger,
	c client.Client,
	desiredIssuer *certv1alpha2.Issuer,
) (currentIssuer *certv1alpha2.Issuer, err error) {
	nn := types.NamespacedName{
		Name:      desiredIssuer.Name,
		Namespace: desiredIssuer.Namespace,
	}
	currentIssuer = &certv1alpha2.Issuer{}
	err = c.Get(ctx, nn, currentIssuer)
	if err != nil && !errors.IsNotFound(err) {
		return currentIssuer, fmt.Errorf("getting Issuer: %w", err)
	}
	if errors.IsNotFound(err) {
		// Create missing Issuer
		log.V(1).Info("creating", "Issuer", nn.String())
		if err = c.Create(ctx, desiredIssuer); err != nil {
			return currentIssuer, fmt.Errorf("creating Issuer: %w", err)
		}
	}
	return currentIssuer, nil
}

// Certificate reconciles a cert-manager.io/v1alpha2, Kind=Certificate.
func Certificate(
	ctx context.Context,
	log logr.Logger,
	c client.Client,
	desiredCertificate *certv1alpha2.Certificate,
) (currentCertificate *certv1alpha2.Certificate, err error) {
	nn := types.NamespacedName{
		Name:      desiredCertificate.Name,
		Namespace: desiredCertificate.Namespace,
	}
	currentCertificate = &certv1alpha2.Certificate{}
	err = c.Get(ctx, nn, currentCertificate)
	if err != nil && !errors.IsNotFound(err) {
		return currentCertificate, fmt.Errorf("getting Certificate: %w", err)
	}
	if errors.IsNotFound(err) {
		// Create missing Certificate
		log.V(1).Info("creating", "Certificate", nn.String())
		if err = c.Create(ctx, desiredCertificate); err != nil {
			return currentCertificate, fmt.Errorf("creating Certificate: %w", err)
		}
	}
	return currentCertificate, nil
}

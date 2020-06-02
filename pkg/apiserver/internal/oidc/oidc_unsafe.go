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

/*
This is very very very specific instance when we're using unsafe. We want being as close as possible
to the k8s OIDC integration, thus we're reusing their authenticator and how they are parsing the flags

However their New() function creates the authenticator in async manner, which makes stuff tricky for us.

It's hard verifying the authenticator is initialized (you get a hard-coded error back, but you cannot make the
authentication pass due to asymmetric encryption nature.)

Thus we're re-exporting two private methods to create authenticator in a sync manner, and ensure it's initialized
by fetching the OIDC /.well-known/openid-configuration and letting it configure itself
*/
package oidc

import (
	"context"
	"fmt"
	_ "unsafe"

	"github.com/coreos/go-oidc"
	"github.com/go-logr/logr"
	k8soidc "k8s.io/apiserver/plugin/pkg/authenticator/token/oidc"
)

func newAuthenticator(log logr.Logger, opts k8soidc.Options) (*k8soidc.Authenticator, error) {
	// this is more-or-less similar implementation to k8soidc.New except it's synchronous
	// after this operation completed the authenticator shall be initialized
	var providerErr error
	auth, err := newAuthenticatorUnsafe(opts, func(ctx context.Context, a *k8soidc.Authenticator, config *oidc.Config) {
		// ctx contains http.Client the NewProvider shall use
		provider, err := oidc.NewProvider(ctx, opts.IssuerURL)
		if err != nil {
			providerErr = fmt.Errorf("oidc authenticator: initializing plugin: %w", err)
			return
		}
		verifier := provider.Verifier(config)
		setVerifierUnsafe(a, verifier)
		log.Info("initialized authenticator")
	})
	if err != nil {
		return nil, err
	}
	if providerErr != nil {
		return nil, providerErr
	}
	return auth, nil

}

//go:linkname newAuthenticatorUnsafe k8s.io/apiserver/plugin/pkg/authenticator/token/oidc.newAuthenticator
func newAuthenticatorUnsafe(opts k8soidc.Options, initVerifier func(ctx context.Context, a *k8soidc.Authenticator, config *oidc.Config)) (*k8soidc.Authenticator, error)

//go:linkname setVerifierUnsafe k8s.io/apiserver/plugin/pkg/authenticator/token/oidc.(*Authenticator).setVerifier
func setVerifierUnsafe(a *k8soidc.Authenticator, v *oidc.IDTokenVerifier)

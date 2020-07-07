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

package oidc

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/plugin/pkg/authenticator/token/oidc"
	cliflag "k8s.io/component-base/cli/flag"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"

	"github.com/kubermatic/kubecarrier/pkg/apiserver/auth"
)

func init() {
	auth.RegisterAuthProvider("OIDC", &OIDCAuthenticator{})
}

type OIDCAuthenticator struct {
	APIAudiences  []string
	opts          oidc.Options
	authenticator *oidc.Authenticator
	logr.Logger
}

var _ auth.Provider = (*OIDCAuthenticator)(nil)
var _ inject.Logger = (*OIDCAuthenticator)(nil)

func (O *OIDCAuthenticator) InjectLogger(l logr.Logger) error {
	O.Logger = l
	return nil
}

func (O *OIDCAuthenticator) Init() error {
	authenticator, err := newAuthenticator(O.Logger, O.opts)
	if err != nil {
		return err
	}
	O.Info("setting up OIDC authenticator middleware", "iss", O.opts.IssuerURL)
	O.authenticator = authenticator
	return nil
}

func (O *OIDCAuthenticator) Authenticate(ctx context.Context) (user.Info, error) {
	token, err := grpc_auth.AuthFromMD(ctx, "Bearer")
	if err != nil {
		O.Error(err, "cannot extract token")
		return nil, err
	}
	resp, present, err := O.authenticator.AuthenticateToken(ctx, token)
	if err != nil {
		O.Error(err, "cannot authenticate token", "token", token)
		return nil, status.Error(codes.Unauthenticated, "AuthenticateToken")
	}
	if !present {
		O.Error(fmt.Errorf("missing token"), "Token is not present")
		return nil, status.Error(codes.Unauthenticated, "AuthenticateToken")
	}
	return resp.User, nil
}

func (O *OIDCAuthenticator) AddFlags(fs *pflag.FlagSet) {
	opts := &O.opts
	fs.StringVar(&opts.IssuerURL, "oidc-issuer-url", opts.IssuerURL, ""+
		"The URL of the OpenID issuer, only HTTPS scheme will be accepted. "+
		"If set, it will be used to verify the OIDC JSON Web Token (JWT).")

	fs.StringVar(&opts.ClientID, "oidc-client-id", opts.ClientID,
		"The client ID for the OpenID Connect client, must be set if oidc-issuer-url is set.")

	fs.StringVar(&opts.CAFile, "oidc-ca-file", opts.CAFile, ""+
		"If set, the OpenID server'opts certificate will be verified by one of the authorities "+
		"in the oidc-ca-file, otherwise the host'opts root CA set will be used.")

	fs.StringVar(&opts.UsernameClaim, "oidc-username-claim", "sub", ""+
		"The OpenID claim to use as the user name. Note that claims other than the default ('sub') "+
		"is not guaranteed to be unique and immutable. This flag is experimental, please see "+
		"the authentication documentation for further details.")

	fs.StringVar(&opts.UsernamePrefix, "oidc-username-prefix", "", ""+
		"If provided, all usernames will be prefixed with this value. If not provided, "+
		"username claims other than 'email' are prefixed by the issuer URL to avoid "+
		"clashes. To skip any prefixing, provide the value '-'.")

	fs.StringVar(&opts.GroupsClaim, "oidc-groups-claim", "", ""+
		"If provided, the name of a custom OpenID Connect claim for specifying user groups. "+
		"The claim value is expected to be a string or array of strings. This flag is experimental, "+
		"please see the authentication documentation for further details.")

	fs.StringVar(&opts.GroupsPrefix, "oidc-groups-prefix", "", ""+
		"If provided, all groups will be prefixed with this value to prevent conflicts with "+
		"other authentication strategies.")

	fs.StringSliceVar(&opts.SupportedSigningAlgs, "oidc-signing-algs", []string{"RS256"}, ""+
		"Comma-separated list of allowed JOSE asymmetric signing algorithms. JWTs with a "+
		"'alg' header value not in this list will be rejected. "+
		"Values are defined by RFC 7518 https://tools.ietf.org/html/rfc7518#section-3.1.")

	fs.Var(cliflag.NewMapStringStringNoSplit(&opts.RequiredClaims), "oidc-required-claim", ""+
		"A key=value pair that describes a required claim in the ID Token. "+
		"If set, the claim is verified to be present in the ID Token with a matching value. "+
		"Repeat this flag to specify multiple claims.")
}

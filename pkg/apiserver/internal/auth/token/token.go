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

package anonymous

import (
	"context"

	"github.com/go-logr/logr"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	authenticationv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/authentication/user"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"

	"github.com/kubermatic/kubecarrier/pkg/apiserver/auth"
)

type Auth struct {
	logr.Logger
	client       client.Client
	APIAudiences []string
}

func init() {
	auth.RegisterAuthProvider("Token", &Auth{})
}

var _ auth.Provider = (*Auth)(nil)
var _ inject.Logger = (*Auth)(nil)
var _ inject.Client = (*Auth)(nil)

func (a *Auth) InjectLogger(l logr.Logger) error {
	a.Logger = l
	return nil
}

func (a *Auth) InjectClient(c client.Client) error {
	a.client = c
	return nil
}

func (a *Auth) AddFlags(fs *pflag.FlagSet) {
	fs.StringArrayVar(&a.APIAudiences, "--token-api-audiences ", a.APIAudiences,
		"Identifiers of the API. The service account token authenticator will "+
			"validate that tokens used against the API are bound to at least one of these "+
			"audiences",
	)
}

func (a *Auth) Init() error {
	return nil
}

// +kubebuilder:rbac:groups="authentication.k8s.io",resources=tokenreviews,verbs=create

func (a *Auth) Authenticate(ctx context.Context) (user.Info, error) {
	token, err := grpc_auth.AuthFromMD(ctx, "Bearer")
	if err != nil {
		a.Error(err, "cannot extract token")
		return nil, err
	}
	tokenReview := &authenticationv1.TokenReview{
		Spec: authenticationv1.TokenReviewSpec{
			Token: token,
		},
	}
	if err := a.client.Create(ctx, tokenReview); err != nil {
		return nil, err
	}

	if !tokenReview.Status.Authenticated {
		return nil, status.Error(codes.Unauthenticated, tokenReview.Status.Error)
	}

	if len(a.APIAudiences) > 0 && sets.NewString(a.APIAudiences...).Intersection(sets.NewString(tokenReview.Status.Audiences...)).Len() == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "wrong API audience")
	}

	userInfo := &user.DefaultInfo{
		Name:   tokenReview.Status.User.Username,
		UID:    tokenReview.Status.User.UID,
		Groups: tokenReview.Status.User.Groups,
	}
	if tokenReview.Status.User.Extra != nil {
		userInfo.Extra = make(map[string][]string)
		for k, v := range tokenReview.Status.User.Extra {
			userInfo.Extra[k] = v
		}
	}
	return userInfo, nil
}

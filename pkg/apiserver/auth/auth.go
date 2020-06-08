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

package auth

import (
	"context"
	"fmt"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
)

type contextKey string

type AuthProvider interface {
	RegisterPFlags(fs *pflag.FlagSet)
	Init() error
	Authenticate(ctx context.Context) (user.Info, error)
}

var authProviderFactory = make(map[string]AuthProvider)

func RegisterAuthProvider(name string, provider AuthProvider) {
	authProviderFactory[name] = provider
}

func RegisteredAuthProviders() (out []string) {
	for k := range authProviderFactory {
		out = append(out, k)
	}
	return
}

func RegisterPFlags(fs *pflag.FlagSet) {
	for _, provider := range authProviderFactory {
		provider.RegisterPFlags(fs)
	}
}

func newAuthProvider(name string) (AuthProvider, error) {
	authProvider, ok := authProviderFactory[name]
	if !ok {
		return nil, fmt.Errorf("unknown authorization mode: %v", name)
	}
	return authProvider, nil
}

const (
	userInfoKey contextKey = "userinfo.kubecarrier.io"
)

func ExtractUserInfo(ctx context.Context) (user.Info, bool) {
	u, ok := ctx.Value(userInfoKey).(*authenticator.Response)
	return u.User, ok
}

func CreateAuthFunction(authorizationModes []string) (grpc_auth.AuthFunc, error) {
	authProviders := make([]AuthProvider, 0, len(authorizationModes))
	for _, mode := range authorizationModes {
		authProvider, err := newAuthProvider(mode)
		if err != nil {
			return nil, err
		}
		if err := authProvider.Init(); err != nil {
			return nil, fmt.Errorf("cannot init auth provider: %s: %w", mode, err)
		}
		authProviders = append(authProviders, authProvider)
	}

	return func(ctx context.Context) (context.Context, error) {
		for _, provider := range authProviders {
			userInfo, err := provider.Authenticate(ctx)
			if err != nil {
				s, ok := status.FromError(err)
				if !ok {
					s = status.New(codes.Unknown, err.Error())
				}
				if s.Code() == codes.Unauthenticated {
					continue
				}
				return ctx, err
			}
			return context.WithValue(ctx, userInfoKey, userInfo), nil
		}
		return ctx, status.Error(codes.Unauthenticated, "no auth plugin successfully authenticated the user")
	}, nil
}

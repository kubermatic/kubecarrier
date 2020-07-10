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

	"github.com/spf13/pflag"
	"k8s.io/apiserver/pkg/authentication/user"

	"github.com/kubermatic/kubecarrier/pkg/apiserver/auth"
)

type Auth struct{}

var _ auth.Provider = (*Auth)(nil)

func init() {
	auth.RegisterAuthProvider(auth.ProviderAnynymous, &Auth{})
}

func (a Auth) AddFlags(fs *pflag.FlagSet) {}

func (a Auth) Init() error {
	return nil
}

func (a Auth) Authenticate(ctx context.Context) (user.Info, error) {
	return &user.DefaultInfo{
		Name:   "system:anonymous",
		Groups: []string{"system:unauthenticated"},
	}, nil
}

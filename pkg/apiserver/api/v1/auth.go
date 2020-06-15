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

package v1

import (
	context "context"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/apiserver/internal/authorizer"
)

func (req *ListRequest) Authorize(ctx context.Context, a authorizer.Authorizer) error {
	return a.Authorize(ctx, &catalogv1alpha1.Offering{}, authorizer.AuthorizationOption{
		Namespace: req.Account,
		Verb:      authorizer.RequestList,
	})
}

func (req *GetRequest) Authorize(ctx context.Context, a authorizer.Authorizer) error {
	return a.Authorize(ctx, &catalogv1alpha1.Offering{}, authorizer.AuthorizationOption{
		Name:      req.Name,
		Namespace: req.Account,
		Verb:      authorizer.RequestGet,
	})
}

func (req *WatchRequest) Authorize(ctx context.Context, a authorizer.Authorizer) error {
	return a.Authorize(ctx, &catalogv1alpha1.Offering{}, authorizer.AuthorizationOption{
		Namespace: req.Account,
		Verb:      authorizer.RequestWatch,
	})
}

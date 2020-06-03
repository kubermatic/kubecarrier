/*
Copyright 2020 The KubeCarrier Authors.

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
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	v1 "github.com/kubermatic/kubecarrier/pkg/apiserver/api/v1"
	"github.com/kubermatic/kubecarrier/pkg/apiserver/internal/authorizer"
)

type offeringAuthWrapper struct {
	service    v1.OfferingServiceServer
	authorizer authorizer.Authorizer
}

func NewOfferingAuthWrapper(service v1.OfferingServiceServer, authorizer authorizer.Authorizer) v1.OfferingServiceServer {
	return &offeringAuthWrapper{
		service:    service,
		authorizer: authorizer,
	}
}

func (w offeringAuthWrapper) List(ctx context.Context, req *v1.ListRequest) (res *v1.OfferingList, err error) {
	if err := w.authorizer.Authorize(ctx, &catalogv1alpha1.Offering{}, authorizer.AuthorizationOption{
		Namespace: req.Account,
		Verb:      authorizer.RequestList,
	}); err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return w.service.List(ctx, req)
}

func (w offeringAuthWrapper) Get(ctx context.Context, req *v1.GetRequest) (res *v1.Offering, err error) {
	if err := w.authorizer.Authorize(ctx, &catalogv1alpha1.Offering{}, authorizer.AuthorizationOption{
		Name:      req.Name,
		Namespace: req.Account,
		Verb:      authorizer.RequestGet,
	}); err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return w.service.Get(ctx, req)
}

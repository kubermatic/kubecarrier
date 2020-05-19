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
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	v1 "github.com/kubermatic/kubecarrier/pkg/apiserver/api/v1"
	"github.com/kubermatic/kubecarrier/pkg/apiserver/internal/util"
)

type providerServer struct {
	client client.Client
}

var _ v1.ProviderServiceServer = (*providerServer)(nil)

// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=providers,verbs=get;list

func NewProviderServiceServer(c client.Client) v1.ProviderServiceServer {
	return &providerServer{
		client: c,
	}
}

func (o providerServer) List(ctx context.Context, req *v1.ListRequest) (res *v1.ProviderList, err error) {
	listOptions, err := validateListRequest(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	providerList := &catalogv1alpha1.ProviderList{}
	if err := o.client.List(ctx, providerList, listOptions...); err != nil {
		return nil, fmt.Errorf("listing providers: %w", err)
	}

	res, err = o.convertProviderList(providerList)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("converting ProviderList: %s", err.Error()))
	}
	return
}

func (o providerServer) Get(ctx context.Context, req *v1.GetRequest) (res *v1.Provider, err error) {
	if err := validateGetRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	provider := &catalogv1alpha1.Provider{}
	if err = o.client.Get(ctx, types.NamespacedName{
		Name:      req.Name,
		Namespace: req.Account,
	}, provider); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("getting provider: %s", err.Error()))
	}
	res, err = o.convertProvider(provider)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("converting Provider: %s", err.Error()))
	}
	return
}

func (o providerServer) convertProvider(in *catalogv1alpha1.Provider) (out *v1.Provider, err error) {
	creationTimestamp, err := util.TimestampProto(&in.ObjectMeta.CreationTimestamp)
	if err != nil {
		return nil, err
	}
	deletionTimestamp, err := util.TimestampProto(in.ObjectMeta.DeletionTimestamp)
	if err != nil {
		return nil, err
	}
	out = &v1.Provider{
		Metadata: &v1.ObjectMeta{
			Uid:               string(in.UID),
			Name:              in.Name,
			Account:           in.Namespace,
			CreationTimestamp: creationTimestamp,
			DeletionTimestamp: deletionTimestamp,
			ResourceVersion:   in.ResourceVersion,
			Labels:            in.Labels,
			Annotations:       in.Annotations,
			Generation:        in.Generation,
		},
		Spec: &v1.ProviderSpec{
			Metadata: &v1.AccountMetadata{
				DisplayName: in.Spec.Metadata.DisplayName,
				Description: in.Spec.Metadata.Description,
			},
		},
	}
	return
}

func (o providerServer) convertProviderList(in *catalogv1alpha1.ProviderList) (out *v1.ProviderList, err error) {
	out = &v1.ProviderList{
		Metadata: &v1.ListMeta{
			Continue:        in.Continue,
			ResourceVersion: in.ResourceVersion,
		},
	}
	for _, inProvider := range in.Items {
		provider, err := o.convertProvider(&inProvider)
		if err != nil {
			return nil, err
		}
		out.Items = append(out.Items, provider)
	}
	return
}

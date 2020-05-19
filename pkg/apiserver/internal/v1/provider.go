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
	"k8s.io/apimachinery/pkg/labels"
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

func (o providerServer) validateListRequest(req *v1.ProviderListRequest) ([]client.ListOption, error) {
	var listOptions []client.ListOption
	if req.Account == "" {
		return listOptions, fmt.Errorf("missing namespace")
	}
	listOptions = append(listOptions, client.InNamespace(req.Account))
	if req.Limit < 0 {
		return listOptions, fmt.Errorf("invalid limit: should not be negative number")
	}
	listOptions = append(listOptions, client.Limit(req.Limit))
	if req.LabelSelector != "" {
		selector, err := labels.Parse(req.LabelSelector)
		if err != nil {
			return listOptions, fmt.Errorf("invalid LabelSelector: %w", err)
		}
		listOptions = append(listOptions, client.MatchingLabelsSelector{
			Selector: selector,
		})
	}
	if req.Continue != "" {
		listOptions = append(listOptions, client.Continue(req.Continue))
	}
	return listOptions, nil
}

func (o providerServer) List(ctx context.Context, req *v1.ProviderListRequest) (res *v1.ProviderList, err error) {
	listOptions, err := o.validateListRequest(req)
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

func (o providerServer) Get(ctx context.Context, req *v1.ProviderGetRequest) (res *v1.Provider, err error) {
	if err := o.validateGetRequest(req); err != nil {
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

func (o providerServer) validateGetRequest(req *v1.ProviderGetRequest) error {
	if req.Name == "" {
		return fmt.Errorf("missing name")
	}
	if req.Account == "" {
		return fmt.Errorf("missing namespace")
	}
	return nil
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
			Metadata: &v1.ProviderMetadata{
				DisplayName: in.Spec.Metadata.DisplayName,
				Description: in.Spec.Metadata.Description,
			},
		},
	}
	if in.Spec.Metadata.Logo != nil {
		out.Spec.Metadata.Logo = convertImage(in.Spec.Metadata.Logo)
	}
	if in.Spec.Metadata.Icon != nil {
		out.Spec.Metadata.Icon = convertImage(in.Spec.Metadata.Icon)
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

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
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	v1 "github.com/kubermatic/kubecarrier/pkg/apiserver/api/v1"
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

func (o providerServer) validateListRequest(req *v1.ProviderListRequest) error {
	if req.Tenant == "" {
		return errors.New("missing tenant")
	}
	if req.Limit < 0 {
		return errors.New("invalid limit: should not be negative number")
	}
	if req.LabelSelector != "" {
		if _, err := labels.Parse(req.LabelSelector); err != nil {
			return errors.New("invalid LabelSelector: unable to parse requirement: found '==', expected: identifier")
		}
	}
	return nil
}

func (o providerServer) getListOptions(req *v1.ProviderListRequest) ([]client.ListOption, error) {
	if err := o.validateListRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	listOptions := []client.ListOption{}
	listOptions = append(listOptions, client.InNamespace(req.Tenant))
	listOptions = append(listOptions, client.Limit(req.Limit))
	if req.LabelSelector != "" {
		labelSelector, err := labels.Parse(req.LabelSelector)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("invalid LabelSelector: %w", err).Error())
		}
		listOptions = append(listOptions, client.MatchingLabelsSelector{Selector: labelSelector})
	}
	return listOptions, nil
}

func (o providerServer) List(ctx context.Context, req *v1.ProviderListRequest) (res *v1.ProviderList, err error) {
	listOptions, err := o.getListOptions(req)
	if err != nil {
		return nil, err
	}
	providerList := &catalogv1alpha1.ProviderList{}
	if err := o.client.List(ctx, providerList, listOptions...); err != nil {
		return nil, fmt.Errorf("listing provider: %w", err)
	}

	res = &v1.ProviderList{}
	for _, catalogProvider := range providerList.Items {
		res.Items = append(res.Items, o.convertProvider(&catalogProvider))

	}
	res.Continue = providerList.Continue
	return
}

func validateProviderRequest(req *v1.ProviderRequest) error {
	if req.Name == "" {
		return errors.New("missing name")
	}
	if req.Tenant == "" {
		return errors.New("missing tenant")
	}
	return nil
}

func (o providerServer) Get(ctx context.Context, req *v1.ProviderRequest) (res *v1.Provider, err error) {
	if err := validateProviderRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	provider := &catalogv1alpha1.Provider{}
	if err = o.client.Get(ctx, types.NamespacedName{
		Name:      req.Name,
		Namespace: req.Tenant,
	}, provider); err != nil {
		return
	}

	return o.convertProvider(provider), nil

}

func (o providerServer) convertProvider(in *catalogv1alpha1.Provider) (out *v1.Provider) {
	return &v1.Provider{
		Name: in.Name,
		Metadata: &v1.AccountMetadata{
			DisplayName: in.Spec.Metadata.DisplayName,
			Description: in.Spec.Metadata.Description,
		},
	}
}

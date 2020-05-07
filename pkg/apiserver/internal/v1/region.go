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
)

type regionServer struct {
	client client.Client
}

var _ v1.RegionServiceServer = (*regionServer)(nil)

// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=regions,verbs=get;list

func NewRegionServiceServer(c client.Client) v1.RegionServiceServer {
	return &regionServer{
		client: c,
	}
}

func (o regionServer) List(ctx context.Context, req *v1.RegionListRequest) (res *v1.RegionList, err error) {
	var listOptions []client.ListOption
	listOptions, err = o.validateListRequest(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	regionList := &catalogv1alpha1.RegionList{}
	if err := o.client.List(ctx, regionList, listOptions...); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("listing regions: %s", err.Error()))
	}

	res = &v1.RegionList{}
	for _, catalogRegion := range regionList.Items {
		res.Items = append(res.Items, o.convertRegion(&catalogRegion))
	}
	res.Continue = regionList.Continue
	return
}

func (o regionServer) Get(ctx context.Context, req *v1.RegionRequest) (res *v1.Region, err error) {
	if err = o.validateGetRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	region := &catalogv1alpha1.Region{}
	if err = o.client.Get(ctx, types.NamespacedName{
		Name:      req.Name,
		Namespace: req.Tenant,
	}, region); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("getting region: %s", err.Error()))
	}
	return o.convertRegion(region), nil

}

func (o regionServer) validateListRequest(req *v1.RegionListRequest) ([]client.ListOption, error) {
	var listOptions []client.ListOption
	if req.Tenant == "" {
		return listOptions, fmt.Errorf("missing namespace")
	}
	listOptions = append(listOptions, client.InNamespace(req.Tenant))
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

func (o regionServer) validateGetRequest(req *v1.RegionRequest) error {
	if req.Name == "" {
		return fmt.Errorf("missing name")
	}
	if req.Tenant == "" {
		return fmt.Errorf("missing namespace")
	}
	return nil
}

func (o regionServer) convertRegion(in *catalogv1alpha1.Region) (out *v1.Region) {
	return &v1.Region{
		Name: in.Name,
		Metadata: &v1.RegionMetadata{
			DisplayName: in.Spec.Metadata.DisplayName,
			Description: in.Spec.Metadata.Description,
		},
		Provider: &v1.ObjectReference{
			Name: in.Spec.Provider.Name,
		},
	}
}

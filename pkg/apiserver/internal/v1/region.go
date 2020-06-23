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
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	v1 "github.com/kubermatic/kubecarrier/pkg/apiserver/api/v1"
)

type regionServer struct {
	client        client.Client
	dynamicClient dynamic.Interface
	scheme        *runtime.Scheme

	gvr schema.GroupVersionResource
}

var _ v1.RegionServiceServer = (*regionServer)(nil)

// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=regions,verbs=get;list

func NewRegionServiceServer(c client.Client, dynamicClient dynamic.Interface, restMapper meta.RESTMapper, scheme *runtime.Scheme) (v1.RegionServiceServer, error) {
	regionServer := &regionServer{
		client:        c,
		dynamicClient: dynamicClient,
		scheme:        scheme,
	}
	objGVK, err := apiutil.GVKForObject(&catalogv1alpha1.Region{}, regionServer.scheme)
	if err != nil {
		return nil, err
	}
	restMapping, err := restMapper.RESTMapping(objGVK.GroupKind(), objGVK.Version)
	if err != nil {
		return nil, err
	}
	regionServer.gvr = restMapping.Resource
	return regionServer, nil
}

func (o regionServer) GetGVR() schema.GroupVersionResource {
	return o.gvr
}

func (o regionServer) List(ctx context.Context, req *v1.ListRequest) (res *v1.RegionList, err error) {
	listOptions, err := req.GetListOptions()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	regionList := &catalogv1alpha1.RegionList{}
	if err := o.client.List(ctx, regionList, listOptions); err != nil {
		return nil, status.Errorf(codes.Internal, "listing regions: %s", err.Error())
	}

	res, err = o.convertRegionList(regionList)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting RegionList: %s", err.Error())
	}
	return
}

func (o regionServer) Get(ctx context.Context, req *v1.GetRequest) (res *v1.Region, err error) {
	region := &catalogv1alpha1.Region{}
	if err = o.client.Get(ctx, types.NamespacedName{
		Name:      req.Name,
		Namespace: req.Account,
	}, region); err != nil {
		return nil, status.Errorf(codes.Internal, "getting region: %s", err.Error())
	}
	res, err = o.convertRegion(region)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting Region: %s", err.Error())
	}
	return
}

func (o regionServer) convertEvent(event runtime.Object) (*any.Any, error) {
	catalogRegion := &catalogv1alpha1.Region{}
	if err := o.scheme.Convert(event, catalogRegion, nil); err != nil {
		return nil, status.Errorf(codes.Internal, "converting event.Object to Region: %s", err.Error())
	}
	region, err := o.convertRegion(catalogRegion)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting Region: %s", err.Error())
	}
	any, err := ptypes.MarshalAny(region)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "marshalling Region to Any: %s", err.Error())
	}
	return any, nil
}

func (o regionServer) Watch(req *v1.WatchRequest, stream v1.RegionService_WatchServer) error {
	listOptions, err := req.GetListOptions()
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return watch(o.dynamicClient, o.gvr, req.Account, *listOptions.AsListOptions(), stream, o.convertEvent)
}

func (o regionServer) convertRegion(in *catalogv1alpha1.Region) (out *v1.Region, err error) {
	metadata, err := convertObjectMeta(in.ObjectMeta)
	if err != nil {
		return nil, err
	}
	out = &v1.Region{
		Metadata: metadata,
		Spec: &v1.RegionSpec{
			Metadata: &v1.RegionMetadata{
				DisplayName: in.Spec.Metadata.DisplayName,
				Description: in.Spec.Metadata.Description,
			},
			Provider: &v1.ObjectReference{
				Name: in.Spec.Provider.Name,
			},
		},
	}
	return
}

func (o regionServer) convertRegionList(in *catalogv1alpha1.RegionList) (out *v1.RegionList, err error) {
	out = &v1.RegionList{
		Metadata: convertListMeta(in.ListMeta),
	}
	for _, inRegion := range in.Items {
		region, err := o.convertRegion(&inRegion)
		if err != nil {
			return nil, err
		}
		out.Items = append(out.Items, region)
	}
	return
}

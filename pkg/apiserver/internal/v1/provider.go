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

type providerServer struct {
	client        client.Client
	dynamicClient dynamic.Interface
	scheme        *runtime.Scheme

	gvr schema.GroupVersionResource
}

var _ v1.ProviderServiceServer = (*providerServer)(nil)

// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=providers,verbs=get;list

func NewProviderServiceServer(c client.Client, dynamicClient dynamic.Interface, restMapper meta.RESTMapper, scheme *runtime.Scheme) (v1.ProviderServiceServer, error) {
	providerServer := &providerServer{
		client:        c,
		dynamicClient: dynamicClient,
		scheme:        scheme,
	}
	objGVK, err := apiutil.GVKForObject(&catalogv1alpha1.Provider{}, providerServer.scheme)
	if err != nil {
		return nil, err
	}
	restMapping, err := restMapper.RESTMapping(objGVK.GroupKind(), objGVK.Version)
	if err != nil {
		return nil, err
	}
	providerServer.gvr = restMapping.Resource
	return providerServer, nil
}

func (o providerServer) GetGVR() schema.GroupVersionResource {
	return o.gvr
}

func (o providerServer) List(ctx context.Context, req *v1.ListRequest) (res *v1.ProviderList, err error) {
	listOptions, err := req.GetListOptions()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	providerList := &catalogv1alpha1.ProviderList{}
	if err := o.client.List(ctx, providerList, listOptions); err != nil {
		return nil, status.Errorf(codes.Internal, "listing providers: %s", err.Error())
	}

	res, err = o.convertProviderList(providerList)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting ProviderList: %s", err.Error())
	}
	return
}

func (o providerServer) Get(ctx context.Context, req *v1.GetRequest) (res *v1.Provider, err error) {
	provider := &catalogv1alpha1.Provider{}
	if err = o.client.Get(ctx, types.NamespacedName{
		Name:      req.Name,
		Namespace: req.Account,
	}, provider); err != nil {
		return nil, status.Errorf(codes.Internal, "getting provider: %s", err.Error())
	}
	res, err = o.convertProvider(provider)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting Provider: %s", err.Error())
	}
	return
}

func (o providerServer) convertEvent(event runtime.Object) (*any.Any, error) {
	catalogProvider := &catalogv1alpha1.Provider{}
	if err := o.scheme.Convert(event, catalogProvider, nil); err != nil {
		return nil, status.Errorf(codes.Internal, "converting event.Object to Provider: %s", err.Error())
	}
	provider, err := o.convertProvider(catalogProvider)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting Provider: %s", err.Error())
	}
	any, err := ptypes.MarshalAny(provider)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "marshalling Provider to Any: %s", err.Error())
	}
	return any, nil
}

func (o providerServer) Watch(req *v1.WatchRequest, stream v1.ProviderService_WatchServer) error {
	listOptions, err := req.GetListOptions()
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return watch(o.dynamicClient, o.gvr, req.Account, *listOptions.AsListOptions(), stream, o.convertEvent)
}

func (o providerServer) convertProvider(in *catalogv1alpha1.Provider) (out *v1.Provider, err error) {
	metadata, err := convertObjectMeta(in.ObjectMeta)
	if err != nil {
		return nil, err
	}
	out = &v1.Provider{
		Metadata: metadata,
		Spec: &v1.ProviderSpec{
			Metadata: &v1.ProviderMetadata{
				DisplayName:      in.Spec.Metadata.DisplayName,
				Description:      in.Spec.Metadata.Description,
				ShortDescription: in.Spec.Metadata.ShortDescription,
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
		Metadata: convertListMeta(in.ListMeta),
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

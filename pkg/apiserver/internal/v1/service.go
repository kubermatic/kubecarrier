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
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/golang/protobuf/ptypes/empty"

	v1 "github.com/kubermatic/kubecarrier/pkg/apiserver/api/v1"
)

type serviceServer struct {
	client client.Client
	mapper meta.RESTMapper
}

var _ v1.ServicesServer = (*serviceServer)(nil)

func NewServicesServer(c client.Client, mapper meta.RESTMapper) v1.ServicesServer {
	return &serviceServer{
		client: c,
		mapper: mapper,
	}
}
func (o serviceServer) Create(ctx context.Context, req *v1.ServiceCreateRequest) (res *v1.Service, err error) {
	obj := &unstructured.Unstructured{}

	gvk, err := o.gvkFromService(o.mapper, req.Service, req.Version)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "creating service: unable to get Kind: %s", err.Error())
	}
	obj.SetGroupVersionKind(gvk)
	val := map[string]interface{}{}
	if err := json.Unmarshal([]byte(req.Spec.Spec), &val); err != nil {
		return nil, status.Error(codes.Internal, "creating service: spec should be type of map[string]intreface{}")
	}
	if err := unstructured.SetNestedMap(obj.Object, val, "spec"); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("creating services: %s", err.Error()))
	}
	// force account from request
	req.Spec.Metadata.Account = req.Account
	if err := SetMetadata(obj, req.Spec.Metadata); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("creating services: %s", err.Error()))
	}
	if err := o.client.Create(ctx, obj); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("creating services: %s", err.Error()))
	}
	res, err = o.convertService(obj)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("converting Service: %s", err.Error()))
	}
	return
}

func (o serviceServer) List(ctx context.Context, req *v1.ServiceListRequest) (res *v1.ServiceList, err error) {
	var listOptions []client.ListOption
	listOptions, err = o.validateListRequest(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	obj := &unstructured.UnstructuredList{}
	gvk, err := o.gvkFromService(o.mapper, req.Service, req.Version)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "listing service: unable to get Kind: %s", err.Error())
	}
	obj.SetGroupVersionKind(gvk)
	if err := o.client.List(ctx, obj, listOptions...); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("listing services: %s", err.Error()))
	}

	res, err = o.convertServiceList(obj)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("converting ServiceList: %s", err.Error()))
	}
	return
}

func (o serviceServer) gvkFromService(mapper meta.RESTMapper, service string, version string) (schema.GroupVersionKind, error) {
	parts := strings.SplitN(service, ".", 2)
	gvr := schema.GroupVersionResource{
		Resource: parts[0],
		Group:    parts[1],
		Version:  version,
	}
	return o.mapper.KindFor(gvr)
}

func (o serviceServer) Get(ctx context.Context, req *v1.ServiceGetRequest) (res *v1.Service, err error) {
	if err = o.validateGetRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	obj := &unstructured.Unstructured{}
	gvk, err := o.gvkFromService(o.mapper, req.Service, req.Version)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "getting service: unable to get Kind: %s", err.Error())
	}
	obj.SetGroupVersionKind(gvk)
	if err = o.client.Get(ctx, types.NamespacedName{
		Name:      req.Name,
		Namespace: req.Account,
	}, obj); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("getting service: %s", err.Error()))
	}
	res, err = o.convertService(obj)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("converting Service: %s", err.Error()))
	}
	return
}

func (o serviceServer) Delete(ctx context.Context, req *v1.ServiceDeleteRequest) (*empty.Empty, error) {
	obj := &unstructured.Unstructured{}
	gvk, err := o.gvkFromService(o.mapper, req.Service, req.Version)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "deleting service: unable to get Kind: %s", err.Error())
	}
	obj.SetGroupVersionKind(gvk)
	obj.SetNamespace(req.Account)
	obj.SetName(req.Name)
	if err := o.client.Delete(ctx, obj); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("delete service: %s", err.Error()))
	}
	return &empty.Empty{}, nil
}

func (o serviceServer) validateListRequest(req *v1.ServiceListRequest) ([]client.ListOption, error) {
	var listOptions []client.ListOption
	if req.Account == "" {
		return listOptions, fmt.Errorf("missing namespace")
	}
	if req.Version == "" {
		return listOptions, fmt.Errorf("missing version")
	}
	if req.Service == "" {
		return listOptions, fmt.Errorf("missing service")
	}
	if len(strings.SplitN(req.Service, ".", 2)) < 2 {
		return listOptions, fmt.Errorf("service should have format: {kind}.{apiGroup}")
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

func (o serviceServer) validateGetRequest(req *v1.ServiceGetRequest) error {
	if req.Name == "" {
		return fmt.Errorf("missing name")
	}
	if req.Service == "" {
		return fmt.Errorf("missing service")
	}
	if len(strings.SplitN(req.Service, ".", 2)) < 2 {
		return fmt.Errorf("service should have format: {kind}.{apiGroup}")
	}
	if req.Version == "" {
		return fmt.Errorf("missing version")
	}
	if req.Account == "" {
		return fmt.Errorf("missing namespace")
	}
	return nil
}

func (o serviceServer) convertService(in *unstructured.Unstructured) (out *v1.Service, err error) {
	metadata, err := FromUnstructured(in)
	if err != nil {
		return nil, err
	}
	out = &v1.Service{Metadata: metadata}
	spec, _, err := unstructured.NestedMap(in.Object, "spec")
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}
	out.Spec = string(data)
	status, ok, err := unstructured.NestedMap(in.Object, "status")
	if err != nil {
		return nil, err
	}
	if ok {
		data, err = json.Marshal(status)
		if err != nil {
			return nil, err
		}
		out.Status = string(data)
	}
	return
}

func (o serviceServer) convertServiceList(in *unstructured.UnstructuredList) (out *v1.ServiceList, err error) {
	out = &v1.ServiceList{
		Metadata: &v1.ListMeta{
			Continue:        in.GetContinue(),
			ResourceVersion: in.GetResourceVersion(),
		},
	}
	for _, inService := range in.Items {
		service, err := o.convertService(&inService)
		if err != nil {
			return nil, err
		}
		out.Items = append(out.Items, service)
	}
	return
}

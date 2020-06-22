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

	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"

	v1 "github.com/kubermatic/kubecarrier/pkg/apiserver/api/v1"
)

type instanceServer struct {
	client        client.Client
	mapper        meta.RESTMapper
	dynamicClient dynamic.Interface
	scheme        *runtime.Scheme
}

var _ v1.InstancesServiceServer = (*instanceServer)(nil)

func NewInstancesServer(c client.Client, dynamicClient dynamic.Interface, mapper meta.RESTMapper, scheme *runtime.Scheme) v1.InstancesServiceServer {
	return &instanceServer{
		client:        c,
		dynamicClient: dynamicClient,
		mapper:        mapper,
		scheme:        scheme,
	}
}
func (o instanceServer) Create(ctx context.Context, req *v1.InstanceCreateRequest) (res *v1.Instance, err error) {
	obj := &unstructured.Unstructured{}

	gvk, err := o.getGVK(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "creating instance: unable to get Kind: %s", err.Error())
	}
	obj.SetGroupVersionKind(gvk)
	val := map[string]interface{}{}
	rawObject, err := v1.NewRawObject(req.Spec.Spec.Encoding, req.Spec.Spec.Data)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "creating instance: spec format: %s", err.Error())
	}
	if err := rawObject.Unmarshal(&val); err != nil {
		return nil, status.Error(codes.Internal, "creating instance: spec should be type of map[string]intreface{}")
	}
	if err := unstructured.SetNestedMap(obj.Object, val, "spec"); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("creating instances: %s", err.Error()))
	}
	// force account from request
	req.Spec.Metadata.Account = req.Account
	if err := SetMetadata(obj, req.Spec.Metadata); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("creating instances: %s", err.Error()))
	}
	if err := o.client.Create(ctx, obj); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("creating instances: %s", err.Error()))
	}
	res, err = o.convertInstance(obj)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("converting Instance: %s", err.Error()))
	}
	res.Offering = req.Offering
	return
}

func (o instanceServer) List(ctx context.Context, req *v1.InstanceListRequest) (res *v1.InstanceList, err error) {
	listOptions, err := req.GetListOptions()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	obj := &unstructured.UnstructuredList{}
	gvk, err := o.getGVK(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "listing instance: unable to get Kind: %s", err.Error())
	}
	obj.SetGroupVersionKind(gvk)
	if err := o.client.List(ctx, obj, listOptions); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("listing instances: %s", err.Error()))
	}

	res, err = o.convertInstanceList(obj, req.Offering)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("converting InstanceList: %s", err.Error()))
	}
	return
}

func (o instanceServer) getGVK(req v1.OfferingVersionGetter) (schema.GroupVersionKind, error) {
	gvr := v1.GetOfferingGVR(req)
	kind, err := o.mapper.KindFor(gvr)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}
	return kind, nil
}

func (o instanceServer) Get(ctx context.Context, req *v1.InstanceGetRequest) (res *v1.Instance, err error) {
	obj := &unstructured.Unstructured{}
	gvk, err := o.getGVK(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "getting instance: unable to get Kind: %s", err.Error())
	}
	obj.SetGroupVersionKind(gvk)
	if err = o.client.Get(ctx, types.NamespacedName{
		Name:      req.Name,
		Namespace: req.Account,
	}, obj); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("getting instance: %s", err.Error()))
	}
	res, err = o.convertInstance(obj)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("converting Instance: %s", err.Error()))
	}
	res.Offering = req.Offering
	return
}

func (o instanceServer) Delete(ctx context.Context, req *v1.InstanceDeleteRequest) (*empty.Empty, error) {
	obj := &unstructured.Unstructured{}
	gvk, err := o.getGVK(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "deleting instance: unable to get Kind: %s", err.Error())
	}
	obj.SetGroupVersionKind(gvk)
	obj.SetNamespace(req.Account)
	obj.SetName(req.Name)
	if err := o.client.Delete(ctx, obj); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("delete instance: %s", err.Error()))
	}
	return &empty.Empty{}, nil
}

func (o instanceServer) convertEvent(event runtime.Object) (*any.Any, error) {
	obj := &unstructured.Unstructured{}
	if err := o.scheme.Convert(event, obj, nil); err != nil {
		return nil, status.Errorf(codes.Internal, "converting event.Object to service instance: %s", err.Error())
	}
	instance, err := o.convertInstance(obj)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting Offering: %s", err.Error())
	}
	any, err := ptypes.MarshalAny(instance)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "marshalling instance to Any: %s", err.Error())
	}
	return any, nil
}

func (o instanceServer) Watch(req *v1.InstanceWatchRequest, stream v1.InstancesService_WatchServer) error {
	listOptions, err := req.GetListOptions()
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	gvr := v1.GetOfferingGVR(req)
	return watch(o.dynamicClient, gvr, req.Account, *listOptions.AsListOptions(), stream, o.convertEvent)
}

func (o instanceServer) convertInstance(in *unstructured.Unstructured) (out *v1.Instance, err error) {
	metadata, err := FromUnstructured(in)
	if err != nil {
		return nil, err
	}
	out = &v1.Instance{Metadata: metadata}
	spec, _, err := unstructured.NestedMap(in.Object, "spec")
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}
	out.Spec = v1.NewJSONRawObject(data)
	status, ok, err := unstructured.NestedMap(in.Object, "status")
	if err != nil {
		return nil, err
	}
	if ok {
		data, err = json.Marshal(status)
		if err != nil {
			return nil, err
		}
		out.Status = v1.NewJSONRawObject(data)
	}
	return
}

func (o instanceServer) convertInstanceList(in *unstructured.UnstructuredList, offering string) (out *v1.InstanceList, err error) {
	out = &v1.InstanceList{
		Metadata: &v1.ListMeta{
			Continue:        in.GetContinue(),
			ResourceVersion: in.GetResourceVersion(),
		},
	}
	for _, inInstance := range in.Items {
		instance, err := o.convertInstance(&inInstance)
		if err != nil {
			return nil, err
		}
		instance.Offering = offering
		out.Items = append(out.Items, instance)
	}
	return
}

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

	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
	if err = o.validateCreateRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	obj := &unstructured.Unstructured{}

	gvk, err := o.gvkFromOffering(req.Offering, req.Version)
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
	var listOptions []client.ListOption
	listOptions, err = o.validateListRequest(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	obj := &unstructured.UnstructuredList{}
	gvk, err := o.gvkFromOffering(req.Offering, req.Version)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "listing instance: unable to get Kind: %s", err.Error())
	}
	obj.SetGroupVersionKind(gvk)
	if err := o.client.List(ctx, obj, listOptions...); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("listing instances: %s", err.Error()))
	}

	res, err = o.convertInstanceList(obj, req.Offering)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("converting InstanceList: %s", err.Error()))
	}
	return
}

func (o instanceServer) Get(ctx context.Context, req *v1.InstanceGetRequest) (res *v1.Instance, err error) {
	if err = o.validateGetRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	obj := &unstructured.Unstructured{}
	gvk, err := o.gvkFromOffering(req.Offering, req.Version)
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
	if err := o.validateDeleteRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	obj := &unstructured.Unstructured{}
	gvk, err := o.gvkFromOffering(req.Offering, req.Version)
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

func (o instanceServer) Watch(req *v1.InstanceWatchRequest, stream v1.InstancesService_WatchServer) error {
	listOptions, err := o.validateWatchRequest(req)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	gvr := o.gvrFromOffering(req.Offering, req.Version)
	watcher, err := o.dynamicClient.Resource(gvr).Namespace(req.Account).Watch(listOptions)
	if err != nil {
		return status.Errorf(codes.Internal, "watching service instances: %s", err.Error())
	}
	defer watcher.Stop()
	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return status.Error(codes.Internal, "watch event channel was closed")
			}
			obj := &unstructured.Unstructured{}
			if err := o.scheme.Convert(event.Object, obj, nil); err != nil {
				return status.Errorf(codes.Internal, "converting event.Object to service instance: %s", err.Error())
			}
			instance, err := o.convertInstance(obj)
			if err != nil {
				return status.Errorf(codes.Internal, "converting instance: %s", err.Error())
			}
			any, err := ptypes.MarshalAny(instance)
			if err != nil {
				return status.Errorf(codes.Internal, "marshalling instance to Any: %s", err.Error())
			}
			err = stream.Send(&v1.WatchEvent{
				Type:   string(event.Type),
				Object: any,
			})
			if grpcStatus, ok := err.(toGRPCStatus); ok {
				return status.Error(grpcStatus.GRPCStatus().Code(), grpcStatus.GRPCStatus().Message())
			} else if err != nil {
				return status.Errorf(codes.Internal, "sending instance stream: %s", err.Error())
			}
		}
	}
}

func (o instanceServer) gvrFromOffering(offering string, version string) schema.GroupVersionResource {
	parts := strings.SplitN(offering, ".", 2)
	gvr := schema.GroupVersionResource{
		Resource: parts[0],
		Group:    parts[1],
		Version:  version,
	}
	return gvr
}

func (o instanceServer) gvkFromOffering(offering string, version string) (schema.GroupVersionKind, error) {
	gvr := o.gvrFromOffering(offering, version)
	kind, err := o.mapper.KindFor(gvr)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}
	return kind, nil
}

func (o instanceServer) validateDeleteRequest(req *v1.InstanceDeleteRequest) error {
	if req.Name == "" {
		return fmt.Errorf("missing name")
	}
	if req.Offering == "" {
		return fmt.Errorf("missing offering")
	}
	if len(strings.SplitN(req.Offering, ".", 2)) < 2 {
		return fmt.Errorf("offering should have format: {kind}.{apiGroup}")
	}
	if req.Version == "" {
		return fmt.Errorf("missing version")
	}
	if req.Account == "" {
		return fmt.Errorf("missing namespace")
	}
	return nil
}

func (o instanceServer) validateListRequest(req *v1.InstanceListRequest) ([]client.ListOption, error) {
	var listOptions []client.ListOption
	if req.Account == "" {
		return listOptions, fmt.Errorf("missing namespace")
	}
	if req.Version == "" {
		return listOptions, fmt.Errorf("missing version")
	}
	if req.Offering == "" {
		return listOptions, fmt.Errorf("missing offering")
	}
	if len(strings.SplitN(req.Offering, ".", 2)) < 2 {
		return listOptions, fmt.Errorf("offering should have format: {kind}.{apiGroup}")
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

func (o instanceServer) validateGetRequest(req *v1.InstanceGetRequest) error {
	if req.Name == "" {
		return fmt.Errorf("missing name")
	}
	if req.Offering == "" {
		return fmt.Errorf("missing offering")
	}
	if len(strings.SplitN(req.Offering, ".", 2)) < 2 {
		return fmt.Errorf("offering should have format: {kind}.{apiGroup}")
	}
	if req.Version == "" {
		return fmt.Errorf("missing version")
	}
	if req.Account == "" {
		return fmt.Errorf("missing namespace")
	}
	return nil
}

func (o instanceServer) validateCreateRequest(req *v1.InstanceCreateRequest) error {
	if req.Offering == "" {
		return fmt.Errorf("missing offering")
	}
	if len(strings.SplitN(req.Offering, ".", 2)) < 2 {
		return fmt.Errorf("offering should have format: {kind}.{apiGroup}")
	}
	if req.Version == "" {
		return fmt.Errorf("missing version")
	}
	if req.Account == "" {
		return fmt.Errorf("missing namespace")
	}
	if req.Spec == nil {
		return fmt.Errorf("missing spec")
	}
	if req.Spec.Metadata == nil && req.Spec.Metadata.Name == "" {
		return fmt.Errorf("missing metadata name")
	}

	return nil
}

func (o instanceServer) validateWatchRequest(req *v1.InstanceWatchRequest) (metav1.ListOptions, error) {
	var listOptions metav1.ListOptions
	if req.Account == "" {
		return listOptions, fmt.Errorf("missing namespace")
	}
	if req.Offering == "" {
		return listOptions, fmt.Errorf("missing offering")
	}
	if len(strings.SplitN(req.Offering, ".", 2)) < 2 {
		return listOptions, fmt.Errorf("offering should have format: {kind}.{apiGroup}")
	}
	if req.Version == "" {
		return listOptions, fmt.Errorf("missing version")
	}
	if req.LabelSelector != "" {
		_, err := labels.Parse(req.LabelSelector)
		if err != nil {
			return listOptions, fmt.Errorf("invalid LabelSelector: %w", err)
		}
		listOptions.LabelSelector = req.LabelSelector
	}
	listOptions.ResourceVersion = req.ResourceVersion
	return listOptions, nil
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

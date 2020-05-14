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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	v1 "github.com/kubermatic/kubecarrier/pkg/apiserver/api/v1"
	"github.com/kubermatic/kubecarrier/pkg/apiserver/internal/util"
)

type offeringServer struct {
	client        client.Client
	dynamicClient dynamic.Interface
	restMapper    meta.RESTMapper
	scheme        *runtime.Scheme

	gvr schema.GroupVersionResource
}

var _ v1.OfferingServiceServer = (*offeringServer)(nil)

// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=offerings,verbs=get;list;watch

func NewOfferingServiceServer(c client.Client, dynamicClient dynamic.Interface, restMapper meta.RESTMapper, scheme *runtime.Scheme) (v1.OfferingServiceServer, error) {
	offeringServer := &offeringServer{
		client:        c,
		dynamicClient: dynamicClient,
		restMapper:    restMapper,
		scheme:        scheme,
	}
	objGVK, err := apiutil.GVKForObject(&catalogv1alpha1.Offering{}, offeringServer.scheme)
	if err != nil {
		return nil, err
	}
	restMapping, err := offeringServer.restMapper.RESTMapping(objGVK.GroupKind(), objGVK.Version)
	if err != nil {
		return nil, err
	}
	offeringServer.gvr = restMapping.Resource
	return offeringServer, nil
}

func (o offeringServer) List(ctx context.Context, req *v1.OfferingListRequest) (res *v1.OfferingList, err error) {
	var listOptions []client.ListOption
	listOptions, err = o.validateListRequest(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	offeringList := &catalogv1alpha1.OfferingList{}
	if err := o.client.List(ctx, offeringList, listOptions...); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("listing offerings: %s", err.Error()))
	}

	res, err = o.convertOfferingList(offeringList)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("converting OfferingList: %s", err.Error()))
	}
	return
}

func (o offeringServer) Get(ctx context.Context, req *v1.OfferingGetRequest) (res *v1.Offering, err error) {
	if err = o.validateGetRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	offering := &catalogv1alpha1.Offering{}
	if err = o.client.Get(ctx, types.NamespacedName{
		Name:      req.Name,
		Namespace: req.Account,
	}, offering); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("getting offering: %s", err.Error()))
	}
	res, err = o.convertOffering(offering)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("converting Offering: %s", err.Error()))
	}
	return
}

func (o offeringServer) Watch(req *v1.OfferingWatchRequest, stream v1.OfferingService_WatchServer) error {
	listOptions, err := o.validateWatchRequest(req)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, err.Error())
	}
	watcher, err := o.dynamicClient.Resource(o.gvr).Namespace(req.Account).Watch(listOptions)
	if err != nil {
		return status.Errorf(codes.Internal, err.Error())
	}
	for {
		select {
		case <-stream.Context().Done():
			return status.Errorf(codes.Internal, "server is down")
		case event := <-watcher.ResultChan():
			catalogOffering := &catalogv1alpha1.Offering{}
			if err := o.scheme.Convert(event.Object, catalogOffering, nil); err != nil {
				return status.Error(codes.Internal, err.Error())
			}
			offering, err := o.convertOffering(catalogOffering)
			if err != nil {
				return status.Error(codes.Internal, err.Error())
			}
			any, err := ptypes.MarshalAny(offering)
			if err != nil {
				return status.Error(codes.Internal, err.Error())
			}
			if req.OperationType == "" ||
				req.OperationType == string(event.Type) {
				if err := stream.Send(&v1.Event{
					Type:   string(event.Type),
					Object: any,
				}); err != nil {
					return status.Errorf(codes.Internal, err.Error())
				}
			}
		}
	}
}

func (o offeringServer) validateListRequest(req *v1.OfferingListRequest) ([]client.ListOption, error) {
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

func (o offeringServer) validateGetRequest(req *v1.OfferingGetRequest) error {
	if req.Name == "" {
		return fmt.Errorf("missing name")
	}
	if req.Account == "" {
		return fmt.Errorf("missing namespace")
	}
	return nil
}

func (o offeringServer) validateWatchRequest(req *v1.OfferingWatchRequest) (metav1.ListOptions, error) {
	var listOptions metav1.ListOptions
	if req.Account == "" {
		return listOptions, fmt.Errorf("missing namespace")
	}
	if err := util.ValidateWatchOperation(req.OperationType); err != nil {
		return listOptions, err
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

func (o offeringServer) convertOffering(in *catalogv1alpha1.Offering) (out *v1.Offering, err error) {
	var versions []*v1.CRDVersion
	for _, catalogCRDVersion := range in.Spec.CRD.Versions {
		schemaBytes, _ := json.Marshal(catalogCRDVersion.Schema)
		versions = append(versions, &v1.CRDVersion{
			Name:   catalogCRDVersion.Name,
			Schema: string(schemaBytes),
		})
	}

	creationTimestamp, err := util.TimestampProto(&in.ObjectMeta.CreationTimestamp)
	if err != nil {
		return nil, err
	}
	deletionTimestamp, err := util.TimestampProto(in.ObjectMeta.DeletionTimestamp)
	if err != nil {
		return nil, err
	}
	out = &v1.Offering{
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
		Spec: &v1.OfferingSpec{
			Metadata: &v1.OfferingMetadata{
				DisplayName: in.Spec.Metadata.DisplayName,
				Description: in.Spec.Metadata.Description,
			},
			Provider: &v1.ObjectReference{
				Name: in.Spec.Provider.Name,
			},
			Crd: &v1.CRDInformation{
				Name:     in.Spec.CRD.Name,
				ApiGroup: in.Spec.CRD.APIGroup,
				Kind:     in.Spec.CRD.Kind,
				Plural:   in.Spec.CRD.Plural,
				Versions: versions,
				Region: &v1.ObjectReference{
					Name: in.Spec.CRD.Region.Name,
				},
			},
		},
	}
	return
}

func (o offeringServer) convertOfferingList(in *catalogv1alpha1.OfferingList) (out *v1.OfferingList, err error) {
	out = &v1.OfferingList{
		Metadata: &v1.ListMeta{
			Continue:        in.Continue,
			ResourceVersion: in.ResourceVersion,
		},
	}
	for _, inOffering := range in.Items {
		offering, err := o.convertOffering(&inOffering)
		if err != nil {
			return nil, err
		}
		out.Items = append(out.Items, offering)
	}
	return
}

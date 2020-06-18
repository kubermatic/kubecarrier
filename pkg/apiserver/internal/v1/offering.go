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

	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	v1 "github.com/kubermatic/kubecarrier/pkg/apiserver/api/v1"
)

type offeringServer struct {
	client        client.Client
	dynamicClient dynamic.Interface
	scheme        *runtime.Scheme

	gvr schema.GroupVersionResource
}

var _ v1.OfferingServiceServer = (*offeringServer)(nil)

// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=offerings,verbs=get;list;watch

func NewOfferingServiceServer(c client.Client, dynamicClient dynamic.Interface, restMapper meta.RESTMapper, scheme *runtime.Scheme) (v1.OfferingServiceServer, error) {
	offeringServer := &offeringServer{
		client:        c,
		dynamicClient: dynamicClient,
		scheme:        scheme,
	}
	objGVK, err := apiutil.GVKForObject(&catalogv1alpha1.Offering{}, offeringServer.scheme)
	if err != nil {
		return nil, err
	}
	restMapping, err := restMapper.RESTMapping(objGVK.GroupKind(), objGVK.Version)
	if err != nil {
		return nil, err
	}
	offeringServer.gvr = restMapping.Resource
	return offeringServer, nil
}

func (o offeringServer) GetGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    catalogv1alpha1.GroupVersion.Group,
		Version:  catalogv1alpha1.GroupVersion.Version,
		Resource: "offerings",
	}
}

func (o offeringServer) List(ctx context.Context, req *v1.ListRequest) (res *v1.OfferingList, err error) {
	listOptions, err := req.GetListOptions()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	offeringList := &catalogv1alpha1.OfferingList{}
	if err := o.client.List(ctx, offeringList, listOptions); err != nil {
		return nil, status.Errorf(codes.Internal, "listing offerings: %s", err.Error())
	}

	res, err = o.convertOfferingList(offeringList)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting OfferingList: %s", err.Error())
	}
	return
}

func (o offeringServer) Get(ctx context.Context, req *v1.GetRequest) (res *v1.Offering, err error) {
	offering := &catalogv1alpha1.Offering{}
	if err = o.client.Get(ctx, types.NamespacedName{
		Name:      req.Name,
		Namespace: req.Account,
	}, offering); err != nil {
		return nil, status.Errorf(codes.Internal, "getting offering: %s", err.Error())
	}
	res, err = o.convertOffering(offering)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting Offering: %s", err.Error())
	}
	return
}

func (o offeringServer) Watch(req *v1.WatchRequest, stream v1.OfferingService_WatchServer) error {
	listOptions, err := req.GetListOptions()
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	watcher, err := o.dynamicClient.Resource(o.gvr).Namespace(req.Account).Watch(*listOptions.AsListOptions())
	if err != nil {
		return status.Errorf(codes.Internal, "watching offerings: %s", err.Error())
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
			catalogOffering := &catalogv1alpha1.Offering{}
			if err := o.scheme.Convert(event.Object, catalogOffering, nil); err != nil {
				return status.Errorf(codes.Internal, "converting event.Object to Offering: %s", err.Error())
			}
			offering, err := o.convertOffering(catalogOffering)
			if err != nil {
				return status.Errorf(codes.Internal, "converting Offering: %s", err.Error())
			}
			any, err := ptypes.MarshalAny(offering)
			if err != nil {
				return status.Errorf(codes.Internal, "marshalling Offering to Any: %s", err.Error())
			}
			err = stream.Send(&v1.WatchEvent{
				Type:   string(event.Type),
				Object: any,
			})
			if grpcStatus, ok := err.(toGRPCStatus); ok {
				return status.Error(grpcStatus.GRPCStatus().Code(), grpcStatus.GRPCStatus().Message())
			} else if err != nil {
				return status.Errorf(codes.Internal, "sending Offering stream: %s", err.Error())
			}
		}
	}
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
	metadata, err := convertObjectMeta(in.ObjectMeta)
	if err != nil {
		return nil, err
	}

	out = &v1.Offering{
		Metadata: metadata,
		Spec: &v1.OfferingSpec{
			Metadata: &v1.OfferingMetadata{
				DisplayName:      in.Spec.Metadata.DisplayName,
				Description:      in.Spec.Metadata.Description,
				ShortDescription: in.Spec.Metadata.ShortDescription,
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
	if in.Spec.Metadata.Logo != nil {
		out.Spec.Metadata.Logo = convertImage(in.Spec.Metadata.Logo)
	}
	if in.Spec.Metadata.Icon != nil {
		out.Spec.Metadata.Icon = convertImage(in.Spec.Metadata.Icon)
	}
	return
}

func (o offeringServer) convertOfferingList(in *catalogv1alpha1.OfferingList) (out *v1.OfferingList, err error) {
	out = &v1.OfferingList{
		Metadata: convertListMeta(in.ListMeta),
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

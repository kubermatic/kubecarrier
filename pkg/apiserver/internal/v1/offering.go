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

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	v1 "github.com/kubermatic/kubecarrier/pkg/apiserver/api/v1"
	"github.com/kubermatic/kubecarrier/pkg/apiserver/internal/util"
)

type offeringServer struct {
	client client.Client
}

var _ v1.OfferingServiceServer = (*offeringServer)(nil)

// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=offerings,verbs=get;list

func NewOfferingServiceServer(c client.Client) v1.OfferingServiceServer {
	return &offeringServer{
		client: c,
	}
}

func (o offeringServer) List(ctx context.Context, req *v1.ListRequest) (res *v1.OfferingList, err error) {
	var listOptions []client.ListOption
	listOptions, err = util.ValidateListRequest(req)
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

func (o offeringServer) Get(ctx context.Context, req *v1.GetRequest) (res *v1.Offering, err error) {
	if err = util.ValidateGetRequest(req); err != nil {
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

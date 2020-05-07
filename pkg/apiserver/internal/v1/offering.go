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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	v1 "github.com/kubermatic/kubecarrier/pkg/apiserver/api/v1"
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

func (o offeringServer) List(ctx context.Context, req *v1.OfferingListRequest) (res *v1.OfferingList, err error) {
	var listOptions []client.ListOption
	listOptions, err = o.validateListRequest(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	offeringList := &catalogv1alpha1.OfferingList{}
	if err := o.client.List(ctx, offeringList, listOptions...); err != nil {
		return nil, err
	}

	res = &v1.OfferingList{}
	for _, catalogOffering := range offeringList.Items {
		res.Items = append(res.Items, o.convertOffering(&catalogOffering))
	}
	res.Continue = offeringList.Continue
	return
}

func (o offeringServer) Get(ctx context.Context, req *v1.OfferingRequest) (res *v1.Offering, err error) {
	if err = o.validateGetRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	offering := &catalogv1alpha1.Offering{}
	if err = o.client.Get(ctx, types.NamespacedName{
		Name:      req.Name,
		Namespace: req.Tenant,
	}, offering); err != nil {
		return nil, err
	}
	return o.convertOffering(offering), nil

}

func (o offeringServer) validateListRequest(req *v1.OfferingListRequest) ([]client.ListOption, error) {
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

func (o offeringServer) validateGetRequest(req *v1.OfferingRequest) error {
	if req.Name == "" {
		return fmt.Errorf("missing name")
	}
	if req.Tenant == "" {
		return fmt.Errorf("missing namespace")
	}
	return nil
}

func (o offeringServer) convertOffering(in *catalogv1alpha1.Offering) (out *v1.Offering) {
	var versions []*v1.CRDVersion
	for _, catalogCRDVersion := range in.Spec.CRD.Versions {
		schemaBytes, _ := json.Marshal(catalogCRDVersion.Schema)
		versions = append(versions, &v1.CRDVersion{
			Name:   catalogCRDVersion.Name,
			Schema: string(schemaBytes),
		})
	}
	return &v1.Offering{
		Name: in.Name,
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
	}
}

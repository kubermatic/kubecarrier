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

package services

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubermatic/kubecarrier/pkg/api/v1alpha1"
	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
)

type offeringServer struct {
	client client.Client
}

var _ v1alpha1.OfferingServiceServer = (*offeringServer)(nil)

// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=offerings,verbs=get;list

func NewOfferingServiceServer(c client.Client) v1alpha1.OfferingServiceServer {
	return &offeringServer{
		client: c,
	}
}

func (o offeringServer) List(ctx context.Context, req *v1alpha1.OfferingListRequest) (res *v1alpha1.OfferingList, err error) {
	var selector labels.Selector
	offeringList := &catalogv1alpha1.OfferingList{}
	if req.Selector != "" {
		selector, err = labels.Parse(req.Selector)
		if err != nil {
			return nil, fmt.Errorf("invalid selector: %w", err)
		}
		if err := o.client.List(ctx, offeringList, client.MatchingLabelsSelector{
			Selector: selector,
		}, client.InNamespace(req.Namespace)); err != nil {
			return nil, fmt.Errorf("listing offering: %w", err)
		}
	} else {
		if err := o.client.List(ctx, offeringList, client.InNamespace(req.Namespace)); err != nil {
			return nil, fmt.Errorf("listing offering: %w", err)
		}
	}

	res = &v1alpha1.OfferingList{}
	for _, catalogOffering := range offeringList.Items {
		res.Items = append(res.Items, o.convertOffering(&catalogOffering))

	}
	return
}

func (o offeringServer) Get(ctx context.Context, req *v1alpha1.Offering) (res *v1alpha1.Offering, err error) {
	offering := &catalogv1alpha1.Offering{}
	if err = o.client.Get(ctx, types.NamespacedName{
		Name:      req.Name,
		Namespace: req.Namespace,
	}, offering); err != nil {
		return
	}

	return o.convertOffering(offering), nil

}

func (o offeringServer) convertOffering(in *catalogv1alpha1.Offering) (out *v1alpha1.Offering) {
	var versions []*v1alpha1.CRDVersion
	for _, catalogCRDVersion := range in.Spec.CRD.Versions {
		versions = append(versions, &v1alpha1.CRDVersion{
			Name:   catalogCRDVersion.Name,
			Schema: catalogCRDVersion.Schema.String(),
		})
	}
	return &v1alpha1.Offering{
		Name: in.Name,
		Metadata: &v1alpha1.OfferingMetadata{
			DisplayName: in.Spec.Metadata.DisplayName,
			Description: in.Spec.Metadata.Description,
		},
		Provider: &v1alpha1.ObjectReference{
			Name: in.Spec.Provider.Name,
		},
		Crd: &v1alpha1.CRDInformation{
			Name:     in.Spec.CRD.Name,
			ApiGroup: in.Spec.CRD.APIGroup,
			Kind:     in.Spec.CRD.Kind,
			Plural:   in.Spec.CRD.Plural,
			Versions: versions,
			Region: &v1alpha1.ObjectReference{
				Name: in.Spec.CRD.Name,
			},
		},
	}
}

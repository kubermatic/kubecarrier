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
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	v1 "github.com/kubermatic/kubecarrier/pkg/apiserver/api/v1"
	"github.com/kubermatic/kubecarrier/pkg/apiserver/internal/util"
)

type accountServer struct {
	client client.Client
}

var _ v1.AccountServiceServer = (*accountServer)(nil)

// +kubebuilder:rbac:groups=catalog.kubecarrier.io,resources=accounts,verbs=get;list

func NewAccountServiceServer(c client.Client) v1.AccountServiceServer {
	return &accountServer{
		client: c,
	}
}

func (o accountServer) List(ctx context.Context, req *v1.AccountListRequest) (res *v1.AccountList, err error) {
	var listOptions []client.ListOption
	listOptions, err = o.validateListRequest(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	accountList := &catalogv1alpha1.AccountList{}
	if err := o.client.List(ctx, accountList, listOptions...); err != nil {
		return nil, status.Errorf(codes.Internal, "listing accounts: %s", err.Error())
	}

	res, err = o.convertAccountList(accountList)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting AccountList: %s", err.Error())
	}
	return
}

func (o accountServer) Get(ctx context.Context, req *v1.AccountGetRequest) (res *v1.Account, err error) {
	if err = o.validateGetRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	account := &catalogv1alpha1.Account{}
	if err = o.client.Get(ctx, types.NamespacedName{
		Name: req.Name,
	}, account); err != nil {
		return nil, status.Errorf(codes.Internal, "getting account: %s", err.Error())
	}
	res, err = o.convertAccount(account)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting Account: %s", err.Error())
	}
	return
}

func (o accountServer) validateListRequest(req *v1.AccountListRequest) ([]client.ListOption, error) {
	var listOptions []client.ListOption
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

func (o accountServer) validateGetRequest(req *v1.AccountGetRequest) error {
	if req.Name == "" {
		return fmt.Errorf("missing name")
	}
	return nil
}

func (o accountServer) convertAccount(in *catalogv1alpha1.Account) (out *v1.Account, err error) {
	creationTimestamp, err := util.TimestampProto(&in.ObjectMeta.CreationTimestamp)
	if err != nil {
		return nil, err
	}
	deletionTimestamp, err := util.TimestampProto(in.ObjectMeta.DeletionTimestamp)
	if err != nil {
		return nil, err
	}
	out = &v1.Account{
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
		Spec: &v1.AccountSpec{
			Metadata: &v1.AccountMetadata{
				DisplayName:      in.Spec.Metadata.DisplayName,
				Description:      in.Spec.Metadata.Description,
				ShortDescription: in.Spec.Metadata.ShortDescription,
			},
		},
		Status: &v1.AccountStatus{
			ObservedGeneration: in.Status.ObservedGeneration,
			Phase:              string(in.Status.Phase),
		},
	}
	if in.Spec.Metadata.Logo != nil {
		out.Spec.Metadata.Logo = convertImage(in.Spec.Metadata.Logo)
	}
	if in.Spec.Metadata.Icon != nil {
		out.Spec.Metadata.Icon = convertImage(in.Spec.Metadata.Icon)
	}
	for _, accountRole := range in.Spec.Roles {
		out.Spec.Roles = append(out.Spec.Roles, string(accountRole))
	}
	for _, subject := range in.Spec.Subjects {
		out.Spec.Subjects = append(out.Spec.Subjects, &v1.Subject{
			Kind:      subject.Kind,
			ApiGroup:  subject.APIGroup,
			Name:      subject.Name,
			Namespace: subject.Namespace,
		})
	}
	if in.Status.Namespace != nil {
		out.Status.Namespace = &v1.ObjectReference{
			Name: in.Status.Namespace.Name,
		}
	}
	for _, condition := range in.Status.Conditions {
		lastTransitionTime, err := util.TimestampProto(&condition.LastTransitionTime)
		if err != nil {
			return nil, err
		}
		out.Status.Conditions = append(out.Status.Conditions, &v1.AccountCondition{
			Type:               string(condition.Type),
			Status:             string(condition.Status),
			LastTransitionTime: lastTransitionTime,
			Reason:             condition.Reason,
			Message:            condition.Message,
		})
	}
	return
}

func (o accountServer) convertAccountList(in *catalogv1alpha1.AccountList) (out *v1.AccountList, err error) {
	out = &v1.AccountList{
		Metadata: &v1.ListMeta{
			Continue:        in.Continue,
			ResourceVersion: in.ResourceVersion,
		},
	}
	for _, inAccount := range in.Items {
		account, err := o.convertAccount(&inAccount)
		if err != nil {
			return nil, err
		}
		out.Items = append(out.Items, account)
	}
	return
}

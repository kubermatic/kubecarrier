/*
Copyright 2019 The KubeCarrier Authors.

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

package authorizer

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/gobuffalo/flect"
	"google.golang.org/grpc/metadata"
	authv1 "k8s.io/api/authorization/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

const (
	userHeaderKey = "authorization"
)

type AuthorizationClient struct {
	Scheme *runtime.Scheme
	Log    logr.Logger
	client.Client
}

func (a AuthorizationClient) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
	decision, err := a.authorize(ctx, obj, GetAuthorization{
		Name:      key.Name,
		Namespace: key.Namespace,
	})
	if err != nil {
		return fmt.Errorf("authorizing request: %w", err)
	}
	if decision != DecisionAllowed {
		return fmt.Errorf("not allowed: %w", err)
	}
	return a.Client.Get(ctx, key, obj)
}

func (a AuthorizationClient) List(ctx context.Context, list runtime.Object, opts ...client.ListOption) error {
	var authOpt ListAuthorization
	for _, opt := range opts {
		if namespace, isInNamespace := opt.(client.InNamespace); isInNamespace {
			authOpt.Namespace = string(namespace)
		}
	}
	decision, err := a.authorize(ctx, list, authOpt)
	if err != nil {
		return fmt.Errorf("authorizing request: %w", err)
	}
	if decision != DecisionAllowed {
		return fmt.Errorf("not allowed: %w", err)
	}
	return a.Client.List(ctx, list, opts...)
}

type AuthorizationDecision string

const (
	DecisionAllowed   AuthorizationDecision = "allowed"
	DecisionDenied    AuthorizationDecision = "denied"
	DecisionNoOpinion AuthorizationDecision = "no opinion"
)

// +kubebuilder:rbac:groups=authorization.k8s.io,resources=subjectaccessreviews,verbs=create

func (a AuthorizationClient) authorize(
	ctx context.Context,
	objType runtime.Object,
	option AuthorizationOption,
) (AuthorizationDecision, error) {
	userInfo, err := a.getUserInfo(ctx)
	if err != nil {
		return DecisionDenied, fmt.Errorf("cannot get the user infomation from the request: %w", err)
	}

	gvk, err := apiutil.GVKForObject(objType, a.Scheme)
	if err != nil {
		return DecisionDenied, fmt.Errorf("cannot get GVK for %T: %w", objType, err)
	}
	var resource string
	if meta.IsListType(objType) {
		resource = flect.Pluralize(strings.ToLower(strings.TrimSuffix(gvk.Kind, "List")))

	} else {
		resource = flect.Pluralize(strings.ToLower(gvk.Kind))
	}

	review := &authv1.SubjectAccessReview{
		Spec: authv1.SubjectAccessReviewSpec{
			ResourceAttributes: &authv1.ResourceAttributes{
				Group:    gvk.Group,
				Version:  gvk.Version,
				Resource: resource,
			},
			User: userInfo.Name,
		},
	}
	option.apply(review)
	if err := a.Create(ctx, review); err != nil {
		return DecisionDenied, fmt.Errorf("creating SubjectAccessReview: %s", err)
	}
	switch {
	case review.Status.Denied:
		return DecisionDenied, nil
	case review.Status.Allowed:
		return DecisionAllowed, nil
	default:
		return DecisionNoOpinion, nil
	}
}

func (a AuthorizationClient) getUserInfo(ctx context.Context) (UserInfo, error) {
	var userInfo UserInfo
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return userInfo, fmt.Errorf("request doesn't contain user info")
	}
	userNames := md.Get(userHeaderKey)
	if len(userNames) != 1 {
		return userInfo, fmt.Errorf("user header is invalid in request")
	}
	userInfo.Name = userNames[0]
	return userInfo, nil
}

type RequestOperation string

const (
	RequestGet  RequestOperation = "get"
	RequestList RequestOperation = "list"
)

type AuthorizationOption interface {
	apply(review *authv1.SubjectAccessReview)
}

type GetAuthorization struct {
	Namespace string
	Name      string
}

func (a GetAuthorization) apply(review *authv1.SubjectAccessReview) {
	review.Spec.ResourceAttributes.Verb = string(RequestGet)
	review.Spec.ResourceAttributes.Namespace = a.Namespace
	review.Spec.ResourceAttributes.Name = a.Name
}

type ListAuthorization struct {
	Namespace string
}

func (a ListAuthorization) apply(review *authv1.SubjectAccessReview) {
	review.Spec.ResourceAttributes.Verb = string(RequestList)
	review.Spec.ResourceAttributes.Namespace = a.Namespace
}

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

	"github.com/go-logr/logr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	authv1 "k8s.io/api/authorization/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubermatic/kubecarrier/pkg/apiserver/auth"
)

type Authorizer struct {
	scheme     *runtime.Scheme
	log        logr.Logger
	client     client.Client
	restMapper meta.RESTMapper
}

func NewAuthorizer(log logr.Logger, scheme *runtime.Scheme, client client.Client, restMapper meta.RESTMapper) Authorizer {
	return Authorizer{
		scheme:     scheme,
		log:        log,
		client:     client,
		restMapper: restMapper,
	}
}

type authRequest interface {
	GetAuthOption() AuthorizationOption
	GetGVR(server interface{}) schema.GroupVersionResource
}

func (a Authorizer) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if authReq, ok := req.(authRequest); ok {
			opts := authReq.GetAuthOption()
			if gvr := authReq.GetGVR(info.Server); !gvr.Empty() {
				if err := a.Authorize(ctx, gvr, opts); err != nil {
					return nil, status.Error(codes.Unauthenticated, err.Error())
				}
			}
		}
		return handler(ctx, req)
	}
}

func (a Authorizer) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrapper := &recvWrapper{stream, a, srv}
		return handler(srv, wrapper)
	}
}

type recvWrapper struct {
	grpc.ServerStream
	a   Authorizer
	srv interface{}
}

func (s *recvWrapper) RecvMsg(m interface{}) error {
	if err := s.ServerStream.RecvMsg(m); err != nil {
		return err
	}
	if authReq, ok := m.(authRequest); ok {
		opts := authReq.GetAuthOption()
		gvr := authReq.GetGVR(s.srv)
		if err := s.a.Authorize(s.Context(), gvr, opts); err != nil {
			return status.Error(codes.Unauthenticated, err.Error())
		}
	}
	return nil
}

// +kubebuilder:rbac:groups=authorization.k8s.io,resources=subjectaccessreviews,verbs=create

func (a Authorizer) Authorize(
	ctx context.Context,
	gvr schema.GroupVersionResource,
	option AuthorizationOption,
) error {
	user, err := auth.ExtractUserInfo(ctx)
	if err != nil {
		return err
	}
	review := &authv1.SubjectAccessReview{
		Spec: authv1.SubjectAccessReviewSpec{
			ResourceAttributes: &authv1.ResourceAttributes{
				Group:    gvr.Group,
				Version:  gvr.Version,
				Resource: gvr.Resource,
			},
			User:   user.GetName(),
			Groups: user.GetGroups(),
			UID:    user.GetUID(),
			Extra:  make(map[string]authv1.ExtraValue),
		},
	}
	for key, val := range user.GetExtra() {
		review.Spec.Extra[key] = val
	}
	option.apply(review)
	if err := a.client.Create(ctx, review); err != nil {
		return fmt.Errorf("creating SubjectAccessReview: %s", err)
	}
	if !review.Status.Allowed {
		return fmt.Errorf("permission denied")
	}
	return nil
}

type RequestOperation string

const (
	RequestGet    RequestOperation = "get"
	RequestList   RequestOperation = "list"
	RequestWatch  RequestOperation = "watch"
	RequestCreate RequestOperation = "create"
	RequestDelete RequestOperation = "delete"
)

type AuthorizationOption struct {
	Name      string
	Namespace string
	Verb      RequestOperation
}

func (a AuthorizationOption) apply(review *authv1.SubjectAccessReview) {
	review.Spec.ResourceAttributes.Name = a.Name
	review.Spec.ResourceAttributes.Namespace = a.Namespace
	review.Spec.ResourceAttributes.Verb = string(a.Verb)
}

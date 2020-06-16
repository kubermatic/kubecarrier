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
	authv1 "k8s.io/api/authorization/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

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

// +kubebuilder:rbac:groups=authorization.k8s.io,resources=subjectaccessreviews,verbs=create

func (a Authorizer) Authorize(
	ctx context.Context,
	objType runtime.Object,
	option AuthorizationOption,
) error {
	user, err := auth.ExtractUserInfo(ctx)
	if err != nil {
		return err
	}
	gvk, err := apiutil.GVKForObject(objType, a.scheme)
	if err != nil {
		return fmt.Errorf("cannot get GVK for %T: %w", objType, err)
	}
	restMapping, err := a.restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return fmt.Errorf("cannot get resources for %T: %w", objType, err)
	}

	review := &authv1.SubjectAccessReview{
		Spec: authv1.SubjectAccessReviewSpec{
			ResourceAttributes: &authv1.ResourceAttributes{
				Group:    restMapping.Resource.Group,
				Version:  restMapping.Resource.Version,
				Resource: restMapping.Resource.Resource,
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

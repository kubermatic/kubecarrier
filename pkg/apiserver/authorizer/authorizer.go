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
	authv1 "k8s.io/api/authorization/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

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

func (a AuthorizationOption) Apply(review *authv1.SubjectAccessReview) {
	review.Spec.ResourceAttributes.Name = a.Name
	review.Spec.ResourceAttributes.Namespace = a.Namespace
	review.Spec.ResourceAttributes.Verb = string(a.Verb)
}

type AuthRequest interface {
	GetAuthOption() AuthorizationOption
	GetGVR(server interface{}) schema.GroupVersionResource
}

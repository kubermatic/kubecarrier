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
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	v1 "github.com/kubermatic/kubecarrier/pkg/apiserver/api/v1"
)

var (
	serviceGVK = schema.GroupVersionKind{
		Group:   "eu-west-1.team-a",
		Version: "v1alpha1",
		Kind:    "CouchDB",
	}
	serviceListGVK = schema.GroupVersionKind{
		Group:   "eu-west-1.team-a",
		Version: "v1alpha1",
		Kind:    "CouchDBList",
	}
)

type fakeRESTMapper struct {
	kind string
	meta.RESTMapper
}

func newFakeRESTMapper(kind string) meta.RESTMapper {
	return &fakeRESTMapper{kind: kind}
}
func (frm *fakeRESTMapper) KindFor(resource schema.GroupVersionResource) (schema.GroupVersionKind, error) {
	return schema.GroupVersionKind{
		Group:   resource.Group,
		Version: resource.Version,
		Kind:    frm.kind,
	}, nil
}

func newService(name, namespace string, spec, status, labels map[string]string) (*unstructured.Unstructured, error) {
	service := &unstructured.Unstructured{}
	metadata := &v1.ObjectMeta{
		Name:    name,
		Account: namespace,
		Labels:  labels,
	}
	service.SetUnstructuredContent(map[string]interface{}{})
	if err := unstructured.SetNestedStringMap(service.Object, spec, "spec"); err != nil {
		return nil, err
	}
	if err := unstructured.SetNestedStringMap(service.Object, status, "status"); err != nil {
		return nil, err
	}
	service.SetLabels(labels)
	service.SetGroupVersionKind(serviceGVK)
	service.SetName(name)
	service.SetNamespace(namespace)
	return service, SetMetadata(service, metadata)
}

func TestGetService(t *testing.T) {
	service, err := newService("test-service", "test-namespace",
		map[string]string{"username": "username", "password": "password"},
		map[string]string{"status": "ready"}, map[string]string{})
	assert.Nil(t, err)

	client := fakeclient.NewFakeClientWithScheme(testScheme, service)
	serviceServer := NewServicesServer(client, newFakeRESTMapper("CouchDB"))
	ctx := context.Background()
	tests := []struct {
		name           string
		req            *v1.ServiceGetRequest
		expectedError  error
		expectedResult *v1.Service
	}{
		{
			name: "missing namespace",
			req: &v1.ServiceGetRequest{
				Name:    "test-service",
				Account: "",
				Service: "couchdb.eu-west-1.team-a",
				Version: "v1alpha1",
			},
			expectedError:  status.Errorf(codes.InvalidArgument, "missing namespace"),
			expectedResult: nil,
		},
		{
			name: "missing name",
			req: &v1.ServiceGetRequest{
				Account: "test-namespace",
				Service: "couchdb.eu-west-1.team-a",
				Version: "v1alpha1",
			},
			expectedError:  status.Errorf(codes.InvalidArgument, "missing name"),
			expectedResult: nil,
		},
		{
			name: "missing service",
			req: &v1.ServiceGetRequest{
				Name:    "test-service",
				Account: "test-namespace",
				Version: "v1alpha1",
			},
			expectedError:  status.Errorf(codes.InvalidArgument, "missing service"),
			expectedResult: nil,
		},
		{
			name: "missing version",
			req: &v1.ServiceGetRequest{
				Name:    "test-service",
				Account: "test-namespace",
				Service: "couchdb.eu-west-1.team-a",
			},
			expectedError:  status.Errorf(codes.InvalidArgument, "missing version"),
			expectedResult: nil,
		},
		{
			name: "wrong service name",
			req: &v1.ServiceGetRequest{
				Name:    "test-service",
				Account: "test-namespace",
				Service: "couchdb",
				Version: "v1alpha1",
			},
			expectedError:  status.Errorf(codes.InvalidArgument, "service should have format: {kind}.{apiGroup}"),
			expectedResult: nil,
		},
		{
			name: "valid request",
			req: &v1.ServiceGetRequest{
				Name:    "test-service",
				Account: "test-namespace",
				Service: "couchdb.eu-west-1.team-a",
				Version: "v1alpha1",
			},
			expectedError: nil,
			expectedResult: &v1.Service{
				Metadata: &v1.ObjectMeta{
					Name:    "test-service",
					Account: "test-namespace",
					Labels:  map[string]string{},
				},
				Spec:   "{\"password\":\"password\",\"username\":\"username\"}",
				Status: "{\"status\":\"ready\"}",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service, err := serviceServer.Get(ctx, test.req)
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedResult, service)
		})
	}
}

func TestCreateService(t *testing.T) {
	spec := "{\"password\":\"password\",\"username\":\"username\"}"
	client := fakeclient.NewFakeClientWithScheme(testScheme)
	serviceServer := NewServicesServer(client, newFakeRESTMapper("CouchDB"))
	ctx := context.Background()
	tests := []struct {
		name           string
		req            *v1.ServiceCreateRequest
		expectedError  error
		expectedResult *v1.Service
	}{
		{
			name: "valid request",
			req: &v1.ServiceCreateRequest{
				Account: "test-namespace",
				Service: "couchdb.eu-west-1.team-a",
				Version: "v1alpha1",
				Spec: &v1.Service{
					Metadata: &v1.ObjectMeta{Name: "test-service"},
					Spec:     spec,
				},
			},
			expectedError: nil,
			expectedResult: &v1.Service{
				Metadata: &v1.ObjectMeta{
					Name:            "test-service",
					Account:         "test-namespace",
					ResourceVersion: "1",
				},
				Spec: spec,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service, err := serviceServer.Create(ctx, test.req)
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedResult, service)
		})
	}
}

func TestDeleteService(t *testing.T) {
	service, err := newService("test-service", "test-namespace",
		map[string]string{"username": "username", "password": "password"},
		map[string]string{"status": "ready"}, map[string]string{})
	assert.Nil(t, err)
	client := fakeclient.NewFakeClientWithScheme(testScheme)
	serviceServer := NewServicesServer(client, newFakeRESTMapper("CouchDB"))
	ctx := context.Background()
	err = client.Create(ctx, service)
	assert.Nil(t, err)
	tests := []struct {
		name           string
		req            *v1.ServiceDeleteRequest
		expectedError  error
		expectedResult *empty.Empty
	}{
		{
			name: "valid request",
			req: &v1.ServiceDeleteRequest{
				Name:    "test-service",
				Account: "test-namespace",
				Service: "couchdb.eu-west-1.team-a",
				Version: "v1alpha1",
			},
			expectedError:  nil,
			expectedResult: &empty.Empty{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			empty, err := serviceServer.Delete(ctx, test.req)
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedResult, empty)
		})
	}
}

func TestListService(t *testing.T) {
	spec := map[string]string{"username": "username", "password": "password"}
	st := map[string]string{"status": "ready"}
	labels := map[string]string{
		"test-label": "service1",
	}
	services := &unstructured.UnstructuredList{}
	services.SetGroupVersionKind(serviceListGVK)
	service, err := newService("test-service-1", "test-namespace", spec, st, labels)
	assert.Nil(t, err)
	services.Items = append(services.Items, *service)
	labels = map[string]string{
		"test-label": "service2",
	}
	service, err = newService("test-service-2", "test-namespace", spec, st, labels)
	assert.Nil(t, err)
	services.Items = append(services.Items, *service)

	client := fakeclient.NewFakeClientWithScheme(testScheme, services)
	testScheme.AddKnownTypeWithName(serviceListGVK, services)
	serviceServer := NewServicesServer(client, newFakeRESTMapper("CouchDBList"))
	ctx := context.Background()
	// err = client.Create(ctx, service)
	// assert.Nil(t, err)
	tests := []struct {
		name           string
		req            *v1.ServiceListRequest
		expectedError  error
		expectedResult *v1.ServiceList
	}{
		{
			name: "valid request",
			req: &v1.ServiceListRequest{
				Account: "",
				Service: "couchdb.eu-west-1.team-a",
				Version: "v1alpha1",
			},
			expectedError:  status.Errorf(codes.InvalidArgument, "missing namespace"),
			expectedResult: nil,
		},
		{
			name: "invalid limit",
			req: &v1.ServiceListRequest{
				Account: "test-namespace",
				Service: "couchdb.eu-west-1.team-a",
				Version: "v1alpha1",
				Limit:   -1,
			},
			expectedError:  status.Errorf(codes.InvalidArgument, "invalid limit: should not be negative number"),
			expectedResult: nil,
		},
		{
			name: "invalid label selector",
			req: &v1.ServiceListRequest{
				Account:       "test-namespace",
				Service:       "couchdb.eu-west-1.team-a",
				Version:       "v1alpha1",
				LabelSelector: "test-label=====service1",
			},
			expectedError:  status.Errorf(codes.InvalidArgument, "invalid LabelSelector: unable to parse requirement: found '==', expected: identifier"),
			expectedResult: nil,
		},
		{
			name: "valid request",
			req: &v1.ServiceListRequest{
				Account: "test-namespace",
				Service: "couchdb.eu-west-1.team-a",
				Version: "v1alpha1",
			},
			expectedError: nil,
			expectedResult: &v1.ServiceList{
				Metadata: &v1.ListMeta{
					Continue:        "",
					ResourceVersion: "",
				},
				Items: []*v1.Service{
					{
						Metadata: &v1.ObjectMeta{
							Name:    "test-service-1",
							Account: "test-namespace",
							Labels: map[string]string{
								"test-label": "service1",
							},
						},
						Spec:   "{\"password\":\"password\",\"username\":\"username\"}",
						Status: "{\"status\":\"ready\"}",
					},
					{
						Metadata: &v1.ObjectMeta{
							Name:    "test-service-2",
							Account: "test-namespace",
							Labels: map[string]string{
								"test-label": "service2",
							},
						},
						Spec:   "{\"password\":\"password\",\"username\":\"username\"}",
						Status: "{\"status\":\"ready\"}",
					},
				},
			},
		},
		{
			name: "LabelSelector works",
			req: &v1.ServiceListRequest{
				Account:       "test-namespace",
				Service:       "couchdb.eu-west-1.team-a",
				Version:       "v1alpha1",
				LabelSelector: "test-label=service1",
			},
			expectedError: nil,
			expectedResult: &v1.ServiceList{
				Metadata: &v1.ListMeta{
					Continue:        "",
					ResourceVersion: "",
				},
				Items: []*v1.Service{
					{
						Metadata: &v1.ObjectMeta{
							Name:    "test-service-1",
							Account: "test-namespace",
							Labels: map[string]string{
								"test-label": "service1",
							},
						},
						Spec:   "{\"password\":\"password\",\"username\":\"username\"}",
						Status: "{\"status\":\"ready\"}",
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			services, err := serviceServer.List(ctx, test.req)
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedResult, services)
		})
	}
}

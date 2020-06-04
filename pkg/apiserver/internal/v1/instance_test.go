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
	instanceGVK = schema.GroupVersionKind{
		Group:   "eu-west-1.team-a",
		Version: "v1alpha1",
		Kind:    "CouchDB",
	}
	instanceListGVK = schema.GroupVersionKind{
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

func newInstance(name, namespace string, spec, status, labels map[string]string) (*unstructured.Unstructured, error) {
	instance := &unstructured.Unstructured{}
	metadata := &v1.ObjectMeta{
		Name:    name,
		Account: namespace,
		Labels:  labels,
	}
	instance.SetUnstructuredContent(map[string]interface{}{})
	if err := unstructured.SetNestedStringMap(instance.Object, spec, "spec"); err != nil {
		return nil, err
	}
	if err := unstructured.SetNestedStringMap(instance.Object, status, "status"); err != nil {
		return nil, err
	}
	instance.SetLabels(labels)
	instance.SetGroupVersionKind(instanceGVK)
	instance.SetName(name)
	instance.SetNamespace(namespace)
	return instance, SetMetadata(instance, metadata)
}

func TestGetInstance(t *testing.T) {
	instance, err := newInstance("test-instance", "test-namespace",
		map[string]string{"username": "username", "password": "password"},
		map[string]string{"status": "ready"}, map[string]string{})
	assert.Nil(t, err)

	client := fakeclient.NewFakeClientWithScheme(testScheme, instance)
	instanceServer := NewInstancesServer(client, nil, newFakeRESTMapper("CouchDB"), testScheme)
	ctx := context.Background()
	tests := []struct {
		name           string
		req            *v1.InstanceGetRequest
		expectedError  error
		expectedResult *v1.Instance
	}{
		{
			name: "missing namespace",
			req: &v1.InstanceGetRequest{
				Name:     "test-instance",
				Account:  "",
				Offering: "couchdb.eu-west-1.team-a",
				Version:  "v1alpha1",
			},
			expectedError:  status.Errorf(codes.InvalidArgument, "missing namespace"),
			expectedResult: nil,
		},
		{
			name: "missing name",
			req: &v1.InstanceGetRequest{
				Account:  "test-namespace",
				Offering: "couchdb.eu-west-1.team-a",
				Version:  "v1alpha1",
			},
			expectedError:  status.Errorf(codes.InvalidArgument, "missing name"),
			expectedResult: nil,
		},
		{
			name: "missing offering name",
			req: &v1.InstanceGetRequest{
				Name:    "test-instance",
				Account: "test-namespace",
				Version: "v1alpha1",
			},
			expectedError:  status.Errorf(codes.InvalidArgument, "missing offering"),
			expectedResult: nil,
		},
		{
			name: "missing version",
			req: &v1.InstanceGetRequest{
				Name:     "test-instance",
				Account:  "test-namespace",
				Offering: "couchdb.eu-west-1.team-a",
			},
			expectedError:  status.Errorf(codes.InvalidArgument, "missing version"),
			expectedResult: nil,
		},
		{
			name: "wrong offering name",
			req: &v1.InstanceGetRequest{
				Name:     "test-instance",
				Account:  "test-namespace",
				Offering: "couchdb",
				Version:  "v1alpha1",
			},
			expectedError:  status.Errorf(codes.InvalidArgument, "offering should have format: {kind}.{apiGroup}"),
			expectedResult: nil,
		},
		{
			name: "valid request",
			req: &v1.InstanceGetRequest{
				Name:     "test-instance",
				Account:  "test-namespace",
				Offering: "couchdb.eu-west-1.team-a",
				Version:  "v1alpha1",
			},
			expectedError: nil,
			expectedResult: &v1.Instance{
				Metadata: &v1.ObjectMeta{
					Name:    "test-instance",
					Account: "test-namespace",
					Labels:  map[string]string{},
				},
				Offering: "couchdb.eu-west-1.team-a",
				Spec:     v1.NewJSONRawObject([]byte("{\"password\":\"password\",\"username\":\"username\"}")),
				Status:   v1.NewJSONRawObject([]byte("{\"status\":\"ready\"}")),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			instance, err := instanceServer.Get(ctx, test.req)
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedResult, instance)
		})
	}
}

func TestCreateInstance(t *testing.T) {
	spec := v1.NewJSONRawObject([]byte("{\"password\":\"password\",\"username\":\"username\"}"))
	client := fakeclient.NewFakeClientWithScheme(testScheme)
	instanceServer := NewInstancesServer(client, nil, newFakeRESTMapper("CouchDB"), testScheme)
	ctx := context.Background()
	tests := []struct {
		name           string
		req            *v1.InstanceCreateRequest
		expectedError  error
		expectedResult *v1.Instance
	}{
		{
			name: "valid request",
			req: &v1.InstanceCreateRequest{
				Account:  "test-namespace",
				Offering: "couchdb.eu-west-1.team-a",
				Version:  "v1alpha1",
				Spec: &v1.Instance{
					Metadata: &v1.ObjectMeta{Name: "test-instance"},
					Spec:     spec,
				},
			},
			expectedError: nil,
			expectedResult: &v1.Instance{
				Metadata: &v1.ObjectMeta{
					Name:            "test-instance",
					Account:         "test-namespace",
					ResourceVersion: "1",
				},
				Spec:     spec,
				Offering: "couchdb.eu-west-1.team-a",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			instance, err := instanceServer.Create(ctx, test.req)
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedResult, instance)
		})
	}
}

func TestDeleteInstance(t *testing.T) {
	instance, err := newInstance("test-instance", "test-namespace",
		map[string]string{"username": "username", "password": "password"},
		map[string]string{"status": "ready"}, map[string]string{})
	assert.Nil(t, err)
	client := fakeclient.NewFakeClientWithScheme(testScheme)
	instanceServer := NewInstancesServer(client, nil, newFakeRESTMapper("CouchDB"), testScheme)
	ctx := context.Background()
	err = client.Create(ctx, instance)
	assert.Nil(t, err)
	tests := []struct {
		name           string
		req            *v1.InstanceDeleteRequest
		expectedError  error
		expectedResult *empty.Empty
	}{
		{
			name: "valid request",
			req: &v1.InstanceDeleteRequest{
				Name:     "test-instance",
				Account:  "test-namespace",
				Offering: "couchdb.eu-west-1.team-a",
				Version:  "v1alpha1",
			},
			expectedError:  nil,
			expectedResult: &empty.Empty{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			empty, err := instanceServer.Delete(ctx, test.req)
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedResult, empty)
		})
	}
}

func TestListInstance(t *testing.T) {
	spec := map[string]string{"username": "username", "password": "password"}
	st := map[string]string{"status": "ready"}
	labels := map[string]string{
		"test-label": "instance1",
	}
	instances := &unstructured.UnstructuredList{}
	instances.SetGroupVersionKind(instanceListGVK)
	instance, err := newInstance("test-instance-1", "test-namespace", spec, st, labels)
	assert.Nil(t, err)
	instances.Items = append(instances.Items, *instance)
	labels = map[string]string{
		"test-label": "instance2",
	}
	instance, err = newInstance("test-instance-2", "test-namespace", spec, st, labels)
	assert.Nil(t, err)
	instances.Items = append(instances.Items, *instance)

	client := fakeclient.NewFakeClientWithScheme(testScheme, instances)
	testScheme.AddKnownTypeWithName(instanceListGVK, instances)
	instanceServer := NewInstancesServer(client, nil, newFakeRESTMapper("CouchDBList"), testScheme)
	ctx := context.Background()
	tests := []struct {
		name           string
		req            *v1.InstanceListRequest
		expectedError  error
		expectedResult *v1.InstanceList
	}{
		{
			name: "valid request",
			req: &v1.InstanceListRequest{
				Account:  "",
				Offering: "couchdb.eu-west-1.team-a",
				Version:  "v1alpha1",
			},
			expectedError:  status.Errorf(codes.InvalidArgument, "missing namespace"),
			expectedResult: nil,
		},
		{
			name: "invalid limit",
			req: &v1.InstanceListRequest{
				Account:  "test-namespace",
				Offering: "couchdb.eu-west-1.team-a",
				Version:  "v1alpha1",
				Limit:    -1,
			},
			expectedError:  status.Errorf(codes.InvalidArgument, "invalid limit: should not be negative number"),
			expectedResult: nil,
		},
		{
			name: "invalid label selector",
			req: &v1.InstanceListRequest{
				Account:       "test-namespace",
				Offering:      "couchdb.eu-west-1.team-a",
				Version:       "v1alpha1",
				LabelSelector: "test-label=====instance1",
			},
			expectedError:  status.Errorf(codes.InvalidArgument, "invalid LabelSelector: unable to parse requirement: found '==', expected: identifier"),
			expectedResult: nil,
		},
		{
			name: "valid request",
			req: &v1.InstanceListRequest{
				Account:  "test-namespace",
				Offering: "couchdb.eu-west-1.team-a",
				Version:  "v1alpha1",
			},
			expectedError: nil,
			expectedResult: &v1.InstanceList{
				Metadata: &v1.ListMeta{
					Continue:        "",
					ResourceVersion: "",
				},
				Items: []*v1.Instance{
					{
						Metadata: &v1.ObjectMeta{
							Name:    "test-instance-1",
							Account: "test-namespace",
							Labels: map[string]string{
								"test-label": "instance1",
							},
						},
						Spec:     v1.NewJSONRawObject([]byte("{\"password\":\"password\",\"username\":\"username\"}")),
						Offering: "couchdb.eu-west-1.team-a",
						Status:   v1.NewJSONRawObject([]byte("{\"status\":\"ready\"}")),
					},
					{
						Metadata: &v1.ObjectMeta{
							Name:    "test-instance-2",
							Account: "test-namespace",
							Labels: map[string]string{
								"test-label": "instance2",
							},
						},
						Spec:     v1.NewJSONRawObject([]byte("{\"password\":\"password\",\"username\":\"username\"}")),
						Offering: "couchdb.eu-west-1.team-a",
						Status:   v1.NewJSONRawObject([]byte("{\"status\":\"ready\"}")),
					},
				},
			},
		},
		{
			name: "LabelSelector works",
			req: &v1.InstanceListRequest{
				Account:       "test-namespace",
				Offering:      "couchdb.eu-west-1.team-a",
				Version:       "v1alpha1",
				LabelSelector: "test-label=instance1",
			},
			expectedError: nil,
			expectedResult: &v1.InstanceList{
				Metadata: &v1.ListMeta{
					Continue:        "",
					ResourceVersion: "",
				},
				Items: []*v1.Instance{
					{
						Metadata: &v1.ObjectMeta{
							Name:    "test-instance-1",
							Account: "test-namespace",
							Labels: map[string]string{
								"test-label": "instance1",
							},
						},
						Spec:     v1.NewJSONRawObject([]byte("{\"password\":\"password\",\"username\":\"username\"}")),
						Status:   v1.NewJSONRawObject([]byte("{\"status\":\"ready\"}")),
						Offering: "couchdb.eu-west-1.team-a",
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			instances, err := instanceServer.List(ctx, test.req)
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedResult, instances)
		})
	}
}

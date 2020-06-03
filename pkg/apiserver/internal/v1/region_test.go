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

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	v1 "github.com/kubermatic/kubecarrier/pkg/apiserver/api/v1"
)

func TestListRegion(t *testing.T) {
	regions := &catalogv1alpha1.RegionList{
		Items: []catalogv1alpha1.Region{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-region-1",
					Namespace: "test-namespace",
					Labels: map[string]string{
						"test-label": "region1",
					},
				},
				Spec: catalogv1alpha1.RegionSpec{
					Metadata: corev1alpha1.ServiceClusterMetadata{
						Description: "Test Region",
						DisplayName: "Test Region",
					},
					Provider: catalogv1alpha1.ObjectReference{
						Name: "test-provider",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-region-2",
					Namespace: "test-namespace",
					Labels: map[string]string{
						"test-label": "region2",
					},
				},
				Spec: catalogv1alpha1.RegionSpec{
					Metadata: corev1alpha1.ServiceClusterMetadata{
						Description: "Test Region",
						DisplayName: "Test Region",
					},
					Provider: catalogv1alpha1.ObjectReference{
						Name: "test-provider",
					},
				},
			},
		},
	}
	client := fakeclient.NewFakeClientWithScheme(testScheme, regions)
	regionServer := regionServer{
		client: client,
	}
	ctx := context.Background()
	tests := []struct {
		name           string
		req            *v1.ListRequest
		expectedError  error
		expectedResult *v1.RegionList
	}{
		{
			name: "missing namespace",
			req: &v1.ListRequest{
				Account: "",
			},
			expectedError:  status.Errorf(codes.InvalidArgument, "missing namespace"),
			expectedResult: nil,
		},
		{
			name: "invalid limit",
			req: &v1.ListRequest{
				Account: "test-namespace",
				Limit:   -1,
			},
			expectedError:  status.Errorf(codes.InvalidArgument, "invalid limit: should not be negative number"),
			expectedResult: nil,
		},
		{
			name: "invalid label selector",
			req: &v1.ListRequest{
				Account:       "test-namespace",
				LabelSelector: "test-label=====region1",
			},
			expectedError:  status.Errorf(codes.InvalidArgument, "invalid LabelSelector: unable to parse requirement: found '==', expected: identifier"),
			expectedResult: nil,
		},
		{
			name: "valid request",
			req: &v1.ListRequest{
				Account: "test-namespace",
			},
			expectedError: nil,
			expectedResult: &v1.RegionList{
				Metadata: &v1.ListMeta{
					Continue:        "",
					ResourceVersion: "",
				},
				Items: []*v1.Region{
					{
						Metadata: &v1.ObjectMeta{
							Name:    "test-region-1",
							Account: "test-namespace",
							Labels: map[string]string{
								"test-label": "region1",
							},
						},
						Spec: &v1.RegionSpec{
							Metadata: &v1.RegionMetadata{
								Description: "Test Region",
								DisplayName: "Test Region",
							},
							Provider: &v1.ObjectReference{
								Name: "test-provider",
							},
						},
					},
					{
						Metadata: &v1.ObjectMeta{
							Name:    "test-region-2",
							Account: "test-namespace",
							Labels: map[string]string{
								"test-label": "region2",
							},
						},
						Spec: &v1.RegionSpec{
							Metadata: &v1.RegionMetadata{
								Description: "Test Region",
								DisplayName: "Test Region",
							},
							Provider: &v1.ObjectReference{
								Name: "test-provider",
							},
						},
					},
				},
			},
		},
		{
			name: "LabelSelector works",
			req: &v1.ListRequest{
				Account:       "test-namespace",
				LabelSelector: "test-label=region1",
			},
			expectedError: nil,
			expectedResult: &v1.RegionList{
				Metadata: &v1.ListMeta{
					Continue:        "",
					ResourceVersion: "",
				},
				Items: []*v1.Region{
					{
						Metadata: &v1.ObjectMeta{
							Name:    "test-region-1",
							Account: "test-namespace",
							Labels: map[string]string{
								"test-label": "region1",
							},
						},
						Spec: &v1.RegionSpec{
							Metadata: &v1.RegionMetadata{
								Description: "Test Region",
								DisplayName: "Test Region",
							},
							Provider: &v1.ObjectReference{
								Name: "test-provider",
							},
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			regions, err := regionServer.handleListRequest(ctx, test.req)
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedResult, regions)
		})
	}
}

func TestGetRegion(t *testing.T) {
	region := &catalogv1alpha1.Region{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-region",
			Namespace: "test-namespace",
		},
		Spec: catalogv1alpha1.RegionSpec{
			Metadata: corev1alpha1.ServiceClusterMetadata{
				Description: "Test Region",
				DisplayName: "Test Region",
			},
			Provider: catalogv1alpha1.ObjectReference{
				Name: "test-provider",
			},
		},
	}
	client := fakeclient.NewFakeClientWithScheme(testScheme, region)
	regionServer := regionServer{
		client: client,
	}
	ctx := context.Background()
	tests := []struct {
		name           string
		req            *v1.GetRequest
		expectedError  error
		expectedResult *v1.Region
	}{
		{
			name: "missing namespace",
			req: &v1.GetRequest{
				Name:    "test-region",
				Account: "",
			},
			expectedError:  status.Errorf(codes.InvalidArgument, "missing namespace"),
			expectedResult: nil,
		},
		{
			name: "missing name",
			req: &v1.GetRequest{
				Account: "test-namespace",
			},
			expectedError:  status.Errorf(codes.InvalidArgument, "missing name"),
			expectedResult: nil,
		},
		{
			name: "valid request",
			req: &v1.GetRequest{
				Name:    "test-region",
				Account: "test-namespace",
			},
			expectedError: nil,
			expectedResult: &v1.Region{
				Metadata: &v1.ObjectMeta{
					Name:    "test-region",
					Account: "test-namespace",
				},
				Spec: &v1.RegionSpec{
					Metadata: &v1.RegionMetadata{
						Description: "Test Region",
						DisplayName: "Test Region",
					},
					Provider: &v1.ObjectReference{
						Name: "test-provider",
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			region, err := regionServer.handleGetRequest(ctx, test.req)
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedResult, region)
		})
	}
}

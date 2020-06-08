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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	v1 "github.com/kubermatic/kubecarrier/pkg/apiserver/api/v1"
)

func TestListOffering(t *testing.T) {
	offerings := &catalogv1alpha1.OfferingList{
		Items: []catalogv1alpha1.Offering{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-offering-1",
					Namespace: "test-namespace",
					Labels: map[string]string{
						"test-label": "offering1",
					},
				},
				Spec: catalogv1alpha1.OfferingSpec{
					Metadata: catalogv1alpha1.OfferingMetadata{
						CommonMetadata: catalogv1alpha1.CommonMetadata{
							Description: "Test Offering",
							DisplayName: "Test Offering",
						},
					},
					Provider: catalogv1alpha1.ObjectReference{
						Name: "test-provider",
					},
					CRD: catalogv1alpha1.CRDInformation{
						Name:     "test-crd",
						APIGroup: "test-crd-group",
						Kind:     "test-kind",
						Plural:   "test-plural",
						Versions: []catalogv1alpha1.CRDVersion{
							{
								Name: "test-version",
								Schema: &apiextensionsv1.CustomResourceValidation{
									OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
										Properties: map[string]apiextensionsv1.JSONSchemaProps{
											"apiVersion": {Type: "string"},
										},
										Type: "object",
									},
								},
							},
						},
						Region: catalogv1alpha1.ObjectReference{
							Name: "test-region",
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-offering-2",
					Namespace: "test-namespace",
					Labels: map[string]string{
						"test-label": "offering2",
					},
				},
				Spec: catalogv1alpha1.OfferingSpec{
					Metadata: catalogv1alpha1.OfferingMetadata{
						CommonMetadata: catalogv1alpha1.CommonMetadata{
							Description: "Test Offering",
							DisplayName: "Test Offering",
						},
					},
					Provider: catalogv1alpha1.ObjectReference{
						Name: "test-provider",
					},
					CRD: catalogv1alpha1.CRDInformation{
						Name:     "test-crd",
						APIGroup: "test-crd-group",
						Kind:     "test-kind",
						Plural:   "test-plural",
						Versions: []catalogv1alpha1.CRDVersion{
							{
								Name: "test-version",
								Schema: &apiextensionsv1.CustomResourceValidation{
									OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
										Properties: map[string]apiextensionsv1.JSONSchemaProps{
											"apiVersion": {Type: "string"},
										},
										Type: "object",
									},
								},
							},
						},
						Region: catalogv1alpha1.ObjectReference{
							Name: "test-region",
						},
					},
				},
			},
		},
	}
	client := fakeclient.NewFakeClientWithScheme(testScheme, offerings)
	offeringServer := offeringServer{
		client: client,
	}
	ctx := context.Background()
	tests := []struct {
		name           string
		req            *v1.ListRequest
		expectedError  error
		expectedResult *v1.OfferingList
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
				LabelSelector: "test-label=====offering1",
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
			expectedResult: &v1.OfferingList{
				Metadata: &v1.ListMeta{
					Continue:        "",
					ResourceVersion: "",
				},
				Items: []*v1.Offering{
					{
						Metadata: &v1.ObjectMeta{
							Name:    "test-offering-1",
							Account: "test-namespace",
							Labels: map[string]string{
								"test-label": "offering1",
							},
						},
						Spec: &v1.OfferingSpec{
							Metadata: &v1.OfferingMetadata{
								Description: "Test Offering",
								DisplayName: "Test Offering",
							},
							Provider: &v1.ObjectReference{
								Name: "test-provider",
							},
							Crd: &v1.CRDInformation{
								Name:     "test-crd",
								ApiGroup: "test-crd-group",
								Kind:     "test-kind",
								Plural:   "test-plural",
								Versions: []*v1.CRDVersion{
									{
										Name:   "test-version",
										Schema: `{"openAPIV3Schema":{"type":"object","properties":{"apiVersion":{"type":"string"}}}}`,
									},
								},
								Region: &v1.ObjectReference{
									Name: "test-region",
								},
							},
						},
					},
					{
						Metadata: &v1.ObjectMeta{
							Name:    "test-offering-2",
							Account: "test-namespace",
							Labels: map[string]string{
								"test-label": "offering2",
							},
						},
						Spec: &v1.OfferingSpec{
							Metadata: &v1.OfferingMetadata{
								Description: "Test Offering",
								DisplayName: "Test Offering",
							},
							Provider: &v1.ObjectReference{
								Name: "test-provider",
							},
							Crd: &v1.CRDInformation{
								Name:     "test-crd",
								ApiGroup: "test-crd-group",
								Kind:     "test-kind",
								Plural:   "test-plural",
								Versions: []*v1.CRDVersion{
									{
										Name:   "test-version",
										Schema: `{"openAPIV3Schema":{"type":"object","properties":{"apiVersion":{"type":"string"}}}}`,
									},
								},
								Region: &v1.ObjectReference{
									Name: "test-region",
								},
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
				LabelSelector: "test-label=offering1",
			},
			expectedError: nil,
			expectedResult: &v1.OfferingList{
				Metadata: &v1.ListMeta{
					Continue:        "",
					ResourceVersion: "",
				},
				Items: []*v1.Offering{
					{
						Metadata: &v1.ObjectMeta{
							Name:    "test-offering-1",
							Account: "test-namespace",
							Labels: map[string]string{
								"test-label": "offering1",
							},
						},
						Spec: &v1.OfferingSpec{
							Metadata: &v1.OfferingMetadata{
								Description: "Test Offering",
								DisplayName: "Test Offering",
							},
							Provider: &v1.ObjectReference{
								Name: "test-provider",
							},
							Crd: &v1.CRDInformation{
								Name:     "test-crd",
								ApiGroup: "test-crd-group",
								Kind:     "test-kind",
								Plural:   "test-plural",
								Versions: []*v1.CRDVersion{
									{
										Name:   "test-version",
										Schema: `{"openAPIV3Schema":{"type":"object","properties":{"apiVersion":{"type":"string"}}}}`,
									},
								},
								Region: &v1.ObjectReference{
									Name: "test-region",
								},
							},
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			offerings, err := offeringServer.List(ctx, test.req)
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedResult, offerings)
		})
	}
}

func TestGetOffering(t *testing.T) {
	offering := &catalogv1alpha1.Offering{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-offering",
			Namespace: "test-namespace",
		},
		Spec: catalogv1alpha1.OfferingSpec{
			Metadata: catalogv1alpha1.OfferingMetadata{
				CommonMetadata: catalogv1alpha1.CommonMetadata{
					Description: "Test Offering",
					DisplayName: "Test Offering",
				},
			},
			Provider: catalogv1alpha1.ObjectReference{
				Name: "test-provider",
			},
			CRD: catalogv1alpha1.CRDInformation{
				Name:     "test-crd",
				APIGroup: "test-crd-group",
				Kind:     "test-kind",
				Plural:   "test-plural",
				Versions: []catalogv1alpha1.CRDVersion{
					{
						Name: "test-version",
						Schema: &apiextensionsv1.CustomResourceValidation{
							OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
								Properties: map[string]apiextensionsv1.JSONSchemaProps{
									"apiVersion": {Type: "string"},
								},
								Type: "object",
							},
						},
					},
				},
				Region: catalogv1alpha1.ObjectReference{
					Name: "test-region",
				},
			},
		},
	}
	client := fakeclient.NewFakeClientWithScheme(testScheme, offering)
	offeringServer := offeringServer{
		client: client,
	}
	ctx := context.Background()
	tests := []struct {
		name           string
		req            *v1.GetRequest
		expectedError  error
		expectedResult *v1.Offering
	}{
		{
			name: "valid request",
			req: &v1.GetRequest{
				Name:    "test-offering",
				Account: "test-namespace",
			},
			expectedError: nil,
			expectedResult: &v1.Offering{
				Metadata: &v1.ObjectMeta{
					Name:    "test-offering",
					Account: "test-namespace",
				},
				Spec: &v1.OfferingSpec{
					Metadata: &v1.OfferingMetadata{
						Description: "Test Offering",
						DisplayName: "Test Offering",
					},
					Provider: &v1.ObjectReference{
						Name: "test-provider",
					},
					Crd: &v1.CRDInformation{
						Name:     "test-crd",
						ApiGroup: "test-crd-group",
						Kind:     "test-kind",
						Plural:   "test-plural",
						Versions: []*v1.CRDVersion{
							{
								Name:   "test-version",
								Schema: `{"openAPIV3Schema":{"type":"object","properties":{"apiVersion":{"type":"string"}}}}`,
							},
						},
						Region: &v1.ObjectReference{
							Name: "test-region",
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			offering, err := offeringServer.Get(ctx, test.req)
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedResult, offering)
		})
	}
}

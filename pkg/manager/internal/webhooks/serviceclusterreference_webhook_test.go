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

package webhooks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func TestServiceClusterReferenceValidatingCreate(t *testing.T) {
	serviceClusterReferenceWebhookHandler := ServiceClusterReferenceWebhookHandler{
		Log: testutil.NewLogger(t),
	}

	tests := []struct {
		name          string
		object        *catalogv1alpha1.ServiceClusterReference
		expectedError bool
	}{
		{
			name: "invalid serviceClusterReference name",
			object: &catalogv1alpha1.ServiceClusterReference{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "-test-serviceClusterReference",
					Namespace: "test-namespace",
				},
				Spec: catalogv1alpha1.ServiceClusterReferenceSpec{
					Metadata: corev1alpha1.ServiceClusterMetadata{
						Description: "Test ServiceClusterReference",
						DisplayName: "Test ServiceClusterReference",
					},
					Provider: catalogv1alpha1.ObjectReference{
						Name: "Provider",
					},
				},
			},
			expectedError: true,
		},
		{
			name: "serviceClusterReference data missing",
			object: &catalogv1alpha1.ServiceClusterReference{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-serviceclusterreference",
					Namespace: "test-namespace",
				},
			},
			expectedError: true,
		},
		{
			name: "metadata missing",
			object: &catalogv1alpha1.ServiceClusterReference{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-serviceclusterreference",
					Namespace: "test-namespace",
				},
				Spec: catalogv1alpha1.ServiceClusterReferenceSpec{
					Provider: catalogv1alpha1.ObjectReference{
						Name: "Provider",
					},
				},
			},
			expectedError: true,
		},
		{
			name: "metadata description missing",
			object: &catalogv1alpha1.ServiceClusterReference{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-serviceclusterreference",
					Namespace: "test-namespace",
				},
				Spec: catalogv1alpha1.ServiceClusterReferenceSpec{
					Metadata: corev1alpha1.ServiceClusterMetadata{
						DisplayName: "Test ServiceClusterReference",
					},
					Provider: catalogv1alpha1.ObjectReference{
						Name: "Provider",
					},
				},
			},
			expectedError: true,
		},
		{
			name: "metadata displayName missing",
			object: &catalogv1alpha1.ServiceClusterReference{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-serviceclusterreference",
					Namespace: "test-namespace",
				},
				Spec: catalogv1alpha1.ServiceClusterReferenceSpec{
					Metadata: corev1alpha1.ServiceClusterMetadata{
						DisplayName: "Test ServiceClusterReference",
					},
					Provider: catalogv1alpha1.ObjectReference{
						Name: "Provider",
					},
				},
			},
			expectedError: true,
		},
		{
			name: "provider missing",
			object: &catalogv1alpha1.ServiceClusterReference{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-serviceclusterreference",
					Namespace: "test-namespace",
				},
				Spec: catalogv1alpha1.ServiceClusterReferenceSpec{
					Metadata: corev1alpha1.ServiceClusterMetadata{
						Description: "Test ServiceClusterReference",
						DisplayName: "Test ServiceClusterReference",
					},
				},
			},
			expectedError: true,
		},
		{
			name: "can pass validating create",
			object: &catalogv1alpha1.ServiceClusterReference{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-serviceclusterreference",
					Namespace: "test-namespace",
				},
				Spec: catalogv1alpha1.ServiceClusterReferenceSpec{
					Metadata: corev1alpha1.ServiceClusterMetadata{
						Description: "Test ServiceClusterReference",
						DisplayName: "Test ServiceClusterReference",
					},
					Provider: catalogv1alpha1.ObjectReference{
						Name: "Provider",
					},
				},
			},
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedError, serviceClusterReferenceWebhookHandler.validateCreate(test.object) != nil)
		})
	}
}

func TestServiceClusterReferenceValidatingUpdate(t *testing.T) {
	serviceClusterReferenceWebhookHandler := ServiceClusterReferenceWebhookHandler{
		Log: testutil.NewLogger(t),
	}

	oldObj := &catalogv1alpha1.ServiceClusterReference{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-serviceclusterreference",
			Namespace: "test-namespace",
		},
		Spec: catalogv1alpha1.ServiceClusterReferenceSpec{
			Metadata: corev1alpha1.ServiceClusterMetadata{
				Description: "Test ServiceClusterReference",
				DisplayName: "Test ServiceClusterReference",
			},
			Provider: catalogv1alpha1.ObjectReference{
				Name: "Provider",
			},
		},
	}

	tests := []struct {
		name          string
		object        *catalogv1alpha1.ServiceClusterReference
		expectedError bool
	}{
		{
			name: "provider immutable",
			object: &catalogv1alpha1.ServiceClusterReference{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-serviceclusterreference",
					Namespace: "test-namespace",
				},
				Spec: catalogv1alpha1.ServiceClusterReferenceSpec{
					Metadata: corev1alpha1.ServiceClusterMetadata{
						Description: "Test ServiceClusterReference",
						DisplayName: "Test ServiceClusterReference",
					},
					Provider: catalogv1alpha1.ObjectReference{
						Name: "Provider2",
					},
				},
			},
			expectedError: true,
		},
		{
			name: "can pass validating update",
			object: &catalogv1alpha1.ServiceClusterReference{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-serviceclusterreference",
					Namespace: "test-namespace",
				},
				Spec: catalogv1alpha1.ServiceClusterReferenceSpec{
					Metadata: corev1alpha1.ServiceClusterMetadata{
						Description: "Test ServiceClusterReference",
						DisplayName: "Test ServiceClusterReference new displayName",
					},
					Provider: catalogv1alpha1.ObjectReference{
						Name: "Provider",
					},
				},
			},
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedError, serviceClusterReferenceWebhookHandler.validateUpdate(test.object, oldObj) != nil)
		})
	}
}

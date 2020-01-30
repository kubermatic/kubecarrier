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
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func TestProviderReferenceValidatingMetadata(t *testing.T) {
	providerReferenceWebhookHandler := ProviderReferenceWebhookHandler{
		Log: testutil.NewLogger(t),
	}

	tests := []struct {
		name          string
		object        *catalogv1alpha1.ProviderReference
		expectedError bool
	}{
		{
			name: "no metadata",
			object: &catalogv1alpha1.ProviderReference{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-providerReference",
					Namespace: "test-providerReference-namespace",
				},
			},
			expectedError: true,
		},
		{
			name: "no description",
			object: &catalogv1alpha1.ProviderReference{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-providerReference",
					Namespace: "test-providerReference-namespace",
				},
				Spec: catalogv1alpha1.ProviderReferenceSpec{
					Metadata: catalogv1alpha1.ProviderMetadata{
						DisplayName: "test ProviderReference",
					},
				},
			},
			expectedError: true,
		},
		{
			name: "no displayName",
			object: &catalogv1alpha1.ProviderReference{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-providerReference",
					Namespace: "test-providerReference-namespace",
				},
				Spec: catalogv1alpha1.ProviderReferenceSpec{
					Metadata: catalogv1alpha1.ProviderMetadata{
						Description: "test ProviderReference",
					},
				},
			},
			expectedError: true,
		},
		{
			name: "metadata",
			object: &catalogv1alpha1.ProviderReference{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-providerReference",
					Namespace: "test-providerReference-namespace",
				},
				Spec: catalogv1alpha1.ProviderReferenceSpec{
					Metadata: catalogv1alpha1.ProviderMetadata{
						Description: "test ProviderReference",
						DisplayName: "test ProviderReference",
					},
				},
			},
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedError, providerReferenceWebhookHandler.validateMetadata(test.object) != nil)
		})
	}
}

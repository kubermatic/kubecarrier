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

	"github.com/kubermatic/utils/pkg/testutil"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
)

func TestProviderValidatingCreate(t *testing.T) {
	providerWebhookHandler := ProviderWebhookHandler{
		Log: testutil.NewLogger(t),
	}

	tests := []struct {
		name          string
		object        *catalogv1alpha1.Provider
		expectedError bool
	}{
		{
			name: "invalid Provider name",
			object: &catalogv1alpha1.Provider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test.provider",
					Namespace: "test-provider-namespace",
				},
				Spec: catalogv1alpha1.ProviderSpec{
					Metadata: catalogv1alpha1.AccountMetadata{
						CommonMetadata: catalogv1alpha1.CommonMetadata{
							Description: "test Provider",
							DisplayName: "test Provider",
						},
					},
				},
			},
			expectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedError, providerWebhookHandler.validateCreate(test.object) != nil)
		})
	}
}

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
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/manager/internal/controllers"
)

func TestCatalogEntryDefaultMatchLabels(t *testing.T) {
	catalogEntryWebhookHandler := CatalogEntryWebhookHandler{
		ProviderLabel: controllers.ProviderLabel,
	}
	catalogEntry := &catalogv1alpha1.CatalogEntry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-catalogEntry",
			Namespace: "test-provider-namespace",
		},
	}
	provider := &catalogv1alpha1.Provider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-provider",
			Namespace: "kubecarrier-system",
		},
	}

	require.NoError(t, catalogEntryWebhookHandler.defaultMatchLabels(catalogEntry, provider))
	assert.Equal(t, catalogEntry.Spec.CRDSelector.MatchLabels[controllers.ProviderLabel], provider.Name)

}

func TestCatalogEntryValidatingMetadata(t *testing.T) {
	catalogEntryWebhookHandler := CatalogEntryWebhookHandler{
		ProviderLabel: controllers.ProviderLabel,
	}

	tests := []struct {
		name          string
		object        *catalogv1alpha1.CatalogEntry
		expectedError bool
	}{
		{
			name: "metadata missing",
			object: &catalogv1alpha1.CatalogEntry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-catalogEntry",
					Namespace: "test-provider-namespace",
				},
			},
			expectedError: true,
		},
		{
			name: "description missing",
			object: &catalogv1alpha1.CatalogEntry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-catalogEntry",
					Namespace: "test-provider-namespace",
				},
				Spec: catalogv1alpha1.CatalogEntrySpec{
					Metadata: catalogv1alpha1.CatalogEntryMetadata{
						DisplayName: "test CatalogEntry",
					},
				},
			},
			expectedError: true,
		},
		{
			name: "displayName missing",
			object: &catalogv1alpha1.CatalogEntry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-catalogEntry",
					Namespace: "test-provider-namespace",
				},
				Spec: catalogv1alpha1.CatalogEntrySpec{
					Metadata: catalogv1alpha1.CatalogEntryMetadata{
						Description: "test CatalogEntry",
					},
				},
			},
			expectedError: true,
		},
		{
			name: "can pass validating metadata",
			object: &catalogv1alpha1.CatalogEntry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-catalogEntry",
					Namespace: "test-provider-namespace",
				},
				Spec: catalogv1alpha1.CatalogEntrySpec{
					Metadata: catalogv1alpha1.CatalogEntryMetadata{
						Description: "test CatalogEntry",
						DisplayName: "test CatalogEntry",
					},
				},
			},
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedError, catalogEntryWebhookHandler.validateMetadata(test.object) != nil)
		})
	}
}

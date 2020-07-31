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

	"k8c.io/utils/pkg/testutil"

	catalogv1alpha1 "k8c.io/kubecarrier/pkg/apis/catalog/v1alpha1"
)

func TestCatalogEntryValidatingUpdate(t *testing.T) {
	catalogEntryWebhookHandler := CatalogEntryWebhookHandler{
		Log: testutil.NewLogger(t),
	}

	oldObj := &catalogv1alpha1.CatalogEntry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-catalogEntry",
			Namespace: "test-provider-namespace",
		},
		Spec: catalogv1alpha1.CatalogEntrySpec{
			Metadata: catalogv1alpha1.CatalogEntryMetadata{
				CommonMetadata: catalogv1alpha1.CommonMetadata{
					Description: "test CatalogEntry",
					DisplayName: "test CatalogEntry",
				},
			},
			BaseCRD: catalogv1alpha1.ObjectReference{
				Name: "test-crd",
			},
		},
	}

	tests := []struct {
		name          string
		object        *catalogv1alpha1.CatalogEntry
		expectedError bool
	}{
		{
			name: "referenced crd immutable",
			object: &catalogv1alpha1.CatalogEntry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-catalogEntry",
					Namespace: "test-provider-namespace",
				},
				Spec: catalogv1alpha1.CatalogEntrySpec{
					Metadata: catalogv1alpha1.CatalogEntryMetadata{
						CommonMetadata: catalogv1alpha1.CommonMetadata{
							Description: "test CatalogEntry",
							DisplayName: "test CatalogEntry",
						},
					},
					BaseCRD: catalogv1alpha1.ObjectReference{
						Name: "test-crd2",
					},
				},
			},
			expectedError: true,
		},
		{
			name: "can pass validating update",
			object: &catalogv1alpha1.CatalogEntry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-catalogEntry",
					Namespace: "test-provider-namespace",
				},
				Spec: catalogv1alpha1.CatalogEntrySpec{
					Metadata: catalogv1alpha1.CatalogEntryMetadata{
						CommonMetadata: catalogv1alpha1.CommonMetadata{
							Description: "test CatalogEntry",
							DisplayName: "test CatalogEntry",
						},
					},
					BaseCRD: catalogv1alpha1.ObjectReference{
						Name: "test-crd",
					},
				},
			},
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedError, catalogEntryWebhookHandler.validateUpdate(test.object, oldObj) != nil)
		})
	}
}

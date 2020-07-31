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

	catalogv1alpha1 "k8c.io/kubecarrier/pkg/apis/catalog/v1alpha1"
)

func TestOfferingValidatingUpdate(t *testing.T) {
	offeringWebhookHandler := OfferingWebhookHandler{
		Log: testutil.NewLogger(t),
	}

	oldObj := &catalogv1alpha1.Offering{
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
				Name: "Provider",
			},
		},
	}

	tests := []struct {
		name          string
		object        *catalogv1alpha1.Offering
		expectedError bool
	}{
		{
			name: "provider immutable",
			object: &catalogv1alpha1.Offering{
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
						Name: "Provider2",
					},
				},
			},
			expectedError: true,
		},
		{
			name: "can pass validating update",
			object: &catalogv1alpha1.Offering{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-offering",
					Namespace: "test-namespace",
				},
				Spec: catalogv1alpha1.OfferingSpec{
					Metadata: catalogv1alpha1.OfferingMetadata{
						CommonMetadata: catalogv1alpha1.CommonMetadata{
							Description: "Test Offering",
							DisplayName: "Test Offering new displayName",
						},
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
			assert.Equal(t, test.expectedError, offeringWebhookHandler.validateUpdate(test.object, oldObj) != nil)
		})
	}
}

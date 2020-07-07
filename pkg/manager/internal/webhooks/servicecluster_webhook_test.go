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

	covev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/utils/pkg/testutil"
)

func TestServiceClusterValidatingCreate(t *testing.T) {
	serviceClusterWebhookHandler := ServiceClusterWebhookHandler{
		Log: testutil.NewLogger(t),
	}

	tests := []struct {
		name          string
		object        *covev1alpha1.ServiceCluster
		expectedError bool
	}{
		{
			name: "invalid serviceCluster name",
			object: &covev1alpha1.ServiceCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test.serviceCluster",
					Namespace: "test-serviceCluster-namespace",
				},
			},
			expectedError: true,
		},
		{
			name: "can pass validate create",
			object: &covev1alpha1.ServiceCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service-cluster",
					Namespace: "test-serviceCluster-namespace",
				},
			},
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedError, serviceClusterWebhookHandler.validateCreate(test.object) != nil)
		})
	}
}

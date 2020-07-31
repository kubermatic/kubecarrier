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

func TestTenantValidatingCreate(t *testing.T) {
	tenantWebhookHandler := TenantWebhookHandler{
		Log: testutil.NewLogger(t),
	}

	tests := []struct {
		name          string
		object        *catalogv1alpha1.Tenant
		expectedError bool
	}{
		{
			name: "invalid tenant name",
			object: &catalogv1alpha1.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test.tenant",
					Namespace: "test-tenant-namespace",
				},
			},
			expectedError: true,
		},
		{
			name: "can pass validate create",
			object: &catalogv1alpha1.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-tenant",
					Namespace: "test-tenant-namespace",
				},
			},
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedError, tenantWebhookHandler.validateCreate(test.object) != nil)
		})
	}
}

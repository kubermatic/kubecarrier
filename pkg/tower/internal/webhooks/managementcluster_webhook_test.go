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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	masterv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/master/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/constants"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func TestManagementClusterValidatingCreate(t *testing.T) {
	managementClusterWebhookHandler := ManagementClusterWebhookHandler{
		Log: testutil.NewLogger(t),
	}

	tests := []struct {
		name          string
		object        *masterv1alpha1.ManagementCluster
		expectedError error
	}{
		{
			name: "invalid KubeconfigSecret",
			object: &masterv1alpha1.ManagementCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-kubecarrier",
				},
			},
			expectedError: fmt.Errorf("ManagementCluster.Spec.KubeconfigSecret should not be nil"),
		},
		{
			name: "can pass validate create",
			object: &masterv1alpha1.ManagementCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-kubecarrier",
				},
				Spec: masterv1alpha1.ManagementClusterSpec{
					KubeconfigSecret: &masterv1alpha1.ObjectReference{
						Name: "test-kubeconfig-secret",
					},
				},
			},
			expectedError: nil,
		},
		{
			name: "local management cluster can pass validate create",
			object: &masterv1alpha1.ManagementCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: constants.LocalManagementClusterName,
				},
			},
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedError, managementClusterWebhookHandler.validateCreate(test.object))
		})
	}
}

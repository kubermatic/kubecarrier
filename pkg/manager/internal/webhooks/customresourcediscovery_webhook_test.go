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

	corev1alpha1 "k8c.io/kubecarrier/pkg/apis/core/v1alpha1"
)

func TestCustomResourceDiscoveryValidatingUpdate(t *testing.T) {
	customResourceDiscoveryWebhookHandler := CustomResourceDiscoveryWebhookHandler{
		Log: testutil.NewLogger(t),
	}

	oldObj := &corev1alpha1.CustomResourceDiscovery{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-crdiscovery",
			Namespace: "test-namespace",
		},
		Spec: corev1alpha1.CustomResourceDiscoverySpec{
			ServiceCluster: corev1alpha1.ObjectReference{
				Name: "ServiceCluster",
			},
			CRD: corev1alpha1.ObjectReference{
				Name: "CRD",
			},
			KindOverride: "KindOverride",
		},
	}

	tests := []struct {
		name          string
		object        *corev1alpha1.CustomResourceDiscovery
		expectedError bool
	}{
		{
			name: "kind override immutable",
			object: &corev1alpha1.CustomResourceDiscovery{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-crdiscovery",
					Namespace: "test-namespace",
				},
				Spec: corev1alpha1.CustomResourceDiscoverySpec{
					ServiceCluster: corev1alpha1.ObjectReference{
						Name: "ServiceCluster",
					},
					CRD: corev1alpha1.ObjectReference{
						Name: "CRD",
					},
					KindOverride: "KindOverride2",
				},
			},
			expectedError: true,
		},
		{
			name: "servicecluster immutable",
			object: &corev1alpha1.CustomResourceDiscovery{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-crdiscovery",
					Namespace: "test-namespace",
				},
				Spec: corev1alpha1.CustomResourceDiscoverySpec{
					ServiceCluster: corev1alpha1.ObjectReference{
						Name: "ServiceCluster2",
					},
					CRD: corev1alpha1.ObjectReference{
						Name: "CRD",
					},
					KindOverride: "KindOverride",
				},
			},
			expectedError: true,
		},
		{
			name: "crd immutable",
			object: &corev1alpha1.CustomResourceDiscovery{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-crdiscovery",
					Namespace: "test-namespace",
				},
				Spec: corev1alpha1.CustomResourceDiscoverySpec{
					ServiceCluster: corev1alpha1.ObjectReference{
						Name: "ServiceCluster",
					},
					CRD: corev1alpha1.ObjectReference{
						Name: "CRD2",
					},
					KindOverride: "KindOverride",
				},
			},
			expectedError: true,
		},
		{
			name: "can pass validating update",
			object: &corev1alpha1.CustomResourceDiscovery{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-crdiscovery",
					Namespace: "test-namespace",
				},
				Spec: corev1alpha1.CustomResourceDiscoverySpec{
					ServiceCluster: corev1alpha1.ObjectReference{
						Name: "ServiceCluster",
					},
					CRD: corev1alpha1.ObjectReference{
						Name: "CRD",
					},
					KindOverride: "KindOverride",
				},
			},
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedError, customResourceDiscoveryWebhookHandler.validateUpdate(test.object, oldObj) != nil)
		})
	}
}

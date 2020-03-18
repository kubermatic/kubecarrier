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

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func TestServiceClusterAssignmentValidatingCreate(t *testing.T) {
	serviceClusterAssignmentWebhookHandler := ServiceClusterAssignmentWebhookHandler{
		Log: testutil.NewLogger(t),
	}

	tests := []struct {
		name          string
		object        *corev1alpha1.ServiceClusterAssignment
		expectedError bool
	}{
		{
			name: "serviceClusterAssignment name incorrect",
			object: &corev1alpha1.ServiceClusterAssignment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "management-namespace.service-cluster2",
					Namespace: "test-namespace",
				},
				Spec: corev1alpha1.ServiceClusterAssignmentSpec{
					ServiceCluster: corev1alpha1.ObjectReference{
						Name: "service-cluster",
					},
					ManagementClusterNamespace: corev1alpha1.ObjectReference{
						Name: "management-namespace",
					},
				},
			},
			expectedError: true,
		},
		{
			name: "can pass validating create",
			object: &corev1alpha1.ServiceClusterAssignment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "management-namespace.service-cluster",
					Namespace: "test-namespace",
				},
				Spec: corev1alpha1.ServiceClusterAssignmentSpec{
					ServiceCluster: corev1alpha1.ObjectReference{
						Name: "service-cluster",
					},
					ManagementClusterNamespace: corev1alpha1.ObjectReference{
						Name: "management-namespace",
					},
				},
			},
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedError, serviceClusterAssignmentWebhookHandler.validateCreate(test.object) != nil)
		})
	}
}

func TestServiceClusterAssignmentValidatingUpdate(t *testing.T) {
	serviceClusterAssignmentWebhookHandler := ServiceClusterAssignmentWebhookHandler{
		Log: testutil.NewLogger(t),
	}

	oldObj := &corev1alpha1.ServiceClusterAssignment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "management-namespace.service-cluster",
			Namespace: "test-namespace",
		},
		Spec: corev1alpha1.ServiceClusterAssignmentSpec{
			ServiceCluster: corev1alpha1.ObjectReference{
				Name: "service-cluster",
			},
			ManagementClusterNamespace: corev1alpha1.ObjectReference{
				Name: "management-namespace",
			},
		},
	}

	tests := []struct {
		name          string
		object        *corev1alpha1.ServiceClusterAssignment
		expectedError bool
	}{
		{
			name: "service cluster immutable",
			object: &corev1alpha1.ServiceClusterAssignment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "management-namespace.service-cluster",
					Namespace: "test-namespace",
				},
				Spec: corev1alpha1.ServiceClusterAssignmentSpec{
					ServiceCluster: corev1alpha1.ObjectReference{
						Name: "service-cluster2",
					},
					ManagementClusterNamespace: corev1alpha1.ObjectReference{
						Name: "management-namespace",
					},
				},
			},
			expectedError: true,
		},
		{
			name: "management cluster namespace immutable",
			object: &corev1alpha1.ServiceClusterAssignment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "management-namespace.service-cluster",
					Namespace: "test-namespace",
				},
				Spec: corev1alpha1.ServiceClusterAssignmentSpec{
					ServiceCluster: corev1alpha1.ObjectReference{
						Name: "service-cluster",
					},
					ManagementClusterNamespace: corev1alpha1.ObjectReference{
						Name: "management-namespace2",
					},
				},
			},
			expectedError: true,
		},
		{
			name: "can pass validating update",
			object: &corev1alpha1.ServiceClusterAssignment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "management-namespace.service-cluster",
					Namespace: "test-namespace",
				},
				Spec: corev1alpha1.ServiceClusterAssignmentSpec{
					ServiceCluster: corev1alpha1.ObjectReference{
						Name: "service-cluster",
					},
					ManagementClusterNamespace: corev1alpha1.ObjectReference{
						Name: "management-namespace",
					},
				},
			},
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedError, serviceClusterAssignmentWebhookHandler.validateUpdate(test.object, oldObj) != nil)
		})
	}
}

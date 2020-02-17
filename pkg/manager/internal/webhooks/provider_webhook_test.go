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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
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
			name: "invalid provider name",
			object: &catalogv1alpha1.Provider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test.provider",
					Namespace: "test-provider-namespace",
				},
				Spec: catalogv1alpha1.ProviderSpec{
					Metadata: catalogv1alpha1.ProviderMetadata{
						Description: "test Provider",
						DisplayName: "test Provider",
					},
				},
			},
			expectedError: true,
		},
		{
			name: "metadata missing",
			object: &catalogv1alpha1.Provider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-provider",
					Namespace: "test-provider-namespace",
				},
			},
			expectedError: true,
		},
		{
			name: "description missing",
			object: &catalogv1alpha1.Provider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-provider",
					Namespace: "test-provider-namespace",
				},
				Spec: catalogv1alpha1.ProviderSpec{
					Metadata: catalogv1alpha1.ProviderMetadata{
						DisplayName: "test Provider",
					},
				},
			},
			expectedError: true,
		},
		{
			name: "displayName missing",
			object: &catalogv1alpha1.Provider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-provider",
					Namespace: "test-provider-namespace",
				},
				Spec: catalogv1alpha1.ProviderSpec{
					Metadata: catalogv1alpha1.ProviderMetadata{
						Description: "test Provider",
					},
				},
			},
			expectedError: true,
		},
		{
			name: "can pass validate create",
			object: &catalogv1alpha1.Provider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-provider",
					Namespace: "test-provider-namespace",
				},
				Spec: catalogv1alpha1.ProviderSpec{
					Metadata: catalogv1alpha1.ProviderMetadata{
						Description: "test Provider",
						DisplayName: "test Provider",
					},
				},
			},
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedError, providerWebhookHandler.validateCreate(test.object) != nil)
		})
	}
}

func TestProviderValidatingDelete(t *testing.T) {
	ctx := context.Background()
	provider := &catalogv1alpha1.Provider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-provider",
			Namespace: "test-provider-namespace",
		},
		Spec: catalogv1alpha1.ProviderSpec{
			Metadata: catalogv1alpha1.ProviderMetadata{
				Description: "test Provider",
				DisplayName: "test Provider",
			},
		},
		Status: catalogv1alpha1.ProviderStatus{
			NamespaceName: "default",
		},
	}
	for _, test := range []struct {
		name          string
		client        client.Client
		expectedError bool
	}{
		{
			name:   "simple clean namespace",
			client: fakeclient.NewFakeClientWithScheme(testScheme),
		},
		{
			name: "extra derived custom resource",
			client: fakeclient.NewFakeClientWithScheme(testScheme, &catalogv1alpha1.DerivedCustomResource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dummy",
					Namespace: "default",
				},
			}),
			expectedError: true,
		},
		{
			name: "extra CustomResourceDiscovery",
			client: fakeclient.NewFakeClientWithScheme(testScheme, &corev1alpha1.CustomResourceDiscovery{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dummy",
					Namespace: "default",
				},
			}),
			expectedError: true,
		},
		{
			name: "extra CustomResourceDiscoverySet",
			client: fakeclient.NewFakeClientWithScheme(testScheme, &corev1alpha1.CustomResourceDiscoverySet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dummy",
					Namespace: "default",
				},
			}),
			expectedError: true,
		},
		{
			name: "extra CustomResourceDiscoverySet",
			client: fakeclient.NewFakeClientWithScheme(testScheme, &corev1alpha1.CustomResourceDiscoverySet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dummy",
					Namespace: "default",
				},
			},
				&corev1alpha1.CustomResourceDiscovery{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dummy",
						Namespace: "default",
					},
				}),
			expectedError: true,
		},
	} {
		providerWebhookHandler := ProviderWebhookHandler{
			Log:    testutil.NewLogger(t),
			Client: test.client,
			Scheme: testScheme,
		}
		t.Run(test.name, func(t *testing.T) {
			if test.expectedError {
				err := providerWebhookHandler.validateDelete(ctx, provider)
				assert.Error(t, err)
				t.Log(err)
			} else {
				assert.NoError(t, providerWebhookHandler.validateDelete(ctx, provider))
			}
		})
	}
}

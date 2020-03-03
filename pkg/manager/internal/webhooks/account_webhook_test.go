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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func TestAccountValidatingCreate(t *testing.T) {

	tests := []struct {
		name            string
		object          *catalogv1alpha1.Account
		existingObjects []runtime.Object
		expectedError   error
	}{
		{
			name: "invalid account name",
			object: &catalogv1alpha1.Account{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test.account",
				},
				Spec: catalogv1alpha1.AccountSpec{
					Metadata: catalogv1alpha1.AccountMetadata{
						Description: "test Account",
						DisplayName: "test Account",
					},
				},
			},
			expectedError: fmt.Errorf("account name: test.account is not a valid DNS 1123 Label, A DNS-1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character. (e.g. 'my-name',  or '123-abc', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?'"),
		},
		{
			name: "missing roles",
			object: &catalogv1alpha1.Account{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-account",
				},
			},
			expectedError: fmt.Errorf("no roles assigned"),
		},
		{
			name: "metadata missing",
			object: &catalogv1alpha1.Account{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-account",
				},
				Spec: catalogv1alpha1.AccountSpec{
					Roles: []catalogv1alpha1.AccountRole{
						catalogv1alpha1.ProviderRole,
					},
				},
			},
			expectedError: fmt.Errorf("the description or the display name of an Account with Provider role cannot be empty"),
		},
		{
			name: "metadata description missing",
			object: &catalogv1alpha1.Account{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-account",
				},
				Spec: catalogv1alpha1.AccountSpec{
					Roles: []catalogv1alpha1.AccountRole{
						catalogv1alpha1.ProviderRole,
					},
					Metadata: catalogv1alpha1.AccountMetadata{
						DisplayName: "test Account",
					},
				},
			},
			expectedError: fmt.Errorf("the description or the display name of an Account with Provider role cannot be empty"),
		},
		{
			name: "displayName missing",
			object: &catalogv1alpha1.Account{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-account",
				},
				Spec: catalogv1alpha1.AccountSpec{
					Roles: []catalogv1alpha1.AccountRole{
						catalogv1alpha1.ProviderRole,
					},
					Metadata: catalogv1alpha1.AccountMetadata{
						Description: "test Account",
					},
				},
			},
			expectedError: fmt.Errorf("the description or the display name of an Account with Provider role cannot be empty"),
		},
		{
			name: "duplicate roles",
			object: &catalogv1alpha1.Account{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-account",
				},
				Spec: catalogv1alpha1.AccountSpec{
					Metadata: catalogv1alpha1.AccountMetadata{
						Description: "test Account",
						DisplayName: "test Account",
					},
					Roles: []catalogv1alpha1.AccountRole{
						catalogv1alpha1.ProviderRole,
						catalogv1alpha1.ProviderRole,
					},
				},
			},
			expectedError: fmt.Errorf("role Provider is duplicated"),
		},
		{
			name: "namespace already exists",
			object: &catalogv1alpha1.Account{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-account",
				},
				Spec: catalogv1alpha1.AccountSpec{
					Metadata: catalogv1alpha1.AccountMetadata{
						Description: "test Account",
						DisplayName: "test Account",
					},
					Roles: []catalogv1alpha1.AccountRole{
						catalogv1alpha1.ProviderRole,
					},
				},
			},
			existingObjects: []runtime.Object{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-account"}},
			},
			expectedError: fmt.Errorf("namespace test-account already exists"),
		},
		{
			name: "can pass validate create",
			object: &catalogv1alpha1.Account{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-account",
				},
				Spec: catalogv1alpha1.AccountSpec{
					Metadata: catalogv1alpha1.AccountMetadata{
						Description: "test Account",
						DisplayName: "test Account",
					},
					Roles: []catalogv1alpha1.AccountRole{
						catalogv1alpha1.ProviderRole,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			accountWebhookHandler := AccountWebhookHandler{
				Log:    testutil.NewLogger(t),
				Client: fakeclient.NewFakeClientWithScheme(testScheme, test.existingObjects...),
			}
			if test.expectedError == nil {
				assert.NoError(t, accountWebhookHandler.validateCreate(context.Background(), test.object))
				return
			}

			err := accountWebhookHandler.validateCreate(context.Background(), test.object)
			if assert.Error(t, err) {
				assert.EqualError(t, err, test.expectedError.Error())
			}
		})
	}
}

func TestAccountValidatingDelete(t *testing.T) {
	ctx := context.Background()
	account := &catalogv1alpha1.Account{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-account",
			Namespace: "test-account-namespace",
		},
		Spec: catalogv1alpha1.AccountSpec{
			Metadata: catalogv1alpha1.AccountMetadata{
				Description: "test Account",
				DisplayName: "test Account",
			},
		},
		Status: catalogv1alpha1.AccountStatus{
			Namespace: catalogv1alpha1.ObjectReference{Name: "default"},
		},
	}
	for _, test := range []struct {
		name          string
		client        client.Client
		expectedError error
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
			expectedError: fmt.Errorf("deletion blocking objects found:\nDerivedCustomResource.catalog.kubecarrier.io/v1alpha1: dummy\n"),
		},
		{
			name: "extra CustomResourceDiscovery",
			client: fakeclient.NewFakeClientWithScheme(testScheme, &corev1alpha1.CustomResourceDiscovery{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dummy",
					Namespace: "default",
				},
			}),
			expectedError: fmt.Errorf("deletion blocking objects found:\nCustomResourceDiscovery.kubecarrier.io/v1alpha1: dummy\n"),
		},
		{
			name: "extra CustomResourceDiscoverySet",
			client: fakeclient.NewFakeClientWithScheme(testScheme, &corev1alpha1.CustomResourceDiscoverySet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dummy",
					Namespace: "default",
				},
			}),
			expectedError: fmt.Errorf("deletion blocking objects found:\nCustomResourceDiscoverySet.kubecarrier.io/v1alpha1: dummy\n"),
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
			expectedError: fmt.Errorf("deletion blocking objects found:\nCustomResourceDiscovery.kubecarrier.io/v1alpha1: dummy\nCustomResourceDiscoverySet.kubecarrier.io/v1alpha1: dummy\n"),
		},
	} {
		accountWebhookHandler := AccountWebhookHandler{
			Log:    testutil.NewLogger(t),
			Client: test.client,
			Scheme: testScheme,
		}
		t.Run(test.name, func(t *testing.T) {
			if test.expectedError == nil {
				assert.NoError(t, accountWebhookHandler.validateDelete(ctx, account))
				return
			}

			err := accountWebhookHandler.validateDelete(ctx, account)
			if assert.Error(t, err) {
				assert.EqualError(t, err, test.expectedError.Error())
			}
		})
	}
}

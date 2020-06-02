/*
Copyright 2020 The KubeCarrier Authors.

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

package v1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	v1 "github.com/kubermatic/kubecarrier/pkg/apiserver/api/v1"
)

func TestListAccount(t *testing.T) {
	accounts := &catalogv1alpha1.AccountList{
		Items: []catalogv1alpha1.Account{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-account-1",
					Labels: map[string]string{
						"test-label": "account1",
					},
				},
				Spec: catalogv1alpha1.AccountSpec{
					Metadata: catalogv1alpha1.AccountMetadata{
						CommonMetadata: catalogv1alpha1.CommonMetadata{
							DisplayName:      "provider",
							ShortDescription: "provider desc",
						},
					},
					Roles: []catalogv1alpha1.AccountRole{
						catalogv1alpha1.ProviderRole,
					},
					Subjects: []rbacv1.Subject{
						{
							Kind:     rbacv1.GroupKind,
							APIGroup: "rbac.authorization.k8s.io",
							Name:     "provider1",
						},
					},
				},
				Status: catalogv1alpha1.AccountStatus{
					Namespace: &catalogv1alpha1.ObjectReference{
						Name: "test-account-1",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-account-2",
					Labels: map[string]string{
						"test-label": "account2",
					},
				},
				Spec: catalogv1alpha1.AccountSpec{
					Metadata: catalogv1alpha1.AccountMetadata{
						CommonMetadata: catalogv1alpha1.CommonMetadata{
							DisplayName:      "tenant",
							ShortDescription: "tenant desc",
						},
					},
					Roles: []catalogv1alpha1.AccountRole{
						catalogv1alpha1.TenantRole,
					},
					Subjects: []rbacv1.Subject{
						{
							Kind:     rbacv1.GroupKind,
							APIGroup: "rbac.authorization.k8s.io",
							Name:     "tenant1",
						},
					},
				},
				Status: catalogv1alpha1.AccountStatus{
					Namespace: &catalogv1alpha1.ObjectReference{
						Name: "test-account-2",
					},
				},
			},
		},
	}
	client := fakeclient.NewFakeClientWithScheme(testScheme, accounts)
	accountServer := accountServer{
		client: client,
	}
	ctx := context.Background()
	tests := []struct {
		name           string
		req            *v1.ListRequest
		expectedError  error
		expectedResult *v1.AccountList
	}{
		{
			name: "invalid label selector",
			req: &v1.ListRequest{
				LabelSelector: "test-label=====account1",
			},
			expectedError:  status.Errorf(codes.InvalidArgument, "invalid LabelSelector: unable to parse requirement: found '==', expected: identifier"),
			expectedResult: nil,
		},
		{
			name:          "valid request",
			req:           &v1.ListRequest{},
			expectedError: nil,
			expectedResult: &v1.AccountList{
				Metadata: &v1.ListMeta{
					Continue:        "",
					ResourceVersion: "",
				},
				Items: []*v1.Account{
					{
						Metadata: &v1.ObjectMeta{
							Name: "test-account-1",
							Labels: map[string]string{
								"test-label": "account1",
							},
						},
						Spec: &v1.AccountSpec{
							Metadata: &v1.AccountMetadata{
								DisplayName:      "provider",
								ShortDescription: "provider desc",
							},
							Roles: []*v1.AccountRole{
								{
									Type: "Provider",
								},
							},
							Subjects: []*v1.Subject{
								{
									Kind:     "Group",
									ApiGroup: "rbac.authorization.k8s.io",
									Name:     "provider1",
								},
							},
						},
						Status: &v1.AccountStatus{},
					},
					{
						Metadata: &v1.ObjectMeta{
							Name: "test-account-2",
							Labels: map[string]string{
								"test-label": "account2",
							},
						},
						Spec: &v1.AccountSpec{
							Metadata: &v1.AccountMetadata{
								DisplayName:      "tenant",
								ShortDescription: "tenant desc",
							},
							Roles: []*v1.AccountRole{
								{
									Type: "Tenant",
								},
							},
							Subjects: []*v1.Subject{
								{
									Kind:     "Group",
									ApiGroup: "rbac.authorization.k8s.io",
									Name:     "tenant1",
								},
							},
						},
						Status: &v1.AccountStatus{},
					},
				},
			},
		},
		{
			name: "LabelSelector works",
			req: &v1.ListRequest{
				LabelSelector: "test-label=account1",
			},
			expectedError: nil,
			expectedResult: &v1.AccountList{
				Metadata: &v1.ListMeta{
					Continue:        "",
					ResourceVersion: "",
				},
				Items: []*v1.Account{
					{
						Metadata: &v1.ObjectMeta{
							Name: "test-account-1",
							Labels: map[string]string{
								"test-label": "account1",
							},
						},
						Spec: &v1.AccountSpec{
							Metadata: &v1.AccountMetadata{
								DisplayName:      "provider",
								ShortDescription: "provider desc",
							},
							Roles: []*v1.AccountRole{
								{
									Type: "Provider",
								},
							},
							Subjects: []*v1.Subject{
								{
									Kind:     "Group",
									ApiGroup: "rbac.authorization.k8s.io",
									Name:     "provider1",
								},
							},
						},
						Status: &v1.AccountStatus{},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Here just pass empty user since fakeClient doesn't work with field indexer, here we just test the basic functionality.
			accounts, err := accountServer.handleListRequest(ctx, test.req, "")
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedResult, accounts)
		})
	}
}

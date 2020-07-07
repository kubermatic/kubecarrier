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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/kubermatic/utils/pkg/testutil"

	fakev1 "github.com/kubermatic/kubecarrier/pkg/apis/fake/v1"
)

func TestDBValidatingCreate(t *testing.T) {

	tests := []struct {
		name            string
		object          *fakev1.DB
		existingObjects []runtime.Object
		expectedError   error
	}{
		{
			name: "invalid db name",
			object: &fakev1.DB{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test.db",
				},
				Spec: fakev1.DBSpec{
					Config: fakev1.Config{
						Create: fakev1.OperationFlagEnabled,
					},
				},
			},
			expectedError: fmt.Errorf("DB name: test.db is not a valid DNS 1123 Label, A DNS-1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character. (e.g. 'my-name',  or '123-abc', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?'"),
		},
		{
			name: "create operation disabled",
			object: &fakev1.DB{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testdb",
				},
				Spec: fakev1.DBSpec{
					Config: fakev1.Config{
						Create: fakev1.OperationFlagDisabled,
					},
				},
			},
			expectedError: fmt.Errorf("create operation disabled for %s", "testdb"),
		},
		{
			name: "create operation enabled",
			object: &fakev1.DB{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testdb",
				},
				Spec: fakev1.DBSpec{
					Config: fakev1.Config{
						Create: fakev1.OperationFlagEnabled,
					},
				},
			},
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dbWebhookHandler := DBWebhookHandler{
				Log:    testutil.NewLogger(t),
				Client: fakeclient.NewFakeClientWithScheme(testScheme, test.existingObjects...),
			}
			if test.expectedError == nil {
				assert.NoError(t, dbWebhookHandler.validateCreate(context.Background(), test.object))
				return
			}

			err := dbWebhookHandler.validateCreate(context.Background(), test.object)
			if assert.Error(t, err) {
				assert.EqualError(t, err, test.expectedError.Error())
			}
		})
	}
}

func TestDBValidatingDelete(t *testing.T) {

	tests := []struct {
		name            string
		object          *fakev1.DB
		existingObjects []runtime.Object
		expectedError   error
	}{
		{
			name: "delete operation disabled",
			object: &fakev1.DB{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testdb",
				},
				Spec: fakev1.DBSpec{
					Config: fakev1.Config{
						Delete: fakev1.OperationFlagDisabled,
					},
				},
			},
			expectedError: fmt.Errorf("delete operation disabled for %s", "testdb"),
		},
		{
			name: "delete operation enabled",
			object: &fakev1.DB{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testdb",
				},
				Spec: fakev1.DBSpec{
					Config: fakev1.Config{
						Delete: fakev1.OperationFlagEnabled,
					},
				},
			},
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dbWebhookHandler := DBWebhookHandler{
				Log:    testutil.NewLogger(t),
				Client: fakeclient.NewFakeClientWithScheme(testScheme, test.existingObjects...),
			}
			if test.expectedError == nil {
				assert.NoError(t, dbWebhookHandler.validateDelete(context.Background(), test.object))
				return
			}

			err := dbWebhookHandler.validateDelete(context.Background(), test.object)
			if assert.Error(t, err) {
				assert.EqualError(t, err, test.expectedError.Error())
			}
		})
	}
}

func TestDBValidatingUpdate(t *testing.T) {

	oldObj := &fakev1.DB{
		ObjectMeta: metav1.ObjectMeta{
			Name: "testdb",
		},
		Spec: fakev1.DBSpec{
			DatabaseName: "dbname",
			Config: fakev1.Config{
				Update: fakev1.OperationFlagEnabled,
			},
		},
	}

	tests := []struct {
		name            string
		oldObj          *fakev1.DB
		object          *fakev1.DB
		existingObjects []runtime.Object
		expectedError   error
	}{
		{
			name: "update operation disabled",
			oldObj: &fakev1.DB{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testdb",
				},
				Spec: fakev1.DBSpec{
					Config: fakev1.Config{
						Update: fakev1.OperationFlagDisabled,
					},
				},
			},
			object: &fakev1.DB{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testdb",
				},
				Spec: fakev1.DBSpec{
					Config: fakev1.Config{Update: fakev1.OperationFlagDisabled},
				},
			},
			expectedError: fmt.Errorf("update operation disabled for %s", "testdb"),
		},
		{
			name:   "update operation enabled in oldObj",
			oldObj: oldObj,
			object: &fakev1.DB{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testdb",
				},
				Spec: fakev1.DBSpec{
					DatabaseName: "dbname",
					Config: fakev1.Config{
						Update: fakev1.OperationFlagDisabled,
					},
				},
			},
			expectedError: nil,
		},
		{
			name: "update operation enabled in newObj",
			oldObj: &fakev1.DB{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testdb",
				},
				Spec: fakev1.DBSpec{
					Config: fakev1.Config{
						Update: fakev1.OperationFlagDisabled,
					},
				},
			},
			object: &fakev1.DB{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testdb",
				},
				Spec: fakev1.DBSpec{
					Config: fakev1.Config{
						Update: fakev1.OperationFlagEnabled,
					},
				},
			},
			expectedError: nil,
		},
		{
			name:   "change Database name",
			oldObj: oldObj,
			object: &fakev1.DB{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testdb2",
				},
				Spec: fakev1.DBSpec{
					DatabaseName: "new dbname",
					Config:       fakev1.Config{},
				},
			},
			expectedError: fmt.Errorf("the Database name is immutable"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dbWebhookHandler := DBWebhookHandler{
				Log:    testutil.NewLogger(t),
				Client: fakeclient.NewFakeClientWithScheme(testScheme, test.existingObjects...),
			}
			if test.expectedError == nil {
				assert.NoError(t, dbWebhookHandler.validateUpdate(test.oldObj, test.object))
				return
			}

			err := dbWebhookHandler.validateUpdate(test.oldObj, test.object)
			if assert.Error(t, err) {
				assert.EqualError(t, err, test.expectedError.Error())
			}
		})
	}
}

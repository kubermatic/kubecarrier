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

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func TestDerivedCustomResourceDefinitionValidatingCreate(t *testing.T) {
	derivedCustomResourceDefinitionWebhookHandler := DerivedCustomResourceDefinitionWebhookHandler{
		Log: testutil.NewLogger(t),
	}

	tests := []struct {
		name          string
		object        *catalogv1alpha1.DerivedCustomResourceDefinition
		expectedError bool
	}{
		{
			name: "BaseCRD missing",
			object: &catalogv1alpha1.DerivedCustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-derivedCustomResourceDefinition",
					Namespace: "test-namespace",
				},
				Spec: catalogv1alpha1.DerivedCustomResourceDefinitionSpec{
					Expose: []catalogv1alpha1.VersionExposeConfig{
						{
							Versions: []string{
								"v1alpha1",
							},
							Fields: []catalogv1alpha1.FieldPath{
								{JSONPath: ".spec.prop1"},
							},
						},
					},
				},
			},
			expectedError: true,
		},
		{
			name: "Expose missing",
			object: &catalogv1alpha1.DerivedCustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-derivedCustomResourceDefinition",
					Namespace: "test-namespace",
				},
				Spec: catalogv1alpha1.DerivedCustomResourceDefinitionSpec{
					BaseCRD: catalogv1alpha1.ObjectReference{
						Name: "BaseCRD",
					},
				},
			},
			expectedError: true,
		},
		{
			name: "Exposed Version missing",
			object: &catalogv1alpha1.DerivedCustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-derivedCustomResourceDefinition",
					Namespace: "test-namespace",
				},
				Spec: catalogv1alpha1.DerivedCustomResourceDefinitionSpec{
					BaseCRD: catalogv1alpha1.ObjectReference{
						Name: "BaseCRD",
					},
					Expose: []catalogv1alpha1.VersionExposeConfig{
						{
							Fields: []catalogv1alpha1.FieldPath{
								{JSONPath: ".spec.prop1"},
							},
						},
					},
				},
			},
			expectedError: true,
		},
		{
			name: "Exposed Field missing",
			object: &catalogv1alpha1.DerivedCustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-derivedCustomResourceDefinition",
					Namespace: "test-namespace",
				},
				Spec: catalogv1alpha1.DerivedCustomResourceDefinitionSpec{
					BaseCRD: catalogv1alpha1.ObjectReference{
						Name: "BaseCRD",
					},
					Expose: []catalogv1alpha1.VersionExposeConfig{
						{
							Versions: []string{
								"v1alpha1",
							},
						},
					},
				},
			},
			expectedError: true,
		},
		{
			name: "can pass validating create",
			object: &catalogv1alpha1.DerivedCustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-derivedCustomResourceDefinition",
					Namespace: "test-namespace",
				},
				Spec: catalogv1alpha1.DerivedCustomResourceDefinitionSpec{
					BaseCRD: catalogv1alpha1.ObjectReference{
						Name: "BaseCRD",
					},
					Expose: []catalogv1alpha1.VersionExposeConfig{
						{
							Versions: []string{
								"v1alpha1",
							},
							Fields: []catalogv1alpha1.FieldPath{
								{JSONPath: ".spec.prop1"},
							},
						},
					},
				},
			},
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedError, derivedCustomResourceDefinitionWebhookHandler.validateCreate(test.object) != nil)
		})
	}
}

func TestDerivedCustomResourceDefinitionValidatingUpdate(t *testing.T) {
	derivedCustomResourceDefinitionWebhookHandler := DerivedCustomResourceDefinitionWebhookHandler{
		Log: testutil.NewLogger(t),
	}

	oldObject := &catalogv1alpha1.DerivedCustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-derivedCustomResourceDefinition",
			Namespace: "test-namespace",
		},
		Spec: catalogv1alpha1.DerivedCustomResourceDefinitionSpec{
			BaseCRD: catalogv1alpha1.ObjectReference{
				Name: "BaseCRD",
			},
			Expose: []catalogv1alpha1.VersionExposeConfig{
				{
					Versions: []string{
						"v1alpha1",
					},
					Fields: []catalogv1alpha1.FieldPath{
						{JSONPath: ".spec.prop1"},
					},
				},
			},
		},
	}

	tests := []struct {
		name          string
		object        *catalogv1alpha1.DerivedCustomResourceDefinition
		expectedError bool
	}{
		{
			name: "BaseCRD immutable",
			object: &catalogv1alpha1.DerivedCustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-derivedCustomResourceDefinition",
					Namespace: "test-namespace",
				},
				Spec: catalogv1alpha1.DerivedCustomResourceDefinitionSpec{
					BaseCRD: catalogv1alpha1.ObjectReference{
						Name: "BaseCRD2",
					},
					Expose: []catalogv1alpha1.VersionExposeConfig{
						{
							Versions: []string{
								"v1alpha1",
							},
							Fields: []catalogv1alpha1.FieldPath{
								{JSONPath: ".spec.prop1"},
							},
						},
					},
				},
			},
			expectedError: true,
		},
		{
			name: "can pass validating update",
			object: &catalogv1alpha1.DerivedCustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-derivedCustomResourceDefinition",
					Namespace: "test-namespace",
				},
				Spec: catalogv1alpha1.DerivedCustomResourceDefinitionSpec{
					BaseCRD: catalogv1alpha1.ObjectReference{
						Name: "BaseCRD",
					},
					Expose: []catalogv1alpha1.VersionExposeConfig{
						{
							Versions: []string{
								"v1alpha1",
							},
							Fields: []catalogv1alpha1.FieldPath{
								{JSONPath: ".spec.prop1"},
								{JSONPath: ".spec.prop2"},
							},
						},
					},
				},
			},
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedError, derivedCustomResourceDefinitionWebhookHandler.validateUpdate(test.object, oldObject) != nil)
		})
	}
}

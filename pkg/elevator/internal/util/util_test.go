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

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
)

var (
	testScheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(corev1.AddToScheme(testScheme))
	utilruntime.Must(catalogv1alpha1.AddToScheme(testScheme))
}

func Test_splitStatusFields(t *testing.T) {
	fields := []catalogv1alpha1.FieldPath{
		{JSONPath: ".status.observedGeneration"},
		{JSONPath: "status.conditions"},
		{JSONPath: ".spec.setting3000"},
		{JSONPath: ".data.something_else"},
	}

	statusFields, otherFields := SplitStatusFields(fields)
	assert.Equal(t, []catalogv1alpha1.FieldPath{
		{JSONPath: ".status.observedGeneration"},
		{JSONPath: "status.conditions"},
	}, statusFields)
	assert.Equal(t, []catalogv1alpha1.FieldPath{
		{JSONPath: ".spec.setting3000"},
		{JSONPath: ".data.something_else"},
	}, otherFields)
}

func Test_copyFields(t *testing.T) {
	tests := []struct {
		name      string
		src, dest *unstructured.Unstructured
		fields    []catalogv1alpha1.FieldPath
		expected  map[string]interface{}
	}{
		{
			name: "default",
			src: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"spec": map[string]interface{}{
						"field1": "test",
						"field2": "test",
						"field3": "test",
					},
					"status": map[string]interface{}{
						"test": int64(4),
					},
				},
			},
			dest: &unstructured.Unstructured{Object: map[string]interface{}{}},
			fields: []catalogv1alpha1.FieldPath{
				{JSONPath: ".spec.field1"},
				{JSONPath: "spec.field2"},
				{JSONPath: "status.test"},
			},
			expected: map[string]interface{}{
				"spec": map[string]interface{}{
					"field1": "test",
					"field2": "test",
				},
				"status": map[string]interface{}{
					"test": int64(4),
				},
			},
		},

		{
			name: "overrides",
			src: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"spec": map[string]interface{}{
						"field1": "another_value",
						"field2": "hans2000",
						"field3": "test",
					},
					"status": map[string]interface{}{
						"test": int64(14),
					},
				},
			},
			dest: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"spec": map[string]interface{}{
						"field1": "456",
						"field2": "t2",
					},
					"status": map[string]interface{}{
						"test": int64(2),
					},
				},
			},
			fields: []catalogv1alpha1.FieldPath{
				{JSONPath: ".spec.field1"},
				{JSONPath: "spec.field2"},
				{JSONPath: "status.test"},
			},
			expected: map[string]interface{}{
				"spec": map[string]interface{}{
					"field1": "another_value",
					"field2": "hans2000",
				},
				"status": map[string]interface{}{
					"test": int64(14),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := CopyFields(test.src, test.dest, test.fields)
			require.NoError(t, err)

			assert.Equal(
				t, test.expected, test.dest.Object)
		})
	}
}

func TestPatch(t *testing.T) {
	tenantObj := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "eu-west-1.provider/v1alpha1",
		"kind":       "CouchDB",
		"metadata": map[string]interface{}{
			"name":      "test-1",
			"namespace": "default",
			"uid":       "uuid-field",
		},
		"spec": map[string]interface{}{
			"test1": "spec2000",
		},
	}}

	specFields := []catalogv1alpha1.FieldPath{{JSONPath: ".spec.test1"}}
	defaults := map[string]interface{}{
		"spec": map[string]interface{}{
			"test2": "test2",
		},
	}
	providerObj := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "eu-west-1.provider/v1alpha1",
		"kind":       "CouchDBInternal",
	}}

	require.NoError(t, BuildProviderObj(tenantObj, providerObj, testScheme, specFields, defaults))

	wantedProviderObj := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "eu-west-1.provider/v1alpha1",
		"kind":       "CouchDBInternal",
		"metadata": map[string]interface{}{
			"name":      "test-1",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"test1": "spec2000",
			"test2": "test2",
		},
	}}
	assert.Equal(t, wantedProviderObj, providerObj)
}

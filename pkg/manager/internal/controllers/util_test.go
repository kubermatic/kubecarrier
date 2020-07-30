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

package controllers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

const exampleObjectSchema string = `
type: object
properties:
  apiVersion:
    type: string
  kind:
    type: string
  spec:
    type: object
    properties:
      prop1:
        type: string
      prop2:
        type: string
        default: test123
      stringArray:
        type: array
        items:
          type: string
      roles:
        type: array
        items:
          properties:
            name:
              type: string
  status:
    type: object
    properties:
      phase:
        type: string
      conditions:
        type: array
        items:
          properties:
            message:
              type: string
            reason:
              type: string
`

func TestExampleObject(t *testing.T) {
	jsonSchema := apiextensionsv1.JSONSchemaProps{}
	require.NoError(t, yaml.Unmarshal([]byte(exampleObjectSchema), &jsonSchema), "unmarshal schema yaml")

	gvk := schema.GroupVersionKind{
		Group:   "test.kubecarrier.io",
		Version: "v1alpha1",
		Kind:    "Test",
	}

	obj, err := exampleObjectFromJSONSchema(gvk, jsonSchema)
	require.NoError(t, err)

	// Resulting Example Object
	// apiVersion: test.kubecarrier.io/v1alpha1
	// kind: Test
	// metadata:
	//   name: test
	// spec:
	//   prop1: ""
	//   prop2: test123
	//   roles:
	//   - name: ""
	//   stringArray: []
	assert.Equal(t, map[string]interface{}{
		"apiVersion": "test.kubecarrier.io/v1alpha1",
		"kind":       "Test",
		"metadata": map[string]interface{}{
			"name": "test",
		},
		"spec": map[string]interface{}{
			"prop1": "",
			"prop2": "test123", // from field defaults
			"roles": []interface{}{
				map[string]interface{}{
					"name": "",
				},
			},
			"stringArray": []interface{}{},
		},
	}, obj.Object)

	j, _ := yaml.Marshal(obj)
	fmt.Println(string(j))
	t.Fail()
}

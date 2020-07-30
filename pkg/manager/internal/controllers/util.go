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
	"encoding/json"
	"strings"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	ServiceClusterLabel  = "kubecarrier.io/service-cluster"
	OriginNamespaceLabel = "kubecarrier.io/origin-namespace"
)

func getStorageVersion(crd *apiextensionsv1.CustomResourceDefinition) string {
	for _, version := range crd.Spec.Versions {
		if version.Storage {
			return version.Name
		}
	}
	return ""
}

// object generic k8s object with metav1 and runtime Object interfaces implemented
type object interface {
	runtime.Object
	metav1.Object
}

func exampleObjectFromJSONSchema(
	gvk schema.GroupVersionKind,
	jsonSchema apiextensionsv1.JSONSchemaProps) (*unstructured.Unstructured, error) {

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{},
	}
	walkExampleObject(jsonSchema, true, obj.Object)
	obj.SetGroupVersionKind(gvk)
	obj.SetName(strings.ToLower(gvk.Kind))

	return obj, nil
}

// walkExampleObject builds an example object from a JSONSchema.
func walkExampleObject(in apiextensionsv1.JSONSchemaProps, firstLevel bool, obj map[string]interface{}) {

	for field, props := range in.Properties {
		if firstLevel {
			// skip common fields added by SetGroupVersionKind
			switch field {
			case "apiVersion", "metadata", "kind", "status":
				continue
			}
		}
		if props.Default != nil {
			var defaultValue interface{}
			_ = json.Unmarshal(props.Default.Raw, &defaultValue)
			obj[field] = defaultValue
			continue
		}

		switch props.Type {
		case "string":
			obj[field] = ""
		case "boolean":
			obj[field] = false
		case "integer", "number":
			obj[field] = 1
		case "array":
			obj[field] = []interface{}{}

			// array sub object
			if props.Items.Schema != nil && props.Items.Schema.Properties != nil {
				item := map[string]interface{}{}
				walkExampleObject(*props.Items.Schema, false, item)
				obj[field] = []interface{}{item}
			}

		case "object":
			obj[field] = map[string]interface{}{}
			walkExampleObject(props, false, obj[field].(map[string]interface{}))
		default:
			obj[field] = nil
		}
	}
}

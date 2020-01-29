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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
)

var (
	testScheme = runtime.NewScheme()

	providerGVK = schema.GroupVersionKind{
		Kind:    "CouchDBInternal",
		Version: "v1alpha1",
		Group:   "eu-west-1.provider",
	}
	providerType = &unstructured.Unstructured{}

	tenantGVK = schema.GroupVersionKind{
		Kind:    "CouchDB",
		Version: "v1alpha1",
		Group:   "eu-west-1.provider",
	}
	tenantType = &unstructured.Unstructured{}

	providerNamespace = "extreme-cloud"
	dcrd              = &catalogv1alpha1.DerivedCustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "couchdbs.eu-west-1",
			Namespace: providerNamespace,
		},
		Spec: catalogv1alpha1.DerivedCustomResourceDefinitionSpec{
			Expose: []catalogv1alpha1.VersionExposeConfig{
				{
					Versions: []string{"v1alpha1"},
					Fields: []catalogv1alpha1.FieldPath{
						{JSONPath: ".spec.test1"},
						{JSONPath: ".status.test1"},
					},
				},
			},
		},
	}
)

func init() {
	// setup scheme for all tests
	if err := corev1.AddToScheme(testScheme); err != nil {
		panic(err)
	}
	if err := catalogv1alpha1.AddToScheme(testScheme); err != nil {
		panic(err)
	}

	// Test Setup
	providerType.SetGroupVersionKind(providerGVK)
	tenantType.SetGroupVersionKind(tenantGVK)
}

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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
)

var (
	testScheme = runtime.NewScheme()

	masterClusterGVK = schema.GroupVersionKind{
		Kind:    "CouchDBInternal",
		Version: "v1alpha1",
		Group:   "eu-west-1.provider",
	}
	masterClusterType = &unstructured.Unstructured{}

	serviceClusterGVK = schema.GroupVersionKind{
		Kind:    "CouchDB",
		Version: "v1alpha1",
		Group:   "couchdb.io",
	}
	serviceClusterType = &unstructured.Unstructured{}

	providerNamespace = "extreme-cloud"
)

func init() {
	// setup scheme for all tests
	if err := corev1.AddToScheme(testScheme); err != nil {
		panic(err)
	}
	if err := corev1alpha1.AddToScheme(testScheme); err != nil {
		panic(err)
	}

	// Test Setup
	masterClusterType.SetGroupVersionKind(masterClusterGVK)
	serviceClusterType.SetGroupVersionKind(serviceClusterGVK)
}

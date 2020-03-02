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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	ownerhelpers "github.com/kubermatic/kubecarrier/pkg/internal/owner"
)

// object generic k8s object with metav1 and runtime Object interfaces implemented
type object interface {
	runtime.Object
	metav1.Object
}

// addOwnerReference adds an OwnerReference to an object.
func addOwnerReference(owner object, object *unstructured.Unstructured, scheme *runtime.Scheme) error {
	switch object.GetKind() {
	case "ClusterRole", "ClusterRoleBinding",
		"CustomResourceDefinition",
		"MutatingWebhookConfiguration", "ValidatingWebhookConfiguration":
		// Non-Namespaced objects
		ownerhelpers.SetOwnerReference(owner, object, scheme)
	default:
		if err := controllerutil.SetControllerReference(owner, object, scheme); err != nil {
			return fmt.Errorf("set ownerReference: %w, obj: %s, %s", err, object.GetKind(), object.GetName())
		}
	}
	return nil
}

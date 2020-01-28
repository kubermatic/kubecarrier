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
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
)

func TestCustomResourceDefinitionDiscoveryReconciler(t *testing.T) {
	const (
		serviceClusterName = "eu-west-1"
	)

	crdRef := &corev1alpha1.CustomResourceDefinitionDiscovery{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster.redis",
			Namespace: "tenant-hans",
		},
		Spec: corev1alpha1.CustomResourceDefinitionDiscoverySpec{
			CRD:            corev1alpha1.ObjectReference{Name: "cluster.redis"},
			ServiceCluster: corev1alpha1.ObjectReference{Name: serviceClusterName},
		},
	}
	crdRefNN := types.NamespacedName{
		Namespace: crdRef.Namespace,
		Name:      crdRef.Name,
	}

	crd := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster.redis",
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "redis",
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{Name: "corev1alpha1"},
			},
			Scope: "Namespaced",
		},
	}

	r := &CustomResourceDefinitionDiscoveryReconciler{
		Client:             fakeclient.NewFakeClientWithScheme(testScheme, crdRef),
		ServiceClient:      fakeclient.NewFakeClientWithScheme(testScheme, crd),
		Scheme:             testScheme,
		ServiceClusterName: serviceClusterName,
	}
	_, _ = r, crdRefNN
}

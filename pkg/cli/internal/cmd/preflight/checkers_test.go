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

package preflight

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	testScheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(appsv1.AddToScheme(testScheme))
	utilruntime.Must(corev1.AddToScheme(testScheme))
}

func TestKubernetesVersionChecker(t *testing.T) {
	tests := []struct {
		name          string
		checker       kubernetesVersionChecker
		expectedError error
	}{
		{
			name: "kubernetes version is higher than first supported version",
			checker: kubernetesVersionChecker{
				kubernetesVersion:     "v1.16.4",
				firstSupportedVersion: firstSupportedKubernetesVersion,
			},
			expectedError: nil,
		},
		{
			name: "kubernetes version equals than first supported version",
			checker: kubernetesVersionChecker{
				kubernetesVersion:     "v1.16.0",
				firstSupportedVersion: firstSupportedKubernetesVersion,
			},
			expectedError: nil,
		},
		{
			name: "kubernetes version is lower than first supported version",
			checker: kubernetesVersionChecker{
				kubernetesVersion:     "v1.15.4",
				firstSupportedVersion: firstSupportedKubernetesVersion,
			},
			expectedError: fmt.Errorf("kubernetes version is lower than the oldest version that KubeCarrier supports, requrires: >= v1.16.0, found: v1.15.4"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedError, test.checker.check())
		})
	}
}

func TestCertManagerChecker(t *testing.T) {
	namespace := &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: certManagerNamespaceName,
		},
	}
	certManagerDeployment := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      certManagerDeploymentName,
			Namespace: certManagerNamespaceName,
		},
		Status: appsv1.DeploymentStatus{
			Conditions: []appsv1.DeploymentCondition{
				{

					Type:   appsv1.DeploymentAvailable,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}
	certManagerWebhookDeployment := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      certManagerWebhookDeploymentName,
			Namespace: certManagerNamespaceName,
		},
		Status: appsv1.DeploymentStatus{
			Conditions: []appsv1.DeploymentCondition{
				{

					Type:   appsv1.DeploymentAvailable,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}
	certManagerCAInjectorDeployment := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      certManagerCAInjectorDeploymentName,
			Namespace: certManagerNamespaceName,
		},
		Status: appsv1.DeploymentStatus{
			Conditions: []appsv1.DeploymentCondition{
				{

					Type:   appsv1.DeploymentAvailable,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}
	unavailableCertManagerCAInjectorDeployment := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      certManagerCAInjectorDeploymentName,
			Namespace: certManagerNamespaceName,
		},
	}
	tests := []struct {
		name          string
		checker       certManagerChecker
		expectedError error
	}{
		{
			name: "check can pass",
			checker: certManagerChecker{
				client:                 fake.NewFakeClientWithScheme(testScheme, namespace, certManagerDeployment, certManagerCAInjectorDeployment, certManagerWebhookDeployment),
				certManagerNamespace:   certManagerNamespaceName,
				certManagerDeployments: []string{certManagerDeploymentName, certManagerCAInjectorDeploymentName, certManagerWebhookDeploymentName},
			},
			expectedError: nil,
		},
		{
			name: "nothing found",
			checker: certManagerChecker{
				client:                 fake.NewFakeClientWithScheme(testScheme),
				certManagerNamespace:   certManagerNamespaceName,
				certManagerDeployments: []string{certManagerDeploymentName, certManagerCAInjectorDeploymentName, certManagerWebhookDeploymentName},
			},
			expectedError: fmt.Errorf(`namespaces "cert-manager" not found
deployments.apps "cert-manager" not found
deployments.apps "cert-manager-cainjector" not found
deployments.apps "cert-manager-webhook" not found
`),
		},
		{
			name: "deployments not found",
			checker: certManagerChecker{
				client:                 fake.NewFakeClientWithScheme(testScheme, namespace),
				certManagerNamespace:   certManagerNamespaceName,
				certManagerDeployments: []string{certManagerDeploymentName, certManagerCAInjectorDeploymentName, certManagerWebhookDeploymentName},
			},
			expectedError: fmt.Errorf(`deployments.apps "cert-manager" not found
deployments.apps "cert-manager-cainjector" not found
deployments.apps "cert-manager-webhook" not found
`),
		},
		{
			name: "deployments not available",
			checker: certManagerChecker{
				client:                 fake.NewFakeClientWithScheme(testScheme, namespace, certManagerDeployment, certManagerWebhookDeployment, unavailableCertManagerCAInjectorDeployment),
				certManagerNamespace:   certManagerNamespaceName,
				certManagerDeployments: []string{certManagerDeploymentName, certManagerCAInjectorDeploymentName, certManagerWebhookDeploymentName},
			},
			expectedError: fmt.Errorf(`deployment cert-manager-cainjector is not available
`),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedError, test.checker.check())
		})
	}
}

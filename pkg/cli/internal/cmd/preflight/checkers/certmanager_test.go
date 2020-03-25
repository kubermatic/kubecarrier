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

package checkers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestCertManagerChecker(t *testing.T) {
	certManagerCertificatesCRD := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: v1.ObjectMeta{
			Name: certManagerCertificatesCRDName,
		},
		Status: apiextensionsv1.CustomResourceDefinitionStatus{
			Conditions: []apiextensionsv1.CustomResourceDefinitionCondition{
				{

					Type:   apiextensionsv1.Established,
					Status: apiextensionsv1.ConditionTrue,
				},
			},
		},
	}
	certManagerIssuersCRD := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: v1.ObjectMeta{
			Name: certManagerIssuersCRDName,
		},
		Status: apiextensionsv1.CustomResourceDefinitionStatus{
			Conditions: []apiextensionsv1.CustomResourceDefinitionCondition{
				{

					Type:   apiextensionsv1.Established,
					Status: apiextensionsv1.ConditionTrue,
				},
			},
		},
	}
	unestablishedCertManagerIssuersCRD := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: v1.ObjectMeta{
			Name: certManagerIssuersCRDName,
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
				client:          fake.NewFakeClientWithScheme(testScheme, certManagerCertificatesCRD, certManagerIssuersCRD),
				certManagerCRDs: []string{certManagerIssuersCRDName, certManagerCertificatesCRDName},
			},
			expectedError: nil,
		},
		{
			name: "nothing found",
			checker: certManagerChecker{
				client:          fake.NewFakeClientWithScheme(testScheme),
				certManagerCRDs: []string{certManagerIssuersCRDName, certManagerCertificatesCRDName},
			},
			expectedError: fmt.Errorf(`customresourcedefinitions.apiextensions.k8s.io "issuers.cert-manager.io" not found
customresourcedefinitions.apiextensions.k8s.io "certificates.cert-manager.io" not found
`),
		},
		{
			name: "crd not established",
			checker: certManagerChecker{
				client:          fake.NewFakeClientWithScheme(testScheme, certManagerCertificatesCRD, unestablishedCertManagerIssuersCRD),
				certManagerCRDs: []string{certManagerIssuersCRDName, certManagerCertificatesCRDName},
			},
			expectedError: fmt.Errorf(`crd issuers.cert-manager.io is not established
`),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedError, test.checker.check())
		})
	}
}

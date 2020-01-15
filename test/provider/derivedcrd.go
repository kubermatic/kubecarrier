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

package provider

import (
	"context"

	"github.com/stretchr/testify/suite"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
	"github.com/kubermatic/kubecarrier/test/framework"
	"github.com/kubermatic/kubecarrier/test/helpers"
)

var (
	_ suite.SetupAllSuite    = (*DerivedCRDSuite)(nil)
	_ suite.TearDownAllSuite = (*DerivedCRDSuite)(nil)
)

var derivedCRDBaseCRD = &apiextensionsv1.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{
		Name: "catapults.test.kubecarrier.io",
		Labels: map[string]string{
			"kubecarrier.io/service-cluster": "eu-west-1",
			"kubecarrier.io/provider":        "dcrd",
		},
	},
	Spec: apiextensionsv1.CustomResourceDefinitionSpec{
		Group: "test.kubecarrier.io",
		Names: apiextensionsv1.CustomResourceDefinitionNames{
			Kind:     "Catapult",
			ListKind: "CatapultList",
			Plural:   "catapults",
			Singular: "catapult",
		},
		Scope: apiextensionsv1.NamespaceScoped,
		Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
			{
				Name:    "v1alpha1",
				Served:  true,
				Storage: true,
				Schema: &apiextensionsv1.CustomResourceValidation{
					OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
						Properties: map[string]apiextensionsv1.JSONSchemaProps{
							"apiVersion": {Type: "string"},
							"kind":       {Type: "string"},
							"metadata":   {Type: "object"},
							"spec": {
								Type: "object",
								Properties: map[string]apiextensionsv1.JSONSchemaProps{
									"prop1": {Type: "string"},
									"prop2": {Type: "string"},
								},
							},
						},
						Type: "object",
					},
				},
			},
		},
	},
}

type DerivedCRDSuite struct {
	suite.Suite
	*framework.Framework
	helpers.ProviderMixin

	masterClient client.Client
	provider     *catalogv1alpha1.Provider
	crd          *apiextensionsv1.CustomResourceDefinition
}

func (s *DerivedCRDSuite) SetupSuite() {
	var err error
	s.masterClient, err = s.MasterClient()
	s.Require().NoError(err, "creating master client")

	// init provider mixin
	s.ProviderMixin.Client = s.masterClient
	s.ProviderMixin.Suite = s.Suite

	// Create a Provider to execute our tests in
	s.provider = s.CreateProvider("dcrd")

	// Create example CRD
	s.crd = derivedCRDBaseCRD.DeepCopy()
	ctx := context.Background()
	s.Require().NoError(s.masterClient.Create(ctx, s.crd), "creating test CRD")
}

func (s *DerivedCRDSuite) TearDownSuite() {
	ctx := context.Background()
	s.DeleteProvider(s.provider)
	s.Require().NoError(s.masterClient.Delete(ctx, s.crd), "deleting test CRD")
}

func (s *DerivedCRDSuite) TestDerivedCRD() {
	ctx := context.Background()

	dcrd := &catalogv1alpha1.DerivedCustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: s.provider.Status.NamespaceName,
		},
		Spec: catalogv1alpha1.DerivedCustomResourceDefinitionSpec{
			BaseCRD: catalogv1alpha1.ObjectReference{
				Name: s.crd.Name,
			},
			KindOverride: "TestResource",
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
	s.Require().NoError(s.masterClient.Create(ctx, dcrd), "creating DerivedCustomResourceDefinition")

	// Wait for DCRD to be ready
	s.Require().NoError(testutil.WaitUntilReady(s.masterClient, dcrd))

	// Check reported status
	if s.NotNil(dcrd.Status.DerivedCRD, ".status.derivedCRD should be set") {
		s.Equal("testresources.eu-west-1.dcrd", dcrd.Status.DerivedCRD.Name)
		s.Equal("eu-west-1.dcrd", dcrd.Status.DerivedCRD.Group)
		s.Equal("TestResource", dcrd.Status.DerivedCRD.Kind)
		s.Equal("testresources", dcrd.Status.DerivedCRD.Plural)
		s.Equal("testresource", dcrd.Status.DerivedCRD.Singular)
	}

	// Check created CRD
	crd := &apiextensionsv1.CustomResourceDefinition{}
	s.Require().NoError(s.masterClient.Get(ctx, types.NamespacedName{
		Name: dcrd.Status.DerivedCRD.Name,
	}, crd), "getting derived CRD")

	schemaYaml, _ := yaml.Marshal(crd.Spec.Versions[0].Schema.OpenAPIV3Schema)
	s.Equal(`properties:
  apiVersion:
    type: string
  kind:
    type: string
  metadata:
    type: object
  spec:
    properties:
      prop1:
        type: string
    type: object
type: object
`, string(schemaYaml))

	// Cleanup
	s.Require().NoError(s.masterClient.Delete(ctx, dcrd), "deleting the DerivedCustomResourceDefinition object")

	// Wait for DerivedCustomResourceDefinition to be gone
	s.Require().NoError(testutil.WaitUntilNotFound(s.masterClient, dcrd))
}

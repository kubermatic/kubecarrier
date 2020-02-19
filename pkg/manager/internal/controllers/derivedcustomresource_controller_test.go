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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func Test_DerivedCustomResourceReconciler(t *testing.T) {
	baseCRD := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "catapults.test.kubecarrier.io",
			Labels: map[string]string{
				"kubecarrier.io/service-cluster": "eu-west-1",
				"kubecarrier.io/provider":        "dcr",
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
	baseCRD.Status.AcceptedNames = baseCRD.Spec.Names

	provider := &catalogv1alpha1.Account{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dcr",
		},
		Spec: catalogv1alpha1.AccountSpec{
			Roles: []catalogv1alpha1.AccountRole{
				catalogv1alpha1.ProviderRole,
			},
		},
		Status: catalogv1alpha1.AccountStatus{
			NamespaceName: "provider-dcr",
		},
	}

	providerNS := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "provider-dcr"}}
	_, err := util.InsertOwnerReference(provider, providerNS, testScheme)
	require.NoError(t, err)

	derivedCR := &catalogv1alpha1.DerivedCustomResource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: provider.Status.NamespaceName,
		},
		Spec: catalogv1alpha1.DerivedCustomResourceSpec{
			BaseCRD: catalogv1alpha1.ObjectReference{
				Name: baseCRD.Name,
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

	t.Run("Reconcile", func(t *testing.T) {
		derivedCR := derivedCR.DeepCopy()

		client := fakeclient.NewFakeClientWithScheme(testScheme, baseCRD, provider, derivedCR, providerNS)
		log := testutil.NewLogger(t)
		r := &DerivedCustomResourceReconciler{
			Client: client,
			Log:    log,
			Scheme: testScheme,
		}
		ctx := context.Background()

		// 1st Reconcile call
		// creating the derived CRD
		_, err := r.Reconcile(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      derivedCR.Name,
				Namespace: derivedCR.Namespace,
			},
		})
		require.NoError(t, err)

		// Check CRD
		checkDerivedCR := &apiextensionsv1.CustomResourceDefinition{}
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name: "testresources.eu-west-1.dcr",
		}, checkDerivedCR))

		schemaYaml, _ := yaml.Marshal(checkDerivedCR.Spec.Versions[0].Schema.OpenAPIV3Schema)
		assert.Equal(t, `properties:
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

		assert.Equal(t, "eu-west-1.dcr", checkDerivedCR.Spec.Group)
		assert.Equal(t, "TestResource", checkDerivedCR.Spec.Names.Kind)
		assert.Equal(t, "testresources", checkDerivedCR.Spec.Names.Plural)
		assert.Equal(t, "testresource", checkDerivedCR.Spec.Names.Singular)

		// check ready condition
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name:      derivedCR.Name,
			Namespace: derivedCR.Namespace,
		}, derivedCR))

		readyCond, ok := derivedCR.Status.GetCondition(catalogv1alpha1.DerivedCustomResourceReady)
		if assert.True(t, ok, "ready condition should be set") {
			assert.Equal(t, catalogv1alpha1.ConditionFalse, readyCond.Status)
			assert.Equal(t, "CRDNotEstablished", readyCond.Reason)
		}
		crdCond, ok := derivedCR.Status.GetCondition(catalogv1alpha1.DerivedCustomResourceEstablished)
		if assert.True(t, ok, "ready condition should be set") {
			assert.Equal(t, catalogv1alpha1.ConditionFalse, crdCond.Status)
			assert.Equal(t, "Establishing", crdCond.Reason)
		}

		// 2nd Reconcile call
		// after CRD is established
		checkDerivedCR.Status.AcceptedNames = checkDerivedCR.Spec.Names
		checkDerivedCR.Status.Conditions = []apiextensionsv1.CustomResourceDefinitionCondition{
			{
				Type:   apiextensionsv1.Established,
				Status: apiextensionsv1.ConditionTrue,
			},
		}
		require.NoError(t, client.Status().Update(ctx, checkDerivedCR))

		_, err = r.Reconcile(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      derivedCR.Name,
				Namespace: derivedCR.Namespace,
			},
		})
		require.NoError(t, err)

		// check ready condition
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name:      derivedCR.Name,
			Namespace: derivedCR.Namespace,
		}, derivedCR))

		readyCond, ok = derivedCR.Status.GetCondition(catalogv1alpha1.DerivedCustomResourceEstablished)
		if assert.True(t, ok, "ready condition should be set") {
			assert.Equal(t, catalogv1alpha1.ConditionTrue, readyCond.Status)
			assert.Equal(t, "Established", readyCond.Reason)
		}

		// Check Elevator
		checkElevator := &operatorv1alpha1.Elevator{}
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name:      derivedCR.Name,
			Namespace: derivedCR.Namespace,
		}, checkElevator))
		assert.Equal(t, operatorv1alpha1.ElevatorSpec{
			ProviderCRD: operatorv1alpha1.CRDReference{
				Kind:    "Catapult",
				Version: "v1alpha1",
				Group:   "test.kubecarrier.io",
				Plural:  "catapults",
			},
			TenantCRD: operatorv1alpha1.CRDReference{
				Kind:    "TestResource",
				Version: "v1alpha1",
				Group:   "eu-west-1.dcr",
				Plural:  "testresources",
			},
			DerivedCR: operatorv1alpha1.ObjectReference{
				Name: derivedCR.Name,
			},
		}, checkElevator.Spec)

		// 3rd Reconcile call
		// after Elevator is ready
		checkElevator.Status.Conditions = []operatorv1alpha1.ElevatorCondition{
			{
				Type:   operatorv1alpha1.ElevatorReady,
				Status: operatorv1alpha1.ConditionTrue,
			},
		}
		require.NoError(t, client.Status().Update(ctx, checkElevator))

		_, err = r.Reconcile(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      derivedCR.Name,
				Namespace: derivedCR.Namespace,
			},
		})
		require.NoError(t, err)

		// check ready condition
		require.NoError(t, client.Get(ctx, types.NamespacedName{
			Name:      derivedCR.Name,
			Namespace: derivedCR.Namespace,
		}, derivedCR))

		controllerCond, ok := derivedCR.Status.GetCondition(catalogv1alpha1.DerivedCustomResourceControllerReady)
		if assert.True(t, ok, "ready condition should be set") {
			assert.Equal(t, catalogv1alpha1.ConditionTrue, controllerCond.Status)
			assert.Equal(t, "Ready", controllerCond.Reason)
		}
		readyCond, ok = derivedCR.Status.GetCondition(catalogv1alpha1.DerivedCustomResourceReady)
		if assert.True(t, ok, "ready condition should be set") {
			assert.Equal(t, catalogv1alpha1.ConditionTrue, readyCond.Status)
			assert.Equal(t, "ComponentsReady", readyCond.Reason)
		}
		crdCond, ok = derivedCR.Status.GetCondition(catalogv1alpha1.DerivedCustomResourceEstablished)
		if assert.True(t, ok, "ready condition should be set") {
			assert.Equal(t, catalogv1alpha1.ConditionTrue, crdCond.Status)
			assert.Equal(t, "Established", crdCond.Reason)
		}
	})
}

func Test_dummyObject(t *testing.T) {
	t.Run("IsArray", func(t *testing.T) {
		assert.True(t, dummyObject{
			arrayKey: {},
		}.IsArray(), "IsArray should be true")
	})
}

const testObjectSchema string = `
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

func Test_walkDummyObject(t *testing.T) {
	schema := apiextensionsv1.JSONSchemaProps{}
	require.NoError(t, yaml.Unmarshal([]byte(testObjectSchema), &schema), "unmarshal schema yaml")

	obj := dummyObject{}
	walkDummyObject(schema, obj)

	assert.Equal(t, dummyObject{
		"apiVersion": {},
		"kind":       {},
		"spec": {
			"prop1": {},
			"prop2": {},
		},
		"status": {
			"phase": {},
			"conditions": {
				arrayKey: {
					"message": {},
					"reason":  {},
				},
			},
		},
	}, obj)
}

func Test_filterSchema(t *testing.T) {
	schema := apiextensionsv1.JSONSchemaProps{}
	require.NoError(t, yaml.Unmarshal([]byte(testObjectSchema), &schema), "unmarshal schema yaml")

	exposeConfig := catalogv1alpha1.VersionExposeConfig{
		Fields: []catalogv1alpha1.FieldPath{
			{JSONPath: ".spec.prop1"},
			{JSONPath: ".status.phase"},
			{JSONPath: ".status.conditions[].message"},
		},
	}

	filteredSchema, err := filterSchema(schema, exposeConfig)
	require.NoError(t, err)

	filteredSchameYaml, err := yaml.Marshal(filteredSchema)
	require.NoError(t, err, "remarshal filtered schema")

	assert.Equal(t, `properties:
  apiVersion:
    type: string
  kind:
    type: string
  spec:
    properties:
      prop1:
        type: string
    type: object
  status:
    properties:
      conditions:
        items:
          properties:
            message:
              type: string
        type: array
      phase:
        type: string
    type: object
`, string(filteredSchameYaml))
}

func Test_markDummyObject(t *testing.T) {
	exposeConfig := catalogv1alpha1.VersionExposeConfig{
		Fields: []catalogv1alpha1.FieldPath{
			{JSONPath: ".kind"},
			{JSONPath: ".status.phase"},
			{JSONPath: ".status.conditions[].message"},
		},
	}
	marked, err := markDummyObject(exposeConfig, dummyObject{
		"apiVersion": {},
		"kind":       {},
		"metadata":   {},
		"status": {
			"phase": {},
			"conditions": {
				arrayKey: {
					"message": {},
					"reason":  {},
				},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, dummyObject{
		"apiVersion": nil,
		"kind":       nil,
		"metadata":   nil,
		"status": {
			"phase": nil,
			"conditions": {
				arrayKey: {
					"message": nil,
					"reason":  {},
				},
			},
		},
	}, marked)
}

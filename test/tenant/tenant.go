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

package tenant

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
	"github.com/kubermatic/kubecarrier/test/framework"
)

func NewTenantSuite(f *framework.Framework) func(t *testing.T) {
	return func(t *testing.T) {
		managementClient, err := f.ManagementClient()
		require.NoError(t, err, "creating management client")
		defer managementClient.CleanUp(t)
		ctx := context.Background()
		provider := &catalogv1alpha1.Account{
			ObjectMeta: metav1.ObjectMeta{
				Name: "other-cloud",
			},
			Spec: catalogv1alpha1.AccountSpec{
				Metadata: catalogv1alpha1.AccountMetadata{
					DisplayName: "provider1",
					Description: "provider1 test description",
				},
				Roles: []catalogv1alpha1.AccountRole{
					catalogv1alpha1.ProviderRole,
				},
			},
		}
		require.NoError(t, managementClient.Create(ctx, provider))
		require.NoError(t, testutil.WaitUntilReady(managementClient, provider))

		crd := &apiextensionsv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "couchdbs.eu-east-1.example.cloud",
				Labels: map[string]string{
					"kubecarrier.io/origin-namespace": provider.Status.Namespace.Name,
					"kubecarrier.io/service-cluster":  "eu-east-1",
				},
			},
			Spec: apiextensionsv1.CustomResourceDefinitionSpec{
				Group: "eu-east-1.example.cloud",
				Names: apiextensionsv1.CustomResourceDefinitionNames{
					Plural: "couchdbs",
					Kind:   "CouchDB",
				},
				Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
					{
						Name:    "v1alpha1",
						Storage: true,
						Schema: &apiextensionsv1.CustomResourceValidation{
							OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
								Type: "object",
							},
						},
					},
				},
				Scope: apiextensionsv1.NamespaceScoped,
			},
		}

		require.NoError(t, managementClient.Create(ctx, crd))

		catalogEntry := &catalogv1alpha1.CatalogEntry{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "couchdbs",
				Namespace: provider.Status.Namespace.Name,
			},
			Spec: catalogv1alpha1.CatalogEntrySpec{
				Metadata: catalogv1alpha1.CatalogEntryMetadata{
					DisplayName: "Couch DB",
					Description: "The comfy nosql database",
				},
				BaseCRD: catalogv1alpha1.ObjectReference{
					Name: crd.Name,
				},
			},
		}
		require.NoError(t, managementClient.Create(ctx, catalogEntry), "creating catalogEntry error")
		assert.NoError(t, testutil.WaitUntilReady(managementClient, catalogEntry), "catalog entry not ready within time limit")
	}
}

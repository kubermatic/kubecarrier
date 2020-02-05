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
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
	"github.com/kubermatic/kubecarrier/test/framework"
)

func NewFerrySuite(
	f *framework.Framework,
	provider *catalogv1alpha1.Provider,
) func(t *testing.T) {
	return func(t *testing.T) {
		serviceClient, err := f.ServiceClient()
		require.NoError(t, err, "creating service client")
		defer serviceClient.CleanUp(t)

		masterClient, err := f.MasterClient()
		require.NoError(t, err, "creating master client")
		defer masterClient.CleanUp(t)

		ctx := context.Background()
		serviceKubeconfig, err := ioutil.ReadFile(f.Config().ServiceInternalKubeconfigPath)
		require.NoError(t, err, "cannot read service internal kubeconfig")

		var (
			tenant = &catalogv1alpha1.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "otto",
					Namespace: "kubecarrier-system",
				},
			}
			serviceClusterSecret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "us-east-1",
					Namespace: provider.Status.NamespaceName,
				},
				Data: map[string][]byte{
					"kubeconfig": serviceKubeconfig,
				},
			}
			serviceClusterRegistration = &operatorv1alpha1.ServiceClusterRegistration{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "us-east-1",
					Namespace: provider.Status.NamespaceName,
				},
				Spec: operatorv1alpha1.ServiceClusterRegistrationSpec{
					KubeconfigSecret: operatorv1alpha1.ObjectReference{
						Name: "us-east-1",
					},
				},
			}
			crd = &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: "redis.test.kubecarrier.io",
				},
				Spec: apiextensionsv1.CustomResourceDefinitionSpec{
					Group: "test.kubecarrier.io",
					Names: apiextensionsv1.CustomResourceDefinitionNames{
						Singular: "redis",
						Plural:   "redis",
						Kind:     "Redis",
						ListKind: "RedisList",
					},
					Scope: "Namespaced",
					Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
						{
							Name:    "v1alpha1",
							Served:  true,
							Storage: true,
							Schema: &apiextensionsv1.CustomResourceValidation{
								OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
									Description: "schema",
									Type:        "object",
								},
							},
						},
					},
				},
			}
		)

		require.NoError(t, masterClient.Create(ctx, tenant))
		require.NoError(t, testutil.WaitUntilReady(masterClient, tenant))

		require.NoError(t, masterClient.Create(ctx, serviceClusterSecret))
		require.NoError(t, masterClient.Create(ctx, serviceClusterRegistration))
		require.NoError(t, testutil.WaitUntilReady(masterClient, serviceClusterRegistration), "service cluster not ready")

		require.NoError(t, serviceClient.Create(ctx, crd))

		// Check if the ServiceCluster becomes ready
		//
		serviceCluster := &corev1alpha1.ServiceCluster{}
		serviceCluster.SetName(serviceClusterRegistration.GetName())
		serviceCluster.SetNamespace(provider.Status.NamespaceName)
		require.NoError(t, testutil.WaitUntilReady(masterClient, serviceCluster))

		t.Run("parallel-group", func(t *testing.T) {
			t.Run("CustomResourceDefinitionDiscovery", func(t *testing.T) {
				t.Parallel()

				crdd := &corev1alpha1.CustomResourceDefinitionDiscovery{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "redis",
						Namespace: provider.Status.NamespaceName,
					},
					Spec: corev1alpha1.CustomResourceDefinitionDiscoverySpec{
						CRD: corev1alpha1.ObjectReference{
							Name: crd.GetName(),
						},
						ServiceCluster: corev1alpha1.ObjectReference{
							Name: serviceClusterRegistration.GetName(),
						},
					},
				}
				// make sure the object is not present, before we try to create it
				require.NoError(t, client.IgnoreNotFound(masterClient.Delete(ctx, crdd)))
				require.NoError(t, testutil.WaitUntilNotFound(masterClient, crdd))

				// create CustomResourceDefinitionDiscovery object
				require.NoError(t, masterClient.Create(ctx, crdd))
				if assert.NoError(t, testutil.WaitUntilReady(masterClient, crdd)) {
					assert.Equal(t, crd.Name, crdd.Status.CRD.Name)
				}

				// wait till ready
				if assert.NoError(t, testutil.WaitUntilReady(masterClient, crdd)) {
					assert.Equal(t, crd.Name, crdd.Status.CRD.Name)
				}
			})

			t.Run("ServiceClusterAssignment", func(t *testing.T) {
				t.Parallel()
				serviceClusterAssignment := &corev1alpha1.ServiceClusterAssignment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      tenant.Name + "." + serviceClusterRegistration.Name,
						Namespace: provider.Status.NamespaceName,
					},
					Spec: corev1alpha1.ServiceClusterAssignmentSpec{
						ServiceCluster: corev1alpha1.ObjectReference{
							Name: serviceClusterRegistration.Name,
						},
					},
				}
				require.NoError(t, client.IgnoreNotFound(masterClient.Delete(ctx, serviceClusterAssignment)))
				require.NoError(t, testutil.WaitUntilNotFound(masterClient, serviceClusterAssignment))
				require.NoError(t, masterClient.Create(ctx, serviceClusterAssignment))

				ns := &corev1.Namespace{}
				if assert.NoError(t, testutil.WaitUntilReady(masterClient, serviceClusterAssignment), "serviceClusterAssignment never became ready") {
					assert.NoError(t,
						serviceClient.Get(
							ctx,
							types.NamespacedName{Name: serviceClusterAssignment.Status.ServiceClusterNamespace.Name},
							ns,
						),
						"serviceCluster's namespace not created",
					)
				}

				assert.NoError(t, client.IgnoreNotFound(masterClient.Delete(ctx, serviceClusterAssignment)))
				assert.NoError(t, testutil.WaitUntilNotFound(masterClient, serviceClusterAssignment), "serviceClusterAssignment never succesfully got deleted")

				if ns.Name != "" {
					assert.NoError(t, testutil.WaitUntilNotFound(serviceClient, ns), "serviceCluster's namespace not cleared")
				}
			})
		})
	}
}

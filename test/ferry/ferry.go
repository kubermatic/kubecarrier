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

package ferry

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
	"github.com/kubermatic/kubecarrier/test/framework"
)

func NewFerrySuite(f *framework.Framework) func(t *testing.T) {
	return func(t *testing.T) {
		serviceClient, err := f.ServiceClient()
		require.NoError(t, err, "creating service client")
		masterClient, err := f.MasterClient()
		require.NoError(t, err, "creating master client")

		serviceKubeconfig, err := ioutil.ReadFile(f.Config().ServiceInternalKubeconfigPath)
		require.NoError(t, err, "cannot read service internal kubeconfig")

		ctx := context.Background()

		var (
			provider = &catalogv1alpha1.Provider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "steel-inquisitor",
					Namespace: "kubecarrier-system",
				},
			}
			tenant = &catalogv1alpha1.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "otto",
					Namespace: "kubecarrier-system",
				},
			}
			serviceClusterSecret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "eu-west-1",
				},
				Data: map[string][]byte{
					"kubeconfig": serviceKubeconfig,
				},
			}
			serviceClusterRegistration = &operatorv1alpha1.ServiceClusterRegistration{
				ObjectMeta: metav1.ObjectMeta{
					Name: "eu-west-1",
				},
				Spec: operatorv1alpha1.ServiceClusterRegistrationSpec{
					KubeconfigSecret: operatorv1alpha1.ObjectReference{
						Name: "eu-west-1",
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
						Plural: "redis",
						Kind:   "Redis",
					},
					Scope: "Namespaced",
					Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
						{
							Name:    "corev1alpha1",
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

		cleanUp := func() {
			for _, obj := range []runtime.Object{
				provider,
				// These two are automatically deleted when the provider is deleted...at least they should be
				// serviceClusterSecret,
				// serviceClusterRegistration,
				tenant,
			} {
				require.NoError(t, client.IgnoreNotFound(masterClient.Delete(ctx, obj.DeepCopyObject())))
				require.NoError(t, testutil.WaitUntilNotFound(masterClient, obj.DeepCopyObject()))
			}

			for _, obj := range []runtime.Object{
				crd,
			} {
				require.NoError(t, client.IgnoreNotFound(serviceClient.Delete(ctx, obj.DeepCopyObject())))
				require.NoError(t, testutil.WaitUntilNotFound(serviceClient, obj.DeepCopyObject()))
			}

		}
		// clean up before and after completion
		cleanUp()
		// defer cleanUp()

		require.NoError(t, masterClient.Create(ctx, provider))
		require.NoError(t, masterClient.Create(ctx, tenant))

		require.NoError(t, testutil.WaitUntilReady(masterClient, provider))
		require.NoError(t, testutil.WaitUntilReady(masterClient, tenant))
		for _, obj := range []metav1.Object{
			serviceClusterSecret,
			serviceClusterRegistration,
		} {
			obj.SetNamespace(provider.Status.NamespaceName)
		}

		require.NoError(t, masterClient.Create(ctx, serviceClusterSecret))
		require.NoError(t, masterClient.Create(ctx, serviceClusterRegistration))
		require.NoError(t, testutil.WaitUntilReady(masterClient, serviceClusterRegistration))
		require.NoError(t, serviceClient.Create(ctx, crd))

		t.Run("ServiceCluster", func(t *testing.T) {
			t.Parallel()
			serviceCluster := &corev1alpha1.ServiceCluster{}
			serviceCluster.SetName(serviceClusterRegistration.GetName())
			serviceCluster.SetNamespace(provider.Status.NamespaceName)
			require.NoError(t, testutil.WaitUntilReady(masterClient, serviceCluster))
		})

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
			require.NoError(t, client.IgnoreNotFound(masterClient.Delete(ctx, crdd)))
			require.NoError(t, testutil.WaitUntilNotFound(masterClient, crdd))

			require.NoError(t, masterClient.Create(ctx, crdd))
			if assert.NoError(t, testutil.WaitUntilReady(masterClient, crdd)) {
				assert.Equal(t, crd.Name, crdd.Status.CRD.Name)
			}

			// clean up
			assert.NoError(t, masterClient.Delete(ctx, crdd))
			assert.NoError(t, testutil.WaitUntilNotFound(masterClient, crdd))
		})
		t.Run("TenantAssignment", func(t *testing.T) {
			t.Parallel()
			tenantAssignment := &corev1alpha1.TenantAssignment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tenant.Name + "." + serviceClusterRegistration.Name,
					Namespace: provider.Status.NamespaceName,
				},
				Spec: corev1alpha1.TenantAssignmentSpec{
					Tenant: corev1alpha1.ObjectReference{
						Name: tenant.Name,
					},
					ServiceCluster: corev1alpha1.ObjectReference{
						Name: serviceClusterRegistration.Name,
					},
				},
			}
			require.NoError(t, client.IgnoreNotFound(masterClient.Delete(ctx, tenantAssignment)))
			require.NoError(t, testutil.WaitUntilNotFound(masterClient, tenantAssignment))
			require.NoError(t, masterClient.Create(ctx, tenantAssignment))

			ns := &corev1.Namespace{}
			if assert.NoError(t, testutil.WaitUntilReady(masterClient, tenantAssignment)) {
				assert.NoError(t,
					serviceClient.Get(
						ctx,
						types.NamespacedName{Name: tenantAssignment.Status.NamespaceName},
						ns,
					),
				)
			}

			assert.NoError(t, client.IgnoreNotFound(masterClient.Delete(ctx, tenantAssignment)))
			assert.NoError(t, testutil.WaitUntilNotFound(masterClient, tenantAssignment))

			if ns.Name != "" {
				assert.NoError(t, testutil.WaitUntilNotFound(serviceClient, ns))
			}
		})
	}
}

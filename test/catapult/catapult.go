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

package catapult

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/test/framework"
)

func NewCatapultSuit(f *framework.Framework) func(t *testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()
		fctx, err := f.NewFrameworkContext()
		require.NoError(t, err, "cannot create framework context")
		defer fctx.CleanUp(t)

		serviceKubeconfig, err := ioutil.ReadFile(f.Config().ServiceInternalKubeconfigPath)
		require.NoError(t, err, "cannot read service internal kubeconfig")

		var (
			provider = &catalogv1alpha1.Provider{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "provider-",
					Namespace:    "kubecarrier-system",
				},
			}
			tenant = &catalogv1alpha1.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "tenant-",
					Namespace:    "kubecarrier-system",
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

		require.NoError(t, fctx.MasterClient.CreateAndWaitUntilReady(ctx, provider))
		require.NoError(t, fctx.MasterClient.CreateAndWaitUntilReady(ctx, tenant))

		for _, obj := range []metav1.Object{
			serviceClusterSecret,
			serviceClusterRegistration,
		} {
			obj.SetNamespace(provider.Status.NamespaceName)
		}

		require.NoError(t, fctx.MasterClient.Create(ctx, serviceClusterSecret))
		require.NoError(t, fctx.MasterClient.CreateAndWaitUntilReady(ctx, serviceClusterRegistration))
		require.NoError(t, fctx.ServiceClient.Create(ctx, crd))

		serviceClusterAssignment := &corev1alpha1.ServiceClusterAssignment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tenant.Name,
				Namespace: provider.Namespace,
			},
			Spec: corev1alpha1.ServiceClusterAssignmentSpec{
				ServiceCluster: corev1alpha1.ObjectReference{
					serviceClusterRegistration.Name,
				},
			},
		}
		require.NoError(t, fctx.MasterClient.CreateAndWaitUntilReady(ctx, serviceClusterAssignment))
		catapult := &operatorv1alpha1.Catapult{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-catapult",
				Namespace: provider.GetNamespace(),
			},
			Spec: operatorv1alpha1.CatapultSpec{
				ServiceCluster: operatorv1alpha1.ObjectReference{
					Name: serviceClusterRegistration.Name,
				},
				CatapultMappingSpec: operatorv1alpha1.CatapultMappingSpec{
					MasterGroup:   "", // legacy for core group
					MasterKind:    "ConfigMap",
					ServiceGroup:  "", // legacy for core group
					ServiceKind:   "ConfigMap",
					ObjectVersion: "v1",
				},
			},
		}
		require.NoError(t, fctx.MasterClient.CreateAndWaitUntilReady(ctx, catapult))
	}
}

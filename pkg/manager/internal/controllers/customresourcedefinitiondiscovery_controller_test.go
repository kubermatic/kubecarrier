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
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

type providerGetterMock struct {
	provider *v1alpha1.Provider
}

func (p providerGetterMock) GetProviderByProviderNamespace(ctx context.Context, c client.Client, namespace string) (*v1alpha1.Provider, error) {
	if p.provider == nil {
		return nil, fmt.Errorf("notSet")
	}
	return p.provider, nil
}

func TestCustomResourceDefinitionDiscoveryReconciler(t *testing.T) {
	const (
		serviceClusterName = "eu-west-1"
	)
	provider := &v1alpha1.Provider{
		ObjectMeta: metav1.ObjectMeta{
			Name: "extreme-cloud",
		},
	}

	crdDiscovery := &corev1alpha1.CustomResourceDefinitionDiscovery{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "redis.cloud",
			Namespace: "tenant-hans",
		},
		Spec: corev1alpha1.CustomResourceDefinitionDiscoverySpec{
			CRD:            corev1alpha1.ObjectReference{Name: "redis.cloud"},
			ServiceCluster: corev1alpha1.ObjectReference{Name: serviceClusterName},
		},
	}
	crdDiscoveryNN := types.NamespacedName{
		Namespace: crdDiscovery.Namespace,
		Name:      crdDiscovery.Name,
	}

	crd := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "redis.cloud",
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "cloud",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Plural:   "redis",
				Singular: "redis",
				Kind:     "Redis",
				ListKind: "RedisList",
			},
			Scope: "Namespaced",
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{Name: "corev1alpha1"},
			},
		},
	}
	r := &CustomResourceDefinitionDiscoveryReconciler{
		Log:            testutil.NewLogger(t),
		Client:         fakeclient.NewFakeClientWithScheme(testScheme, crdDiscovery, provider),
		Scheme:         testScheme,
		ProviderGetter: providerGetterMock{provider: provider},
	}
	ctx := context.Background()
	reconcileLoop := func() {
		for i := 0; i < 3; i++ {
			_, err := r.Reconcile(ctrl.Request{
				NamespacedName: crdDiscoveryNN,
			})
			require.NoError(t, err)
			require.NoError(t, r.Client.Get(ctx, crdDiscoveryNN, crdDiscovery))
		}
	}

	reconcileLoop()
	cond, ok := crdDiscovery.Status.GetCondition(corev1alpha1.CustomResourceDefinitionDiscoveryDiscovered)
	assert.True(t, ok, "CustomResourceDefinitionDiscoveryDiscovered condition should be present")
	assert.Equal(t, corev1alpha1.ConditionFalse, cond.Status, "CustomResourceDefinitionDiscoveryDiscovered condition should False")

	crdDiscovery.Status.CRD = crd
	crdDiscovery.Status.SetCondition(corev1alpha1.CustomResourceDefinitionDiscoveryCondition{
		Type:   corev1alpha1.CustomResourceDefinitionDiscoveryReady,
		Status: corev1alpha1.ConditionTrue,
	})
	require.NoError(t, r.Client.Status().Update(ctx, crdDiscovery))

	reconcileLoop()
	cond, ok = crdDiscovery.Status.GetCondition(corev1alpha1.CustomResourceDefinitionDiscoveryDiscovered)
	assert.True(t, ok, "CustomResourceDefinitionDiscoveryDiscovered condition should be present")
	assert.Equal(t, corev1alpha1.ConditionTrue, cond.Status, "CustomResourceDefinitionDiscoveryDiscovered condition should True")

	internalCRD := &apiextensionsv1.CustomResourceDefinition{}
	require.NoError(t, r.Client.Get(ctx, types.NamespacedName{
		Name: strings.Join([]string{"redis", serviceClusterName, provider.Name}, "."),
	}, internalCRD))
}

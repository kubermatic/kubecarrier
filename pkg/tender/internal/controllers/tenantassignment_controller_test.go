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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func TestTenantAssignmentReconciler(t *testing.T) {
	tenantAssignment := &corev1alpha1.TenantAssignment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo.eu-west-1",
			Namespace: "provider-bar",
		},
		Spec: corev1alpha1.TenantAssignmentSpec{
			Tenant: corev1alpha1.ObjectReference{
				Name: "foo",
			},
			ServiceCluster: corev1alpha1.ObjectReference{
				Name: "eu-west-1",
			},
		},
	}

	r := TenantAssignmentReconciler{
		Log:           testutil.NewLogger(t),
		MasterClient:  fakeclient.NewFakeClientWithScheme(testScheme, tenantAssignment),
		MasterScheme:  testScheme,
		ServiceClient: fakeclient.NewFakeClientWithScheme(testScheme),
	}
	_, err := r.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      tenantAssignment.Name,
			Namespace: tenantAssignment.Namespace,
		},
	})
	require.NoError(t, err, "Reconcile")

	ctx := context.Background()
	namespaceList := &corev1.NamespaceList{}
	require.NoError(t, r.ServiceClient.List(ctx, namespaceList), "listing Namespaces")
	if assert.Len(t, namespaceList.Items, 1) {
		assert.Equal(t, "TenantAssignment-", namespaceList.Items[0].GenerateName)
	}
}

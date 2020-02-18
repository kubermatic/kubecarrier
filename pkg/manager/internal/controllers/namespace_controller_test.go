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

func TestNamespaceReconciler(t *testing.T) {
	// Setup
	ns1 := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ns1",
			Labels: map[string]string{
				"assignment.kubecarrier.io/eu-west-1.example-cloud": "true",
				"assignment.kubecarrier.io/us-east-1.example-cloud": "true",
			},
		},
	}
	ns2 := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ns2",
			Labels: map[string]string{
				"assignment.kubecarrier.io/eu-west-1.example-cloud": "true",
			},
		},
	}

	r := &NamespaceReconciler{
		Log:    testutil.NewLogger(t),
		Client: fakeclient.NewFakeClientWithScheme(testScheme, ns1, ns2),
		Scheme: testScheme,
	}

	_, err := r.Reconcile(ctrl.Request{})
	require.NoError(t, err)

	// Assert
	ctx := context.Background()
	scaList := &corev1alpha1.ServiceClusterAssignmentList{}
	require.NoError(t, r.List(ctx, scaList))
	assert.Len(t, scaList.Items, 3)

	sca1 := &corev1alpha1.ServiceClusterAssignment{}
	require.NoError(t, r.Get(ctx, types.NamespacedName{
		Name:      "ns1.eu-west-1",
		Namespace: "example-cloud",
	}, sca1))
	assert.Equal(t, corev1alpha1.ServiceClusterAssignmentSpec{
		ServiceCluster: corev1alpha1.ObjectReference{
			Name: "eu-west-1",
		},
		MasterClusterNamespace: corev1alpha1.ObjectReference{
			Name: "ns1",
		},
	}, sca1.Spec)

	sca2 := &corev1alpha1.ServiceClusterAssignment{}
	require.NoError(t, r.Get(ctx, types.NamespacedName{
		Name:      "ns1.us-east-1",
		Namespace: "example-cloud",
	}, sca2))
	assert.Equal(t, corev1alpha1.ServiceClusterAssignmentSpec{
		ServiceCluster: corev1alpha1.ObjectReference{
			Name: "us-east-1",
		},
		MasterClusterNamespace: corev1alpha1.ObjectReference{
			Name: "ns1",
		},
	}, sca2.Spec)

	sca3 := &corev1alpha1.ServiceClusterAssignment{}
	require.NoError(t, r.Get(ctx, types.NamespacedName{
		Name:      "ns2.eu-west-1",
		Namespace: "example-cloud",
	}, sca3))
	assert.Equal(t, corev1alpha1.ServiceClusterAssignmentSpec{
		ServiceCluster: corev1alpha1.ObjectReference{
			Name: "eu-west-1",
		},
		MasterClusterNamespace: corev1alpha1.ObjectReference{
			Name: "ns2",
		},
	}, sca3.Spec)
}

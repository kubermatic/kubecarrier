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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	masterv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/master/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/constants"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func TestManagementClusterReconciler(t *testing.T) {
	localManagementCluster := &masterv1alpha1.ManagementCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.LocalManagementClusterName,
		},
	}
	client := fakeclient.NewFakeClientWithScheme(testScheme, localManagementCluster)
	log := testutil.NewLogger(t)
	r := &ManagementClusterReconciler{
		Client: client,
		Log:    log,
		Scheme: testScheme,
	}
	ctx := context.Background()
	reconcileLoop := func() {
		for i := 0; i < 5; i++ {
			_, err := r.Reconcile(ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name: localManagementCluster.Name,
				},
			})
			require.NoError(t, err, "unexpected error returned by Reconcile")
			require.NoError(t, client.Get(ctx, types.NamespacedName{
				Name: localManagementCluster.Name,
			}, localManagementCluster))
		}
	}

	t.Run("update the local ManagementCluster object status", func(t *testing.T) {
		reconcileLoop()
		assert.Equal(t, localManagementCluster.Generation, localManagementCluster.Status.ObservedGeneration)
		cond, present := localManagementCluster.Status.GetCondition(masterv1alpha1.ManagementClusterReady)
		if assert.True(t, present, "ManagementCluster object ready condition missing") {
			assert.Equal(t, masterv1alpha1.ConditionTrue, cond.Status)
		}

	})
}

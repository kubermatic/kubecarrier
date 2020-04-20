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

package installation

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"

	masterv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/master/v1alpha1"
	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func newMasterCluster(f *testutil.Framework) func(t *testing.T) {
	return func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		t.Cleanup(cancel)

		c := exec.CommandContext(ctx, "kubectl", "kubecarrier", "setup", "--kubeconfig", f.Config().MasterExternalKubeconfigPath, "--master")
		out, err := c.CombinedOutput()
		t.Log(string(out))
		require.NoError(t, err)

		masterClient, err := f.MasterClient(t)
		require.NoError(t, err, "creating master client")

		tower := &operatorv1alpha1.Tower{}
		assert.NoError(t, masterClient.Get(ctx, types.NamespacedName{
			Name:      kubeCarrierName,
			Namespace: kubecarrierSystem,
		}, tower), "getting the Tower object error")
		require.True(t, tower.IsReady(), "tower is not ready")

		localManagementCluster := &masterv1alpha1.ManagementCluster{}
		assert.NoError(t, masterClient.Get(ctx, types.NamespacedName{
			Name: localManagementClusterName,
		}, localManagementCluster), "getting the local ManagementCluster object error")
		assert.True(t, localManagementCluster.IsReady(), "local ManagementCluster is not ready")
	}
}

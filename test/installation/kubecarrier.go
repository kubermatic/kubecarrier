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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func newKubeCarrier(f *testutil.Framework) func(t *testing.T) {
	return func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		t.Cleanup(cancel)

		c := exec.CommandContext(ctx, "kubectl", "kubecarrier", "setup", "--kubeconfig", f.Config().ManagementExternalKubeconfigPath)
		out, err := c.CombinedOutput()
		t.Log(string(out))
		require.NoError(t, err)

		managementClient, err := f.ManagementClient(t)
		require.NoError(t, err, "creating management client")

		testutil.KubeCarrierOperatorCheck(ctx, t, managementClient, f.ManagementScheme)

		kubeCarrier := &operatorv1alpha1.KubeCarrier{ObjectMeta: metav1.ObjectMeta{
			Name: "kubecarrier1",
		}}

		err = managementClient.Create(ctx, kubeCarrier)
		if assert.Error(t, err,
			"KubeCarrier object with name other than 'kubecarrier' should not be allowed to be created",
		) {
			assert.Contains(
				t,
				err.Error(),
				"KubeCarrier object name should be 'kubecarrier', found: kubecarrier1",
				"KubeCarrier creation webhook should error out on incorrect KubeCarrier object name",
			)
		}
		kubeCarrier.Name = "kubecarrier"
		require.NoError(t, testutil.WaitUntilReady(ctx, managementClient, kubeCarrier))
		testutil.KubeCarrierCheck(ctx, t, managementClient, f.ManagementScheme)
	}
}

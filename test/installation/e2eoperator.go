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

	"github.com/stretchr/testify/require"

	"k8c.io/kubecarrier/pkg/testutil"
)

func newE2EOperator(f *testutil.Framework) func(t *testing.T) {
	return func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		t.Cleanup(cancel)

		c := exec.CommandContext(ctx, "kubectl", "kubecarrier", "e2e-test", "setup-e2e-operator", "--kubeconfig", f.Config().ServiceExternalKubeconfigPath)
		out, err := c.CombinedOutput()
		t.Log(string(out))
		require.NoError(t, err)

		serviceClient, err := f.ServiceClient(t)
		require.NoError(t, err, "creating service client")
		testutil.E2EOperatorCheck(ctx, t, serviceClient, f.ServiceScheme)
	}
}

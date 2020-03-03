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

package verify

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

// VerifySuite verifies if we can reach both kubernetes clusters (management and service).
// and whether they are configured for our e2e tests.
func NewVerifySuite(f *testutil.Framework) func(t *testing.T) {
	return func(t *testing.T) {
		// Setup
		logger := testutil.NewLogger(t)
		managementClient, err := f.ManagementClient(logger)
		require.NoError(t, err)

		serviceClient, err := f.ServiceClient(logger)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		t.Cleanup(cancel)
		t.Run("validate management cluster connection", func(t *testing.T) {
			cm := &corev1.ConfigMap{}
			require.NoError(t, managementClient.Get(ctx, types.NamespacedName{
				Name:      "cluster-info",
				Namespace: "kube-public",
			}, cm), "cannot fetch cluster-info")
		})

		t.Run("validate service connection", func(t *testing.T) {
			cm := &corev1.ConfigMap{}
			require.NoError(t, serviceClient.Get(ctx, types.NamespacedName{
				Name:      "cluster-info",
				Namespace: "kube-public",
			}, cm), "cannot fetch cluster-info")
		})
		t.Run("internal service cluster config validation", func(t *testing.T) {
			loader := clientcmd.NewDefaultClientConfigLoadingRules()
			loader.ExplicitPath = f.Config().ServiceInternalKubeconfigPath
			clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
				loader,
				&clientcmd.ConfigOverrides{},
			)
			cfg, err := clientConfig.ClientConfig()
			require.NoError(t, err)
			assert.Equal(t, "system:serviceaccount:default:kubecarrier", cfg.Impersonate.UserName, "internal service cluster kubeconfig has wrong impersonation")
		})
	}
}

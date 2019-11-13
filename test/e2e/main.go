/*
Copyright 2019 The Kubecarrier Authors.

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

package e2e

import (
	"context"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

var (
	AllTests []testing.InternalTest

	MasterExternalKubeconfigPath  string
	MasterInternalKubeconfigPath  string
	ServiceExternalKubeconfigPath string
	ServiceInternalKubeconfigPath string
)

func init() {
	AllTests = append(AllTests, testing.InternalTest{
		Name: "KubecarrierE2ESuite",
		F: func(t *testing.T) {
			suite.Run(t, new(KubecarrierE2ESuite))
		},
	})
}

type KubecarrierE2ESuite struct {
	suite.Suite
	masterClient  client.Client
	serviceClient client.Client
}

var _ suite.SetupAllSuite = (*KubecarrierE2ESuite)(nil)

func (suite *KubecarrierE2ESuite) SetupSuite() {
	t := suite.T()
	t.Logf("master cluster external kubeconfig location: %s", MasterExternalKubeconfigPath)
	t.Logf("master cluster internal kubeconfig location: %s", MasterInternalKubeconfigPath)
	t.Logf("svc cluster external kubeconfig location: %s", ServiceExternalKubeconfigPath)
	t.Logf("svc cluster internal kubeconfig location: %s", ServiceInternalKubeconfigPath)

	t.Log("==== installing kubecarrier in the master cluster ====")
	cmd := exec.Command("anchor", "setup", "--kubeconfig", MasterExternalKubeconfigPath)
	out, err := cmd.CombinedOutput()
	t.Log("\n" + string(out))
	require.NoError(t, err, "cannot install kubecarrier in the master cluster")
	t.Log("==== successfully installed kubecarrier in the master cluster")

	scheme := runtime.NewScheme()
	require.NoError(t, clientgoscheme.AddToScheme(scheme), "adding native k8s scheme")

	{
		cfg, err := clientcmd.BuildConfigFromFlags("", MasterExternalKubeconfigPath)
		t.Logf("master external kubeconfig location: %s", MasterExternalKubeconfigPath)
		require.NoError(t, err, "building rest config")
		suite.masterClient, err = client.New(cfg, client.Options{
			Scheme: scheme,
		})
		require.NoError(t, err)
	}

	{
		cfg, err := clientcmd.BuildConfigFromFlags("", ServiceExternalKubeconfigPath)
		require.NoError(t, err, "building rest config")
		mapper, err := apiutil.NewDiscoveryRESTMapper(cfg)
		require.NoError(t, err)
		suite.serviceClient, err = client.New(cfg, client.Options{
			Scheme: scheme,
			Mapper: mapper,
		})
		require.NoError(t, err)
	}
}

func (suite *KubecarrierE2ESuite) TestValidMasterKubeconfig() {
	cm := &corev1.ConfigMap{}
	require.NoError(suite.T(), suite.masterClient.Get(context.Background(), types.NamespacedName{
		Name:      "cluster-info",
		Namespace: "kube-public",
	}, cm), "cannot fetch cluster-info")
	suite.T().Logf("cluster-info kubeconfig:\n%s", cm.Data["kubeconfig"])
}

func (suite *KubecarrierE2ESuite) TestValidServiceKubeconfig() {
	cm := &corev1.ConfigMap{}
	require.NoError(suite.T(), suite.serviceClient.Get(context.Background(), types.NamespacedName{
		Name:      "cluster-info",
		Namespace: "kube-public",
	}, cm), "cannot fetch cluster-info")
	suite.T().Logf("cluster-info kubeconfig:\n%s", cm.Data["kubeconfig"])
}

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

package e2e

import (
	"context"
	"testing"

	"github.com/go-logr/logr"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
	//"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
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
		Name: "E2ESuite",
		F: func(t *testing.T) {
			suite.Run(t, new(E2ESuite))
		},
	})
}

type E2ESuite struct {
	suite.Suite
	log logr.Logger

	masterClient  client.Client
	serviceClient client.Client

	masterScheme  *runtime.Scheme
	serviceScheme *runtime.Scheme
}

var _ suite.SetupAllSuite = (*E2ESuite)(nil)

func (suite *E2ESuite) SetupSuite() {
	t := suite.T()
	t.Logf("master cluster external kubeconfig location: %s", MasterExternalKubeconfigPath)
	t.Logf("master cluster internal kubeconfig location: %s", MasterInternalKubeconfigPath)
	t.Logf("svc cluster external kubeconfig location: %s", ServiceExternalKubeconfigPath)
	t.Logf("svc cluster internal kubeconfig location: %s", ServiceInternalKubeconfigPath)

	suite.masterScheme = runtime.NewScheme()
	require.NoError(t, clientgoscheme.AddToScheme(suite.masterScheme), "adding native k8s scheme")
	require.NoError(t, operatorv1alpha1.AddToScheme(suite.masterScheme), "adding KubeCarrier operator scheme")
	require.NoError(t, rbacv1.AddToScheme(suite.masterScheme), "adding KubeCarrier operator scheme")

	//suite.serviceScheme = runtime.NewScheme()
	//require.NoError(t, clientgoscheme.AddToScheme(suite.serviceScheme), "adding native k8s scheme")

	{
		cfg, err := clientcmd.BuildConfigFromFlags("", MasterExternalKubeconfigPath)
		t.Logf("master external kubeconfig location: %s", MasterExternalKubeconfigPath)
		require.NoError(t, err, "building rest config")
		suite.masterClient, err = client.New(cfg, client.Options{
			Scheme: suite.masterScheme,
		})
		require.NoError(t, err)
	}

	//{
	//	cfg, err := clientcmd.BuildConfigFromFlags("", ServiceExternalKubeconfigPath)
	//	require.NoError(t, err, "building rest config")
	//	mapper, err := apiutil.NewDiscoveryRESTMapper(cfg)
	//	require.NoError(t, err)
	//	suite.serviceClient, err = client.New(cfg, client.Options{
	//		Scheme: suite.serviceScheme,
	//		Mapper: mapper,
	//	})
	//	require.NoError(t, err)
	//}
	//
	suite.log = testutil.NewLogger(t)
}

func (suite *E2ESuite) TestValidMasterKubeconfig() {
	cm := &corev1.ConfigMap{}
	require.NoError(suite.T(), suite.masterClient.Get(context.Background(), types.NamespacedName{
		Name:      "cluster-info",
		Namespace: "kube-public",
	}, cm), "cannot fetch cluster-info")
	suite.T().Logf("cluster-info kubeconfig:\n%s", cm.Data["kubeconfig"])
}

//func (suite *E2ESuite) TestValidServiceKubeconfig() {
//	cm := &corev1.ConfigMap{}
//	require.NoError(suite.T(), suite.serviceClient.Get(context.Background(), types.NamespacedName{
//		Name:      "cluster-info",
//		Namespace: "kube-public",
//	}, cm), "cannot fetch cluster-info")
//	suite.T().Logf("cluster-info kubeconfig:\n%s", cm.Data["kubeconfig"])
//}

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

	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubermatic/kubecarrier/test/framework"
)

var _ suite.SetupAllSuite = (*VerifySuite)(nil)

// VerifySuite verifies if we can reach both kubernetes clusters (master and service).
// and whether they are configured for our e2e tests.
type VerifySuite struct {
	suite.Suite
	*framework.Framework

	masterClient  client.Client
	serviceClient client.Client
}

func (s *VerifySuite) SetupSuite() {
	t := s.T()
	t.Logf("master cluster external kubeconfig location: %s", s.Framework.Config().MasterExternalKubeconfigPath)
	t.Logf("master cluster internal kubeconfig location: %s", s.Framework.Config().MasterInternalKubeconfigPath)
	t.Logf("svc cluster external kubeconfig location: %s", s.Framework.Config().ServiceExternalKubeconfigPath)
	t.Logf("svc cluster internal kubeconfig location: %s", s.Framework.Config().ServiceInternalKubeconfigPath)

	var err error
	s.Require().NoError(err, "creating testing framework")
	s.masterClient, err = s.MasterClient()
	s.Require().NoError(err, "creating master client")
	s.serviceClient, err = s.ServiceClient()
	s.Require().NoError(err, "creating service client")
}

func (s *VerifySuite) TestValidMasterKubeconfig() {
	cm := &corev1.ConfigMap{}
	s.Require().NoError(s.masterClient.Get(context.Background(), types.NamespacedName{
		Name:      "cluster-info",
		Namespace: "kube-public",
	}, cm), "cannot fetch cluster-info")
	s.T().Logf("cluster-info kubeconfig:\n%s", cm.Data["kubeconfig"])
}

func (s *VerifySuite) TestValidServiceKubeconfig() {
	cm := &corev1.ConfigMap{}
	s.Require().NoError(s.serviceClient.Get(context.Background(), types.NamespacedName{
		Name:      "cluster-info",
		Namespace: "kube-public",
	}, cm), "cannot fetch cluster-info")
	s.T().Logf("cluster-info kubeconfig:\n%s", cm.Data["kubeconfig"])
}

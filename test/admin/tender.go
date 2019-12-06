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

package admin

import (
	"context"
	"io/ioutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func (s *AdminSuite) TestTenderCreationAndDeletion() {
	t := s.T()
	t.Parallel()
	ctx := context.Background()
	const (
		namespace = "default"
	)

	serviceKubeconfig, err := ioutil.ReadFile(s.Framework.Config().ServiceInternalKubeconfigPath)
	require.NoError(t, err, "cannot read service internal kubeconfig")

	sec := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      "eu-west-1",
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"kubeconfig": serviceKubeconfig,
		},
	}
	tender := &operatorv1alpha1.Tender{
		ObjectMeta: v1.ObjectMeta{
			Name:      "eu-west-1",
			Namespace: namespace,
		},
		Spec: operatorv1alpha1.TenderSpec{
			KubeconfigSecret: operatorv1alpha1.ObjectReference{
				Name: "eu-west-1",
			},
		},
	}
	require.NoError(t, client.IgnoreNotFound(s.masterClient.Delete(ctx, sec.DeepCopy())))
	require.NoError(t, client.IgnoreNotFound(s.masterClient.Delete(ctx, tender.DeepCopy())))

	require.NoError(t, s.masterClient.Create(ctx, sec))
	require.NoError(t, s.masterClient.Create(ctx, tender))

	require.NoError(t, testutil.WaitUntilCondition(
		s.masterClient,
		tender,
		operatorv1alpha1.TenderReady,
		operatorv1alpha1.ConditionTrue,
	), "tender object not ready within time limit")

	require.NoError(t, s.masterClient.Delete(ctx, tender))
	require.NoError(t, testutil.WaitUntilNotFound(s.masterClient, tender), "tender object not cleared within time limit")
	assert.NoError(t, s.masterClient.Delete(ctx, sec))
}

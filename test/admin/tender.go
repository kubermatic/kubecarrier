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
	"fmt"
	"io/ioutil"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	assert.NoError(t, wait.Poll(time.Second, 30*time.Second, func() (done bool, err error) {
		if err := s.masterClient.Get(ctx, types.NamespacedName{
			Name:      tender.Name,
			Namespace: tender.Namespace,
		}, tender); err != nil {
			return false, fmt.Errorf("get: %w", err)
		}
		cond, ok := tender.Status.GetCondition(operatorv1alpha1.TenderReady)
		if !ok {
			return false, nil
		}
		return cond.Status == operatorv1alpha1.ConditionTrue, nil
	}), "tender object not ready within time limit")

	require.NoError(t, s.masterClient.Delete(ctx, tender))
	assert.NoError(t, wait.Poll(time.Second, 30*time.Second, func() (done bool, err error) {
		err = s.masterClient.Get(ctx, types.NamespacedName{
			Name:      tender.Name,
			Namespace: tender.Namespace,
		}, tender)
		switch {
		case err == nil:
			return false, nil
		case errors.IsNotFound(err):
			return true, nil
		default:
			return false, err
		}
	}), "tender object not cleared within time limit")
	assert.NoError(t, s.masterClient.Delete(ctx, sec))
}

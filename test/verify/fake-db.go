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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	fakev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/fake/v1alpha1"
)

func (s *VerifySuite) TestJokeOperator() {
	t := s.T()
	t.Parallel()
	s.EnsureFakeOperator(t)
	s.SetupSuite() // requires client reinit due to new CRDs
	ctx := context.Background()

	t.Run("HappyPath", func(t *testing.T) {
		db := &fakev1alpha1.DB{
			ObjectMeta: v1.ObjectMeta{
				Name:      "dummy",
				Namespace: "default",
			},
			Spec: fakev1alpha1.DBSpec{
				Size:         "small",
				DatabaseName: "test",
				DatabaseUser: "username",
			},
		}
		_ = s.serviceClient.Delete(ctx, db)
		require.NoError(t, s.serviceClient.Create(ctx, db))
		assert.NoError(t, wait.Poll(time.Second, 30*time.Second, func() (done bool, err error) {
			require.NoError(t, s.serviceClient.Get(ctx, types.NamespacedName{
				Namespace: "default",
				Name:      "dummy",
			}, db))

			if db.Generation != db.Status.ObservedGeneration {
				return false, nil
			}

			cond, _ := db.Status.GetCondition(fakev1alpha1.DBReady)
			return cond.Status == fakev1alpha1.ConditionTrue, nil

		}), "couldn't reach the DB ready state")
	})
}

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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	e2ev1alpha2 "github.com/kubermatic/kubecarrier/pkg/apis/e2e/v1alpha2"
)

func (s *VerifySuite) TestJokeOperator() {
	t := s.T()
	t.Parallel()
	s.EnsureJokeOperator(t)
	s.SetupSuite() // requires client reinit due to new CRDs

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		jokes := []e2ev1alpha2.JokeItem{
			{
				// https://twitter.com/wm/status/1172654176742105089?lang=en
				Text: "A devops engineer walks into a bar, puts the bartender in a docker container, put kubernetes behind the bar, spins up 1000 bartenders, orders 1 beer.",
				Type: "kubernetes",
			},
			{
				// https://www.reddit.com/r/sysadmin/comments/625mk9/sysadmindevops_jokes/dfjwac5/
				Text: "I'd tell you the one about UDP, but you wouldn't get it.",
				Type: "devops",
			},
		}

		ctx := context.Background()
		c := s.serviceClient
		joke := &e2ev1alpha2.Joke{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dummy",
				Namespace: "default",
			},
			Spec: e2ev1alpha2.JokeSpec{
				JokeDatabase: jokes,
			},
		}
		require.NoError(t, c.Create(ctx, joke))
		assert.NoError(t, wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
			if err := c.Get(ctx, types.NamespacedName{
				Namespace: joke.Namespace,
				Name:      joke.Name,
			}, joke); err != nil {
				if errors.IsNotFound(err) {
					t.Log("joke not yet found")
					return false, nil
				}
				return false, err
			}
			cond, ok := joke.Status.GetCondition(e2ev1alpha2.JokeReady)
			return ok && cond.Status == e2ev1alpha2.ConditionTrue && joke.Status.ObservedGeneration == joke.Generation, nil
		}), "joke wasn't ready within the timeframe")
		t.Log("selected joke: " + joke.Status.SelectedJoke.Text)
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()
		jokes := []e2ev1alpha2.JokeItem{}

		ctx := context.Background()
		c := s.serviceClient
		joke := &e2ev1alpha2.Joke{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dummy-2",
				Namespace: "default",
			},
			Spec: e2ev1alpha2.JokeSpec{
				JokeDatabase: jokes,
			},
		}
		require.NoError(t, c.Create(ctx, joke))
		require.NoError(t, wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
			if err := c.Get(ctx, types.NamespacedName{
				Namespace: joke.Namespace,
				Name:      joke.Name,
			}, joke); err != nil {
				t.Log("joke not yet found")
				if errors.IsNotFound(err) {
					return false, nil
				}
				return false, err
			}
			cond, ok := joke.Status.GetCondition(e2ev1alpha2.JokeReady)
			return ok && cond.Status == e2ev1alpha2.ConditionFalse && joke.Status.ObservedGeneration == joke.Generation, nil
		}), "joke wasn't marked as failed within the timeframe")
		cond, _ := joke.Status.GetCondition(e2ev1alpha2.JokeReady)
		t.Log("joke status message" + cond.Message)
	})
}

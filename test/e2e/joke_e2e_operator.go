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
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubermatic/kubecarrier/pkg/apis/e2e/v1alpha2"
)

func (suite *KubeCarrierE2ESuite) TestJokeOperatorSuccess() {
	suite.T().Parallel()
	jokes := []v1alpha2.JokeItem{
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
	c := suite.serviceClient
	joke := &v1alpha2.Joke{
		ObjectMeta: v1.ObjectMeta{
			Name:      "dummy",
			Namespace: "default",
		},
		Spec: v1alpha2.JokeSpec{
			JokeDatabase: jokes,
		},
	}
	defer suite.Assert().NoError(c.Delete(ctx, joke))
	suite.Require().NoError(c.Create(ctx, joke))
	suite.Assert().NoError(wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
		if err := c.Get(ctx, types.NamespacedName{
			Namespace: joke.Namespace,
			Name:      joke.Name,
		}, joke); err != nil {
			return false, err
		}
		cond, ok := joke.Status.GetCondition(v1alpha2.JokeReady)
		return ok && cond.Status == v1alpha2.ConditionTrue && joke.Status.ObservedGeneration == joke.Generation, nil
	}), "joke wasn't ready within the timeframe")
	suite.Info("selected joke: " + joke.Status.SelectedJoke.Text)
}

func (suite *KubeCarrierE2ESuite) TestJokeFailure() {
	suite.T().Parallel()
	jokes := []v1alpha2.JokeItem{}

	ctx := context.Background()
	c := suite.serviceClient
	joke := &v1alpha2.Joke{
		ObjectMeta: v1.ObjectMeta{
			Name:      "dummy-2",
			Namespace: "default",
		},
		Spec: v1alpha2.JokeSpec{
			JokeDatabase: jokes,
		},
	}
	defer suite.Assert().NoError(c.Delete(ctx, joke))
	suite.Require().NoError(c.Create(ctx, joke))
	suite.Assert().NoError(wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
		if err := c.Get(ctx, types.NamespacedName{
			Namespace: joke.Namespace,
			Name:      joke.Name,
		}, joke); err != nil {
			return false, err
		}
		cond, ok := joke.Status.GetCondition(v1alpha2.JokeReady)
		return ok && cond.Status == v1alpha2.ConditionFalse && joke.Status.ObservedGeneration == joke.Generation, nil
	}), "joke wasn't marked as failed within the timeframe")
	cond, _ := joke.Status.GetCondition(v1alpha2.JokeReady)
	suite.Info("joke status message" + cond.Message)
}

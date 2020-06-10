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

package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func newFakeDB(f *testutil.Framework) func(t *testing.T) {
	return func(t *testing.T) {
		t.Log("testing if we can create FakeDB in service cluster")
		ctx, _ := testutil.LoggingContext(t, context.Background())
		serviceClient, err := f.ServiceClient(t)
		require.NoError(t, err, "creating service client")
		t.Cleanup(serviceClient.CleanUpFunc(ctx))

		testName := strings.Replace(strings.ToLower(t.Name()), "/", "-", -1)
		testNamespace := testName + "-namespace"
		someNamespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: testNamespace,
			},
		}
		require.NoError(
			t, serviceClient.Create(ctx, someNamespace), "creating a Namespace")
		fakeDB := testutil.NewFakeDB(testName, testNamespace)
		t.Log("adding fakeDB")
		require.NoError(t, serviceClient.Create(ctx, fakeDB), "creating FakeDB")
		require.NoError(t, testutil.WaitUntilReady(ctx, serviceClient, fakeDB))
		t.Log("deleting fakeDB")
		require.NoError(t, testutil.DeleteAndWaitUntilNotFound(ctx, serviceClient, fakeDB))
	}
}

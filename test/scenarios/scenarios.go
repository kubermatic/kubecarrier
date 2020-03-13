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

package scenarios

import (
	"testing"

	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func NewSuite(f *testutil.Framework) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		for name, testFn := range map[string]func(f *testutil.Framework) func(t *testing.T){
			"simple": newSimpleScenario,
		} {
			name := name
			testFn := testFn

			t.Run(name, func(t *testing.T) {
				t.Parallel()
				testFn(f)(t)
			})
		}
	}
}

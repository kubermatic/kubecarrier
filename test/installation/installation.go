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

package installation

import (
	"testing"

	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

const (
	kubecarrierSystem          = "kubecarrier-system"
	kubeCarrierName            = "kubecarrier"
	prefix                     = "kubecarrier-manager"
	localManagementClusterName = "local"
)

func NewInstallationSuite(f *testutil.Framework) func(t *testing.T) {
	return func(t *testing.T) {
		for _, test := range []struct {
			name   string
			testFn func(t *testutil.Framework) func(t *testing.T)
		}{
			{
				name:   "MasterCluster",
				testFn: newMasterCluster,
			},
			{
				name:   "ManagementCluster",
				testFn: newManagementCluster,
			},
			{
				name:   "ServiceCluster",
				testFn: newServiceCluster,
			},
		} {
			t.Run(test.name, func(t *testing.T) {
				test.testFn(f)(t)
			})
		}
	}
}

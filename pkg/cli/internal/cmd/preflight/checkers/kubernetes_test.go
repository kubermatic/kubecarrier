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

package checkers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKubernetesVersionChecker(t *testing.T) {
	tests := []struct {
		name                     string
		checker                  kubernetesVersionChecker
		currentKubernetesVersion string
		expectedError            error
	}{
		{
			name: "kubernetes version is higher than first supported version",
			checker: kubernetesVersionChecker{
				firstSupportedVersion: firstSupportedKubernetesVersion,
			},
			currentKubernetesVersion: "v1.16.4",
			expectedError:            nil,
		},
		{
			name: "kubernetes version equals than first supported version",
			checker: kubernetesVersionChecker{
				firstSupportedVersion: firstSupportedKubernetesVersion,
			},
			currentKubernetesVersion: "v1.16.0",
			expectedError:            nil,
		},
		{
			name: "kubernetes version is lower than first supported version",
			checker: kubernetesVersionChecker{
				firstSupportedVersion: firstSupportedKubernetesVersion,
			},
			currentKubernetesVersion: "v1.15.4",
			expectedError:            fmt.Errorf("kubernetes version is lower than the oldest version that KubeCarrier supports, requrires: >= v1.16.0, found: v1.15.4"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedError, test.checker.checkKubernetesVersion(test.currentKubernetesVersion))
		})
	}
}

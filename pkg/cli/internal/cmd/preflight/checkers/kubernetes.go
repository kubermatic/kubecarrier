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

	versionutil "k8s.io/apimachinery/pkg/util/version"
)

// kubernetesVersionChecker checks if the Kubernetes version of the cluster meets the requirement to deploy KubeCarrier.
type kubernetesVersionChecker struct {
	firstSupportedVersion string
	kubernetesVersion     string
}

func (c *kubernetesVersionChecker) check() error {
	firstSupportedVersion, err := versionutil.ParseSemantic(c.firstSupportedVersion)
	if err != nil {
		return err
	}
	kubernetesGitVersion, err := versionutil.ParseSemantic(c.kubernetesVersion)
	if err != nil {
		return err
	}
	if kubernetesGitVersion.LessThan(firstSupportedVersion) {
		return fmt.Errorf("kubernetes version is lower than the oldest version that KubeCarrier supports, requrires: >= v%s, found: v%s", firstSupportedVersion.String(), kubernetesGitVersion.String())
	}
	return nil
}

func (c *kubernetesVersionChecker) name() string {
	return "KubernetesVersion"
}

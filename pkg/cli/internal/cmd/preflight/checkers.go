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

package preflight

import (
	"fmt"

	versionutil "k8s.io/apimachinery/pkg/util/version"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

// checker checks if the state of the system meets KubeCarrier installation requirements
type checker interface {
	check() error
	name() string
}

func RunCheckers(c *rest.Config) error {
	checkers := []checker{
		&kubernetesVersionChecker{
			config:                c,
			firstSupportedVersion: "v1.16.0",
		},
	}
	for _, checker := range checkers {
		if err := checker.check(); err != nil {
			return err
		}
	}
	return nil
}

// kubernetesVersionChecker
type kubernetesVersionChecker struct {
	config                *rest.Config
	firstSupportedVersion string
}

func (c *kubernetesVersionChecker) check() error {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(c.config)
	if err != nil {
		return fmt.Errorf("cannot create discovery client: %w", err)
	}
	kubernetesVersion, err := discoveryClient.ServerVersion()
	if err != nil {
		return fmt.Errorf("can not get the kubernetesVersion: %w", err)
	}
	firstSupportedVersion, err := versionutil.ParseSemantic(c.firstSupportedVersion)
	if err != nil {
		return err
	}
	kubernetesGitVersion, err := versionutil.ParseSemantic(kubernetesVersion.String())
	if err != nil {
		return err
	}
	if kubernetesGitVersion.LessThan(firstSupportedVersion) {
		return fmt.Errorf("kubernetes version is lower than the oldest version that KubeCarrier supports, requrires: >= %s, found: %s", kubernetesGitVersion.String(), firstSupportedVersion.String())
	}
	return nil
}

func (c *kubernetesVersionChecker) name() string {
	return "KubernetesVersion"
}

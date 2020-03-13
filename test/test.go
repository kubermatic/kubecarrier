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

package test

import (
	"fmt"
	"testing"

	"github.com/kubermatic/kubecarrier/pkg/testutil"
	"github.com/kubermatic/kubecarrier/test/installation"
	"github.com/kubermatic/kubecarrier/test/integration"
	"github.com/kubermatic/kubecarrier/test/scenarios"
	"github.com/kubermatic/kubecarrier/test/verify"
)

func AllTests(config testutil.FrameworkConfig) ([]testing.InternalTest, error) {
	f, err := testutil.New(config)
	if err != nil {
		return nil, fmt.Errorf("creating test framework:%w", err)
	}

	var tests []testing.InternalTest
	tests = append(tests,
		testing.InternalTest{
			Name: "VerifySuite",
			F:    verify.NewVerifySuite(f),
		},
		testing.InternalTest{
			Name: "InstallationSuite",
			F:    installation.NewInstallationSuite(f),
		},
		testing.InternalTest{
			Name: "Integration",
			F:    integration.NewIntegrationSuite(f),
		},
		testing.InternalTest{
			Name: "Scenarios",
			F:    scenarios.NewSuite(f),
		},
	)

	return tests, nil
}

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

package cmd

import (
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"

	e2e_test "github.com/kubermatic/kubecarrier/pkg/cli/internal/cmd/e2e-test"
	"github.com/kubermatic/kubecarrier/pkg/cli/internal/cmd/setup"
	"github.com/kubermatic/kubecarrier/pkg/cli/internal/cmd/sut"
	"github.com/kubermatic/kubecarrier/pkg/cli/internal/cmd/version"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

// NewKubecarrierCLI creates the root command for the KubeCarrier CLI.
func NewKubecarrierCLI() *cobra.Command {
	log := ctrl.Log.WithName("kubecarrier")
	cmd := &cobra.Command{
		Use:   "kubecarrier",
		Short: "The CLI tool for managing KubeCarrier",
		Long: `The CLI tool for managing KubeCarrier,
Documentation is available in the project's repository:
https://github.com/kubermatic/kubecarrier`,
	}

	cmd.AddCommand(
		e2e_test.NewCommand(log),
		setup.NewCommand(log),
		version.NewCommand(log),
		sut.NewCommand(log),
	)

	return util.CmdLogMixin(cmd)
}

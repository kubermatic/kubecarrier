/*
Copyright 2019 The Kubecarrier Authors.

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

package anchor

import (
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"github.com/kubermatic/kubecarrier/pkg/anchor/cmd/setup"
	"github.com/kubermatic/kubecarrier/pkg/anchor/cmd/version"
)

// NewAnchor creates the root command for the anchor CLI.
func NewAnchor(log logr.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "anchor",
		Short: "Anchor is the CLI tool for managing Kubecarrier",
		Long: `Anchor is a CLI library for managing Kubecarrier,
Documentation is available in the project's repository:
https://github.com/kubermatic/kubecarrier`,
	}

	cmd.AddCommand(
		version.NewCommand(log),
		setup.NewCommand(log))

	return cmd
}

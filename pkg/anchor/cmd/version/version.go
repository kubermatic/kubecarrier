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

package version

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/kubermatic/kubecarrier/pkg/version"
)

type versionFlagpole struct {
	Full bool
}

// NewCommand returns the version subcommand for anchor.
func NewCommand(log logr.Logger) *cobra.Command {
	var flags versionFlagpole
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "version",
		Short: "prints the CLI version",
		Long:  "prints the CLI version",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			v := version.Get()
			if !flags.Full {
				fmt.Fprintln(cmd.OutOrStdout(), v.Version)
				return
			}

			y, err := yaml.Marshal(v)
			if err != nil {
				return fmt.Errorf("marshalling version information: %w", err)
			}
			fmt.Fprint(cmd.OutOrStdout(), string(y))
			return
		},
	}
	cmd.Flags().BoolVar(&flags.Full, "full", false, "print all build details, not only the version number")
	return cmd
}

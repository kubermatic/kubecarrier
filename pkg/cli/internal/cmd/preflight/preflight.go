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
	"time"

	"github.com/gernest/wow"
	"github.com/gernest/wow/spin"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// NewPreflightCommand returns the preflight checking subcommand for KubeCarrier CLI.
func NewPreflightCommand(log logr.Logger) *cobra.Command {
	flags := genericclioptions.NewConfigFlags(false)
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "preflight",
		Short: "preflight checks for KubeCarrier",
		Long:  "preflight checks for KubeCarrier",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			cfg, err := flags.ToRESTConfig()
			if err != nil {
				return err
			}
			s := wow.New(cmd.OutOrStdout(), spin.Get(spin.Dots), "")
			startTime := time.Now()
			return RunCheckers(cfg, s, startTime, log)
		},
	}
	return cmd
}

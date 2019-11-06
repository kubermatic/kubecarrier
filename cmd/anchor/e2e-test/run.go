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

package e2e_test

import (
	"flag"
	"os"
	"testing"

	"github.com/spf13/cobra"

	"github.com/kubermatic/kubecarrier/test/e2e"
)

func newRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "run all end2end test",
		PreRun: func(cmd *cobra.Command, args []string) {
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			e2e.MasterInternalKubeconfigPath = cfg.masterInternalKubeconfigFile
			e2e.MasterExternalKubeconfigPath = cfg.masterExternalKubeconfigFile
			e2e.ServiceExternalKubeconfigPath = cfg.serviceExternalKubeconfigFile
			e2e.ServiceInternalKubeconfigPath = cfg.serviceInternalKubeconfigFile

			testing.Main(func(pat, str string) (b bool, e error) {
				return true, nil
			}, e2e.AllTests, nil, nil)
			return nil
		},
	}

	// hackery for the go test command to work properly
	oldCommandLine := flag.CommandLine
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// NOTE: this function has sync.Once like semantics; thus any subsequent call doesn't do anything
	testing.Init()
	cmd.Flags().AddGoFlagSet(flag.CommandLine)
	flag.CommandLine = oldCommandLine

	return cmd
}

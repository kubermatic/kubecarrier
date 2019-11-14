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

package e2e_test

import (
	"flag"
	"os"
	"os/exec"
	"testing"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"github.com/kubermatic/kubecarrier/test/e2e"
)

func newRunCommand(log logr.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "run all end2end tests",
		RunE: func(cmd *cobra.Command, args []string) error {
			e2e.MasterInternalKubeconfigPath = cfg.masterInternalKubeconfigFile
			e2e.MasterExternalKubeconfigPath = cfg.masterExternalKubeconfigFile
			e2e.ServiceExternalKubeconfigPath = cfg.serviceExternalKubeconfigFile
			e2e.ServiceInternalKubeconfigPath = cfg.serviceInternalKubeconfigFile

			log.Info("running \"anchor setup\" to install KubeCarrier in the master cluster")
			c := exec.Command("anchor", "setup", "--kubeconfig", e2e.MasterExternalKubeconfigPath)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				log.Error(err, "cannot install kubecarrier in the master cluster")
				os.Exit(2)
			}
			log.Info("installed kubecarrier in the master cluster")
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

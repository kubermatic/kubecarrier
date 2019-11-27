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
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"github.com/kubermatic/kubecarrier/test/framework"
)

func NewCommand(log logr.Logger) *cobra.Command {
	cfg := &framework.Config{}
	cmd := &cobra.Command{
		Use:   "e2e-test",
		Short: "end2end testing utilities",
	}

	cmd.PersistentPreRunE = func(_ *cobra.Command, args []string) error {
		for p := cmd.Parent(); p != nil; p = p.Parent() {
			if p.PersistentPreRunE != nil {
				if err := p.PersistentPreRunE(p, args); err != nil {
					return err
				}
				break
			}
			if p.PersistentPreRun != nil {
				p.PersistentPreRun(p, args)
				break
			}
		}
		cfg.Default()
		return nil
	}

	cmd.AddCommand(
		newRunCommand(log, cfg),
		newSetupE2EOperator(log),
	)

	cmd.PersistentFlags().StringVar(&cfg.TestID, "test-id", "", "unique e2e test id")
	cmd.PersistentFlags().StringVar(&cfg.MasterExternalKubeconfigPath, "master-external-kubeconfig", "", "master cluster external (reachable outside cluster/docker) kubeconfig file")
	cmd.PersistentFlags().StringVar(&cfg.MasterInternalKubeconfigPath, "master-internal-kubeconfig", "", "master cluster internal (reachable within cluster/docker) kubeconfig file")
	cmd.PersistentFlags().StringVar(&cfg.ServiceExternalKubeconfigPath, "service-external-kubeconfig", "", "service cluster external (reachable outside cluster/docker) kubeconfig file")
	cmd.PersistentFlags().StringVar(&cfg.ServiceInternalKubeconfigPath, "service-internal-kubeconfig", "", "service cluster internal (reachable within cluster/docker) kubeconfig file")
	return cmd
}

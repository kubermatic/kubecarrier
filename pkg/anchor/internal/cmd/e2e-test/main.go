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
	"os"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Config struct {
	log logr.Logger

	// Current e2e-test id. The created clusters shall be prefixed accordingly
	testID string

	masterInternalKubeconfigFile  string
	masterExternalKubeconfigFile  string
	serviceInternalKubeconfigFile string
	serviceExternalKubeconfigFile string
}

func (c *Config) masterClusterName() string {
	return "kubecarrier-" + c.testID
}

func (c *Config) serviceClusterName() string {
	return "kubecarrier-svc-" + c.testID
}

func (c *Config) Default() error {
	if c.serviceInternalKubeconfigFile == "" {
		c.serviceInternalKubeconfigFile = os.ExpandEnv("${HOME}/.kube/internal-kind-config-" + c.serviceClusterName())
	}
	if c.serviceExternalKubeconfigFile == "" {
		c.serviceExternalKubeconfigFile = os.ExpandEnv("${HOME}/.kube/kind-config-" + c.serviceClusterName())
	}

	if c.masterInternalKubeconfigFile == "" {
		c.masterInternalKubeconfigFile = os.ExpandEnv("${HOME}/.kube/internal-kind-config-" + c.masterClusterName())
	}
	if c.masterExternalKubeconfigFile == "" {
		c.masterExternalKubeconfigFile = os.ExpandEnv("${HOME}/.kube/kind-config-" + c.masterClusterName())
	}
	if c.log == nil {
		c.log = ctrl.Log.WithValues("e2e-test")
	}

	return nil
}

var cfg Config

func NewCommand(log logr.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "e2e-test",
		Short: "end2end testing utilities",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// hackery up to the 10th
			p := cmd.Parent()
			for p != nil && p.PersistentPreRunE != nil && p.PersistentPreRun != nil {
				p = p.Parent()
			}
			if p != nil && p.PersistentPreRun != nil {
				p.PersistentPreRun(p, args)
			}
			if p != nil && p.PersistentPreRunE != nil {
				if err := p.PersistentPreRunE(p, args); err != nil {
					return err
				}
			}

			cfg.log = log.WithName("e2e-test")
			return cfg.Default()
		},
	}

	cmd.AddCommand(
		newRunCommand(log),
		newSetupE2EOperator(log),
	)

	cmd.PersistentFlags().StringVar(&cfg.testID, "test-id", "", "unique e2e test id")
	cmd.PersistentFlags().StringVar(&cfg.masterExternalKubeconfigFile, "master-external-kubeconfig", "", "master cluster external (reachable outside cluster/docker) kubeconfig file")
	cmd.PersistentFlags().StringVar(&cfg.masterInternalKubeconfigFile, "master-internal-kubeconfig", "", "master cluster internal (reachable within cluster/docker) kubeconfig file")
	cmd.PersistentFlags().StringVar(&cfg.serviceExternalKubeconfigFile, "service-external-kubeconfig", "", "service cluster external (reachable outside cluster/docker) kubeconfig file")
	cmd.PersistentFlags().StringVar(&cfg.serviceInternalKubeconfigFile, "service-internal-kubeconfig", "", "service cluster internal (reachable within cluster/docker) kubeconfig file")
	return cmd
}

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
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cluster/create"
)

type Config struct {
	// Current e2e-test id. The created clusters shall be prefixed accordingly
	testID string

	// reuse existing e2e-test environment, if exists
	reuse bool

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

func (c *Config) SetupKindCluster() error {
	for _, cl := range []struct {
		Name               string
		internalKubeconfig string
		externalKubeconfig string
	}{
		{
			Name:               c.masterClusterName(),
			internalKubeconfig: c.masterInternalKubeconfigFile,
			externalKubeconfig: c.masterExternalKubeconfigFile,
		},
		{
			Name:               c.serviceClusterName(),
			internalKubeconfig: c.serviceInternalKubeconfigFile,
			externalKubeconfig: c.serviceExternalKubeconfigFile,
		},
	} {
		know, err := cluster.IsKnown(cl.Name)
		if err != nil {
			return fmt.Errorf("cannot find out if cluster name %s is known: %w", cl.Name, err)
		}
		ctx := cluster.NewContext(cl.Name)
		switch {
		case !know:
			log.Printf("creating cluster %s", ctx.Name())
			if err := ctx.Create(
				create.Retain(true),
				create.WaitForReady(10*time.Minute),
			); err != nil {
				log.Panic(err)
			}
		case c.reuse && know:
			log.Printf("found existing kind cluster %s, reusing it\n", ctx.Name())
		case !c.reuse && know:
			log.Printf("found existing kind cluster %s, but reuse is disabled\n", ctx.Name())
			return fmt.Errorf("found existing kind cluster %s, but reuse is disabled\n", ctx.Name())
		default:
		}

		log.Printf("for cluster %s kubeconifg is at %s", ctx.Name(), ctx.KubeConfigPath())

		externalKubeconfig, err := ioutil.ReadFile(ctx.KubeConfigPath())
		if err != nil {
			return fmt.Errorf("cannot read kubeconfig: %v", err)
		}
		if err := ioutil.WriteFile(cl.externalKubeconfig, externalKubeconfig, 0600); err != nil {
			return fmt.Errorf("cannot write external kubeconfig: %v", err)
		}

		// List nodes by cluster context name
		n, err := ctx.ListInternalNodes()
		if err != nil {
			return fmt.Errorf("cannot list internal nodes: %w", err)
		}

		cmdNode := n[0].Command("cat", "/etc/kubernetes/admin.conf")
		b := new(bytes.Buffer)
		cmdNode.SetStdout(b)
		if err := cmdNode.Run(); err != nil {
			return fmt.Errorf("cannot read the internal kubeconfig %w", err)
		}

		if err := ioutil.WriteFile(cl.internalKubeconfig, b.Bytes(), 0600); err != nil {
			return fmt.Errorf("cannot write internal kubeconfig: %w", err)
		}
	}
	return nil
}

func (c *Config) TeardownKindCluster() error {
	for _, kindClusterName := range []string{
		c.masterClusterName(),
		c.serviceClusterName(),
	} {
		know, err := cluster.IsKnown(kindClusterName)
		if err != nil {
			return fmt.Errorf("cannot find out if cluster name %s is known: %w", kindClusterName, err)
		}
		if !know {
			log.Printf("cluster %s not known, skipping", kindClusterName)
			continue
		}
		ctx := cluster.NewContext(kindClusterName)
		if err := ctx.Delete(); err != nil {
			return fmt.Errorf("cannot delete cluster %s: %w", kindClusterName, err)
		}

	}
	return nil
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

	return nil
}

var cfg Config

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "e2e-test",
		Short: "end2end testing utilities",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return cfg.Default()
		},
	}

	cmd.AddCommand(
		newRunCommand(),
		newSetupCommand(),
		newTeardownCommand(),
	)

	cmd.PersistentFlags().StringVar(&cfg.testID, "test-id", "", "unique e2e test id")
	cmd.PersistentFlags().BoolVar(&cfg.reuse, "reuse", true, "Reuse existing e2e-test environment if exists")
	return cmd
}

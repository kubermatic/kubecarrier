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

package sut

import (
	"strings"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
)

func NewCommand(log logr.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use: "sut",
		Long: strings.TrimSpace(`
SUT is the primary KubeCarrier debugging utility.

Upon running it "replaces" a running deployment with telepresence and prepares the necessary IDE tasks for its proper running with approximately similar config (mounts are remapped due to different mount points, i.e. containers root isn't your host's root). Currently, the following components should work:

* [x] apiserver
* [x] catapult
* [x] ferry
* [x] eleveator
* [x] manager
* [x] operator

and the following features are implemented:

* [x] RBAC --> running SUT tasks use the kubernetes service account credentials which our component has in the cluster
* [x] mount rebinding --> SUT tasks should have its args/env filepath remapped to different root mount point
* [x] webhook support --> the right ports should be exposed, thus when running SUT task the webhooks reaches the local process
* [x] component auto-discovery; a fzf based picker shows which component exists for faster CLI usage

Currently not implemented:
* [ ] kube DNS & clusterIP routing --> if the host's process tries using in-cluster DNS/clusterIPs this probably won't work for now.

Examples:

kubectl kubecarrier sut manager

the tool shall pause and print the telepresence command you ought to run in separate terminal. After the command reaches running state press any key to continue. The SUT utility shall generate necessary InteliJ/vscode configuration ready for this component.
`),
		Short: "SUT connects to running cluster, replaces target KubeCarrier's component with telepresence, and configures IDE task setup for debugging purposes",
	}

	cmd.AddCommand(
		newSUTSubcommand(log, "apiserver"),
		newSUTSubcommand(log, "catapult", "--enable-leader-election=false"),
		newSUTSubcommand(log, "elevator", "--enable-leader-election=false"),
		newSUTSubcommand(log, "ferry", "--enable-leader-election=false"),
		newSUTSubcommand(log, "manager", "--enable-leader-election=false"),
		newSUTSubcommand(log, "operator", "--enable-leader-election=false"),
	)
	return cmd
}

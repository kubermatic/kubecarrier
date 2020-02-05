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
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func newSUTManagerCommand(log logr.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use: "manager",
		Long: strings.TrimSpace(`

`),
		Short: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetConfig()
			if err != nil {
				return fmt.Errorf("config: %w", err)
			}
			cl, err := client.New(cfg, client.Options{})
			if err != nil {
				return fmt.Errorf("client: %w", err)
			}

			depl := &appsv1.DeploymentList{}
			ctx := context.Background()
			if err := cl.List(ctx, depl, client.MatchingLabels{
				"kubecarrier.io/role": "manager",
			}); err != nil {
				return fmt.Errorf("listing deployments: %w", err)
			}
			choices := bytes.NewBufferString("")
			choiceToDep := make(map[string]*appsv1.Deployment)
			for _, dep := range depl.Items {
				choice := fmt.Sprintf("%-40s %s", dep.Name, dep.Namespace)
				choiceToDep[choice] = &dep
				_, _ = fmt.Fprintln(choices, choice)
			}
			fzfCmd := exec.CommandContext(ctx, "fzf", "--header="+fmt.Sprintf("%-40s %s", "NAME", "NAMESPACE"))
			fzfCmd.Stdin = choices
			fzfCmd.Stderr = os.Stderr
			picker, err := fzfCmd.Output()
			if err != nil {
				return fmt.Errorf("cannot exec fzf", err)
			}
			deployment, ok := choiceToDep[strings.TrimSpace(string(picker))]
			if !ok {
				return fmt.Errorf("invalid choice")
			}
			fmt.Println(deployment)
			return nil
		},
	}
	return cmd
}

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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/kubermatic/kubecarrier/pkg/internal/ide"
)

func newSUTManagerCommand(log logr.Logger) *cobra.Command {
	var (
		extraArgs         []string
		ldFlags           string
		manualTelepresnce bool
	)

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
			ctx, closeCtx := context.WithCancel(context.Background())
			defer closeCtx()

			if err := cl.List(ctx, depl, client.MatchingLabels{
				"kubecarrier.io/role": "manager",
			}); err != nil {
				return fmt.Errorf("listing deployments: %w", err)
			}

			deployment, err := pickDeployment(ctx, depl, cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			if len(deployment.Spec.Template.Spec.Containers) > 1 {
				return fmt.Errorf("only single container allowed")
			}

			tmpDir, err := ioutil.TempDir("", "sut-")
			if err != nil {
				return fmt.Errorf("tmpdir: %w", err)
			}
			log.Info("created temp directory", "dir", tmpDir)
			rootMount := path.Join(tmpDir, "rootfs")
			envJson := path.Join(tmpDir, "env.json")
			logFile := path.Join(tmpDir, "telepresence.log")

			telepresenceArgs := []string{
				"--swap-deployment", deployment.Name,
				"--namespace", deployment.Namespace,
				"--mount", rootMount,
				"--env-json", envJson,
				"--logfile", logFile,
			}

			container := deployment.Spec.Template.Spec.Containers[0]
			for _, portSpec := range container.Ports {
				telepresenceArgs = append(telepresenceArgs, "--expose", fmt.Sprint(portSpec.ContainerPort))
			}

			if manualTelepresnce {
				log.Info("=== manually run telepresence! ===")
				_, _ = fmt.Fprintln(
					cmd.OutOrStdout(),
					"run:", "telepresence", strings.Join(telepresenceArgs, " "),
					"\npress any key to continue...",
				)
				b := make([]byte, 1)
				// pause
				_, _ = cmd.InOrStdin().Read(b)
			} else {
				telepresenceCmd := exec.CommandContext(ctx, "telepresence", telepresenceArgs...)
				telepresenceCmd.Dir = tmpDir
				telepresenceCmd.Stdin = cmd.InOrStdin()
				telepresenceCmd.Stderr = cmd.ErrOrStderr()
				telepresenceCmd.Stdout = cmd.OutOrStdout()
				log.Info("calling telepresence:", "args", telepresenceArgs)
				if err := telepresenceCmd.Start(); err != nil {
					return fmt.Errorf("cannot start telepresence: %w", err)
				}
				// wait for telepresence to start!
				time.Sleep(time.Hour)
			}

			volumeReplacementMap := make(map[string]string)
			for _, mount := range container.VolumeMounts {
				volumeReplacementMap[mount.MountPath] = path.Join(rootMount, mount.MountPath)
			}

			envBytes, err := ioutil.ReadFile(envJson)
			env := make(map[string]string)
			if err := json.Unmarshal(envBytes, env); err != nil {
				return fmt.Errorf("cannot unmarshall the environment")
			}

			hostContainerArgs := make([]string, 0, len(container.Args))
			for _, arg := range container.Args {
				arg, err := k8sExpandEnvArg(arg, env)
				if err != nil {
					return fmt.Errorf("expanding arg: %w", err)
				}
				for containerPath, hostPath := range volumeReplacementMap {
					arg = strings.ReplaceAll(arg, containerPath, hostPath)
				}
				hostContainerArgs = append(hostContainerArgs, arg)
			}

			hostContainerArgs = append(hostContainerArgs, extraArgs...)
			task := ide.Task{
				Name:    "SUT",
				Program: "manager",
				Args:    hostContainerArgs,
				Env:     env,
				LDFlags: ldFlags,
			}
			ide.GenerateIntelijJTasks([]ide.Task{task}, ".")
			ide.GenerateVSCode([]ide.Task{task}, ".")
			return nil
		},
	}
	cmd.Flags().StringVar(&ldFlags, "ld-flags", "", "ld-flags to pass to go compiler upon running this")
	cmd.Flags().StringArrayVar(&extraArgs, "extra-flags", nil, "extra flags to pass to the running task")
	cmd.Flags().BoolVar(&manualTelepresnce, "manual-telepresence", false, "manually run telepresence")
	return cmd
}

func pickDeployment(ctx context.Context, depl *appsv1.DeploymentList, stdErr io.Writer) (*appsv1.Deployment, error) {
	choices := bytes.NewBufferString("")
	choiceToDep := make(map[string]*appsv1.Deployment)
	for _, dep := range depl.Items {
		choice := fmt.Sprintf("%-40s %s", dep.Name, dep.Namespace)
		choiceToDep[choice] = &dep
		_, _ = fmt.Fprintln(choices, choice)
	}

	fzfCmd := exec.CommandContext(ctx, "fzf", "--header="+fmt.Sprintf("%-40s %s", "NAME", "NAMESPACE"))
	fzfCmd.Stdin = choices
	fzfCmd.Stderr = stdErr
	picker, err := fzfCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("cannot exec fzf: %w", err)
	}
	deployment, ok := choiceToDep[strings.TrimSpace(string(picker))]
	if !ok {
		return nil, fmt.Errorf("invalid choice")
	}
	return deployment, nil
}

func k8sExpandEnvArg(arg string, env map[string]string) (string, error) {
	for {
		idx := strings.Index(arg, "$(")
		if idx == -1 {
			return arg, nil
		}

		// handle escaping
		for idx > 0 && arg[idx-1:idx] == "$" {
			newIdx := strings.Index(arg[idx+2:], "$(")
			if newIdx == -1 {
				return arg, nil
			}
			idx += newIdx + 2
		}

		// find closing paren
		delta := strings.Index(arg[idx:], ")")
		if delta == -1 {
			return "", fmt.Errorf("unmatched (")
		}
		envElem := arg[idx+2 : idx+delta]
		envVal, ok := env[envElem]
		if !ok {
			return "", fmt.Errorf("cannot find %s in supplied envVars", envVal)
		}
		arg = arg[:idx] + envVal + arg[idx+delta+1:]
	}
}

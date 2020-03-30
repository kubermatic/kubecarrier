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

package util

import (
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/ssh/terminal"
	ctrl "sigs.k8s.io/controller-runtime"
	corezap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// CmdLogMixin adds necessary CLI flags for logging and setups the controller runtime log
func CmdLogMixin(cmd *cobra.Command) *cobra.Command {
	dev := cmd.PersistentFlags().Bool("development", terminal.IsTerminal(int(os.Stdout.Fd())), "format output for console")
	v := cmd.PersistentFlags().Int8P("verbose", "v", 4, "verbosity level")

	if cmd.PersistentPreRunE != nil {
		parent := cmd.PersistentPreRunE
		cmd.PersistentPreRunE = func(c *cobra.Command, args []string) error {
			ctrl.SetLogger(corezap.New(func(options *corezap.Options) {
				level := zap.NewAtomicLevelAt(zapcore.Level(-*v))
				options.Level = &level
				options.Development = *dev
			}))
			return parent(c, args)
		}
		return cmd
	}

	parent := cmd.PersistentPreRun
	cmd.PersistentPreRun = func(c *cobra.Command, args []string) {
		ctrl.SetLogger(corezap.New(func(options *corezap.Options) {
			level := zap.NewAtomicLevelAt(zapcore.Level(-*v))
			options.Level = &level
			options.Development = *dev
		}))
		if parent != nil {
			parent(c, args)
		}
	}
	return cmd
}

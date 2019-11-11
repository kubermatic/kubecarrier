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

package completion

import (
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kind/cmd/kind/completion/bash"
	"sigs.k8s.io/kind/cmd/kind/completion/zsh"
)

// NewCommand returns a new cobra.Command for shell completion code generation
func NewCommand(log logr.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "completion",
		Short: "Output shell completion code for the specified shell (bash or zsh)",
		Long:  longDescription,
	}
	cmd.AddCommand(
		zsh.NewCommand(),
		bash.NewCommand())
	return cmd
}

const longDescription = `
Output shell completion code for the specified shell (bash or zsh). The shell code must be evaluated to provide interactive completion of anchor commands.  This can be done by sourcing it from the .bash_profile.

 Note for zsh users: [1] zsh completions are only supported in versions of zsh >= 5.2

Examples:
  # Installing bash completion on macOS using homebrew
  ## If running Bash 3.2 included with macOS
  brew install bash-completion
  ## or, if running Bash 4.1+
  brew install bash-completion@2
  ## If anchor is installed via homebrew, this should start working immediately.
  ## If you've installed via other means, you may need add the completion to your completion directory
  anchor completion bash > $(brew --prefix)/etc/bash_completion.d/anchor


  # Installing bash completion on Linux
  ## If bash-completion is not installed on Linux, please install the 'bash-completion' package
  ## via your distribution's package manager.
  ## Load the anchor completion code for bash into the current shell
  source <(anchor completion bash)
  ## Write bash completion code to a file and source if from .bash_profile
  anchor completion bash > ~/.kube/completion.bash.inc
  printf "
  # Anchor shell completion
  source '$HOME/.kube/completion.bash.inc'
  " >> $HOME/.bash_profile
  source $HOME/.bash_profile

  # Load the anchor completion code for zsh[1] into the current shell
  source <(anchor completion zsh)
  # Set the anchor completion code for zsh[1] to autoload on startup
  anchor completion zsh > "${fpath[1]}/_anchor"

Usage:
  anchor completion SHELL [options]
`

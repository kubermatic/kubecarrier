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
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
)

func newRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "run all end2end test",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := ctrl.Log.WithName("e2e-test runner")
			log.Info("running e2e tests")
			return nil
		},
	}
	return cmd
}

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

package deletecmd

import (
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/utils/pkg/util"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(apiextensionsv1.AddToScheme(scheme))
	utilruntime.Must(catalogv1alpha1.AddToScheme(scheme))
	utilruntime.Must(operatorv1alpha1.AddToScheme(scheme))
	utilruntime.Must(corev1alpha1.AddToScheme(scheme))
}

// NewKubecarrierCLI creates the root command for the KubeCarrier CLI.
func NewDeleteCommand(log logr.Logger) *cobra.Command {
	flags := genericclioptions.NewConfigFlags(false)
	var cl = new(util.ClientWatcher)

	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "delete",
		Short: "delete KubeCarrier components",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := flags.ToRESTConfig()
			if err != nil {
				return err
			}
			clL, err := util.NewClientWatcher(
				cfg,
				scheme,
				log,
			)
			if err != nil {
				return err
			}
			*cl = *clL
			return err
		},
	}
	flags.AddFlags(cmd.PersistentFlags())
	cmd.AddCommand(newAccountCommand(log, cl))
	return cmd
}

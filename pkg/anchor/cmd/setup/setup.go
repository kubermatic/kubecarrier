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

package setup

import (
	"context"
	"fmt"
	"os"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/kubermatic/kubecarrier/pkg/kustomize"
	"github.com/kubermatic/kubecarrier/pkg/reconcile"
	"github.com/kubermatic/kubecarrier/pkg/resources/operator"

	"github.com/gernest/wow"
	"github.com/gernest/wow/spin"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/kubermatic/kubecarrier/pkg/anchor/spinner"
)

type flags struct {
	// KubeConfig is the absolute path of the kubeconfig of the kubernetes cluster which you want to deploy kubecarrier.
	KubeConfig string
}

var (
	scheme = runtime.NewScheme()
)

const (
	kubeconfigEnv        = "KUBECONFIG"
	kubecarrierNamespace = "kubecarrier-system"
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
}

func NewCommand(log logr.Logger) *cobra.Command {
	flags := &flags{}
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "setup",
		Short: "Deploy kubecarrier operator",
		Long: `Deploy kubecarrier operator in a kubernetes cluster.
Here are some examples:
- You can specify the kubeconfig absolute path of the cluster that you want to deploy everything in it:
$ anchor setup --kubeconfig=<kubeconfig path>
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runE(flags, log, cmd)
		},
	}

	cmd.Flags().StringVar(&flags.KubeConfig, "kubeconfig", "", "The absolute path of the kubeconfig of kubernetes cluster that set up with. if you don't specify the flag, it will read from the KUBECONFIG environment variable.")
	return cmd
}

func runE(flags *flags, log logr.Logger, cmd *cobra.Command) error {
	ctx := context.Background()
	s := wow.New(cmd.OutOrStdout(), spin.Get(spin.Dots), "")

	// Check the kubeconfig
	if err := spinner.AttachSpinnerTo(s, "Check kubeconfig", func() error {
		kubeconfigPath, err := checkKubeConfig(flags.KubeConfig)
		if err != nil {
			return err
		}

		// Set the kubeconfig environment variable so the client in the following can work with the cluster.
		if err := os.Setenv(kubeconfigEnv, kubeconfigPath); err != nil {
			return nil
		}
		return nil
	}); err != nil {
		return err
	}

	// Get a client from the configuration of the kubernetes cluster.
	conf, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("getting Kubernetes cluster config: %w", err)
	}
	c, err := client.New(conf, client.Options{Scheme: scheme})
	if err != nil {
		return fmt.Errorf("creating Kubernetes client: %w", err)
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: kubecarrierNamespace,
		},
	}
	if err := spinner.AttachSpinnerTo(s, "Create Kubecarrier System Namespace", createNamespace(ctx, c, ns)); err != nil {
		return nil
	}

	if err := spinner.AttachSpinnerTo(s, "Deploy Kubecarrier Operator", reconcileOperator(ctx, log, c, ns)); err != nil {
		return nil
	}

	return nil
}

func createNamespace(ctx context.Context, c client.Client, ns *corev1.Namespace) func() error {
	return func() error {
		if err := c.Create(ctx, ns); err != nil {
			if errors.IsAlreadyExists(err) {
				if err := c.Get(ctx, types.NamespacedName{Name: ns.ObjectMeta.Name}, ns); err != nil {
					return fmt.Errorf("getting Kubecarrier system namespace: %v", err)
				}
				return nil
			} else {
				return fmt.Errorf("creating Kubecarrier system namespace: %v", err)
			}
		}
		return nil
	}
}

func checkKubeConfig(kubeconfig string) (string, error) {
	kubeConfigPath := strings.TrimSpace(kubeconfig)
	if kubeConfigPath == "" {
		kubeConfigPath = strings.TrimSpace(os.Getenv("KUBECONFIG"))
	}

	if kubeConfigPath == "" {
		return "", fmt.Errorf("either $KUBECONFIG or --kubeconfig flag needs to be set")
	}

	kubeConfigStat, err := os.Stat(kubeConfigPath)
	if err != nil {
		return "", fmt.Errorf("checking the kubeconfig path: %w", err)
	}
	// Check the kubeconfig path points to a file
	if !kubeConfigStat.Mode().IsRegular() {
		return "", fmt.Errorf("kubeconfig path %s does not point to a file", kubeConfigPath)
	}
	return kubeConfigPath, nil
}

func reconcileOperator(ctx context.Context, log logr.Logger, c client.Client, kubecarrierNamespace *corev1.Namespace) func() error {
	return func() error {
		// Kustomize Build
		objects, err := operator.Manifests(
			kustomize.NewDefaultKustomize(),
			operator.Config{
				Namespace: kubecarrierNamespace.Name,
			})
		if err != nil {
			return fmt.Errorf("creating operator manifests: %w", err)
		}

		for _, object := range objects {
			if err := controllerutil.SetControllerReference(kubecarrierNamespace, &object, scheme); err != nil {
				return fmt.Errorf("set controller reference: %w", err)
			}
			_, err := reconcile.Unstructured(ctx, log, c, &object)
			if err != nil {
				return fmt.Errorf("reconcile type: %s, err: %w", object.GroupVersionKind().Kind, err)
			}
		}
		return nil
	}
}

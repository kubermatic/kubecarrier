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

package setup

import (
	"context"
	"fmt"
	"time"

	"github.com/gernest/wow"
	"github.com/gernest/wow/spin"
	"github.com/go-logr/logr"
	certv1alpha2 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/cli/internal/cmd/preflight/checkers"
	"github.com/kubermatic/kubecarrier/pkg/cli/internal/spinner"
	"github.com/kubermatic/kubecarrier/pkg/internal/constants"
	"github.com/kubermatic/kubecarrier/pkg/internal/reconcile"
	"github.com/kubermatic/kubecarrier/pkg/internal/resources/operator"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(apiextensionsv1.AddToScheme(scheme))
	utilruntime.Must(operatorv1alpha1.AddToScheme(scheme))
	utilruntime.Must(certv1alpha2.AddToScheme(scheme))
}

func NewCommand(log logr.Logger) *cobra.Command {
	var skipPreflight bool
	flags := genericclioptions.NewConfigFlags(false)
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "setup",
		Short: "Deploy KubeCarrier operator",
		Long: `Deploy KubeCarrier operator in a kubernetes cluster.
Here are some examples:
- You can specify the kubeconfig absolute path of the cluster that you want to deploy everything in it:
$ kubectl kubecarrier setup --kubeconfig=<kubeconfig path>
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := flags.ToRESTConfig()
			if err != nil {
				return err
			}
			return runE(cfg, log, cmd, skipPreflight)
		},
	}
	cmd.Flags().BoolVar(&skipPreflight, "skip-preflight-checks", false, "If true, preflight checks will be skipped")
	flags.AddFlags(cmd.Flags())
	return cmd
}

func runE(conf *rest.Config, log logr.Logger, cmd *cobra.Command, skipPreflight bool) error {
	stopCh := ctrl.SetupSignalHandler()
	ctx, cancelContext := context.WithTimeout(context.Background(), 60*time.Second)
	go func() {
		<-stopCh
		cancelContext()
	}()

	s := wow.New(cmd.OutOrStdout(), spin.Get(spin.Dots), "")
	startTime := time.Now()

	if !skipPreflight {
		if err := checkers.RunChecks(conf, s, startTime, log); err != nil {
			return err
		}
	}
	// Get a client from the configuration of the kubernetes cluster.
	c, err := util.NewClientWatcher(conf, scheme, log)
	if err != nil {
		return fmt.Errorf("creating Kubernetes client: %w", err)
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.KubeCarrierDefaultNamespace,
		},
	}
	if err := spinner.AttachSpinnerTo(s, startTime, fmt.Sprintf("Create %q Namespace", ns.Name), createNamespace(ctx, c, ns)); err != nil {
		return fmt.Errorf("creating KubeCarrier system namespace: %w", err)
	}

	if err := spinner.AttachSpinnerTo(s, startTime, "Deploy KubeCarrier Operator", reconcileOperator(ctx, log, c, ns)); err != nil {
		return fmt.Errorf("deploying KubeCarrier operator: %w", err)
	}

	if err := spinner.AttachSpinnerTo(s, startTime, "Deploy KubeCarrier", deployKubeCarrier(ctx, conf)); err != nil {
		return fmt.Errorf("deploying KubeCarrier controller manager: %w", err)
	}

	return nil
}

func createNamespace(ctx context.Context, c client.Client, ns *corev1.Namespace) func() error {
	return func() error {
		if err := c.Create(ctx, ns); err != nil {
			if errors.IsAlreadyExists(err) {
				if err := c.Get(ctx, types.NamespacedName{Name: ns.ObjectMeta.Name}, ns); err != nil {
					return fmt.Errorf("getting KubeCarrier system namespace: %w", err)
				}
				return nil
			} else {
				return fmt.Errorf("creating KubeCarrier system namespace: %w", err)
			}
		}
		return nil
	}
}

func reconcileOperator(ctx context.Context, log logr.Logger, c *util.ClientWatcher, kubecarrierNamespace *corev1.Namespace) func() error {
	return func() error {
		// Kustomize Build
		objects, err := operator.Manifests(
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
				return fmt.Errorf("reconcile kind: %s, err: %w", object.GroupVersionKind().Kind, err)
			}
		}

		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kubecarrier-operator-manager",
				Namespace: constants.KubeCarrierDefaultNamespace,
			},
		}
		return c.WaitUntil(ctx, deployment, func() (done bool, err error) {
			return util.DeploymentIsAvailable(deployment), nil
		})

	}
}

// deployKubeCarrier deploys the KubeCarrier Object in a kubernetes cluster.
func deployKubeCarrier(ctx context.Context, conf *rest.Config) func() error {
	return func() error {
		// Create another client due to some issues about the restmapper.
		// The issue is that if you use the client that created before, and here try to create the kubeCarrier,
		// it will complain about: `no matches for kind "KubeCarrier" in version "operator.kubecarrier.io/v1alpha1"`,
		// but actually, the scheme is already added to the runtime scheme.
		// And in the following, reinitializing the client solves the issue.

		kubeCarrier := &operatorv1alpha1.KubeCarrier{
			ObjectMeta: metav1.ObjectMeta{
				Name: constants.KubeCarrierDefaultName,
			},
			Spec: operatorv1alpha1.KubeCarrierSpec{
				Master: true,
			},
		}
		w, err := util.NewClientWatcher(conf, scheme, ctrl.Log)
		if err != nil {
			return err
		}
		if _, err := ctrl.CreateOrUpdate(ctx, w, kubeCarrier, func() error {
			return nil
		}); err != nil {
			return fmt.Errorf("cannot create or update KubeCarrier: %w", err)
		}
		return w.WaitUntil(ctx, kubeCarrier, func() (done bool, err error) {
			return kubeCarrier.IsReady(), nil
		})
	}
}

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

package e2e_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gernest/wow"
	"github.com/gernest/wow/spin"
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	certv1alpha2 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/cli/internal/spinner"
	"github.com/kubermatic/kubecarrier/pkg/internal/reconcile"
	"github.com/kubermatic/kubecarrier/pkg/internal/resources/fakeoperator"
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

func newSetupE2EOperator(log logr.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup-e2e-operator",
		Short: "install e2e operator in the given cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig := cmd.Flag("kubeconfig").Value.String()
			namespaceName := cmd.Flag("namespace").Value.String()
			if err := setupE2EOperator(log, kubeconfig, namespaceName, cmd.OutOrStdout()); err != nil {
				return fmt.Errorf("setup-e2e: %w", err)
			}
			return nil
		},
	}
	cmd.Flags().String("kubeconfig", os.Getenv("KUBECONFIG"), "cluster kubeconfig where to install")
	cmd.Flags().StringP("namespace", "n", "kubecarrier-e2e-operator", "namespace where to deploy operator")
	return cmd
}

func setupE2EOperator(log logr.Logger, kubeconfig string, namespaceName string, output io.Writer) error {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}
	if kubeconfig != "" {
		if err := os.Setenv("KUBECONFIG", kubeconfig); err != nil {
			return fmt.Errorf("kubeconfig env setup: %w", err)
		}
	}
	ctx := context.Background()
	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("getting Kubernetes cluster config: %w", err)
	}
	// Get a client from the configuration of the kubernetes cluster.
	c, err := util.NewClientWatcher(cfg, scheme, log)
	if err != nil {
		return fmt.Errorf("creating Kubernetes client: %w", err)
	}

	s := wow.New(output, spin.Get(spin.Dots), "spinner text")

	startTime := time.Now()
	if err := spinner.AttachSpinnerTo(s, startTime, "creating namespace", func() error {
		_, err := controllerutil.CreateOrUpdate(ctx, c, namespace, func() error {
			return nil
		})
		return err

	}); err != nil {
		return err
	}

	if err := spinner.AttachSpinnerTo(s, startTime, "deploy e2e-operator", func() error {
		objects, err := fakeoperator.Manifests(
			fakeoperator.Config{
				Namespace: namespace.Name,
			})
		if err != nil {
			return fmt.Errorf("creating operator manifests: %w", err)
		}

		for _, object := range objects {
			if err := controllerutil.SetControllerReference(namespace, &object, scheme); err != nil {
				return fmt.Errorf("set controller reference: %w", err)
			}
			b, err := yaml.Marshal(object)
			if err != nil {
				// this should never ever happen
				panic(err)
			}
			log.V(9).Info("Creating object\n" + string(b))
			_, err = reconcile.Unstructured(ctx, log, c, &object)
			if err != nil {
				return fmt.Errorf("reconcile kind: %s, err: %w", object.GroupVersionKind().Kind, err)
			}
			log.V(6).Info("reconciled",
				"name", object.GetName(),
				"namespace", object.GetNamespace(),
				"kind", object.GroupVersionKind().Kind,
				"group", object.GroupVersionKind().Group,
				"version", object.GroupVersionKind().Version,
			)
		}
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "e2e-operator-manager",
				Namespace: namespace.Name,
			},
		}
		return c.WaitUntil(ctx, deployment, func() (done bool, err error) {
			return util.DeploymentIsAvailable(deployment), nil
		})
	}); err != nil {
		return err
	}
	return nil

}

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

	"github.com/kubermatic/kubecarrier/pkg/internal/resources/fakeoperator"

	"github.com/gernest/wow"
	"github.com/gernest/wow/spin"
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/kubermatic/kubecarrier/pkg/internal/util"

	"github.com/kubermatic/kubecarrier/pkg/anchor/internal/spinner"
	"github.com/kubermatic/kubecarrier/pkg/internal/kustomize"
	"github.com/kubermatic/kubecarrier/pkg/internal/reconcile"
)

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
	scheme := runtime.NewScheme()

	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return fmt.Errorf("cannot add clientgo scheme: %w", err)
	}
	if err := apiextensionsv1beta1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("cannot add apiextenstions v1beta1 scheme: %w", err)
	}

	c, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return fmt.Errorf("creating Kubernetes client: %w", err)
	}
	s := wow.New(output, spin.Get(spin.Dots), "spinner text")
	s.Start()
	defer s.Stop()

	if err := spinner.AttachSpinnerTo(s, "creating namespace", func() error {
		_, err := controllerutil.CreateOrUpdate(ctx, c, namespace, func() error {
			return nil
		})
		return err

	}); err != nil {
		return err
	}

	if err := spinner.AttachSpinnerTo(s, "deploy e2e-operator", func() error {
		objects, err := fakeoperator.Manifests(
			kustomize.NewDefaultKustomize(),
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
		return nil
	}); err != nil {
		return err
	}

	if err := spinner.AttachSpinnerTo(s, "waiting for available deployment", func() error {
		log.Info("querying deployment")
		return wait.Poll(time.Second, 60*time.Second, func() (done bool, err error) {
			deployment := &appsv1.Deployment{}
			err = c.Get(ctx, types.NamespacedName{
				Name:      "fake-controller-manager",
				Namespace: namespace.Name,
			}, deployment)
			switch {
			case errors.IsNotFound(err):
				return false, nil
			case err != nil:
				return false, err
			default:
				return util.DeploymentIsAvailable(deployment), nil
			}
		})
	}); err != nil {
		return err
	}
	return nil

}

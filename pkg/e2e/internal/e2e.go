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

package e2e

import (
	"os"

	"github.com/kubermatic/kubecarrier/pkg/e2e/internal/controllers"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	corescheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	e2ev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/e2e/v1alpha1"
	e2ev1alpha2 "github.com/kubermatic/kubecarrier/pkg/apis/e2e/v1alpha2"
)

func NewE2E() *cobra.Command {
	var metricsAddr string
	var enableLeaderElection bool

	cmd := &cobra.Command{}
	cmd.Flags().StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	cmd.Flags().BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		var (
			scheme   = runtime.NewScheme()
			setupLog = ctrl.Log.WithName("setup")
		)

		_ = e2ev1alpha1.AddToScheme(scheme)
		_ = e2ev1alpha2.AddToScheme(scheme)
		_ = corescheme.AddToScheme(scheme)

		ctrl.SetLogger(zap.New(func(o *zap.Options) {
			o.Development = true
		}))

		mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
			Scheme:             scheme,
			MetricsBindAddress: metricsAddr,
			LeaderElection:     enableLeaderElection,
			Port:               9443,
		})
		if err != nil {
			setupLog.Error(err, "unable to start manager")
			os.Exit(1)
		}

		if err = (&controllers.JokeReconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("Joke"),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "Joke")
			os.Exit(1)
		}

		// TODO --> try running this & test the conversion webhooks, etc
		if os.Getenv("ENABLE_WEBHOOKS") != "" {
			if err = (&e2ev1alpha2.Joke{}).SetupWebhookWithManager(mgr); err != nil {
				setupLog.Error(err, "unable to create webhook", "webhook", "Joke")
				os.Exit(1)
			}
		}

		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			setupLog.Error(err, "problem running manager")
			os.Exit(1)
		}

		return nil
	}
	return cmd
}

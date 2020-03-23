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

package fakeoperator

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	zap2 "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	corescheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	fakev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/fake/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/fakeoperator/internal/controllers"
)

type flags struct {
	metricsAddr          string
	enableLeaderElection bool
	verbosity            int8
	healthAddr           string
}

func NewFakeOperator() *cobra.Command {
	flags := flags{}

	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "e2e-operator",
		Short: "e2e-operator runs the dummy joke operator for e2e testing purposes",
	}

	cmd.Flags().StringVar(&flags.metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	cmd.Flags().BoolVar(&flags.enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	cmd.Flags().StringVar(&flags.healthAddr, "health-addr", ":9440", "The address the health endpoint binds to.")
	cmd.Flags().Int8VarP(&flags.verbosity, "verbosity", "v", 4, "log level version")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return run(flags, ctrl.Log.WithName("setup"))
	}
	return cmd
}

func run(flags flags, log logr.Logger) error {
	var (
		scheme = runtime.NewScheme()
	)

	_ = fakev1alpha1.AddToScheme(scheme)
	_ = corescheme.AddToScheme(scheme)
	_ = v1beta1.AddToScheme(scheme)

	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
		l := zap2.NewAtomicLevelAt(zapcore.Level(-flags.verbosity))
		o.Level = &l
	}))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     flags.metricsAddr,
		LeaderElection:         flags.enableLeaderElection,
		Port:                   9443,
		HealthProbeBindAddress: flags.healthAddr,
	})
	if err != nil {
		return fmt.Errorf("new manager creation: %w", err)
	}

	if err = (&controllers.DBReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Joke"),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("setup Joke controller: %w", err)
	}

	if err := mgr.AddReadyzCheck("ping", healthz.Ping); err != nil {
		return fmt.Errorf("adding readyz checker: %w", err)
	}

	if err := mgr.AddHealthzCheck("ping", healthz.Ping); err != nil {
		return fmt.Errorf("adding healthz checker: %w", err)
	}

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("manager-runtime: %w", err)
	}
	fmt.Println("exit")
	return nil

}

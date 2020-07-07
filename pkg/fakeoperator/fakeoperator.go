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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	fakev1 "github.com/kubermatic/kubecarrier/pkg/apis/fake/v1"
	fakev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/fake/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/fakeoperator/internal/controllers"
	"github.com/kubermatic/kubecarrier/pkg/fakeoperator/internal/webhooks"
	utilwebhook "github.com/kubermatic/kubecarrier/pkg/internal/util/webhook"
	"github.com/kubermatic/utils/pkg/util"
)

type flags struct {
	metricsAddr          string
	healthAddr           string
	enableLeaderElection bool
	certDir              string
}

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(apiextensionsv1.AddToScheme(scheme))
	utilruntime.Must(fakev1alpha1.AddToScheme(scheme))
	utilruntime.Must(fakev1.AddToScheme(scheme))
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

func NewFakeOperator() *cobra.Command {
	flags := flags{}
	log := ctrl.Log.WithName("manager")
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "e2e-operator",
		Short: "e2e-operator runs the operator for e2e testing purposes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(flags, log)
		},
	}
	cmd.Flags().StringVar(&flags.healthAddr, "health-addr", ":9440", "The address the health endpoint binds to.")
	cmd.Flags().StringVar(&flags.certDir, "cert-dir", "/tmp/k8s-webhook-server/serving-certs", "The webhook TLS certificates directory")
	return util.CmdLogMixin(cmd)
}

func run(flags flags, log logr.Logger) error {

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
		Log:    ctrl.Log.WithName("controllers").WithName("e2e"),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("setup e2e controller: %w", err)
	}

	if err = (&controllers.SnapshotReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("e2e snapshot"),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("setup e2e snapshot controller: %w", err)
	}

	if err = (&controllers.BackupReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("e2e backup"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("setup e2e backup controller: %w", err)
	}

	if err := ctrl.NewWebhookManagedBy(mgr).For(&fakev1alpha1.DB{}).Complete(); err != nil {
		return err
	}
	// Register webhooks as handlers
	wbh := mgr.GetWebhookServer()
	wbh.Register(utilwebhook.GenerateMutateWebhookPath(&fakev1.DB{}, mgr.GetScheme()),
		&webhook.Admission{Handler: &webhooks.DBWebhookHandler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
			Log:    log.WithName("mutating webhooks").WithName("DB"),
		}})

	if err := mgr.AddReadyzCheck("ping", healthz.Ping); err != nil {
		return fmt.Errorf("adding readyz checker: %w", err)
	}

	if err := mgr.AddHealthzCheck("ping", healthz.Ping); err != nil {
		return fmt.Errorf("adding healthz checker: %w", err)
	}

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("manager-runtime: %w", err)
	}
	return nil

}

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

package tower

import (
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	masterv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/master/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
	utilwebhook "github.com/kubermatic/kubecarrier/pkg/internal/util/webhook"
	"github.com/kubermatic/kubecarrier/pkg/tower/internal/controllers"
	"github.com/kubermatic/kubecarrier/pkg/tower/internal/webhooks"
)

type flags struct {
	metricsAddr, healthAddr            string
	enableLeaderElection               bool
	managementClusterHealthCheckPeriod time.Duration
}

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(masterv1alpha1.AddToScheme(scheme))
}

const (
	componentTower = "Tower"
)

func NewTower() *cobra.Command {
	log := ctrl.Log.WithName("tower")
	flags := &flags{}
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   componentTower,
		Short: "KubeCarrier Tower",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(flags, log)
		},
	}
	cmd.Flags().StringVar(&flags.metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	cmd.Flags().StringVar(&flags.healthAddr, "health-addr", ":9440", "The address the health endpoint binds to.")
	cmd.Flags().BoolVar(&flags.enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	cmd.Flags().DurationVar(&flags.managementClusterHealthCheckPeriod, "management-cluster-health-check-period", 5*time.Second,
		"the amount of time to check the status of the management cluster periodically.")
	return util.CmdLogMixin(cmd)
}

func run(flags *flags, log logr.Logger) error {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     flags.metricsAddr,
		LeaderElection:         flags.enableLeaderElection,
		Port:                   9443,
		HealthProbeBindAddress: flags.healthAddr,
	})
	if err != nil {
		return fmt.Errorf("starting manager: %w", err)
	}

	if err = (&controllers.ManagementClusterReconciler{
		Client:                             mgr.GetClient(),
		Log:                                log.WithName("controllers").WithName("ManagementClusterReconciler"),
		Scheme:                             mgr.GetScheme(),
		ManagementClusterHealthCheckPeriod: flags.managementClusterHealthCheckPeriod,
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("creating ManagementCluster controller: %w", err)
	}

	// Register webhooks as handlers
	wbh := mgr.GetWebhookServer()

	// validating webhooks
	wbh.Register(utilwebhook.GenerateValidateWebhookPath(&masterv1alpha1.ManagementCluster{}, mgr.GetScheme()),
		&webhook.Admission{Handler: &webhooks.ManagementClusterWebhookHandler{
			Log: log.WithName("validating webhooks").WithName("ManagementCluster"),
		}})

	if err := mgr.AddReadyzCheck("ping", healthz.Ping); err != nil {
		return fmt.Errorf("adding readyz checker: %w", err)
	}

	if err := mgr.AddHealthzCheck("ping", healthz.Ping); err != nil {
		return fmt.Errorf("adding healthz checker: %w", err)
	}

	log.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("running manager: %w", err)
	}
	return nil
}

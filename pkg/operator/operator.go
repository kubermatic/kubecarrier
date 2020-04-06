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

package operator

import (
	"fmt"

	"github.com/go-logr/logr"
	certv1alpha2 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"github.com/spf13/cobra"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
	utilwebhook "github.com/kubermatic/kubecarrier/pkg/internal/util/webhook"
	"github.com/kubermatic/kubecarrier/pkg/operator/internal/controllers"
	"github.com/kubermatic/kubecarrier/pkg/operator/internal/webhooks"
)

type flags struct {
	metricsAddr, healthAddr string
	enableLeaderElection    bool
}

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(apiextensionsv1.AddToScheme(scheme))
	utilruntime.Must(certv1alpha2.AddToScheme(scheme))
	utilruntime.Must(operatorv1alpha1.AddToScheme(scheme))
}

const (
	componentOperator = "operator"
)

func NewOperatorCommand() *cobra.Command {
	log := ctrl.Log.WithName("operator")
	flags := &flags{}
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   componentOperator,
		Short: "deploy kubecarrier operator",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(flags, log)
		},
	}
	cmd.Flags().StringVar(&flags.metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	cmd.Flags().StringVar(&flags.healthAddr, "health-addr", ":9440", "The address the health endpoint binds to.")
	cmd.Flags().BoolVar(&flags.enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for operator. Enabling this will ensure there is only one active controller manager.")
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

	if err := controllers.NewBaseReconciler(
		&controllers.KubeCarrierStrategy{},
		mgr.GetClient(),
		mgr.GetScheme(),
		mgr.GetRESTMapper(),
		log.WithName("controllers").WithName("KubeCarrier"),
		"KubeCarrier",
		"kubecarrier.kubecarrier.io/controller").SetupWithManager(mgr); err != nil {
		return fmt.Errorf("creating KubeCarrier controller: %w", err)
	}
	if err := controllers.NewBaseReconciler(
		&controllers.FerryStrategy{},
		mgr.GetClient(),
		mgr.GetScheme(),
		mgr.GetRESTMapper(),
		log.WithName("controllers").WithName("Ferry"),
		"Ferry",
		"").SetupWithManager(mgr); err != nil {
		return fmt.Errorf("creating Ferry controller: %w", err)
	}
	if err := controllers.NewBaseReconciler(
		&controllers.CatapultStrategy{Client: mgr.GetClient()},
		mgr.GetClient(),
		mgr.GetScheme(),
		mgr.GetRESTMapper(),
		log.WithName("controllers").WithName("Catapult"),
		"Catapult",
		"catapult.kubecarrier.io/controller",
	).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("creating Catapult controller: %w", err)
	}
	if err := controllers.NewBaseReconciler(
		&controllers.ElevatorStrategy{},
		mgr.GetClient(),
		mgr.GetScheme(),
		mgr.GetRESTMapper(),
		log.WithName("controllers").WithName("Elevator"),
		"Elevator",
		"elevator.kubecarrier.io/controller").SetupWithManager(mgr); err != nil {
		return fmt.Errorf("creating Elevator controller: %w", err)
	}
	if err := controllers.NewBaseReconciler(
		&controllers.TowerStrategy{},
		mgr.GetClient(),
		mgr.GetScheme(),
		mgr.GetRESTMapper(),
		log.WithName("controllers").WithName("Tower"),
		"Tower",
		"tower.kubecarrier.io/controller").SetupWithManager(mgr); err != nil {
		return fmt.Errorf("creating Tower controller: %w", err)
	}
	if err := controllers.NewBaseReconciler(
		&controllers.APIServerStrategy{},
		mgr.GetClient(),
		mgr.GetScheme(),
		mgr.GetRESTMapper(),
		log.WithName("controllers").WithName("API-server"),
		"APIServer",
		"api-server.kubecarrier.io/controller").SetupWithManager(mgr); err != nil {
		return fmt.Errorf("creating API Server controller: %w", err)
	}

	// Register webhooks as handlers
	wbh := mgr.GetWebhookServer()

	// validating webhooks
	wbh.Register(utilwebhook.GenerateValidateWebhookPath(&operatorv1alpha1.KubeCarrier{}, mgr.GetScheme()),
		&webhook.Admission{Handler: &webhooks.KubeCarrierWebhookHandler{
			Log: log.WithName("validating webhooks").WithName("KubeCarrier"),
		}})

	if err := mgr.AddReadyzCheck("ping", healthz.Ping); err != nil {
		return fmt.Errorf("adding readyz checker: %w", err)
	}

	if err := mgr.AddHealthzCheck("ping", healthz.Ping); err != nil {
		return fmt.Errorf("adding healthz checker: %w", err)
	}

	log.Info("starting operator")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("running manager: %w", err)
	}
	return nil
}

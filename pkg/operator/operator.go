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
	"log"
	"net/http"
	_ "net/http/pprof"

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

	"github.com/kubermatic/utils/pkg/util"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	utilwebhook "github.com/kubermatic/kubecarrier/pkg/internal/util/webhook"
	"github.com/kubermatic/kubecarrier/pkg/operator/internal/controllers"
	"github.com/kubermatic/kubecarrier/pkg/operator/internal/webhooks"
)

type flags struct {
	metricsAddr, healthAddr string
	enableLeaderElection    bool
	certDir                 string
}

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(apiextensionsv1.AddToScheme(scheme))
	utilruntime.Must(certv1alpha2.AddToScheme(scheme))
	utilruntime.Must(operatorv1alpha1.AddToScheme(scheme))
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
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
	cmd.Flags().StringVar(&flags.certDir, "cert-dir", "/tmp/k8s-webhook-server/serving-certs", "The webhook TLS certificates directory")
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
		CertDir:                flags.certDir,
	})
	if err != nil {
		return fmt.Errorf("starting manager: %w", err)
	}

	if err = (&controllers.CatapultReconciler{
		Client:     mgr.GetClient(),
		Log:        log.WithName("controllers").WithName("Catapult"),
		Scheme:     mgr.GetScheme(),
		RESTMapper: mgr.GetRESTMapper(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("creating Catapult controller: %w", err)
	}

	if err = (&controllers.ElevatorReconciler{
		Client:     mgr.GetClient(),
		Log:        log.WithName("controllers").WithName("Elevator"),
		Scheme:     mgr.GetScheme(),
		RESTMapper: mgr.GetRESTMapper(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("creating Elevator controller: %w", err)
	}

	if err = (&controllers.FerryReconciler{
		Client:     mgr.GetClient(),
		Log:        log.WithName("controllers").WithName("Elevator"),
		Scheme:     mgr.GetScheme(),
		RESTMapper: mgr.GetRESTMapper(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("creating Ferry controller: %w", err)
	}

	if err = (&controllers.KubeCarrierReconciler{
		Client: mgr.GetClient(),
		Log:    log.WithName("controllers").WithName("KubeCarrier"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("creating KubeCarrier controller: %w", err)
	}

	if err = (&controllers.APIServerReconciler{
		Client:     mgr.GetClient(),
		Log:        log.WithName("controllers").WithName("APIServer"),
		Scheme:     mgr.GetScheme(),
		RESTMapper: mgr.GetRESTMapper(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("creating APIServer controller: %w", err)
	}

	// Register webhooks as handlers
	wbh := mgr.GetWebhookServer()

	// validating webhooks
	wbh.Register(utilwebhook.GenerateMutateWebhookPath(&operatorv1alpha1.KubeCarrier{}, mgr.GetScheme()),
		&webhook.Admission{Handler: &webhooks.KubeCarrierWebhookHandler{
			Log: log.WithName("validating webhooks").WithName("KubeCarrier"),
		}})
	wbh.Register(utilwebhook.GenerateMutateWebhookPath(&operatorv1alpha1.APIServer{}, mgr.GetScheme()),
		&webhook.Admission{Handler: &webhooks.APIServerWebhookHandler{
			Log:    log.WithName("validating webhooks").WithName("APIServer"),
			Client: mgr.GetClient(),
		}})
	wbh.Register(utilwebhook.GenerateMutateWebhookPath(&operatorv1alpha1.Catapult{}, mgr.GetScheme()),
		&webhook.Admission{Handler: &webhooks.CatapultWebhookHandler{
			Log:    log.WithName("validating webhooks").WithName("Catapult"),
			Client: mgr.GetClient(),
		}})
	wbh.Register(utilwebhook.GenerateMutateWebhookPath(&operatorv1alpha1.Elevator{}, mgr.GetScheme()),
		&webhook.Admission{Handler: &webhooks.ElevatorWebhookHandler{
			Log:    log.WithName("validating webhooks").WithName("Elevator"),
			Client: mgr.GetClient(),
		}})
	wbh.Register(utilwebhook.GenerateMutateWebhookPath(&operatorv1alpha1.Ferry{}, mgr.GetScheme()),
		&webhook.Admission{Handler: &webhooks.FerryWebhookHandler{
			Log:    log.WithName("validating webhooks").WithName("Ferry"),
			Client: mgr.GetClient(),
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

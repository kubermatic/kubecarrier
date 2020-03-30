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

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
	"github.com/kubermatic/kubecarrier/pkg/operator/internal/controllers"
)

type flags struct {
	metricsAddr          string
	enableLeaderElection bool
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
	cmd.Flags().BoolVar(&flags.enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for operator. Enabling this will ensure there is only one active controller manager.")
	return util.CmdLogMixin(cmd)
}

func run(flags *flags, log logr.Logger) error {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: flags.metricsAddr,
		LeaderElection:     flags.enableLeaderElection,
		Port:               9443,
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
		Client:     mgr.GetClient(),
		Log:        log.WithName("controllers").WithName("KubeCarrier"),
		Scheme:     mgr.GetScheme(),
		RESTMapper: mgr.GetRESTMapper(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("creating KubeCarrier controller: %w", err)
	}

	if err = (&controllers.TowerReconciler{
		Client:     mgr.GetClient(),
		Log:        log.WithName("controllers").WithName("Tower"),
		Scheme:     mgr.GetScheme(),
		RESTMapper: mgr.GetRESTMapper(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("creating Tower controller: %w", err)
	}

	log.Info("starting operator")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("running manager: %w", err)
	}
	return nil
}

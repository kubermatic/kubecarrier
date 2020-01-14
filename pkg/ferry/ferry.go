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

package ferry

import (
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/ferry/internal/controllers"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

var (
	masterScheme  = runtime.NewScheme()
	serviceScheme = runtime.NewScheme()
)

type flags struct {
	providerNamespace                string
	enableLeaderElection             bool
	serviceClusterStatusUpdatePeriod time.Duration

	// master
	masterMetricsAddr string

	// service
	serviceMetricsAddr string
	serviceKubeconfig  string
	serviceMaster      string
	serviceClusterName string
}

func init() {
	_ = apiextensionsv1.AddToScheme(serviceScheme)
	_ = clientgoscheme.AddToScheme(serviceScheme)
	_ = clientgoscheme.AddToScheme(masterScheme)
	_ = corev1alpha1.AddToScheme(masterScheme)
}

func NewFerryCommand(log logr.Logger) *cobra.Command {
	flags := &flags{}
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "ferry",
		Short: "ServiceClusterRegistration controller manager",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runE(flags, log)
		},
	}

	cmd.Flags().StringVar(&flags.providerNamespace, "provider-namespace", "", "Name of the providers namespace in the master cluster.")
	cmd.Flags().BoolVar(&flags.enableLeaderElection, "enable-leader-election", true,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	cmd.Flags().DurationVar(&flags.serviceClusterStatusUpdatePeriod, "service-cluster-status-update-period", 10*time.Second, "Specifies how often the ferry posts service cluster status to master. Note: must work with service-cluster-monitor-grace-period in kubecarrier-controller-manager.")

	// master
	cmd.Flags().StringVar(&flags.masterMetricsAddr, "master-metrics-addr", ":8080", "The address the metric endpoint binds to.")

	// service cluster client settings
	cmd.Flags().StringVar(&flags.serviceMetricsAddr, "service-cluster-metrics-addr", ":8081", "The address the metric endpoint binds to.")
	cmd.Flags().StringVar(&flags.serviceKubeconfig, "service-cluster-kubeconfig", "", "Path to the Service Cluster kubeconfig.")
	cmd.Flags().StringVar(&flags.serviceClusterName, "service-cluster-name", "", "Name of the Service Cluster the ferry is operating on.")
	for _, flagName := range []string{
		"provider-namespace",
		"service-cluster-name",
		"service-cluster-kubeconfig",
	} {
		if err := cmd.MarkFlagRequired(flagName); err != nil {
			panic(fmt.Errorf("req flag %s: %w", flagName, err))
		}

	}
	return util.CmdLogMixin(cmd)
}

func runE(flags *flags, log logr.Logger) error {
	// KubeCarrier cluster manager
	masterCfg := ctrl.GetConfigOrDie()
	var serviceCfg *rest.Config
	if flags.serviceKubeconfig == "" {
		log.Info("no serviceKubeconfig given, asuming same-cluster")
		serviceCfg = masterCfg
	} else {
		var err error
		serviceCfg, err = clientcmd.BuildConfigFromFlags(flags.serviceMaster, flags.serviceKubeconfig)
		if err != nil {
			return fmt.Errorf("unable to set up service cluster client config: %w", err)
		}
	}

	// Service cluster setup
	serviceMgr, err := ctrl.NewManager(serviceCfg, ctrl.Options{
		Scheme:             serviceScheme,
		MetricsBindAddress: flags.serviceMetricsAddr,
	})
	if err != nil {
		return fmt.Errorf("unable to start manager for service cluster: %w", err)
	}

	// Master cluster setup
	masterMgr, err := ctrl.NewManager(masterCfg, ctrl.Options{
		Scheme:                  masterScheme,
		MetricsBindAddress:      flags.masterMetricsAddr,
		LeaderElection:          flags.enableLeaderElection,
		LeaderElectionNamespace: flags.providerNamespace,
		LeaderElectionID:        "ferry-" + flags.serviceClusterName,
		Namespace:               flags.providerNamespace,
	})
	if err != nil {
		return fmt.Errorf("unable to start manager for master cluster: %w", err)
	}

	if err := util.AddOwnerReverseFieldIndex(
		serviceMgr.GetFieldIndexer(),
		log.WithName("reverseIndex").WithName("namespace"),
		&corev1.Namespace{},
	); err != nil {
		return fmt.Errorf("cannot add Namespace owner field indexer: %w", err)
	}

	serviceClusterDiscoveryClient, err := discovery.NewDiscoveryClientForConfig(serviceCfg)
	if err != nil {
		return fmt.Errorf("cannot create discovery client for service cluster: %w", err)
	}

	if err := (&controllers.ServiceClusterReconciler{
		Log:                       log.WithName("controllers").WithName("ServiceCluster"),
		MasterClient:              masterMgr.GetClient(),
		ServiceClusterVersionInfo: serviceClusterDiscoveryClient,
		ProviderNamespace:         flags.providerNamespace,
		ServiceClusterName:        flags.serviceClusterName,
		StatusUpdatePeriod:        flags.serviceClusterStatusUpdatePeriod,
	}).SetupWithManagers(masterMgr); err != nil {
		return fmt.Errorf("cannot add %s controller: %w", "ServiceCluster", err)
	}

	if err := masterMgr.Add(serviceMgr); err != nil {
		return fmt.Errorf("cannot add service mgr: %w", err)
	}

	if err := masterMgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("error running component: %w", err)
	}
	return nil
}

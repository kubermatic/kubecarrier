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

package tender

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
	"github.com/kubermatic/kubecarrier/pkg/tender/internal/controllers"
)

var (
	masterScheme  = runtime.NewScheme()
	serviceScheme = runtime.NewScheme()
)

type flags struct {
	providerNamespace                      string
	enableLeaderElection                   bool
	serviceClusterStatusUpdatePeriodString string

	// master
	masterMetricsAddr string

	// service
	serviceMetricsAddr string
	serviceKubeConfig  string
	serviceMaster      string
	serviceClusterName string
}

func init() {
	_ = apiextensionsv1beta1.AddToScheme(serviceScheme)
	_ = clientgoscheme.AddToScheme(serviceScheme)
	_ = clientgoscheme.AddToScheme(masterScheme)
	_ = corev1alpha1.AddToScheme(masterScheme)
}

func NewTenderCommand(log logr.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "tender",
		Short: "Tender controller manager",
		Run: func(cmd *cobra.Command, args []string) {
			flags := &flags{}
			run(flags, log)
		},
	}

	flags := &flags{}

	cmd.Flags().StringVar(&flags.providerNamespace, "provider-namespace", "", "Name of the providers namespace in the master cluster.")
	cmd.Flags().BoolVar(&flags.enableLeaderElection, "enable-leader-election", true,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	cmd.Flags().StringVar(&flags.serviceClusterStatusUpdatePeriodString, "service-cluster-status-update-period", "10s", "Specifies how often the tender posts service cluster status to master. Note: must work with service-cluster-monitor-grace-period in kubecarrier-controller-manager.")

	// master
	cmd.Flags().StringVar(&flags.masterMetricsAddr, "master-metrics-addr", ":8080", "The address the metric endpoint binds to.")

	// service cluster client settings
	cmd.Flags().StringVar(&flags.serviceMetricsAddr, "service-cluster-metrics-addr", ":8081", "The address the metric endpoint binds to.")
	cmd.Flags().StringVar(&flags.serviceKubeConfig, "service-cluster-kubeconfig", "", "Path to the Service Cluster kubeconfig.")
	cmd.Flags().StringVar(&flags.serviceClusterName, "service-cluster-name", "", "Name of the Service Cluster the tender is operating on.")
	return cmd
}

func run(flags *flags, log logr.Logger) {
	serviceClusterStatusUpdatePeriod, err := time.ParseDuration(flags.serviceClusterStatusUpdatePeriodString)
	if err != nil {
		log.Error(err, "unable to parse -heartbeat-period")
		os.Exit(1)
	}
	if flags.serviceClusterName == "" {
		log.Info("invalid flag", "error", "-service-cluster-name is required")
		os.Exit(1)
	}
	if flags.providerNamespace == "" {
		log.Info("invalid flag", "error", "-provider-namespace is required")
		os.Exit(1)
	}

	// KubeCarrier cluster manager
	masterCfg := ctrl.GetConfigOrDie()
	var serviceCfg *rest.Config
	if flags.serviceKubeConfig == "" {
		log.Info("no serviceKubeConfig given, asuming same-cluster")
		serviceCfg = masterCfg
	} else {
		serviceCfg, err = clientcmd.BuildConfigFromFlags(flags.serviceMaster, flags.serviceKubeConfig)
		if err != nil {
			log.Error(err, "unable to set up service cluster client config")
			os.Exit(1)
		}
	}

	// Service cluster setup
	serviceMgr, err := ctrl.NewManager(serviceCfg, ctrl.Options{
		Scheme:             serviceScheme,
		MetricsBindAddress: flags.serviceMetricsAddr,
	})
	if err != nil {
		log.Error(err, "unable to start manager for service cluster")
		os.Exit(1)
	}

	// Master cluster setup
	masterMgr, err := ctrl.NewManager(masterCfg, ctrl.Options{
		Scheme:                  masterScheme,
		MetricsBindAddress:      flags.masterMetricsAddr,
		LeaderElection:          flags.enableLeaderElection,
		LeaderElectionNamespace: flags.providerNamespace,
		LeaderElectionID:        "tender-" + flags.serviceClusterName,
		Namespace:               flags.providerNamespace,
	})
	if err != nil {
		log.Error(err, "unable to start manager for master cluster")
		os.Exit(1)
	}

	if err := util.AddOwnerReverseFieldIndex(
		serviceMgr.GetFieldIndexer(),
		log.WithName("reverseIndex").WithName("namespace"),
		&corev1.Namespace{},
	); err != nil {
		log.Error(err, "cannot add Namespace owner field indexer")
		os.Exit(2)
	}

	// Register Controllers
	if err = (&controllers.ServiceClusterReconciler{
		Log: ctrl.Log.WithName("controllers").WithName("ServiceCluster"),

		MasterClient:       masterMgr.GetClient(),
		ServiceClient:      serviceMgr.GetClient(),
		ProviderNamespace:  flags.providerNamespace,
		ServiceClusterName: flags.serviceClusterName,
		StatusUpdatePeriod: serviceClusterStatusUpdatePeriod,
	}).SetupWithManagers(serviceMgr, masterMgr); err != nil {
		log.Error(err, "unable to create controller", "controller", "ServiceCluster")
		os.Exit(1)
	}

	if err = (&controllers.CRDReferenceReconciler{
		Log: ctrl.Log.WithName("controllers").WithName("CRDReference"),

		MasterClient: masterMgr.GetClient(),
		MasterScheme: masterMgr.GetScheme(),

		ServiceClient:      serviceMgr.GetClient(),
		ServiceClusterName: flags.serviceClusterName,
	}).SetupWithManagers(serviceMgr, masterMgr); err != nil {
		log.Error(err, "unable to create controller", "controller", "CRDReference")
		os.Exit(1)
	}

	if err = (&controllers.TenantAssignmentReconciler{
		Log:          ctrl.Log.WithName("controllers").WithName("TenantAssignment"),
		MasterClient: masterMgr.GetClient(),
		MasterScheme: masterMgr.GetScheme(),

		ServiceClient:      serviceMgr.GetClient(),
		ServiceClusterName: flags.serviceClusterName,
	}).SetupWithManagers(serviceMgr, masterMgr); err != nil {
		log.Error(err, "unable to create controller", "controller", "TenantAssignment")
		os.Exit(1)
	}

	var shutdownWG sync.WaitGroup
	shutdownWG.Add(3)
	signalHandler := ctrl.SetupSignalHandler()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer shutdownWG.Done()

		select {
		case <-signalHandler:
			cancel()
		case <-ctx.Done():
		}
	}()

	go func() {
		defer shutdownWG.Done()

		log.Info("starting manager for master cluster")
		if err := masterMgr.Start(ctx.Done()); err != nil {
			log.Error(err, "problem running master cluster manager")
			cancel()
		}
	}()

	go func() {
		defer shutdownWG.Done()

		log.Info("starting manager for service cluster")
		if err := serviceMgr.Start(ctx.Done()); err != nil {
			log.Error(err, "problem running service cluster manager")
			cancel()
		}
	}()

	// wait for all go routines to stop, before exiting main
	shutdownWG.Wait()
}

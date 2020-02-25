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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/ferry/internal/controllers"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

var (
	managementScheme  = runtime.NewScheme()
	serviceScheme = runtime.NewScheme()
)

type flags struct {
	providerNamespace                string
	enableLeaderElection             bool
	serviceClusterStatusUpdatePeriod time.Duration

	// management
	managementMetricsAddr string

	// service
	serviceMetricsAddr string
	serviceKubeconfig  string
	serviceClusterName string
}

func init() {
	_ = apiextensionsv1.AddToScheme(serviceScheme)
	_ = clientgoscheme.AddToScheme(serviceScheme)
	_ = clientgoscheme.AddToScheme(managementScheme)
	_ = corev1alpha1.AddToScheme(managementScheme)
}

func NewFerryCommand(log logr.Logger) *cobra.Command {
	flags := &flags{}
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "ferry",
		Short: "Ferry controller manager",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runE(flags, log)
		},
	}

	cmd.Flags().StringVar(&flags.providerNamespace, "provider-namespace", "", "Name of the providers namespace in the management cluster.")
	cmd.Flags().BoolVar(&flags.enableLeaderElection, "enable-leader-election", true,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	cmd.Flags().DurationVar(&flags.serviceClusterStatusUpdatePeriod, "service-cluster-status-update-period", 10*time.Second, "Specifies how often the ferry posts service cluster status to management. Note: must work with service-cluster-monitor-grace-period in kubecarrier-controller-manager.")

	// management
	cmd.Flags().StringVar(&flags.managementMetricsAddr, "management-metrics-addr", ":8080", "The address the metric endpoint binds to.")

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
	managementCfg := ctrl.GetConfigOrDie()
	mgr, err := ctrl.NewManager(managementCfg, ctrl.Options{
		Scheme:                  managementScheme,
		MetricsBindAddress:      flags.managementMetricsAddr,
		LeaderElection:          flags.enableLeaderElection,
		LeaderElectionNamespace: flags.providerNamespace,
		LeaderElectionID:        "ferry-" + flags.serviceClusterName,
		Namespace:               flags.providerNamespace,
	})
	if err != nil {
		return fmt.Errorf("unable to start manager for management cluster: %w", err)
	}

	// Setup additional client and cache for Service Cluster
	serviceCfg, err := clientcmd.BuildConfigFromFlags(
		"", flags.serviceKubeconfig)
	if err != nil {
		return fmt.Errorf("reading service cluster config: %w", err)
	}
	serviceMapper, err := apiutil.NewDiscoveryRESTMapper(serviceCfg)
	if err != nil {
		return fmt.Errorf("creating service cluster rest mapper: %w", err)
	}
	serviceClient, err := client.New(serviceCfg, client.Options{
		Scheme: serviceScheme,
		Mapper: serviceMapper,
	})
	if err != nil {
		return fmt.Errorf("creating service cluster client: %w", err)
	}
	serviceCache, err := cache.New(serviceCfg, cache.Options{
		Scheme: serviceScheme,
		Mapper: serviceMapper,
	})
	if err != nil {
		return fmt.Errorf("creating service cluster cache: %w", err)
	}
	if err = mgr.Add(serviceCache); err != nil {
		return fmt.Errorf("add service cluster cache to manager: %w", err)
	}
	serviceCachedClient := &client.DelegatingClient{
		Reader:       serviceCache,
		Writer:       serviceClient,
		StatusClient: serviceClient,
	}

	serviceClusterDiscoveryClient, err := discovery.NewDiscoveryClientForConfig(serviceCfg)
	if err != nil {
		return fmt.Errorf("cannot create discovery client for service cluster: %w", err)
	}

	if err := (&controllers.ServiceClusterReconciler{
		Log:                       log.WithName("controllers").WithName("ServiceCluster"),
		ManagementClient:              mgr.GetClient(),
		ServiceClusterVersionInfo: serviceClusterDiscoveryClient,
		ProviderNamespace:         flags.providerNamespace,
		ServiceClusterName:        flags.serviceClusterName,
		StatusUpdatePeriod:        flags.serviceClusterStatusUpdatePeriod,
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("cannot add %s controller: %w", "ServiceCluster", err)
	}

	if err := (&controllers.CustomResourceDiscoveryReconciler{
		Log:                log.WithName("controllers").WithName("CustomResourceDiscovery"),
		ManagementClient:       mgr.GetClient(),
		ManagementScheme:       mgr.GetScheme(),
		ServiceClient:      serviceCachedClient,
		ServiceCache:       serviceCache,
		ServiceClusterName: flags.serviceClusterName,
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("cannot add %s controller: %w", "CustomResourceDiscovery", err)
	}

	if err := (&controllers.ServiceClusterAssignmentReconciler{
		Log:          log.WithName("controllers").WithName("ServiceClusterAssignmentReconciler"),
		ManagementClient: mgr.GetClient(),
		ManagementScheme: mgr.GetScheme(),
		// We need the uncached client here or we might create a second namespace
		// because there is a short timeframe where the cache is not yet synced
		// and the controller would think it did not yet create the namespace.
		ServiceClient:      serviceClient,
		ServiceCache:       serviceCache,
		ServiceClusterName: flags.serviceClusterName,
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("cannot add %s controller: %w", "ServiceClusterAssignmentReconciler", err)
	}

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("error running component: %w", err)
	}
	return nil
}

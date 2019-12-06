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

package dinghy

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
)

var (
	masterScheme  = runtime.NewScheme()
	serviceScheme = runtime.NewScheme()
)

type flags struct {
	enableLeaderElection                                                            bool
	crdVersion, crdKind, crdInternalGroup, crdExternalGroup, crdServiceClusterGroup string
	crdConfigurationName, providerNamespace                                         string

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

func NewDinghyCommand(log logr.Logger) *cobra.Command {
	flags := &flags{}
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "Dinghy",
		Short: "dinghy controller manager",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runE(flags, log)
		},
	}

	cmd.Flags().StringVar(&flags.masterMetricsAddr, "master-metrics-addr", ":8080", "The address the metric endpoint binds to.")
	cmd.Flags().BoolVar(&flags.enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")

	// service cluster client settings
	cmd.Flags().StringVar(&flags.serviceMetricsAddr, "service-cluster-metrics-addr", ":8081", "The address the metric endpoint binds to.")
	cmd.Flags().StringVar(&flags.serviceKubeConfig, "service-cluster-kubeconfig", "", "Path to the Service Cluster kubeconfig.")
	cmd.Flags().StringVar(&flags.serviceMaster, "service-cluster-master", "", "The address of the Service Clusters Kubernetes API server. Overrides any value in kubeconfig. "+
		"Only required if out-of-cluster.")
	cmd.Flags().StringVar(&flags.serviceClusterName, "service-cluster-name", "", "Name of the Service Cluster the tender is operating on.")

	// Settings
	cmd.Flags().StringVar(&flags.crdVersion, "crd-version", "", "The API Version to watch.")
	cmd.Flags().StringVar(&flags.crdKind, "crd-kind", "", "The API Kind to watch.")
	cmd.Flags().StringVar(&flags.crdInternalGroup, "crd-internal-group", "", "The internal API Group to watch.")
	cmd.Flags().StringVar(&flags.crdExternalGroup, "crd-external-group", "", "The external API Group to watch.")
	cmd.Flags().StringVar(&flags.crdServiceClusterGroup, "crd-group", "", "The API Group on the ServiceCluster to watch.")
	cmd.Flags().StringVar(&flags.providerNamespace, "provider-namespace", "", "Name of the providers namespace in the master cluster.")
	cmd.Flags().StringVar(&flags.crdConfigurationName, "crdconfiguration-name", "", "Name of the CRDConfiguration object holding type configuration.")

	for _, flagName := range []string{
		"crd-version",
		"crd-kind",
		"crd-internal-group",
		"crd-external-group",
		"crd-group",
		"provider-namespace",
		"service-cluster-name",
	} {
		if err := cmd.MarkFlagRequired(flagName); err != nil {
			panic(fmt.Errorf("req flag %s: %w", flagName, err))
		}
	}
	return cmd
}

func runE(f *flags, log logr.Logger) error {
	serviceCfg, err := clientcmd.BuildConfigFromFlags(f.serviceMaster, f.serviceKubeConfig)
	if err != nil {
		return fmt.Errorf("cannot setup service cluster config: %w", err)
	}

	serviceMgr, err := ctrl.NewManager(serviceCfg, ctrl.Options{
		Scheme:             serviceScheme,
		MetricsBindAddress: f.serviceMetricsAddr,
	})
	if err != nil {
		return fmt.Errorf("cannot setup service cluster manager: %w", err)
	}

	masterCfg := ctrl.GetConfigOrDie()
	// Master cluster setup
	masterMgr, err := ctrl.NewManager(masterCfg, ctrl.Options{
		Scheme:                  masterScheme,
		LeaderElection:          f.enableLeaderElection,
		LeaderElectionNamespace: f.providerNamespace,
		LeaderElectionID:        "dinghy-" + f.crdConfigurationName,
		Namespace:               f.providerNamespace,
		MetricsBindAddress:      f.masterMetricsAddr,
	})
	if err != nil {
		return fmt.Errorf("create manager for master cluster: %w", err)
	}

	// underlying serviceMgr manager implementation, controllerManager doesn't implement any InjectX
	// interfaces, thus the masterMgr shall only start it, and kill it accordingly
	// It makes for way simpler code then doing this manually
	if err := masterMgr.Add(serviceMgr); err != nil {
		return fmt.Errorf("cannot add service mgr to master: %w", err)
	}
	if err := masterMgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("running master mgr: %w", err)
	}
	return nil
}

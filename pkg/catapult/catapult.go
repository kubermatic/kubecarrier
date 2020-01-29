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

package catapult

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/catapult/internal/controllers"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

type flags struct {
	leaderElectionNamespace string
	enableLeaderElection    bool

	// master
	masterMetricsAddr string
	masterGroup       string
	masterKind        string

	// service
	serviceMetricsAddr     string
	serviceKubeconfig      string
	serviceTargetNamespace string
	serviceGroup           string
	serviceKind            string
	version                string
}

var (
	masterScheme  = runtime.NewScheme()
	serviceScheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(masterScheme)
	_ = apiextensionsv1.AddToScheme(masterScheme)
	_ = corev1alpha1.AddToScheme(masterScheme)

	_ = clientgoscheme.AddToScheme(serviceScheme)
	_ = apiextensionsv1.AddToScheme(serviceScheme)
}

const (
	componentCatapult = "Catapult"
)

func NewCatapult() *cobra.Command {
	// https: //github.com/kubermatic/kubecarrier/issues/64
	// TODO: Implement bare bone flag interface
	// TODO: Implement one-way cross cluster reconciliation
	// TODO: Implement two-way cross cluster reconciliation
	// TODO: Implement cross cluster webhooks support

	log := ctrl.Log.WithName("catapult")
	flags := &flags{}
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   componentCatapult,
		Short: "KubeCarrier Catapult",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(flags, log)
		},
	}
	cmd.Flags().BoolVar(&flags.enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	cmd.Flags().StringVar(&flags.leaderElectionNamespace, "leader-election-namespace", "", "namespace for leader election")

	// master
	cmd.Flags().StringVar(&flags.masterMetricsAddr, "master-metrics-addr", ":8080", "The address the metric endpoint binds to.")
	cmd.Flags().StringVar(&flags.masterGroup, "master-group", "", "The object's group in the master(current) cluster we're catapulting in the service cluster")
	cmd.Flags().StringVar(&flags.masterKind, "master-kind", "", "The object's kind in the master(current) cluster we're catapulting in the service cluster")

	// service cluster client settings
	cmd.Flags().StringVar(&flags.serviceMetricsAddr, "service-cluster-metrics-addr", ":8081", "The address the metric endpoint binds to.")
	cmd.Flags().StringVar(&flags.serviceKubeconfig, "service-cluster-kubeconfig", "", "Path to the Service Cluster kubeconfig.")
	cmd.Flags().StringVar(&flags.serviceTargetNamespace, "service-namespace", "", "Name of the Service Cluster the ferry is operating on.")
	cmd.Flags().StringVar(&flags.serviceGroup, "service-group", "", "The object's group in the service(current) cluster we're catapulting in the service cluster")
	cmd.Flags().StringVar(&flags.serviceKind, "service-kind", "", "The object's kind in the service(current) cluster we're catapulting in the service cluster")

	cmd.Flags().StringVar(&flags.version, "version", "v1", "the objects version to reconcile")
	for _, flagName := range []string{
		"service-namespace",
		"service-cluster-kubeconfig",
		"master-group",
		"master-kind",
		"service-group",
		"service-kind",
		"version",
	} {
		if err := cmd.MarkFlagRequired(flagName); err != nil {
			panic(fmt.Errorf("req flag %s: %w", flagName, err))
		}

	}
	return util.CmdLogMixin(cmd)
}

func run(flags *flags, log logr.Logger) error {
	// KubeCarrier cluster manager
	masterCfg := ctrl.GetConfigOrDie()
	serviceCfg, err := clientcmd.BuildConfigFromFlags("", flags.serviceKubeconfig)
	if err != nil {
		return fmt.Errorf("unable to set up service cluster client config: %w", err)
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
		LeaderElectionNamespace: flags.leaderElectionNamespace,
		LeaderElectionID:        "catapult-" + flags.masterGroup + "-" + flags.masterKind,
	})
	if err != nil {
		return fmt.Errorf("unable to start manager for master cluster: %w", err)
	}

	// Build Dynamic Type
	internalGVK := schema.GroupVersionKind{
		Group:   flags.masterGroup,
		Kind:    flags.masterKind,
		Version: flags.version,
	}

	serviceClusterGVK := schema.GroupVersionKind{
		Group:   flags.serviceGroup,
		Kind:    flags.serviceKind,
		Version: flags.version,
	}

	if err := (&controllers.InternalObjectReconciler{
		MasterClient:                  masterMgr.GetClient(),
		MasterScheme:                  masterMgr.GetScheme(),
		ServiceClient:                 serviceMgr.GetClient(),
		Log:                           ctrl.Log.WithName("controllers").WithName("InternalObjectReconciler"),
		InternalGVK:                   internalGVK,
		ServiceClusterGVK:             serviceClusterGVK,
		ServiceClusterTargetNamespace: flags.serviceTargetNamespace,
	}).SetupWithManagers(serviceMgr, masterMgr); err != nil {
		return fmt.Errorf("setup InternalObjectReconciler: %w", err)
	}

	if err := masterMgr.Add(serviceMgr); err != nil {
		return fmt.Errorf("cannot add service mgr: %w", err)
	}

	if err := masterMgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("error running component: %w", err)
	}
	return nil
}

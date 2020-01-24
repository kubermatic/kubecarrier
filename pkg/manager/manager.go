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

package manager

import (
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
	"github.com/kubermatic/kubecarrier/pkg/manager/internal/controllers"
)

type flags struct {
	kubeCarrierSystemNamespace string
	metricsAddr                string
	enableLeaderElection       bool
}

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = catalogv1alpha1.AddToScheme(scheme)
	_ = corev1alpha1.AddToScheme(scheme)
	_ = apiextensionsv1.AddToScheme(scheme)
}

const (
	componentManager = "manager"
)

func NewManagerCommand() *cobra.Command {
	log := ctrl.Log.WithName("manager")
	flags := &flags{}
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   componentManager,
		Short: "deploy kubecarrier controller manager",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(flags, log)
		},
	}
	cmd.Flags().StringVar(&flags.metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	cmd.Flags().BoolVar(&flags.enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	cmd.Flags().StringVar(&flags.kubeCarrierSystemNamespace, "kubecarrier-system-namespace", os.Getenv("KUBECARRIER_NAMESPACE"), "The namespace that KubeCarrier controller manager deploys to.")
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

	if flags.kubeCarrierSystemNamespace == "" {
		return fmt.Errorf("-kubecarrier-system-namespace or ENVVAR KUBECARRIER_NAMESPACE must be set")
	}

	if err = (&controllers.TenantReconciler{
		Client:                     mgr.GetClient(),
		Log:                        log.WithName("controllers").WithName("Tenant"),
		Scheme:                     mgr.GetScheme(),
		KubeCarrierSystemNamespace: flags.kubeCarrierSystemNamespace,
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("creating Tenant controller: %w", err)
	}

	if err = (&controllers.ProviderReconciler{
		Client: mgr.GetClient(),
		Log:    log.WithName("controllers").WithName("Provider"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("creating Provider controller: %w", err)
	}

	// Register a field index for Provider.Status.NamespaceName
	if err := catalogv1alpha1.RegisterProviderNamespaceFieldIndex(mgr); err != nil {
		return fmt.Errorf("registering ProviderNamespace field index: %w", err)
	}

	// Register Owner field indexes
	fieldIndexerLog := ctrl.Log.WithName("fieldindex")
	if err := util.AddOwnerReverseFieldIndex(
		mgr.GetFieldIndexer(), fieldIndexerLog.WithName("Offering"), &catalogv1alpha1.Offering{},
	); err != nil {
		return fmt.Errorf("registering Offering owner field index: %w", err)
	}
	if err := util.AddOwnerReverseFieldIndex(mgr.GetFieldIndexer(), fieldIndexerLog.WithName("CRD"), &apiextensionsv1.CustomResourceDefinition{}); err != nil {
		return fmt.Errorf("registering CRD owner field index: %w", err)
	}
	if err := util.AddOwnerReverseFieldIndex(
		mgr.GetFieldIndexer(), fieldIndexerLog.WithName("ProviderReference"), &catalogv1alpha1.ProviderReference{},
	); err != nil {
		return fmt.Errorf("registering ProviderReference owner field indexer: %w", err)
	}
	if err := util.AddOwnerReverseFieldIndex(
		mgr.GetFieldIndexer(), fieldIndexerLog.WithName("ServiceClusterReference"), &catalogv1alpha1.ServiceClusterReference{},
	); err != nil {
		return fmt.Errorf("registering ServiceCluster owner field index: %w", err)
	}

	if err = (&controllers.CatalogEntryReconciler{
		Client:                     mgr.GetClient(),
		Log:                        log.WithName("controllers").WithName("CatalogEntry"),
		Scheme:                     mgr.GetScheme(),
		KubeCarrierSystemNamespace: flags.kubeCarrierSystemNamespace,
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("creating CatalogEntry controller: %w", err)
	}

	if err = (&controllers.CatalogReconciler{
		Client:                     mgr.GetClient(),
		Log:                        log.WithName("controllers").WithName("Catalog"),
		Scheme:                     mgr.GetScheme(),
		KubeCarrierSystemNamespace: flags.kubeCarrierSystemNamespace,
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("creating Catalog controller: %w", err)
	}

	if err = (&controllers.DerivedCustomResourceDefinitionReconciler{
		Client:                     mgr.GetClient(),
		Log:                        log.WithName("controllers").WithName("DerivedCustomResourceDefinition"),
		Scheme:                     mgr.GetScheme(),
		KubeCarrierSystemNamespace: flags.kubeCarrierSystemNamespace,
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("creating DerivedCustomResourceDefinition controller: %w", err)
	}

	// Register webhooks
	if err = (&catalogv1alpha1.Tenant{}).SetupWebhookWithManager(mgr); err != nil {
		return fmt.Errorf("registering webhooks for Tenant: %w", err)
	}

	// Register webhooks as handlers
	wbh := mgr.GetWebhookServer()
	wbh.Register("/validate-catalog-kubecarrier-io-v1alpha1-catalogentry", &webhook.Admission{Handler: &catalogv1alpha1.CatalogEntryValidator{}})
	wbh.Register("/mutating-catalog-kubecarrier-io-v1alpha1-catalogentry", &webhook.Admission{Handler: &catalogv1alpha1.CatalogEntryValidator{}})
	log.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("running manager: %w", err)
	}
	return nil
}

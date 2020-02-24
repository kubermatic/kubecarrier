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
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/multiowner"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
	utilwebhook "github.com/kubermatic/kubecarrier/pkg/internal/util/webhook"
	"github.com/kubermatic/kubecarrier/pkg/manager/internal/controllers"
	"github.com/kubermatic/kubecarrier/pkg/manager/internal/webhooks"
)

type flags struct {
	kubeCarrierSystemNamespace string
	metricsAddr                string
	healthAddr                 string
	enableLeaderElection       bool
	certDir                    string
}

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = catalogv1alpha1.AddToScheme(scheme)
	_ = corev1alpha1.AddToScheme(scheme)
	_ = apiextensionsv1.AddToScheme(scheme)
	_ = operatorv1alpha1.AddToScheme(scheme)
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
	cmd.Flags().StringVar(&flags.healthAddr, "health-addr", ":9440", "The address the health endpoint binds to.")
	cmd.Flags().BoolVar(&flags.enableLeaderElection, "enable-leader-election", true,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	cmd.Flags().StringVar(&flags.kubeCarrierSystemNamespace, "kubecarrier-system-namespace", os.Getenv("KUBECARRIER_NAMESPACE"), "The namespace that KubeCarrier controller manager deploys to.")
	cmd.Flags().StringVar(&flags.certDir, "cert-dir", "/tmp/k8s-webhook-server/serving-certs", "The webhook TLS certificates directory")
	return util.CmdLogMixin(cmd)
}

func run(flags *flags, log logr.Logger) error {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                  scheme,
		MetricsBindAddress:      flags.metricsAddr,
		LeaderElection:          flags.enableLeaderElection,
		LeaderElectionID:        "main-controller-manager",
		LeaderElectionNamespace: flags.kubeCarrierSystemNamespace,
		Port:                    9443,
		CertDir:                 flags.certDir,
		HealthProbeBindAddress:  flags.healthAddr,
	})
	if err != nil {
		return fmt.Errorf("starting manager: %w", err)
	}

	if flags.kubeCarrierSystemNamespace == "" {
		return fmt.Errorf("-kubecarrier-system-namespace or ENVVAR KUBECARRIER_NAMESPACE must be set")
	}

	// Register Owner field indexes
	fieldIndexerLog := ctrl.Log.WithName("fieldindex")
	if err := multiowner.AddOwnerReverseFieldIndex(
		mgr.GetFieldIndexer(), fieldIndexerLog.WithName("Offering"), &catalogv1alpha1.Offering{},
	); err != nil {
		return fmt.Errorf("registering Offering owner field index: %w", err)
	}
	if err := multiowner.AddOwnerReverseFieldIndex(
		mgr.GetFieldIndexer(), fieldIndexerLog.WithName("ProviderReference"), &catalogv1alpha1.ProviderReference{},
	); err != nil {
		return fmt.Errorf("registering ProviderReference owner field indexer: %w", err)
	}
	if err := multiowner.AddOwnerReverseFieldIndex(
		mgr.GetFieldIndexer(), fieldIndexerLog.WithName("ServiceClusterReference"), &catalogv1alpha1.ServiceClusterReference{},
	); err != nil {
		return fmt.Errorf("registering ServiceClusterReference owner field index: %w", err)
	}
	if err := multiowner.AddOwnerReverseFieldIndex(
		mgr.GetFieldIndexer(), fieldIndexerLog.WithName("ServiceClusterAssignment"), &corev1alpha1.ServiceClusterAssignment{},
	); err != nil {
		return fmt.Errorf("registering ServiceClusterAssignment owner field index: %w", err)
	}

	if err = (&controllers.AccountReconciler{
		Client: mgr.GetClient(),
		Log:    log.WithName("controllers").WithName("Account"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("creating Account controller: %w", err)
	}
	if err = (&controllers.CustomResourceDiscoveryReconciler{
		Client: mgr.GetClient(),
		Log:    log.WithName("controllers").WithName("CustomResourceDiscovery"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("creating CustomResourceDiscovery controller: %w", err)
	}

	if err = (&controllers.CustomResourceDiscoverySetReconciler{
		Client: mgr.GetClient(),
		Log:    log.WithName("controllers").WithName("CustomResourceDiscoverySet"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("creating CustomResourceDiscoverySet controller: %w", err)
	}

	if err = (&controllers.CatalogEntryReconciler{
		Client: mgr.GetClient(),
		Log:    log.WithName("controllers").WithName("CatalogEntry"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("creating CatalogEntry controller: %w", err)
	}

	if err = (&controllers.CatalogReconciler{
		Client: mgr.GetClient(),
		Log:    log.WithName("controllers").WithName("Catalog"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("creating Catalog controller: %w", err)
	}

	if err = (&controllers.DerivedCustomResourceReconciler{
		Client: mgr.GetClient(),
		Log:    log.WithName("controllers").WithName("DerivedCustomResource"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("creating DerivedCustomResource controller: %w", err)
	}

	if err = (&controllers.ServiceClusterReconciler{
		Client: mgr.GetClient(),
		Log:    log.WithName("controllers").WithName("ServiceCluster"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("creating ServiceCluster controller: %w", err)
	}

	// Register webhooks as handlers
	wbh := mgr.GetWebhookServer()

	// validating webhooks
	wbh.Register(utilwebhook.GenerateValidateWebhookPath(&catalogv1alpha1.CatalogEntry{}, mgr.GetScheme()),
		&webhook.Admission{Handler: &webhooks.CatalogEntryWebhookHandler{
			Log: log.WithName("validating webhooks").WithName("CatalogEntry"),
		}})
	wbh.Register(utilwebhook.GenerateValidateWebhookPath(&catalogv1alpha1.DerivedCustomResource{}, mgr.GetScheme()),
		&webhook.Admission{Handler: &webhooks.DerivedCustomResourceWebhookHandler{
			Log: log.WithName("validating webhooks").WithName("DerivedCustomResource"),
		}})
	wbh.Register(utilwebhook.GenerateValidateWebhookPath(&catalogv1alpha1.Offering{}, mgr.GetScheme()),
		&webhook.Admission{Handler: &webhooks.OfferingWebhookHandler{
			Log: log.WithName("validating webhooks").WithName("Offering"),
		}})
	wbh.Register(utilwebhook.GenerateValidateWebhookPath(&catalogv1alpha1.Account{}, mgr.GetScheme()),
		&webhook.Admission{Handler: &webhooks.AccountWebhookHandler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
			Log:    log.WithName("validating webhooks").WithName("Account"),
		}})
	wbh.Register(utilwebhook.GenerateValidateWebhookPath(&catalogv1alpha1.ProviderReference{}, mgr.GetScheme()),
		&webhook.Admission{Handler: &webhooks.ProviderReferenceWebhookHandler{
			Log: log.WithName("validating webhooks").WithName("ProviderReference"),
		}})
	wbh.Register(utilwebhook.GenerateValidateWebhookPath(&catalogv1alpha1.ServiceClusterReference{}, mgr.GetScheme()),
		&webhook.Admission{Handler: &webhooks.ServiceClusterReferenceWebhookHandler{
			Log: log.WithName("validating webhooks").WithName("ServiceClusterReference"),
		}})
	wbh.Register(utilwebhook.GenerateValidateWebhookPath(&catalogv1alpha1.TenantReference{}, mgr.GetScheme()),
		&webhook.Admission{Handler: &webhooks.TenantReferenceWebhookHandler{
			Log: log.WithName("validating webhooks").WithName("TenantReference"),
		}})
	wbh.Register(utilwebhook.GenerateValidateWebhookPath(&corev1alpha1.CustomResourceDiscovery{}, mgr.GetScheme()),
		&webhook.Admission{Handler: &webhooks.CustomResourceDiscoveryWebhookHandler{
			Log: log.WithName("validating webhooks").WithName("CustomResourceDiscovery"),
		}})
	wbh.Register(utilwebhook.GenerateValidateWebhookPath(&corev1alpha1.ServiceCluster{}, mgr.GetScheme()),
		&webhook.Admission{Handler: &webhooks.ServiceClusterWebhookHandler{
			Log: log.WithName("validating webhooks").WithName("ServiceCluster"),
		}})
	wbh.Register(utilwebhook.GenerateValidateWebhookPath(&corev1alpha1.ServiceClusterAssignment{}, mgr.GetScheme()),
		&webhook.Admission{Handler: &webhooks.ServiceClusterAssignmentWebhookHandler{
			Log: log.WithName("validating webhooks").WithName("ServiceClusterAssignment"),
		}})

	if err := mgr.AddReadyzCheck("ping", healthz.Ping); err != nil {
		return fmt.Errorf("adding readyz checker: %w", err)
	}

	log.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("running manager: %w", err)
	}
	return nil
}

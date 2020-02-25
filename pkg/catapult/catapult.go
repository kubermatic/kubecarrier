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
	"os"
	"strings"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/catapult/internal/controllers"
	"github.com/kubermatic/kubecarrier/pkg/catapult/internal/webhooks"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

type flags struct {
	metricsAddr          string
	healthAddr           string
	enableLeaderElection bool
	certDir              string

	managementClusterKind, managementClusterVersion, managementClusterGroup string
	serviceClusterKind, serviceClusterVersion, serviceClusterGroup          string
	serviceClusterName, serviceClusterKubeconfig                            string
	providerNamespace                                                       string

	mutatingWebhookPath string
	webhookStrategy     string
}

var (
	managementScheme = runtime.NewScheme()
	serviceScheme    = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(managementScheme)
	_ = corev1alpha1.AddToScheme(managementScheme)
	_ = clientgoscheme.AddToScheme(serviceScheme)
}

const (
	componentCatapult = "Catapult"
)

func NewCatapult() *cobra.Command {
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
	cmd.Flags().StringVar(&flags.metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	cmd.Flags().StringVar(&flags.healthAddr, "health-addr", ":9440", "The address the health endpoint binds to.")
	cmd.Flags().BoolVar(&flags.enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	cmd.Flags().StringVar(&flags.certDir, "cert-dir", "/tmp/k8s-webhook-server/serving-certs", "The webhook TLS certificates directory")

	cmd.Flags().StringVar(
		&flags.managementClusterKind, "management-cluster-kind",
		os.Getenv("CATAPULT_MANAGEMENT_CLUSTER_KIND"), "Kind of management cluster CRD.")
	cmd.Flags().StringVar(
		&flags.managementClusterVersion, "management-cluster-version",
		os.Getenv("CATAPULT_MANAGEMENT_CLUSTER_VERSION"), "Version of management cluster CRD.")
	cmd.Flags().StringVar(
		&flags.managementClusterGroup, "management-cluster-group",
		os.Getenv("CATAPULT_MANAGEMENT_CLUSTER_GROUP"), "Group of management cluster CRD.")

	cmd.Flags().StringVar(
		&flags.serviceClusterKind, "service-cluster-kind",
		os.Getenv("CATAPULT_SERVICE_CLUSTER_KIND"), "Kind of service cluster CRD.")
	cmd.Flags().StringVar(
		&flags.serviceClusterVersion, "service-cluster-version",
		os.Getenv("CATAPULT_SERVICE_CLUSTER_VERSION"), "Version of service cluster CRD.")
	cmd.Flags().StringVar(
		&flags.serviceClusterGroup, "service-cluster-group",
		os.Getenv("CATAPULT_SERVICE_CLUSTER_GROUP"), "Group of service cluster CRD.")

	cmd.Flags().StringVar(
		&flags.serviceClusterKubeconfig, "service-cluster-kubeconfig",
		os.Getenv("CATAPULT_SERVICE_CLUSTER_KUBECONFIG"), "Path to service cluster kubeconfig.")
	cmd.Flags().StringVar(
		&flags.serviceClusterName, "service-cluster-name",
		os.Getenv("CATAPULT_SERVICE_CLUSTER_NAME"), "Name of the ServiceCluster.")

	cmd.Flags().StringVar(
		&flags.providerNamespace, "provider-namespace",
		os.Getenv("KUBERNETES_NAMESPACE"), "Name of the provider namespace in the management cluster.")

	cmd.Flags().StringVar(
		&flags.mutatingWebhookPath, "mutating-webhook-path",
		os.Getenv("CATAPULT_MUTATING_WEBHOOK_PATH"), "The URL path of the mutating webhook service.")
	cmd.Flags().StringVar(
		&flags.webhookStrategy, "webhook-strategy",
		os.Getenv("CATAPULT_WEBHOOK_STRATEGY"), "The strategy of deploying the catapult webhook service {None (by default), ServiceCluster}")

	return util.CmdLogMixin(cmd)
}

func run(flags *flags, log logr.Logger) error {
	// validate settings
	checks := []struct {
		value, env, flag string
	}{
		{value: flags.managementClusterKind, env: "CATAPULT_MANAGEMENT_CLUSTER_KIND", flag: "management-cluster-kind"},
		{value: flags.managementClusterVersion, env: "CATAPULT_MANAGEMENT_CLUSTER_VERSION", flag: "management-cluster-version"},
		{value: flags.managementClusterGroup, env: "CATAPULT_MANAGEMENT_CLUSTER_GROUP", flag: "management-cluster-group"},

		{value: flags.serviceClusterKind, env: "CATAPULT_SERVICE_CLUSTER_KIND", flag: "service-cluster-kind"},
		{value: flags.serviceClusterVersion, env: "CATAPULT_SERVICE_CLUSTER_VERSION", flag: "service-cluster-version"},
		{value: flags.serviceClusterGroup, env: "CATAPULT_SERVICE_CLUSTER_GROUP", flag: "service-cluster-group"},

		{value: flags.serviceClusterKubeconfig, env: "CATAPULT_SERVICE_CLUSTER_KUBECONFIG", flag: "service-cluster-kubeconfig"},
		{value: flags.serviceClusterKubeconfig, env: "CATAPULT_SERVICE_CLUSTER_NAME", flag: "service-cluster-name"},

		{value: flags.providerNamespace, env: "KUBERNETES_NAMESPACE", flag: "provider-namespace"},

		{value: flags.mutatingWebhookPath, env: "CATAPULT_MUTATING_WEBHOOK_PATH", flag: "mutating-webhook-path"},
		{value: flags.webhookStrategy, env: "CATAPULT_WEBHOOK_STRATEGY", flag: "webhook-strategy"},
	}
	var errs []string
	for _, check := range checks {
		if check.value == "" {
			errs = append(errs, fmt.Sprintf("flag --%s or envvar %s needs to be non-empty", check.flag, check.env))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf(strings.Join(errs, ", "))
	}

	// Setup Manager
	managementCfg := ctrl.GetConfigOrDie()
	mgr, err := ctrl.NewManager(managementCfg, ctrl.Options{
		Scheme:                 managementScheme,
		MetricsBindAddress:     flags.metricsAddr,
		HealthProbeBindAddress: flags.healthAddr,
		LeaderElection:         flags.enableLeaderElection,
		Port:                   9443,
		CertDir:                flags.certDir,
		NewClient: func(cache cache.Cache, config *rest.Config, options client.Options) (client.Client, error) {
			// Create the Client for Write operations.
			c, err := client.New(config, options)
			if err != nil {
				return nil, err
			}

			// we don't want a client.DelegatingReader here,
			// because we WANT to cache unstructured objects.
			return &client.DelegatingClient{
				Reader:       cache,
				Writer:       c,
				StatusClient: c,
			}, nil
		},
	})
	if err != nil {
		return fmt.Errorf("starting manager: %w", err)
	}

	// Setup additional namespaced client for management cluster
	namespacedCache, err := cache.New(managementCfg, cache.Options{
		Scheme:    mgr.GetScheme(),
		Mapper:    mgr.GetRESTMapper(),
		Namespace: flags.providerNamespace,
	})
	if err != nil {
		return fmt.Errorf(
			"creating namespaced scoped cache: %w", err)
	}
	if err = mgr.Add(namespacedCache); err != nil {
		return fmt.Errorf("add namespaced cache to manager: %w", err)
	}
	namespacedClient := &client.DelegatingClient{
		Reader:       namespacedCache,
		Writer:       mgr.GetClient(),
		StatusClient: mgr.GetClient(),
	}

	// Setup additional client and cache for Service Cluster
	serviceCfg, err := clientcmd.BuildConfigFromFlags(
		"", flags.serviceClusterKubeconfig)
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

	// Setup Types
	managementClusterGVK := schema.GroupVersionKind{
		Kind:    flags.managementClusterKind,
		Version: flags.managementClusterVersion,
		Group:   flags.managementClusterGroup,
	}
	serviceClusterGVK := schema.GroupVersionKind{
		Kind:    flags.serviceClusterKind,
		Version: flags.serviceClusterVersion,
		Group:   flags.serviceClusterGroup,
	}

	// Setup field indexes
	if err := corev1alpha1.RegisterServiceClusterAssignmentNamespaceFieldIndex(namespacedCache); err != nil {
		return fmt.Errorf("registering ServiceClusterAssignment ServiceClusterNamespace field index: %w", err)
	}

	// Setup Controllers
	if err := (&controllers.ManagementClusterObjReconciler{
		Log:              log.WithName("controllers").WithName("ManagementClusterObjReconciler"),
		Client:           mgr.GetClient(),
		NamespacedClient: namespacedClient,

		ServiceClusterClient: serviceCachedClient,
		ServiceClusterCache:  serviceCache,
		ServiceCluster:       flags.serviceClusterName,
		ProviderNamespace:    flags.providerNamespace,

		ManagementClusterGVK: managementClusterGVK,
		ServiceClusterGVK:    serviceClusterGVK,
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("cannot add %s controller: %w", "ManagementClusterObjReconciler", err)
	}

	if err := (&controllers.AdoptionReconciler{
		Log:              log.WithName("controllers").WithName("AdoptionReconciler"),
		Client:           mgr.GetClient(),
		NamespacedClient: namespacedClient,

		ServiceClusterClient: serviceCachedClient,
		ServiceClusterCache:  serviceCache,
		ProviderNamespace:    flags.providerNamespace,

		ManagementClusterGVK: managementClusterGVK,
		ServiceClusterGVK:    serviceClusterGVK,
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("cannot add %s controller: %w", "AdoptionReconciler", err)
	}

	if err := mgr.AddReadyzCheck("ping", healthz.Ping); err != nil {
		return fmt.Errorf("adding readyz checker: %w", err)
	}

	// Register webhooks as handlers
	wbh := mgr.GetWebhookServer()

	// mutating webhook
	wbh.Register(flags.mutatingWebhookPath,
		&webhook.Admission{Handler: &webhooks.ManagementClusterObjWebhookHandler{
			Log:    log.WithName("mutating webhooks").WithName(managementClusterGVK.Kind),
			Scheme: mgr.GetScheme(),

			ManagementClusterClient: namespacedClient,
			ServiceClusterClient:    serviceCachedClient,

			ManagementClusterGVK: managementClusterGVK,
			ServiceClusterGVK:    serviceClusterGVK,

			ProviderNamespace: flags.providerNamespace,
			ServiceCluster:    flags.serviceClusterName,

			WebhookStrategy: corev1alpha1.WebhookStrategyType(flags.webhookStrategy),
		}})

	log.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("running manager: %w", err)
	}
	return nil
}

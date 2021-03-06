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

package elevator

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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"k8c.io/utils/pkg/util"

	catalogv1alpha1 "k8c.io/kubecarrier/pkg/apis/catalog/v1alpha1"
	"k8c.io/kubecarrier/pkg/elevator/internal/controllers"
	"k8c.io/kubecarrier/pkg/elevator/internal/webhooks"
)

type flags struct {
	metricsAddr          string
	healthAddr           string
	enableLeaderElection bool
	certDir              string

	providerKind, providerVersion, providerGroup string
	tenantKind, tenantVersion, tenantGroup       string
	derivedCRName                                string
	providerNamespace                            string

	mutatingWebhookPath string
}

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = catalogv1alpha1.AddToScheme(scheme)
}

const (
	componentElevator = "Elevator"
)

func NewElevator() *cobra.Command {
	log := ctrl.Log.WithName("elevator")
	flags := &flags{}
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   componentElevator,
		Short: "KubeCarrier Elevator",
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
		&flags.providerKind, "provider-kind",
		os.Getenv("ELEVATOR_PROVIDER_KIND"), "Kind of Provider-side CRD.")
	cmd.Flags().StringVar(
		&flags.providerVersion, "provider-version",
		os.Getenv("ELEVATOR_PROVIDER_VERSION"), "Version of Provider-side CRD.")
	cmd.Flags().StringVar(
		&flags.providerGroup, "provider-group",
		os.Getenv("ELEVATOR_PROVIDER_GROUP"), "Group of Provider-side CRD.")

	cmd.Flags().StringVar(
		&flags.tenantKind, "tenant-kind",
		os.Getenv("ELEVATOR_TENANT_KIND"), "Kind of Tenant-side CRD.")
	cmd.Flags().StringVar(
		&flags.tenantVersion, "tenant-version",
		os.Getenv("ELEVATOR_TENANT_VERSION"), "Version of Tenant-side CRD.")
	cmd.Flags().StringVar(
		&flags.tenantGroup, "tenant-group",
		os.Getenv("ELEVATOR_TENANT_GROUP"), "Group of Tenant-side CRD.")

	cmd.Flags().StringVar(
		&flags.derivedCRName, "derived-crd-name",
		os.Getenv("ELEVATOR_DERIVED_CRD_NAME"), "Name of DerivedCR controlling the Tenant-side CRD.")

	cmd.Flags().StringVar(
		&flags.providerNamespace, "provider-namespace",
		os.Getenv("KUBERNETES_NAMESPACE"), "Name of the provider namespace in the management cluster.")

	cmd.Flags().StringVar(
		&flags.mutatingWebhookPath, "mutating-webhook-path",
		os.Getenv("ELEVATOR_MUTATING_WEBHOOK_PATH"), "The URL path of the mutating webhook service.")

	return util.CmdLogMixin(cmd)
}

func run(flags *flags, log logr.Logger) error {
	// validate settings
	checks := []struct {
		value, env, flag string
	}{
		{value: flags.providerKind, env: "ELEVATOR_PROVIDER_KIND", flag: "provider-kind"},
		{value: flags.providerVersion, env: "ELEVATOR_PROVIDER_VERSION", flag: "provider-version"},
		{value: flags.providerGroup, env: "ELEVATOR_PROVIDER_GROUP", flag: "provider-group"},
		{value: flags.tenantKind, env: "ELEVATOR_TENANT_KIND", flag: "tenant-kind"},
		{value: flags.tenantVersion, env: "ELEVATOR_TENANT_VERSION", flag: "tenant-version"},
		{value: flags.tenantGroup, env: "ELEVATOR_TENANT_GROUP", flag: "tenant-group"},
		{value: flags.derivedCRName, env: "ELEVATOR_DERIVED_CRD_NAME", flag: "derived-crd-name"},
		{value: flags.providerNamespace, env: "KUBERNETES_NAMESPACE", flag: "provider-namespace"},
		{value: flags.mutatingWebhookPath, env: "ELEVATOR_MUTATING_WEBHOOK_PATH", flag: "mutating-webhook-path"},
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

	cfg := ctrl.GetConfigOrDie()
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:                  scheme,
		MetricsBindAddress:      flags.metricsAddr,
		HealthProbeBindAddress:  flags.healthAddr,
		LeaderElection:          flags.enableLeaderElection,
		LeaderElectionNamespace: flags.providerNamespace,
		LeaderElectionID:        "elevator-" + flags.derivedCRName,
		CertDir:                 flags.certDir,
		Port:                    9443,
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

	// We only have permissions to access DerivedCustomResources in the provider namespace.
	// So we have to create a new cache, that is limited to this namespace, or we will break on permission errors.
	namespacedCache, err := cache.New(cfg, cache.Options{
		Scheme:    mgr.GetScheme(),
		Mapper:    mgr.GetRESTMapper(),
		Namespace: flags.providerNamespace,
	})
	if err != nil {
		return fmt.Errorf(
			"creating namespaced scoped cache: %w", err)
	}
	if err = mgr.Add(namespacedCache); err != nil {
		return fmt.Errorf(
			"add namespaced cache to manager: %w", err)
	}
	namespacedClient := &client.DelegatingClient{
		Reader:       namespacedCache,
		Writer:       mgr.GetClient(),
		StatusClient: mgr.GetClient(),
	}

	// Setup Types
	providerGVK := schema.GroupVersionKind{
		Kind:    flags.providerKind,
		Version: flags.providerVersion,
		Group:   flags.providerGroup,
	}
	tenantGVK := schema.GroupVersionKind{
		Kind:    flags.tenantKind,
		Version: flags.tenantVersion,
		Group:   flags.tenantGroup,
	}

	// Setup Controllers
	if err := (&controllers.TenantObjReconciler{
		Log:              log.WithName("controllers").WithName("TenantObjReconciler"),
		Client:           mgr.GetClient(),
		Scheme:           mgr.GetScheme(),
		NamespacedClient: namespacedClient,

		ProviderGVK: providerGVK,
		TenantGVK:   tenantGVK,

		DerivedCRName:     flags.derivedCRName,
		ProviderNamespace: flags.providerNamespace,
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("cannot add %s controller: %w", "TenantObjReconciler", err)
	}

	if err := (&controllers.AdoptionReconciler{
		Log:              log.WithName("controllers").WithName("AdoptionReconciler"),
		Client:           mgr.GetClient(),
		Scheme:           mgr.GetScheme(),
		NamespacedClient: namespacedClient,

		ProviderGVK: providerGVK,
		TenantGVK:   tenantGVK,

		DerivedCRName:     flags.derivedCRName,
		ProviderNamespace: flags.providerNamespace,
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("cannot add %s controller: %w", "AdoptionReconciler", err)
	}

	if err := mgr.AddReadyzCheck("ping", healthz.Ping); err != nil {
		return fmt.Errorf("adding readyz checker: %w", err)
	}

	if err := mgr.AddHealthzCheck("ping", healthz.Ping); err != nil {
		return fmt.Errorf("adding healthz checker: %w", err)
	}

	// Register webhooks as handlers
	wbh := mgr.GetWebhookServer()

	// mutating webhook
	wbh.Register(flags.mutatingWebhookPath,
		&webhook.Admission{Handler: &webhooks.TenantObjWebhookHandler{
			Log:    log.WithName("mutating webhooks").WithName(tenantGVK.Kind),
			Scheme: mgr.GetScheme(),

			Client:           mgr.GetClient(),
			NamespacedClient: namespacedClient,

			TenantGVK:   tenantGVK,
			ProviderGVK: providerGVK,

			ProviderNamespace: flags.providerNamespace,
			DerivedCRName:     flags.derivedCRName,
		}})

	log.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("running manager: %w", err)
	}
	return nil
}

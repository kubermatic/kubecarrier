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
	"os"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/kustomize"
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
	_ = clientgoscheme.AddToScheme(scheme)
	_ = operatorv1alpha1.AddToScheme(scheme)
	_ = rbacv1.AddToScheme(scheme)
	_ = apiextensionsv1beta1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

const (
	componentOperator = "operator"
)

func NewOperatorCommand(log logr.Logger) *cobra.Command {
	flags := &flags{}
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   componentOperator,
		Short: "deploy kubecarrier operator",
		Run: func(cmd *cobra.Command, args []string) {
			run(flags, log)
		},
	}
	cmd.Flags().StringVar(&flags.metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	cmd.Flags().BoolVar(&flags.enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for operator. Enabling this will ensure there is only one active controller manager.")
	return cmd
}

func run(flags *flags, log logr.Logger) {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: flags.metricsAddr,
		LeaderElection:     flags.enableLeaderElection,
		Port:               9443,
	})
	if err != nil {
		log.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Field Index
	if err := util.AddOwnerReverseFieldIndex(
		mgr.GetFieldIndexer(), ctrl.Log.WithName("fieldindex").WithName("ClusterRole"), &rbacv1.ClusterRole{},
	); err != nil {
		log.Error(fmt.Errorf("cannot add ClusterRole owner field indexer: %w", err), "unable to start manager")
		os.Exit(1)
	}
	if err := util.AddOwnerReverseFieldIndex(
		mgr.GetFieldIndexer(), ctrl.Log.WithName("fieldindex").WithName("ClusterRoleBinding"), &rbacv1.ClusterRoleBinding{},
	); err != nil {
		log.Error(fmt.Errorf("cannot add ClusterRoleBinding owner field indexer: %w", err), "unable to start manager")
		os.Exit(1)
	}

	kustomize := kustomize.NewDefaultKustomize()
	if err = (&controllers.KubeCarrierReconciler{
		Client:    mgr.GetClient(),
		Log:       log.WithName("controllers").WithName("KubeCarrier"),
		Scheme:    mgr.GetScheme(),
		Kustomize: kustomize,
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create controller", "controller", "KubeCarrier")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	log.Info("starting operator")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "problem running operator")
		os.Exit(1)
	}
}

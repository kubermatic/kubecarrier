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
	"time"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
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
			_ = flags
			return nil
		},
	}
	return cmd
}

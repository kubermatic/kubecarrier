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

package framework

import (
	"fmt"
	"os"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
)

type Config struct {
	TestID string

	MasterExternalKubeconfigPath  string
	MasterInternalKubeconfigPath  string
	ServiceExternalKubeconfigPath string
	ServiceInternalKubeconfigPath string
}

func (c *Config) MasterClusterName() string {
	return "kubecarrier-" + c.TestID
}

func (c *Config) ServiceClusterName() string {
	return "kubecarrier-svc-" + c.TestID
}

func (c *Config) Default() {
	// Master Cluster
	if c.MasterInternalKubeconfigPath == "" {
		c.MasterInternalKubeconfigPath = os.ExpandEnv("${HOME}/.kube/internal-kind-config-" + c.MasterClusterName())
	}
	if c.MasterExternalKubeconfigPath == "" {
		c.MasterExternalKubeconfigPath = os.ExpandEnv("${HOME}/.kube/kind-config-" + c.MasterClusterName())
	}

	// Service Cluster
	if c.ServiceInternalKubeconfigPath == "" {
		c.ServiceInternalKubeconfigPath = os.ExpandEnv("${HOME}/.kube/internal-kind-config-" + c.ServiceClusterName())
	}
	if c.ServiceExternalKubeconfigPath == "" {
		c.ServiceExternalKubeconfigPath = os.ExpandEnv("${HOME}/.kube/kind-config-" + c.ServiceClusterName())
	}
}

type Framework struct {
	masterScheme  *runtime.Scheme
	masterConfig  *restclient.Config
	serviceScheme *runtime.Scheme
	serviceConfig *restclient.Config
	config        Config
}

func New(c Config) (f *Framework, err error) {
	f = &Framework{config: c}

	// Master Setup
	f.masterScheme = runtime.NewScheme()
	if err = clientgoscheme.AddToScheme(f.masterScheme); err != nil {
		return nil, fmt.Errorf("adding clientgo scheme to master scheme: %w", err)
	}
	if err = apiextensionsv1.AddToScheme(f.masterScheme); err != nil {
		return nil, fmt.Errorf("adding apiextensionsv1 scheme to master scheme: %w", err)
	}
	if err = operatorv1alpha1.AddToScheme(f.masterScheme); err != nil {
		return nil, fmt.Errorf("adding operatorv1alpha1 scheme to master scheme: %w", err)
	}
	if err = catalogv1alpha1.AddToScheme(f.masterScheme); err != nil {
		return nil, fmt.Errorf("adding catalogv1alpha1 scheme to master scheme: %w", err)
	}
	if err = corev1alpha1.AddToScheme(f.masterScheme); err != nil {
		return nil, fmt.Errorf("adding corev1alpha1 scheme to master scheme: %w", err)
	}

	f.masterConfig, err = clientcmd.BuildConfigFromFlags("", f.config.MasterExternalKubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("build restconfig for master: %w", err)
	}

	// Service Setup
	f.serviceScheme = runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(f.serviceScheme); err != nil {
		return nil, fmt.Errorf("adding clientgo scheme to service scheme: %w", err)
	}

	f.serviceConfig, err = clientcmd.BuildConfigFromFlags("", f.config.ServiceExternalKubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("build restconfig for service: %w", err)
	}

	return
}

func (f *Framework) MasterClient() (client.Client, error) {
	cfg := f.masterConfig
	return client.New(cfg, client.Options{
		Scheme: f.masterScheme,
	})
}

func (f *Framework) ServiceClient() (client.Client, error) {
	cfg := f.serviceConfig
	return client.New(cfg, client.Options{
		Scheme: f.serviceScheme,
	})
}

func (f *Framework) Config() Config {
	return f.config
}

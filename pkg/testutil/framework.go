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

package testutil

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	certmanagerv1alpha2 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"k8c.io/utils/pkg/testutil"

	catalogv1alpha1 "k8c.io/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "k8c.io/kubecarrier/pkg/apis/core/v1alpha1"
	fakev1 "k8c.io/kubecarrier/pkg/apis/fake/v1"
	fakev1alpha1 "k8c.io/kubecarrier/pkg/apis/fake/v1alpha1"
	operatorv1alpha1 "k8c.io/kubecarrier/pkg/apis/operator/v1alpha1"
)

type FrameworkConfig struct {
	TestID string

	ManagementExternalKubeconfigPath string
	ManagementInternalKubeconfigPath string
	ServiceExternalKubeconfigPath    string
	ServiceInternalKubeconfigPath    string
	CleanUpStrategy                  testutil.CleanUpStrategy
}

func (c *FrameworkConfig) ManagementClusterName() string {
	return "kubecarrier-" + c.TestID
}

func (c *FrameworkConfig) ServiceClusterName() string {
	return "kubecarrier-svc-" + c.TestID
}

func (c *FrameworkConfig) Default() {
	// Management Cluster
	if c.ManagementInternalKubeconfigPath == "" {
		c.ManagementInternalKubeconfigPath = os.ExpandEnv("${HOME}/.kube/internal-kind-config-" + c.ManagementClusterName())
	}
	if c.ManagementExternalKubeconfigPath == "" {
		c.ManagementExternalKubeconfigPath = os.ExpandEnv("${HOME}/.kube/kind-config-" + c.ManagementClusterName())
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
	ManagementScheme *runtime.Scheme
	managementConfig *restclient.Config
	ServiceScheme    *runtime.Scheme
	serviceConfig    *restclient.Config
	config           FrameworkConfig
}

func New(c FrameworkConfig) (f *Framework, err error) {
	if c.CleanUpStrategy != testutil.CleanupAlways && c.CleanUpStrategy != testutil.CleanupOnSuccess && c.CleanUpStrategy != testutil.CleanupNever {
		return nil, fmt.Errorf("unknown clean up strategy: %v", c.CleanUpStrategy)
	}

	f = &Framework{config: c}

	// Management Setup
	f.ManagementScheme = runtime.NewScheme()
	if err = clientgoscheme.AddToScheme(f.ManagementScheme); err != nil {
		return nil, fmt.Errorf("adding clientgo scheme to management scheme: %w", err)
	}
	if err = apiextensionsv1.AddToScheme(f.ManagementScheme); err != nil {
		return nil, fmt.Errorf("adding apiextensionsv1 scheme to management scheme: %w", err)
	}
	if err = operatorv1alpha1.AddToScheme(f.ManagementScheme); err != nil {
		return nil, fmt.Errorf("adding operatorv1alpha1 scheme to management scheme: %w", err)
	}
	if err = catalogv1alpha1.AddToScheme(f.ManagementScheme); err != nil {
		return nil, fmt.Errorf("adding catalogv1alpha1 scheme to management scheme: %w", err)
	}
	if err = corev1alpha1.AddToScheme(f.ManagementScheme); err != nil {
		return nil, fmt.Errorf("adding corev1alpha1 scheme to management scheme: %w", err)
	}
	if err = certmanagerv1alpha2.AddToScheme(f.ManagementScheme); err != nil {
		return nil, fmt.Errorf("adding certmanagerv1alpha3 scheme to management scheme: %w", err)
	}

	f.managementConfig, err = clientcmd.BuildConfigFromFlags("", f.config.ManagementExternalKubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("build restconfig for management: %w", err)
	}

	// Service Setup
	f.ServiceScheme = runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(f.ServiceScheme); err != nil {
		return nil, fmt.Errorf("adding clientgo scheme to service scheme: %w", err)
	}
	if err = apiextensionsv1.AddToScheme(f.ServiceScheme); err != nil {
		return nil, fmt.Errorf("adding apiextensionsv1 scheme to service scheme: %w", err)
	}
	if err = fakev1alpha1.AddToScheme(f.ServiceScheme); err != nil {
		return nil, fmt.Errorf("adding fakev1alpha1 scheme to service scheme: %w", err)
	}
	if err = fakev1.AddToScheme(f.ServiceScheme); err != nil {
		return nil, fmt.Errorf("adding fakev1 scheme to service scheme: %w", err)
	}
	f.serviceConfig, err = clientcmd.BuildConfigFromFlags("", f.config.ServiceExternalKubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("build restconfig for service: %w", err)
	}

	return
}

func (f *Framework) ManagementClient(t *testing.T, options ...func(config *restclient.Config) error) (*testutil.RecordingClient, error) {
	cfg := *f.managementConfig
	for _, f := range options {
		if err := f(&cfg); err != nil {
			return nil, err
		}
	}
	return testutil.NewRecordingClient(t, &cfg, f.ManagementScheme, f.config.CleanUpStrategy), nil
}

func (f *Framework) ServiceClient(t *testing.T, options ...func(config *restclient.Config) error) (*testutil.RecordingClient, error) {
	cfg := *f.serviceConfig
	for _, f := range options {
		if err := f(&cfg); err != nil {
			return nil, err
		}
	}
	return testutil.NewRecordingClient(t, &cfg, f.ServiceScheme, f.config.CleanUpStrategy), nil
}

func (f *Framework) SetupServiceCluster(ctx context.Context, cl *testutil.RecordingClient, t *testing.T, name string, account *catalogv1alpha1.Account) *corev1alpha1.ServiceCluster {
	// Setup
	serviceKubeconfig, err := ioutil.ReadFile(f.Config().ServiceInternalKubeconfigPath)
	require.NoError(t, err, "cannot read service internal kubeconfig")

	serviceClusterSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: account.Status.Namespace.Name,
		},
		Data: map[string][]byte{
			"kubeconfig": serviceKubeconfig,
		},
	}

	serviceCluster := NewServiceCluster(name, account.Status.Namespace.Name, serviceClusterSecret.Name)

	require.NoError(t, cl.Create(ctx, serviceClusterSecret))
	require.NoError(t, cl.Create(ctx, serviceCluster))
	require.NoError(t, testutil.WaitUntilReady(ctx, cl, serviceCluster, testutil.WithTimeout(60*time.Second)))
	t.Logf("service cluster %s successfully created for provider %s", name, account.Name)
	return serviceCluster
}

func (f *Framework) Config() FrameworkConfig {
	return f.config
}

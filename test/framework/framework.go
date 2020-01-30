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
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
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
	MasterScheme  *runtime.Scheme
	masterConfig  *restclient.Config
	ServiceScheme *runtime.Scheme
	serviceConfig *restclient.Config
	config        Config
}

func New(c Config) (f *Framework, err error) {
	f = &Framework{config: c}

	// Master Setup
	f.MasterScheme = runtime.NewScheme()
	if err = clientgoscheme.AddToScheme(f.MasterScheme); err != nil {
		return nil, fmt.Errorf("adding clientgo scheme to master scheme: %w", err)
	}
	if err = apiextensionsv1.AddToScheme(f.MasterScheme); err != nil {
		return nil, fmt.Errorf("adding apiextensionsv1 scheme to master scheme: %w", err)
	}
	if err = operatorv1alpha1.AddToScheme(f.MasterScheme); err != nil {
		return nil, fmt.Errorf("adding operatorv1alpha1 scheme to master scheme: %w", err)
	}
	if err = catalogv1alpha1.AddToScheme(f.MasterScheme); err != nil {
		return nil, fmt.Errorf("adding catalogv1alpha1 scheme to master scheme: %w", err)
	}
	if err = corev1alpha1.AddToScheme(f.MasterScheme); err != nil {
		return nil, fmt.Errorf("adding corev1alpha1 scheme to master scheme: %w", err)
	}

	f.masterConfig, err = clientcmd.BuildConfigFromFlags("", f.config.MasterExternalKubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("build restconfig for master: %w", err)
	}

	// Service Setup
	f.ServiceScheme = runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(f.ServiceScheme); err != nil {
		return nil, fmt.Errorf("adding clientgo scheme to service scheme: %w", err)
	}
	if err = apiextensionsv1.AddToScheme(f.ServiceScheme); err != nil {
		return nil, fmt.Errorf("adding apiextensionsv1 scheme to service scheme: %w", err)
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
		Scheme: f.MasterScheme,
	})
}

func (f *Framework) ServiceClient() (client.Client, error) {
	cfg := f.serviceConfig
	return client.New(cfg, client.Options{
		Scheme: f.ServiceScheme,
	})
}

func (f *Framework) Config() Config {
	return f.config
}

func (f *Framework) NewFrameworkContext() (*FrameworkContext, error) {
	fctx := &FrameworkContext{
		masterTracking:  make([]runtime.Object, 0),
		serviceTracking: make([]runtime.Object, 0),
	}
	mClient, err := f.MasterClient()
	if err != nil {
		return nil, err
	}
	sClient, err := f.ServiceClient()
	if err != nil {
		return nil, err
	}
	fctx.MasterClient = &FrameworkTrackingClient{
		tracking: &fctx.masterTracking,
		Client:   mClient,
	}
	fctx.ServiceClient = &FrameworkTrackingClient{
		tracking: &fctx.serviceTracking,
		Client:   sClient,
	}
	return fctx, nil
}

type FrameworkContext struct {
	MasterClient    *FrameworkTrackingClient
	ServiceClient   *FrameworkTrackingClient
	masterTracking  []runtime.Object
	serviceTracking []runtime.Object
}

func (cl *FrameworkContext) CleanUp(t *testing.T) {
	// cleanup in reverse order of creation
	for i := len(cl.masterTracking) - 1; i >= 0; i-- {
		obj := cl.masterTracking[i]
		objMeta := obj.(metav1.Object)
		assert.NoError(t, testutil.DeleteAndWaitUntilNotFound(cl.MasterClient, obj), "cannot delete %T:%s from master cluster", obj, objMeta.GetName())
	}
	for i := len(cl.serviceTracking) - 1; i >= 0; i-- {
		obj := cl.serviceTracking[i]
		objMeta := obj.(metav1.Object)
		assert.NoError(t, testutil.DeleteAndWaitUntilNotFound(cl.ServiceClient, obj), "cannot delete %T:%s from service cluster", obj, objMeta.GetName())
	}
}

type FrameworkTrackingClient struct {
	tracking *[]runtime.Object
	client.Client
}

var _ client.Client = (*FrameworkTrackingClient)(nil)

func (cl *FrameworkTrackingClient) Create(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) error {
	*cl.tracking = append(*cl.tracking, obj)
	return cl.Client.Create(ctx, obj, opts...)
}

func (cl *FrameworkTrackingClient) CreateAndWaitUntilReady(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) error {
	if err := cl.Create(ctx, obj, opts...); err != nil {
		return err
	}
	return testutil.WaitUntilReady(cl, obj)
}

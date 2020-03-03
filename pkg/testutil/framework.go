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
	"os"
	"sync"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

type FrameworkConfig struct {
	TestID string

	ManagementExternalKubeconfigPath string
	ManagementInternalKubeconfigPath string
	ServiceExternalKubeconfigPath    string
	ServiceInternalKubeconfigPath    string
	CleanUpStrategy                  CleanUpStrategy
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
	if c.CleanUpStrategy != CleanupAlways && c.CleanUpStrategy != CleanupOnSuccess && c.CleanUpStrategy != CleanupNever {
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
	f.serviceConfig, err = clientcmd.BuildConfigFromFlags("", f.config.ServiceExternalKubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("build restconfig for service: %w", err)
	}

	return
}

func (f *Framework) ManagementClient(log logr.Logger) (*RecordingClient, error) {
	cfg := f.managementConfig
	c, err := util.NewClientWatcher(cfg, f.ManagementScheme, log)
	if err != nil {
		return nil, err
	}
	return recordingClient(c, f.ManagementScheme), nil
}

func (f *Framework) ServiceClient(log logr.Logger) (*RecordingClient, error) {
	cfg := f.serviceConfig
	c, err := util.NewClientWatcher(cfg, f.ServiceScheme, log)
	if err != nil {
		return nil, err
	}
	return recordingClient(c, f.ServiceScheme), nil
}

func (f *Framework) NewProviderAccount(name string) *catalogv1alpha1.Account {
	return &catalogv1alpha1.Account{
		ObjectMeta: metav1.ObjectMeta{
			Name: name + "-provider",
		},
		Spec: catalogv1alpha1.AccountSpec{
			Metadata: catalogv1alpha1.AccountMetadata{
				DisplayName: name + " provider",
				Description: name + " provider desc",
			},
			Roles: []catalogv1alpha1.AccountRole{
				catalogv1alpha1.ProviderRole,
			},
		},
	}
}

func (f *Framework) NewTenantAccount(name string) *catalogv1alpha1.Account {
	return &catalogv1alpha1.Account{
		ObjectMeta: metav1.ObjectMeta{
			Name: name + "-tenant",
		},
		Spec: catalogv1alpha1.AccountSpec{
			Metadata: catalogv1alpha1.AccountMetadata{
				DisplayName: name + " tenant",
				Description: name + " tenant desc",
			},
			Roles: []catalogv1alpha1.AccountRole{
				catalogv1alpha1.TenantRole,
			},
		},
	}
}

func (f *Framework) Config() FrameworkConfig {
	return f.config
}

type CleanUpStrategy string

const (
	CleanupAlways    CleanUpStrategy = "always"
	CleanupOnSuccess CleanUpStrategy = "on-success"
	CleanupNever     CleanUpStrategy = "never"
)

type RecordingClient struct {
	*util.ClientWatcher
	scheme  *runtime.Scheme
	objects map[string]runtime.Object
	order   []string
	mux     sync.Mutex
}

func recordingClient(cw *util.ClientWatcher, scheme *runtime.Scheme) *RecordingClient {
	return &RecordingClient{
		ClientWatcher: cw,
		scheme:        scheme,
		objects:       map[string]runtime.Object{},
	}
}

var _ client.Client = (*RecordingClient)(nil)

func (rc *RecordingClient) key(obj runtime.Object) string {
	gvk, err := apiutil.GVKForObject(obj, rc.scheme)
	if err != nil {
		panic(err)
	}

	meta := obj.(metav1.Object)
	key := types.NamespacedName{
		Name:      meta.GetName(),
		Namespace: meta.GetNamespace(),
	}.String()
	return fmt.Sprintf("%s.%s/%s:%s", gvk.Kind, gvk.Group, gvk.Version, key)
}

func (rc *RecordingClient) RegisterForCleanup(obj runtime.Object) {
	rc.mux.Lock()
	defer rc.mux.Unlock()

	key := rc.key(obj)
	rc.objects[key] = obj
	rc.order = append(rc.order, key)
}

func (rc *RecordingClient) UnregisterForCleanup(obj runtime.Object) {
	rc.mux.Lock()
	defer rc.mux.Unlock()

	key := rc.key(obj)
	delete(rc.objects, key)
}

func (rc *RecordingClient) CleanUpFunc(ctx context.Context, t *testing.T, strategy CleanUpStrategy) func() {
	return func() {
		switch strategy {
		case CleanupNever:
			return
		case CleanupOnSuccess:
			if t.Failed() {
				return
			}
		case CleanupAlways:
			break
		default:
			t.Logf("unknown cleanup strategy: %v", strategy)
			t.FailNow()
		}

		// cleanup in reverse order of creation
		for i := len(rc.order) - 1; i >= 0; i-- {
			key := rc.order[i]
			obj, ok := rc.objects[key]
			if !ok {
				continue
			}

			err := DeleteAndWaitUntilNotFound(ctx, rc, obj)
			if err != nil {
				err = fmt.Errorf("cleanup %s: %w", key, err)
			}
			require.NoError(t, err)
		}
	}
}

func (rc *RecordingClient) Create(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) error {
	rc.RegisterForCleanup(obj)
	return rc.Client.Create(ctx, obj, opts...)
}

func (rc *RecordingClient) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOption) error {
	rc.UnregisterForCleanup(obj)
	return rc.Client.Delete(ctx, obj, opts...)
}

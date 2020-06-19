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
	"sync"
	"testing"
	"time"

	certmanagerv1alpha2 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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
	fakev1 "github.com/kubermatic/kubecarrier/pkg/apis/fake/v1"
	fakev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/fake/v1alpha1"
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

func (f *Framework) ManagementClient(t *testing.T, options ...func(config *restclient.Config) error) (*RecordingClient, error) {
	cfg := *f.managementConfig
	for _, f := range options {
		if err := f(&cfg); err != nil {
			return nil, err
		}
	}
	c, err := util.NewClientWatcher(&cfg, f.ManagementScheme, NewLogger(t))
	if err != nil {
		return nil, err
	}
	return recordingClient(c, f.ManagementScheme, t, f.config.CleanUpStrategy), nil
}

func (f *Framework) ServiceClient(t *testing.T, options ...func(config *restclient.Config) error) (*RecordingClient, error) {
	cfg := *f.serviceConfig
	for _, f := range options {
		if err := f(&cfg); err != nil {
			return nil, err
		}
	}
	c, err := util.NewClientWatcher(&cfg, f.ServiceScheme, NewLogger(t))
	if err != nil {
		return nil, err
	}
	return recordingClient(c, f.ServiceScheme, t, f.config.CleanUpStrategy), nil
}

func (f *Framework) SetupServiceCluster(ctx context.Context, cl *RecordingClient, t *testing.T, name string, account *catalogv1alpha1.Account) *corev1alpha1.ServiceCluster {
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
	require.NoError(t, WaitUntilReady(ctx, cl, serviceCluster, WithTimeout(60*time.Second)))
	t.Logf("service cluster %s successfully created for provider %s", name, account.Name)
	return serviceCluster
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
	t *testing.T
	*util.ClientWatcher
	scheme          *runtime.Scheme
	objects         map[string]runtime.Object
	order           []string
	cleanUpStrategy CleanUpStrategy
	mux             sync.Mutex
}

func recordingClient(cw *util.ClientWatcher, scheme *runtime.Scheme, t *testing.T, strategy CleanUpStrategy) *RecordingClient {
	return &RecordingClient{
		ClientWatcher:   cw,
		scheme:          scheme,
		objects:         map[string]runtime.Object{},
		t:               t,
		cleanUpStrategy: strategy,
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

func (rc *RecordingClient) CleanUpFunc(ctx context.Context) func() {
	return func() {
		rc.t.Helper()
		switch rc.cleanUpStrategy {
		case CleanupNever:
			return
		case CleanupOnSuccess:
			if rc.t.Failed() {
				return
			}
		case CleanupAlways:
			break
		default:
			rc.t.Logf("unknown cleanup strategy: %v", rc.cleanUpStrategy)
			rc.t.FailNow()
		}

		// cleanup in reverse order of creation
		for i := len(rc.order) - 1; i >= 0; i-- {
			key := rc.order[i]
			obj, ok := rc.objects[key]
			if !ok {
				continue
			}

			err := DeleteAndWaitUntilNotFound(ctx, rc, obj, WithTimeout(time.Minute))
			if err != nil {
				err = fmt.Errorf("cleanup %s: %w", key, err)
			}
			require.NoError(rc.t, err)
		}
	}
}

func (rc *RecordingClient) Create(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) error {
	rc.t.Helper()
	rc.t.Logf("creating %s", util.MustLogLine(obj, rc.scheme))
	rc.RegisterForCleanup(obj)
	return rc.ClientWatcher.Create(ctx, obj, opts...)
}

func (rc *RecordingClient) EnsureCreated(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) error {
	rc.t.Helper()
	rc.t.Logf("creating %s", util.MustLogLine(obj, rc.scheme))
	rc.RegisterForCleanup(obj)
	oldObj := obj.DeepCopyObject()
	err := rc.ClientWatcher.Create(ctx, obj, opts...)
	if err != nil && errors.IsAlreadyExists(err) {
		rc.t.Logf("alreadyExists, update %s", util.MustLogLine(obj, rc.scheme))
		updateErr := rc.ClientWatcher.Update(ctx, oldObj)
		if err := rc.scheme.Convert(oldObj, obj, nil); err != nil {
			return err
		}
		return updateErr
	}
	return nil
}

func (rc *RecordingClient) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOption) error {
	rc.t.Helper()
	rc.t.Logf("deleting %s", util.MustLogLine(obj, rc.scheme))
	rc.UnregisterForCleanup(obj)
	return rc.ClientWatcher.Delete(ctx, obj, opts...)
}

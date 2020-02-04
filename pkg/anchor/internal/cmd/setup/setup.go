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

package setup

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gernest/wow"
	"github.com/gernest/wow/spin"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	watch2 "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	cache2 "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/watch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/kubermatic/kubecarrier/pkg/anchor/internal/spinner"
	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/reconcile"
	"github.com/kubermatic/kubecarrier/pkg/internal/resources/operator"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

type flags struct {
	// Kubeconfig is the absolute path of the kubeconfig of the kubernetes cluster which you want to deploy kubecarrier.
	Kubeconfig string
}

var (
	scheme = runtime.NewScheme()
)

const (
	kubecarrierNamespaceName = "kubecarrier-system"
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = apiextensionsv1.AddToScheme(scheme)
	_ = operatorv1alpha1.AddToScheme(scheme)
}

func NewCommand(log logr.Logger) *cobra.Command {
	flags := &flags{}
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "setup",
		Short: "Deploy kubecarrier operator",
		Long: `Deploy kubecarrier operator in a kubernetes cluster.
Here are some examples:
- You can specify the kubeconfig absolute path of the cluster that you want to deploy everything in it:
$ anchor setup --kubeconfig=<kubeconfig path>
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runE(flags, log, cmd)
		},
	}

	cmd.Flags().StringVar(&flags.Kubeconfig, "kubeconfig", os.Getenv("KUBECONFIG"), "The absolute path of the kubeconfig of kubernetes cluster that set up with. if you don't specify the flag, it will read from the KUBECONFIG environment variable.")
	return cmd
}

func runE(flags *flags, log logr.Logger, cmd *cobra.Command) error {
	stopCh := ctrl.SetupSignalHandler()
	ctx, cancelContext := context.WithTimeout(context.Background(), 60*time.Second)
	go func() {
		<-stopCh
		cancelContext()
	}()

	s := wow.New(cmd.OutOrStdout(), spin.Get(spin.Dots), "")
	startTime := time.Now()

	// Check the kubeconfig
	if err := spinner.AttachSpinnerTo(s, startTime, "Check kubeconfig", func() error {
		if err := checkKubeconfig(flags.Kubeconfig); err != nil {
			return err
		}

		// Set the kubeconfig environment variable so the client in the following can work with the cluster.
		if err := os.Setenv("KUBECONFIG", flags.Kubeconfig); err != nil {
			return nil
		}
		return nil
	}); err != nil {
		return err
	}

	// Get a client from the configuration of the kubernetes cluster.
	conf, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("getting Kubernetes cluster config: %w", err)
	}
	c, err := newClientWatcher(conf)
	if err != nil {
		return fmt.Errorf("creating Kubernetes client: %w", err)
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: kubecarrierNamespaceName,
		},
	}
	if err := spinner.AttachSpinnerTo(s, startTime, fmt.Sprintf("Create %q Namespace", kubecarrierNamespaceName), createNamespace(ctx, c, ns)); err != nil {
		return fmt.Errorf("creating KubeCarrier system namespace: %w", err)
	}

	if err := spinner.AttachSpinnerTo(s, startTime, "Deploy KubeCarrier Operator", reconcileOperator(ctx, log, c, ns)); err != nil {
		return fmt.Errorf("deploying kubecarrier operator: %w", err)
	}

	if err := spinner.AttachSpinnerTo(s, startTime, "Deploy KubeCarrier", deployKubeCarrier(ctx, ns, conf)); err != nil {
		return fmt.Errorf("deploying kubecarrier controller manager: %w", err)
	}

	return nil
}

func createNamespace(ctx context.Context, c client.Client, ns *corev1.Namespace) func() error {
	return func() error {
		if err := c.Create(ctx, ns); err != nil {
			if errors.IsAlreadyExists(err) {
				if err := c.Get(ctx, types.NamespacedName{Name: ns.ObjectMeta.Name}, ns); err != nil {
					return fmt.Errorf("getting KubeCarrier system namespace: %w", err)
				}
				return nil
			} else {
				return fmt.Errorf("creating KubeCarrier system namespace: %w", err)
			}
		}
		return nil
	}
}

func checkKubeconfig(kubeconfig string) error {
	kubeconfigPath := strings.TrimSpace(kubeconfig)
	if kubeconfigPath == "" {
		return fmt.Errorf("either $KUBECONFIG or --kubeconfig flag needs to be set")
	}

	kubeconfigStat, err := os.Stat(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("checking the kubeconfig path: %w", err)
	}
	// Check the kubeconfig path points to a file
	if !kubeconfigStat.Mode().IsRegular() {
		return fmt.Errorf("kubeconfig path %s does not point to a file", kubeconfigPath)
	}
	return nil
}

func reconcileOperator(ctx context.Context, log logr.Logger, c *clientWatcher, kubecarrierNamespace *corev1.Namespace) func() error {
	return func() error {
		// Kustomize Build
		objects, err := operator.Manifests(
			operator.Config{
				Namespace: kubecarrierNamespace.Name,
			})
		if err != nil {
			return fmt.Errorf("creating operator manifests: %w", err)
		}

		for _, object := range objects {
			if err := controllerutil.SetControllerReference(kubecarrierNamespace, &object, scheme); err != nil {
				return fmt.Errorf("set controller reference: %w", err)
			}
			_, err := reconcile.Unstructured(ctx, log, c, &object)
			if err != nil {
				return fmt.Errorf("reconcile kind: %s, err: %w", object.GroupVersionKind().Kind, err)
			}
		}

		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kubecarrier-operator-manager",
				Namespace: kubecarrierNamespaceName,
			},
		}
		return c.WaitUntil(ctx, deployment, func(obj runtime.Object) (b bool, err error) {
			return util.DeploymentIsAvailable(obj.(*appsv1.Deployment)), nil
		})

	}
}

// deployKubeCarrier deploys the KubeCarrier Object in a kubernetes cluster.
func deployKubeCarrier(ctx context.Context, kubeCarrierNamespace *corev1.Namespace, conf *rest.Config) func() error {
	return func() error {
		// Create another client due to some issues about the restmapper.
		// The issue is that if you use the client that created before, and here try to create the kubeCarrier,
		// it will complain about: `no matches for kind "KubeCarrier" in version "operator.kubecarrier.io/v1alpha1"`,
		// but actually, the scheme is already added to the runtime scheme.
		// And in the following, reinitializing the client solves the issue.

		kubeCarrier := &operatorv1alpha1.KubeCarrier{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kubecarrier",
				Namespace: kubeCarrierNamespace.Name,
			},
		}
		w, err := newClientWatcher(conf)
		if err != nil {
			return err
		}
		if _, err := ctrl.CreateOrUpdate(ctx, w, kubeCarrier, func() error {
			return nil
		}); err != nil {
			return fmt.Errorf("cannot create or update kubecarrier: %w", err)
		}
		return w.WaitUntil(ctx, kubeCarrier, func(obj runtime.Object) (b bool, err error) {
			return obj.(*operatorv1alpha1.KubeCarrier).IsReady(), nil
		})
	}
}

func newClientWatcher(conf *rest.Config) (*clientWatcher, error) {
	mapper, err := apiutil.NewDynamicRESTMapper(conf, apiutil.WithLazyDiscovery)
	if err != nil {
		return nil, fmt.Errorf("rest mapper: %w", err)
	}
	d, err := dynamic.NewForConfig(conf)
	if err != nil {
		return nil, err
	}
	cll, err := client.New(conf, client.Options{
		Scheme: scheme,
		Mapper: mapper,
	})
	if err != nil {
		return nil, err
	}
	return &clientWatcher{
		client:     d,
		restMapper: mapper,
		scheme:     scheme,
		Client:     cll,
	}, nil
}

type clientWatcher struct {
	client     dynamic.Interface
	restMapper meta.RESTMapper
	scheme     *runtime.Scheme
	client.Client
}

func (cl *clientWatcher) WaitUntil(ctx context.Context, obj util.Object, cond ...func(obj runtime.Object) (bool, error)) error {
	objGVK, err := apiutil.GVKForObject(obj, scheme)
	if err != nil {
		return err
	}
	rmap, err := cl.restMapper.RESTMapping(objGVK.GroupKind(), objGVK.Version)
	if err != nil {
		return err
	}
	objNN := types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
	ri := cl.client.Resource(rmap.Resource).Namespace(objNN.Namespace)
	if _, err := watch.ListWatchUntil(ctx, &cache2.ListWatch{
		ListFunc: func(options metav1.ListOptions) (object runtime.Object, err error) {
			return ri.List(options)
		},
		WatchFunc: ri.Watch,
	}, func(event watch2.Event) (b bool, err error) {
		objTmp, err := scheme.New(objGVK)
		if err != nil {
			return false, err
		}
		obj := objTmp.(util.Object)
		if err := scheme.Convert(event.Object, obj, nil); err != nil {
			return false, err
		}
		if obj.GetNamespace() != objNN.Namespace || obj.GetName() != objNN.Name {
			return false, nil
		}
		for _, f := range cond {
			ok, err := f(objTmp)
			if err != nil {
				return false, err
			}
			if !ok {
				return false, nil
			}
		}
		return true, nil
	}); err != nil {
		return err
	}
	return nil
}

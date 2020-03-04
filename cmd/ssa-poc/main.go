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

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(catalogv1alpha1.AddToScheme(scheme))
	utilruntime.Must(operatorv1alpha1.AddToScheme(scheme))
	utilruntime.Must(apiextensionsv1.AddToScheme(scheme))
	utilruntime.Must(corev1alpha1.AddToScheme(scheme))
}

type LogRoundTripper struct {
	Rt http.RoundTripper
}

// RoundTrip performs a round-trip HTTP request and logs relevant information
// about it.
func (lrt *LogRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	defer func() {
		if request.Body != nil {
			request.Body.Close()
		}
	}()

	var err error

	log.Printf("[DEBUG] Request URL: %s %s", request.Method, request.URL)

	if request.Body != nil {
		body, err := request.GetBody()
		if err != nil {
			return nil, err
		}
		if bodyBB, err := ioutil.ReadAll(body); err != nil {
			return nil, err
		} else {
			log.Printf("[DEBUG] Request body\n%s", string(bodyBB))
		}
		if err := body.Close(); err != nil {
			return nil, err
		}
	}

	response, err := lrt.Rt.RoundTrip(request)
	if response == nil {
		return nil, err
	}

	log.Printf("[DEBUG] Response Code: %d", response.StatusCode)
	return response, err
}

func main() {
	flags := genericclioptions.NewConfigFlags(false)
	cmd := &cobra.Command{
		Use: "ssa-poc",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := flags.ToRESTConfig()
			cfg.Wrap(func(rt http.RoundTripper) http.RoundTripper {
				return &LogRoundTripper{rt}
			})
			if err != nil {
				return err
			}
			c, err := client.New(cfg, client.Options{
				Scheme: scheme,
			})
			if err != nil {
				return err
			}
			ctx := context.Background()
			configMaps := &corev1.ConfigMapList{}
			if err := c.List(ctx, configMaps); err != nil {
				return err
			}

			crdiscovery := &corev1alpha1.CustomResourceDiscovery{}
			crdiscovery.Namespace = "default"
			crdiscovery.Name = "test"

			if _, err := ctrl.CreateOrUpdate(ctx, c, crdiscovery, func() error { return nil }); err != nil {
				return err
			}

			for _, cond := range []corev1alpha1.CustomResourceDiscoveryConditionType{
				corev1alpha1.CustomResourceDiscoveryControllerReady,
				//corev1alpha1.CustomResourceDiscoveryDiscovered,
				//corev1alpha1.CustomResourceDiscoveryEstablished,
			} {
				go settingTheSSAStatus(ctx, c, crdiscovery.DeepCopy(), cond)
			}

			<-ctx.Done()
			return nil
		},
	}
	flags.AddFlags(cmd.Flags())
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

func settingTheStatus(ctx context.Context, cl client.Client, obj *corev1alpha1.CustomResourceDiscovery, conditionType corev1alpha1.CustomResourceDiscoveryConditionType) {
	for i := 0; ; i++ {
		time.Sleep(time.Duration(rand.Int63n(int64(time.Second))))
		obj.Status.SetCondition(corev1alpha1.CustomResourceDiscoveryCondition{
			Message: fmt.Sprint(i),
			Reason:  "NewNumberGenerated",
			Status:  corev1alpha1.ConditionTrue,
			Type:    conditionType,
		})
		if err := cl.Status().Update(ctx, obj); err != nil {
			if errors.IsConflict(err) {
				fmt.Printf("%v conflict!\n", conditionType)
				if err := cl.Get(ctx, types.NamespacedName{
					Namespace: obj.Namespace,
					Name:      obj.Name,
				}, obj); err != nil {
					panic(err)
				}
				continue
			}
			fmt.Printf("%v error happened %s\n", conditionType, err.Error())
		}
		fmt.Printf("%v setting to number %d\n", conditionType, i)
	}
}

func settingTheSSAStatus(ctx context.Context, cl client.Client, obj *corev1alpha1.CustomResourceDiscovery, conditionType corev1alpha1.CustomResourceDiscoveryConditionType) {
	for i := 0; ; i++ {
		time.Sleep(time.Duration(rand.Int63n(int64(time.Second))))
		obj.Status.SetCondition(corev1alpha1.CustomResourceDiscoveryCondition{
			Message: fmt.Sprint(i),
			Reason:  "NewNumberGenerated",
			Status:  corev1alpha1.ConditionTrue,
			Type:    conditionType,
		})
		/*
			if err := cl.Status().Update(ctx, u, client.FieldOwner(fmt.Sprint(conditionType))); err != nil {
				if errors.IsConflict(err) {
					fmt.Printf("%v conflict!\n", conditionType)
					if err := cl.Get(ctx, types.NamespacedName{
						Namespace: obj.Namespace,
						Name:      obj.Name,
					}, obj); err != nil {
						panic(err)
					}
					continue
				}
				fmt.Printf("%v error happened %s\n", conditionType, err.Error())
			} else {
				fmt.Printf("%v setting to number %d\n", conditionType, i)
			}
		*/

		if err := cl.Status().Patch(ctx, obj, client.Apply, client.FieldOwner(fmt.Sprint(conditionType)), client.ForceOwnership); err != nil {
			if errors.IsConflict(err) {
				fmt.Printf("%v conflict!\n", conditionType)
				if err := cl.Get(ctx, types.NamespacedName{
					Namespace: obj.Namespace,
					Name:      obj.Name,
				}, obj); err != nil {
					panic(err)
				}
				continue
			}
			fmt.Printf("%v error happened %s\n", conditionType, err.Error())
		} else {
			fmt.Printf("%v setting to number %d\n", conditionType, i)
		}

	}
}

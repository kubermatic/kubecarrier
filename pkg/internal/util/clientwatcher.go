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

package util

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	clientwatch "k8s.io/client-go/tools/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func NewClientWatcher(conf *rest.Config, scheme *runtime.Scheme) (*ClientWatcher, error) {
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
	return &ClientWatcher{
		client:     d,
		restMapper: mapper,
		scheme:     scheme,
		Client:     cll,
	}, nil
}

type ClientWatcher struct {
	client     dynamic.Interface
	restMapper meta.RESTMapper
	scheme     *runtime.Scheme
	client.Client
}

func (cl *ClientWatcher) WaitUntil(ctx context.Context, obj Object, cond ...func(obj runtime.Object) (bool, error)) error {
	objGVK, err := apiutil.GVKForObject(obj, cl.scheme)
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
	if _, err := clientwatch.ListWatchUntil(ctx, &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (object runtime.Object, err error) {
			return ri.List(options)
		},
		WatchFunc: ri.Watch,
	}, func(event watch.Event) (b bool, err error) {
		objTmp, err := cl.scheme.New(objGVK)
		if err != nil {
			return false, err
		}
		obj := objTmp.(Object)
		if err := cl.scheme.Convert(event.Object, obj, nil); err != nil {
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

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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	clientwatch "k8s.io/client-go/tools/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type ClientWatcher struct {
	dynamicClient dynamic.Interface
	restMapper    meta.RESTMapper
	scheme        *runtime.Scheme
	log           logr.Logger
	client.Client
}

var _ client.Client = (*ClientWatcher)(nil)

func NewClientWatcher(conf *rest.Config, scheme *runtime.Scheme, log logr.Logger) (*ClientWatcher, error) {
	mapper, err := apiutil.NewDynamicRESTMapper(conf, apiutil.WithLazyDiscovery)
	if err != nil {
		return nil, fmt.Errorf("rest mapper: %w", err)
	}
	d, err := dynamic.NewForConfig(conf)
	if err != nil {
		return nil, err
	}
	k8sClient, err := client.New(conf, client.Options{
		Scheme: scheme,
		Mapper: mapper,
	})
	if err != nil {
		return nil, err
	}
	return &ClientWatcher{
		dynamicClient: d,
		restMapper:    mapper,
		scheme:        scheme,
		Client:        k8sClient,
		log:           log,
	}, nil
}

// WaitUntil waits until the Object's condition function is true, or the context deadline is reached
func (cw *ClientWatcher) WaitUntil(ctx context.Context, obj runtime.Object, cond ...func(obj runtime.Object, eventType watch.EventType) (bool, error)) error {
	objNN, err := client.ObjectKeyFromObject(obj)
	if err != nil {
		return fmt.Errorf("getting object key: %w", err)
	}
	objGVK, err := apiutil.GVKForObject(obj, cw.scheme)
	if err != nil {
		return err
	}
	lw, err := cw.objListWatch(obj)
	if err != nil {
		return fmt.Errorf("getting objListWatch: %w", err)
	}
	if _, err := clientwatch.ListWatchUntil(ctx, lw, func(event watch.Event) (b bool, err error) {
		typedEventObject, err := cw.scheme.New(objGVK)
		if err == nil {
			if err := cw.scheme.Convert(event.Object, typedEventObject, nil); err != nil {
				return false, err
			}
		} else {
			if runtime.IsNotRegisteredError(err) {
				// the object is unregistered in the scheme, keep it as *unstructured.Unstructured
				typedEventObject = event.Object
			} else {
				return false, err
			}
		}
		for _, f := range cond {
			ok, err := f(typedEventObject, event.Type)
			if err != nil {
				return false, err
			}
			if !ok {
				return false, nil
			}
		}
		if err := cw.Get(ctx, objNN, obj); err != nil {
			return false, err
		}
		return true, nil
	}); err != nil {
		return fmt.Errorf("%s: %w", MustLogLine(obj, cw.scheme), err)
	}
	return nil
}

// WaitUntilNotFound waits until the object is not found or the context deadline is exceeded
func (cw *ClientWatcher) WaitUntilNotFound(ctx context.Context, obj runtime.Object) error {
	// things get a bit tricky with not found watches
	//  clientwatch.UntilWithSync seems useful since it has cache pre-conditions which I can check whether
	// the objects existed in initial list operation. But there are few other issues with it:
	// * it doesn't call condition function with DELETED event types for some reason (nor does it get it from watch interface to my current debugging knowledge)
	// * it doesn't properly update the cache store since the event objects are types to *unstructured.Unstructured instead of GVK schema type
	lw, err := cw.objListWatch(obj)
	if err != nil {
		return fmt.Errorf("getting objListWatch: %w", err)
	}
	list, err := lw.List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	initialItems, err := meta.ExtractList(list)
	if err != nil {
		return err
	}
	if len(initialItems) == 0 {
		return nil
	}

	metaObj, err := meta.ListAccessor(list)
	if err != nil {
		return err
	}
	currResourceVersion := metaObj.GetResourceVersion()
	_, err = clientwatch.Until(ctx, currResourceVersion, lw, func(event watch.Event) (b bool, err error) {
		return event.Type == watch.Deleted, nil
	})
	if err != nil {
		return fmt.Errorf("%s: %w", MustLogLine(obj, cw.scheme), err)
	}
	return nil
}

func (cw *ClientWatcher) objListWatch(obj runtime.Object) (*cache.ListWatch, error) {
	objGVK, err := apiutil.GVKForObject(obj, cw.scheme)
	if err != nil {
		return nil, err
	}
	objNN, err := client.ObjectKeyFromObject(obj)
	if err != nil {
		return nil, fmt.Errorf("getting object key: %w", err)
	}
	restMapping, err := cw.restMapper.RESTMapping(objGVK.GroupKind(), objGVK.Version)
	if err != nil {
		return nil, err
	}
	resourceInterface := cw.dynamicClient.Resource(restMapping.Resource).Namespace(objNN.Namespace)
	return &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (object runtime.Object, err error) {
			options.FieldSelector = "metadata.name=" + objNN.Name
			return resourceInterface.List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (w watch.Interface, err error) {
			options.FieldSelector = "metadata.name=" + objNN.Name
			return resourceInterface.Watch(options)
		},
	}, nil
}

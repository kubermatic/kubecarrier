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
	"time"

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

type clientWatcherOption struct {
	softTimeout time.Duration
	hardTimeout time.Duration
}

func (opt *clientWatcherOption) Defaults() {
	if opt.softTimeout == 0 {
		opt.softTimeout = defaultTimeout
	}

	if opt.hardTimeout == 0 {
		opt.hardTimeout = 10 * opt.softTimeout
	}
}

type ClientWatcherOption func(*clientWatcherOption) error

func WithClientWatcherTimeout(t time.Duration) ClientWatcherOption {
	return func(option *clientWatcherOption) error {
		option.softTimeout = t
		return nil
	}
}

func WithClientWatcherHardTimeout(t time.Duration) ClientWatcherOption {
	return func(option *clientWatcherOption) error {
		option.hardTimeout = t
		return nil
	}
}

func WithoutClientWatcher() ClientWatcherOption {
	return func(option *clientWatcherOption) error {
		option.softTimeout = time.Duration(0)
		return nil
	}
}

const (
	defaultTimeout = 30 * time.Second
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

func (cw *ClientWatcher) init(ctx context.Context, obj runtime.Object, options []ClientWatcherOption) (context.Context, func(), error) {
	cfg := &clientWatcherOption{}
	for _, f := range options {
		if err := f(cfg); err != nil {
			return nil, nil, err
		}
	}
	cfg.Defaults()
	ctx, cancel := context.WithTimeout(ctx, cfg.hardTimeout)
	timer := time.AfterFunc(cfg.softTimeout, func() {
		fmt.Printf("[WARNING][SOFT_LIMIT_EXCEEDED] %s passed soft-limit\n", MustLogLine(obj, cw.scheme))
	})
	return ctx, func() {
		cancel()
		timer.Stop()
	}, nil
}

// WaitUntil waits until the Object's condition function is true, or the context deadline is reached
//
// condition function should operate on the passed object in a closure and should not modify the obj
func (cw *ClientWatcher) WaitUntil(ctx context.Context, obj runtime.Object, cond func() (done bool, err error), options ...ClientWatcherOption) error {
	ctx, closeFn, err := cw.init(ctx, obj, options)
	if err != nil {
		return err
	}
	defer closeFn()

	lw, err := cw.objListWatch(obj)
	if err != nil {
		return fmt.Errorf("getting objListWatch: %w", err)
	}
	if _, err := clientwatch.ListWatchUntil(ctx, lw, func(event watch.Event) (b bool, err error) {
		switch event.Type {
		case watch.Added:
			fallthrough
		case watch.Modified:
		case watch.Deleted:
			return
		case watch.Bookmark:
			return
		case watch.Error:
			return false, fmt.Errorf("watch error")
		}

		if err := cw.scheme.Convert(event.Object, obj, nil); err != nil {
			return false, err
		}
		ok, err := cond()
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
		return true, nil
	}); err != nil {
		return fmt.Errorf("%s (hardTimeout): %w", MustLogLine(obj, cw.scheme), err)
	}
	return nil
}

// WaitUntilNotFound waits until the object is not found or the context deadline is exceeded
func (cw *ClientWatcher) WaitUntilNotFound(ctx context.Context, obj runtime.Object, options ...ClientWatcherOption) error {
	ctx, closeFn, err := cw.init(ctx, obj, options)
	if err != nil {
		return err
	}
	defer closeFn()

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
		return fmt.Errorf("%s: (hardTimeout) %w", MustLogLine(obj, cw.scheme), err)
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
	if objNN.Name == "" {
		return nil, fmt.Errorf("name must not be empty")
	}
	restMapping, err := cw.restMapper.RESTMapping(objGVK.GroupKind(), objGVK.Version)
	if err != nil {
		return nil, err
	}
	if restMapping.Scope.Name() == meta.RESTScopeNameNamespace && objNN.Namespace == "" {
		return nil, fmt.Errorf("namespace must not be empty")
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

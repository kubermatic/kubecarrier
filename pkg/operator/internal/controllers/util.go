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

package controllers

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/kubermatic/utils/pkg/owner"
	"github.com/kubermatic/utils/pkg/util"

	"k8c.io/kubecarrier/pkg/internal/reconcile"
)

// object generic k8s object with metav1 and runtime Object interfaces implemented
type object interface {
	runtime.Object
	metav1.Object
}

// reconcileOwnedObjectsForNamespacedOwner reconciles objects which are owned by a namespaced owner.
// It works as following:
// If owned object is namespace-scoped, and is in the same namespace as the owner is in, it will use
// `controllerutil.SetControllerReference()` to do the owner handling.
// If owned object is cluster-scoped or in another namespace, it will use KubeCarrier customize owner package to perform
// owner handling.
func reconcileOwnedObjectsForNamespacedOwner(
	ctx context.Context,
	log logr.Logger,
	scheme *runtime.Scheme,
	restMapper meta.RESTMapper,
	client client.Client,
	ownerObj object,
	objects []unstructured.Unstructured) (bool, error) {
	var deploymentIsReady int32
	g, ctx := errgroup.WithContext(ctx)
	for _, object := range objects {
		object := object
		g.Go(func() error {
			gvk, err := apiutil.GVKForObject(&object, scheme)
			if err != nil {
				return err
			}
			restMapping, err := restMapper.RESTMapping(schema.GroupKind{
				Group: gvk.Group,
				Kind:  gvk.Kind,
			}, gvk.Version)
			if err != nil {
				return err
			}

			switch restMapping.Scope.Name() {
			case meta.RESTScopeNameNamespace:
				if err := controllerutil.SetControllerReference(ownerObj, &object, scheme); err != nil {
					return err
				}
			case meta.RESTScopeNameRoot:
				if _, err := owner.SetOwnerReference(ownerObj, &object, scheme); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unknown REST scope: %s", restMapping.Scope.Name())
			}
			curObj, err := reconcile.Unstructured(ctx, log, client, &object)
			if err != nil {
				return fmt.Errorf("reconcile kind: %s, err: %w", object.GroupVersionKind().Kind, err)
			}

			switch ctr := curObj.(type) {
			case *appsv1.Deployment:
				if util.DeploymentIsAvailable(ctr) {
					atomic.AddInt32(&deploymentIsReady, 1)
				}
			}
			return nil
		})
	}
	err := g.Wait()
	if err != nil {
		return false, err
	}
	return deploymentIsReady > 0, nil
}

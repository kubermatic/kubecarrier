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

package reconcile

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Unstructured reconciles a unstructured.Unstructured object,
// by calling the right typed reconcile function for the given GVK.
// Returns the "real" type, e.g.: *corev1.Service, *appsv1.Deployment.
func Unstructured(
	ctx context.Context,
	log logr.Logger,
	c client.Client,
	desiredObject *unstructured.Unstructured,
) (current metav1.Object, err error) {

	// lookup reconcile function
	gvk := desiredObject.GroupVersionKind()
	fn, ok := unstructuredReconcilers[gvk]
	if !ok {
		return nil, fmt.Errorf("cannot reconcile unknown type: %s, GVK needs to be registered in the 'unstructuredReconcilers' map in pkg/reconcile/unstructured.go", gvk)
	}

	return fn(ctx, log, c, desiredObject)
}

var unstructuredReconcilers = map[schema.GroupVersionKind]unstructuredReconcileFn{
	// "apps" group
	schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	}: unstructuredReconcileFn(unstructuredDeployment),

	// "core" group
	schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "ServiceAccount",
	}: unstructuredReconcileFn(unstructuredServiceAccount),
	schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Service",
	}: unstructuredReconcileFn(unstructuredService),

	// "rbac.authorization.k8s.io" group
	schema.GroupVersionKind{
		Group:   "rbac.authorization.k8s.io",
		Version: "v1",
		Kind:    "Role",
	}: unstructuredReconcileFn(unstructuredRole),
	schema.GroupVersionKind{
		Group:   "rbac.authorization.k8s.io",
		Version: "v1",
		Kind:    "RoleBinding",
	}: unstructuredReconcileFn(unstructuredRoleBinding),
	schema.GroupVersionKind{
		Group:   "rbac.authorization.k8s.io",
		Version: "v1",
		Kind:    "ClusterRole",
	}: unstructuredReconcileFn(unstructuredClusterRole),
	schema.GroupVersionKind{
		Group:   "rbac.authorization.k8s.io",
		Version: "v1",
		Kind:    "ClusterRoleBinding",
	}: unstructuredReconcileFn(unstructuredClusterRoleBinding),

	// "apiextensions.k8s.io" group
	schema.GroupVersionKind{
		Group:   "apiextensions.k8s.io",
		Version: "v1beta1",
		Kind:    "CustomResourceDefinition",
	}: unstructuredReconcileFn(unstructuredCustomResourceDefinition),
}

type unstructuredReconcileFn func(
	ctx context.Context,
	log logr.Logger,
	c client.Client,
	desiredObject *unstructured.Unstructured,
) (current metav1.Object, err error)

// "apps" group reconcile proxies

func unstructuredDeployment(
	ctx context.Context,
	log logr.Logger,
	c client.Client,
	desiredObj *unstructured.Unstructured,
) (current metav1.Object, err error) {
	// convert to proper type
	obj := &appsv1.Deployment{}
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(desiredObj.Object, obj); err != nil {
		return current, fmt.Errorf("convert from unstructured: %w", err)
	}

	return Deployment(ctx, log, c, obj)
}

// "core" group reconcile proxies

func unstructuredServiceAccount(
	ctx context.Context,
	log logr.Logger,
	c client.Client,
	desiredObj *unstructured.Unstructured,
) (current metav1.Object, err error) {
	// convert to proper type
	obj := &corev1.ServiceAccount{}
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(desiredObj.Object, obj); err != nil {
		return current, fmt.Errorf("convert from unstructured: %w", err)
	}

	return ServiceAccount(ctx, log, c, obj)
}

func unstructuredService(
	ctx context.Context,
	log logr.Logger,
	c client.Client,
	desiredObj *unstructured.Unstructured,
) (current metav1.Object, err error) {
	// convert to proper type
	obj := &corev1.Service{}
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(desiredObj.Object, obj); err != nil {
		return current, fmt.Errorf("convert from unstructured: %w", err)
	}

	return Service(ctx, log, c, obj)
}

// "rbac.authorization.k8s.io" group reconcile proxies

func unstructuredRole(
	ctx context.Context,
	log logr.Logger,
	c client.Client,
	desiredObj *unstructured.Unstructured,
) (current metav1.Object, err error) {
	// convert to proper type
	obj := &rbacv1.Role{}
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(desiredObj.Object, obj); err != nil {
		return current, fmt.Errorf("convert from unstructured: %w", err)
	}

	return Role(ctx, log, c, obj)
}

func unstructuredRoleBinding(
	ctx context.Context,
	log logr.Logger,
	c client.Client,
	desiredObj *unstructured.Unstructured,
) (current metav1.Object, err error) {
	// convert to proper type
	obj := &rbacv1.RoleBinding{}
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(desiredObj.Object, obj); err != nil {
		return current, fmt.Errorf("convert from unstructured: %w", err)
	}

	return RoleBinding(ctx, log, c, obj)
}

func unstructuredClusterRole(
	ctx context.Context,
	log logr.Logger,
	c client.Client,
	desiredObj *unstructured.Unstructured,
) (current metav1.Object, err error) {
	// convert to proper type
	obj := &rbacv1.ClusterRole{}
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(desiredObj.Object, obj); err != nil {
		return current, fmt.Errorf("convert from unstructured: %w", err)
	}

	return ClusterRole(ctx, log, c, obj)
}

func unstructuredClusterRoleBinding(
	ctx context.Context,
	log logr.Logger,
	c client.Client,
	desiredObj *unstructured.Unstructured,
) (current metav1.Object, err error) {
	// convert to proper type
	obj := &rbacv1.ClusterRoleBinding{}
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(desiredObj.Object, obj); err != nil {
		return current, fmt.Errorf("convert from unstructured: %w", err)
	}

	return ClusterRoleBinding(ctx, log, c, obj)
}

func unstructuredCustomResourceDefinition(
	ctx context.Context,
	log logr.Logger,
	c client.Client,
	desiredObj *unstructured.Unstructured,
) (current metav1.Object, err error) {
	// convert to proper type
	obj := &apiextensionsv1beta1.CustomResourceDefinition{}
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(desiredObj.Object, obj); err != nil {
		return current, fmt.Errorf("convert from unstructured: %w", err)
	}

	return CustomResourceDefinition(ctx, log, c, obj)

}

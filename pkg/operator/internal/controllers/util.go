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

	adminv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	ownerhelpers "github.com/kubermatic/kubecarrier/pkg/internal/owner"
)

type generalizedListOption interface {
	client.ListOption
	client.DeleteAllOfOption
}

// object generic k8s object with metav1 and runtime Object interfaces implemented
type object interface {
	runtime.Object
	metav1.Object
}

// addOwnerReference adds an OwnerReference to an object.
func addOwnerReference(owner object, object *unstructured.Unstructured, scheme *runtime.Scheme) error {
	switch object.GetKind() {
	case "ClusterRole", "ClusterRoleBinding",
		"CustomResourceDefinition",
		"MutatingWebhookConfiguration", "ValidatingWebhookConfiguration":
		// Non-Namespaced objects
		ownerhelpers.SetOwnerReference(owner, object, scheme)
	default:
		if err := controllerutil.SetControllerReference(owner, object, scheme); err != nil {
			return fmt.Errorf("set ownerReference: %w, obj: %s, %s", err, object.GetKind(), object.GetName())
		}
	}
	return nil
}

// cleanupClusterRoles deletes owned ClusterRoles
// cleaned is true when all ClusterRoles have been cleaned up.
func cleanupClusterRoles(ctx context.Context, c client.Client, ownedBy generalizedListOption) (cleaned bool, err error) {
	clusterRoleList := &rbacv1.ClusterRoleList{}
	if err := c.List(ctx, clusterRoleList, ownedBy); err != nil {
		return false, fmt.Errorf("listing ClusterRoles: %w", err)
	}
	for _, clusterRole := range clusterRoleList.Items {
		if err := c.Delete(ctx, &clusterRole); err != nil && !errors.IsNotFound(err) {
			return false, fmt.Errorf("deleting ClusterRole: %w", err)
		}
	}
	return len(clusterRoleList.Items) == 0, nil
}

// cleanupClusterRoleBindings deletes owned ClusterRoleBindings
// cleaned is true when all ClusterRoleBindings have been cleaned up.
func cleanupClusterRoleBindings(ctx context.Context, c client.Client, ownedBy generalizedListOption) (cleaned bool, err error) {
	clusterRoleBindingList := &rbacv1.ClusterRoleBindingList{}
	if err := c.List(ctx, clusterRoleBindingList, ownedBy); err != nil {
		return false, fmt.Errorf("listing ClusterRoleBindings: %w", err)
	}
	for _, clusterRoleBinding := range clusterRoleBindingList.Items {
		if err := c.Delete(ctx, &clusterRoleBinding); err != nil && !errors.IsNotFound(err) {
			return false, fmt.Errorf("deleting ClusterRoleBinding: %w", err)
		}
	}
	return len(clusterRoleBindingList.Items) == 0, nil
}

// cleanupCustomResourceDefinitions deletes owned CustomResourceDefinitions
// cleaned is true when all CustomResourceDefinitions have been cleaned up.
func cleanupCustomResourceDefinitions(ctx context.Context, c client.Client, ownedBy generalizedListOption) (cleaned bool, err error) {
	customResourceDefinitionList := &apiextensionsv1.CustomResourceDefinitionList{}
	if err := c.List(ctx, customResourceDefinitionList, ownedBy); err != nil {
		return false, fmt.Errorf("listing CustomResourceDefinitions: %w", err)
	}
	for _, customResourceDefinition := range customResourceDefinitionList.Items {
		if err := c.Delete(ctx, &customResourceDefinition); err != nil && !errors.IsNotFound(err) {
			return false, fmt.Errorf("deleting CustomResourceDefinition: %w", err)
		}
	}
	return len(customResourceDefinitionList.Items) == 0, nil
}

// cleanupMutatingWebhookConfigurations deletes owned MutatingWebhookConfigurations
// cleaned is true when all MutatingWebhookConfigurations have been cleaned up.
func cleanupMutatingWebhookConfigurations(ctx context.Context, c client.Client, ownedBy generalizedListOption) (cleaned bool, err error) {
	mutatingWebhookConfigurationList := &adminv1beta1.MutatingWebhookConfigurationList{}
	if err := c.List(ctx, mutatingWebhookConfigurationList, ownedBy); err != nil {
		return false, fmt.Errorf("listing MutatingWebhookConfigurations: %w", err)
	}
	for _, mutatingWebhookConfiguration := range mutatingWebhookConfigurationList.Items {
		if err := c.Delete(ctx, &mutatingWebhookConfiguration); err != nil && !errors.IsNotFound(err) {
			return false, fmt.Errorf("deleting MutatingWebhookConfiguration: %w", err)
		}
	}
	return len(mutatingWebhookConfigurationList.Items) == 0, nil
}

// cleanupValidatingWebhookConfigurations deletes owned ValidatingWebhookConfigurations
// cleaned is true when all ValidatingWebhookConfigurations have been cleaned up.
func cleanupValidatingWebhookConfigurations(ctx context.Context, c client.Client, ownedBy generalizedListOption) (cleaned bool, err error) {
	validatingWebhookConfigurationList := &adminv1beta1.ValidatingWebhookConfigurationList{}
	if err := c.List(ctx, validatingWebhookConfigurationList, ownedBy); err != nil {
		return false, fmt.Errorf("listing ValidatingWebhookConfigurations: %w", err)
	}
	for _, validatingWebhookConfiguration := range validatingWebhookConfigurationList.Items {
		if err := c.Delete(ctx, &validatingWebhookConfiguration); err != nil && !errors.IsNotFound(err) {
			return false, fmt.Errorf("deleting ValidatingWebhookConfiguration: %w", err)
		}
	}
	return len(validatingWebhookConfigurationList.Items) == 0, nil
}

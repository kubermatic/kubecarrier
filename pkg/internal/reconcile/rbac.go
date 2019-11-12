/*
Copyright 2019 The Kubecarrier Authors.

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
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Role reconciles a rbac.authorization.k8s.io/v1, Kind=Role.
func Role(
	ctx context.Context,
	log logr.Logger,
	c client.Client,
	desiredRole *rbacv1.Role,
) (currentRole *rbacv1.Role, err error) {
	name := types.NamespacedName{
		Name:      desiredRole.Name,
		Namespace: desiredRole.Namespace,
	}

	// Lookup current version of the object
	currentRole = &rbacv1.Role{}
	err = c.Get(ctx, name, currentRole)
	if err != nil && !errors.IsNotFound(err) {
		// unexpected error
		return nil, fmt.Errorf("getting Role: %w", err)
	}

	if errors.IsNotFound(err) {
		// Role needs to be created
		log.V(1).Info("creating", "Role", name.String())
		if err = c.Create(ctx, desiredRole); err != nil {
			return nil, fmt.Errorf("creating Role: %w", err)
		}
		// no need to check for updates, object was created just now
		return desiredRole, nil
	}

	if !equality.Semantic.DeepEqual(desiredRole.Rules, currentRole.Rules) {
		// desired and current Role .Roles are not equal -> trigger an update
		log.V(1).Info("updating", "Role", name.String())
		currentRole.Rules = desiredRole.Rules
		if err = c.Update(ctx, currentRole); err != nil {
			return nil, fmt.Errorf("updating Role: %w", err)
		}
	}

	return currentRole, nil
}

// RoleBinding reconciles a rbac.authorization.k8s.io/v1, Kind=RoleBinding.
func RoleBinding(
	ctx context.Context,
	log logr.Logger,
	c client.Client,
	desiredRoleBinding *rbacv1.RoleBinding,
) (currentRoleBinding *rbacv1.RoleBinding, err error) {
	name := types.NamespacedName{
		Name:      desiredRoleBinding.Name,
		Namespace: desiredRoleBinding.Namespace,
	}

	// Lookup current version of the object
	currentRoleBinding = &rbacv1.RoleBinding{}
	err = c.Get(ctx, name, currentRoleBinding)
	if err != nil && !errors.IsNotFound(err) {
		// unexpected error
		return nil, fmt.Errorf("getting RoleBinding: %w", err)
	}

	if errors.IsNotFound(err) {
		// RoleBinding needs to be created
		log.V(1).Info("creating", "RoleBinding", name.String())
		if err = c.Create(ctx, desiredRoleBinding); err != nil {
			return nil, fmt.Errorf("creating RoleBinding: %w", err)
		}
		// no need to check for updates, object was created just now
		return desiredRoleBinding, nil
	}

	return currentRoleBinding, nil
}

// ClusterRole reconciles a rbac.authorization.k8s.io/v1, Kind=ClusterRole.
func ClusterRole(
	ctx context.Context,
	log logr.Logger,
	c client.Client,
	desiredClusterRole *rbacv1.ClusterRole,
) (currentClusterRole *rbacv1.ClusterRole, err error) {
	name := types.NamespacedName{
		Name:      desiredClusterRole.Name,
		Namespace: desiredClusterRole.Namespace,
	}

	// Lookup current version of the object
	currentClusterRole = &rbacv1.ClusterRole{}
	err = c.Get(ctx, name, currentClusterRole)
	if err != nil && !errors.IsNotFound(err) {
		// unexpected error
		return nil, fmt.Errorf("getting ClusterRole: %w", err)
	}

	if errors.IsNotFound(err) {
		// ClusterRole needs to be created
		log.V(1).Info("creating", "ClusterRole", name.String())
		if err = c.Create(ctx, desiredClusterRole); err != nil {
			return nil, fmt.Errorf("creating ClusterRole: %w", err)
		}
		// no need to check for updates, object was created just now
		return desiredClusterRole, nil
	}

	if !equality.Semantic.DeepEqual(desiredClusterRole.Rules, currentClusterRole.Rules) {
		// desired and current ClusterRole .ClusterRoles are not equal -> trigger an update
		log.V(1).Info("updating", "ClusterRole", name.String())
		currentClusterRole.Rules = desiredClusterRole.Rules
		if err = c.Update(ctx, currentClusterRole); err != nil {
			return nil, fmt.Errorf("updating ClusterRole: %w", err)
		}
	}

	return currentClusterRole, nil
}

// ClusterRoleBinding reconciles a rbac.authorization.k8s.io/v1, Kind=ClusterRoleBinding.
func ClusterRoleBinding(
	ctx context.Context,
	log logr.Logger,
	c client.Client,
	desiredClusterRoleBinding *rbacv1.ClusterRoleBinding,
) (currentClusterRoleBinding *rbacv1.ClusterRoleBinding, err error) {
	name := types.NamespacedName{
		Name:      desiredClusterRoleBinding.Name,
		Namespace: desiredClusterRoleBinding.Namespace,
	}

	// Lookup current version of the object
	currentClusterRoleBinding = &rbacv1.ClusterRoleBinding{}
	err = c.Get(ctx, name, currentClusterRoleBinding)
	if err != nil && !errors.IsNotFound(err) {
		// unexpected error
		return nil, fmt.Errorf("getting ClusterRoleBinding: %w", err)
	}

	if errors.IsNotFound(err) {
		// ClusterRoleBinding needs to be created
		log.V(1).Info("creating", "ClusterRoleBinding", name.String())
		if err = c.Create(ctx, desiredClusterRoleBinding); err != nil {
			return nil, fmt.Errorf("creating ClusterRoleBinding: %w", err)
		}
		// no need to check for updates, object was created just now
		return desiredClusterRoleBinding, nil
	}

	return currentClusterRoleBinding, nil
}

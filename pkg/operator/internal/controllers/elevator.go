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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	resourceselevator "github.com/kubermatic/kubecarrier/pkg/internal/resources/elevator"
)

// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=elevators,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=elevators/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;create;update;patch;delete;escalate;bind
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete;escalate;bind
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete

type ElevatorStrategy struct {
}

func (c *ElevatorStrategy) GetObj() Component {
	return &operatorv1alpha1.Elevator{}
}

func (c *ElevatorStrategy) GetOwnedObjectsTypes() []runtime.Object {
	return []runtime.Object{
		&appsv1.Deployment{},
		&corev1.Service{},
		&rbacv1.Role{},
		&rbacv1.RoleBinding{},
		&corev1.ServiceAccount{},
		&rbacv1.ClusterRole{},
		&rbacv1.ClusterRoleBinding{},
	}
}

func (c *ElevatorStrategy) GetManifests(ctx context.Context, component Component) ([]unstructured.Unstructured, error) {
	elevator, ok := component.(*operatorv1alpha1.Elevator)
	if !ok {
		return nil, fmt.Errorf("can't assert to Elevator: %v", component)
	}
	return resourceselevator.Manifests(
		resourceselevator.Config{
			Name:      elevator.Name,
			Namespace: elevator.Namespace,

			ProviderKind:    elevator.Spec.ProviderCRD.Kind,
			ProviderVersion: elevator.Spec.ProviderCRD.Version,
			ProviderGroup:   elevator.Spec.ProviderCRD.Group,
			ProviderPlural:  elevator.Spec.ProviderCRD.Plural,

			TenantKind:    elevator.Spec.TenantCRD.Kind,
			TenantVersion: elevator.Spec.TenantCRD.Version,
			TenantGroup:   elevator.Spec.TenantCRD.Group,
			TenantPlural:  elevator.Spec.TenantCRD.Plural,

			DerivedCRName: elevator.Spec.DerivedCR.Name,
		})
}

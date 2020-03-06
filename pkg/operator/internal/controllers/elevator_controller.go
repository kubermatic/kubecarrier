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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/source"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/owner"
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

type ElevatorController struct {
	Obj *operatorv1alpha1.Elevator
}

func (c *ElevatorController) GetObj() Component {
	return c.Obj
}

func (c *ElevatorController) GetOwnedObjectsTypes() []runtime.Object {
	return []runtime.Object{
		&rbacv1.ClusterRole{},
		&rbacv1.ClusterRoleBinding{},
		&apiextensionsv1.CustomResourceDefinition{},
	}
}
func (c *ElevatorController) SetupWithManager(builder *builder.Builder, scheme *runtime.Scheme) *builder.Builder {
	enqueuer := owner.EnqueueRequestForOwner(&operatorv1alpha1.Elevator{}, scheme)
	return builder.For(&operatorv1alpha1.Elevator{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Watches(&source.Kind{Type: &corev1.ServiceAccount{}}, enqueuer).
		Watches(&source.Kind{Type: &rbacv1.ClusterRole{}}, enqueuer).
		Watches(&source.Kind{Type: &rbacv1.ClusterRoleBinding{}}, enqueuer)
}

func (c *ElevatorController) GetManifests(ctx context.Context) ([]unstructured.Unstructured, error) {
	return resourceselevator.Manifests(
		resourceselevator.Config{
			Name:      c.Obj.Name,
			Namespace: c.Obj.Namespace,

			ProviderKind:    c.Obj.Spec.ProviderCRD.Kind,
			ProviderVersion: c.Obj.Spec.ProviderCRD.Version,
			ProviderGroup:   c.Obj.Spec.ProviderCRD.Group,
			ProviderPlural:  c.Obj.Spec.ProviderCRD.Plural,

			TenantKind:    c.Obj.Spec.TenantCRD.Kind,
			TenantVersion: c.Obj.Spec.TenantCRD.Version,
			TenantGroup:   c.Obj.Spec.TenantCRD.Group,
			TenantPlural:  c.Obj.Spec.TenantCRD.Plural,

			DerivedCRName: c.Obj.Spec.DerivedCR.Name,
		})
}

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
	resourceferry "github.com/kubermatic/kubecarrier/pkg/internal/resources/ferry"
)

// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete;escalate;bind
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=ferries,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=ferries/status,verbs=get;update;patch

type FerryStrategy struct {
}

func (c *FerryStrategy) GetObj() Component {
	return &operatorv1alpha1.Ferry{}
}

func (c *FerryStrategy) GetOwnedObjectsTypes() []runtime.Object {
	return []runtime.Object{
		&corev1.Service{},
		&corev1.ServiceAccount{},
		&rbacv1.Role{},
		&rbacv1.RoleBinding{},
		&appsv1.Deployment{},
	}

}

func (c *FerryStrategy) GetManifests(ctx context.Context, component Component) ([]unstructured.Unstructured, error) {
	ferry, ok := component.(*operatorv1alpha1.Ferry)
	if !ok {
		return nil, fmt.Errorf("can't assert to Ferry: %v", component)
	}
	return resourceferry.Manifests(
		resourceferry.Config{
			ProviderNamespace:    ferry.Namespace,
			Name:                 ferry.Name,
			KubeconfigSecretName: ferry.Spec.KubeconfigSecret.Name,
		})
}

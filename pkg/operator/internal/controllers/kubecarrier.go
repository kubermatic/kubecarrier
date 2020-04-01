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

	certv1alpha2 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	adminv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/constants"
	"github.com/kubermatic/kubecarrier/pkg/internal/resources/manager"
)

// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=towers,verbs=create
// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=kubecarriers,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=kubecarriers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;create;update;patch;delete;escalate;bind
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete;escalate;bind
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=issuers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete

type KubeCarrierStrategy struct {
}

func (c *KubeCarrierStrategy) GetObj() Component {
	return &operatorv1alpha1.KubeCarrier{}
}

func (c *KubeCarrierStrategy) GetDeletionObjectTypes() []runtime.Object {
	return []runtime.Object{}
}

func (c *KubeCarrierStrategy) GetManifests(ctx context.Context, component Component) ([]unstructured.Unstructured, error) {
	kubeCarrier, ok := component.(*operatorv1alpha1.KubeCarrier)
	if !ok {
		return nil, fmt.Errorf("can't assert to KubeCarrier: %v", component)
	}
	objects, err := manager.Manifests(
		manager.Config{
			Name:      kubeCarrier.Name,
			Namespace: constants.KubeCarrierDefaultNamespace,
		})
	if err != nil {
		return nil, err
	}
	if kubeCarrier.Spec.Master {
		tower := &operatorv1alpha1.Tower{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "operator.kubecarrier.io/v1alpha1",
				Kind:       "Tower",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      kubeCarrier.Name,
				Namespace: kubeCarrier.Namespace,
			},
		}
		towerUnstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(tower)
		if err != nil {
			return nil, fmt.Errorf("convert Tower object to unstructured: %w", err)
		}
		objects = append(objects, unstructured.Unstructured{Object: towerUnstructured})
	}
	return objects, nil
}

func (c *KubeCarrierStrategy) AddWatches(builder *builder.Builder, scheme *runtime.Scheme) *builder.Builder {
	return builder.
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&certv1alpha2.Issuer{}).
		Owns(&certv1alpha2.Certificate{}).
		Owns(&rbacv1.ClusterRole{}).
		Owns(&rbacv1.ClusterRoleBinding{}).
		Owns(&apiextensionsv1.CustomResourceDefinition{}).
		Owns(&adminv1beta1.MutatingWebhookConfiguration{}).
		Owns(&adminv1beta1.ValidatingWebhookConfiguration{})
}

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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/owner"
	resourcescatapult "github.com/kubermatic/kubecarrier/pkg/internal/resources/catapult"
)

// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=catapults,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=catapults/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;create;update;patch;delete;escalate;bind
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete;escalate;bind
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=issuers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete

type CatapultController struct {
	Client client.Client
}

func (c *CatapultController) GetObj() Component {
	return &operatorv1alpha1.Catapult{}
}

func (c *CatapultController) GetOwnedObjectsTypes() []runtime.Object {
	return []runtime.Object{
		&rbacv1.ClusterRole{},
		&rbacv1.ClusterRoleBinding{},
		&apiextensionsv1.CustomResourceDefinition{},
		&adminv1beta1.MutatingWebhookConfiguration{},
		&adminv1beta1.ValidatingWebhookConfiguration{},
	}
}

func (c *CatapultController) SetupWithManager(builder *builder.Builder, scheme *runtime.Scheme) *builder.Builder {
	enqueuer := owner.EnqueueRequestForOwner(&operatorv1alpha1.Catapult{}, scheme)
	return builder.For(&operatorv1alpha1.Catapult{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&certv1alpha2.Issuer{}).
		Owns(&certv1alpha2.Certificate{}).
		Watches(&source.Kind{Type: &rbacv1.ClusterRole{}}, enqueuer).
		Watches(&source.Kind{Type: &rbacv1.ClusterRoleBinding{}}, enqueuer).
		Watches(&source.Kind{Type: &adminv1beta1.MutatingWebhookConfiguration{}}, enqueuer).
		Watches(&source.Kind{Type: &adminv1beta1.ValidatingWebhookConfiguration{}}, enqueuer)
}

func (c *CatapultController) GetManifests(ctx context.Context, component Component) ([]unstructured.Unstructured, error) {

	catapult, ok := component.(*operatorv1alpha1.Catapult)
	if !ok {
		return nil, fmt.Errorf("can't assert to Catapult: %v", component)
	}

	// Lookup Ferry to get name of secret.
	ferry := &operatorv1alpha1.Ferry{}
	if err := c.Client.Get(ctx, types.NamespacedName{
		Name:      catapult.Spec.ServiceCluster.Name,
		Namespace: catapult.Namespace,
	}, ferry); err != nil {
		return nil, fmt.Errorf("getting Ferry: %w", err)
	}
	return resourcescatapult.Manifests(
		resourcescatapult.Config{
			Name:      catapult.Name,
			Namespace: catapult.Namespace,

			ManagementClusterKind:    catapult.Spec.ManagementClusterCRD.Kind,
			ManagementClusterVersion: catapult.Spec.ManagementClusterCRD.Version,
			ManagementClusterGroup:   catapult.Spec.ManagementClusterCRD.Group,
			ManagementClusterPlural:  catapult.Spec.ManagementClusterCRD.Plural,

			ServiceClusterKind:    catapult.Spec.ServiceClusterCRD.Kind,
			ServiceClusterVersion: catapult.Spec.ServiceClusterCRD.Version,
			ServiceClusterGroup:   catapult.Spec.ServiceClusterCRD.Group,
			ServiceClusterPlural:  catapult.Spec.ServiceClusterCRD.Plural,

			ServiceClusterName:   catapult.Spec.ServiceCluster.Name,
			ServiceClusterSecret: ferry.Spec.KubeconfigSecret.Name,
			WebhookStrategy:      string(catapult.Spec.WebhookStrategy),
		})
}

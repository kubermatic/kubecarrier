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

	"github.com/go-logr/logr"
	certv1alpha2 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	adminv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/owner"
	resourcescatapult "github.com/kubermatic/kubecarrier/pkg/internal/resources/catapult"
)

type catapultController struct {
	Obj    *operatorv1alpha1.Catapult
	Client client.Client
}

func (c *catapultController) GetObj() Component {
	return c.Obj
}

func (c *catapultController) GetManifests(ctx context.Context) ([]unstructured.Unstructured, error) {

	// Lookup Ferry to get name of secret.
	ferry := &operatorv1alpha1.Ferry{}
	if err := c.Client.Get(ctx, types.NamespacedName{
		Name:      c.Obj.Spec.ServiceCluster.Name,
		Namespace: c.Obj.Namespace,
	}, ferry); err != nil {
		return nil, fmt.Errorf("getting Ferry: %w", err)
	}
	return resourcescatapult.Manifests(
		resourcescatapult.Config{
			Name:      c.Obj.Name,
			Namespace: c.Obj.Namespace,

			ManagementClusterKind:    c.Obj.Spec.ManagementClusterCRD.Kind,
			ManagementClusterVersion: c.Obj.Spec.ManagementClusterCRD.Version,
			ManagementClusterGroup:   c.Obj.Spec.ManagementClusterCRD.Group,
			ManagementClusterPlural:  c.Obj.Spec.ManagementClusterCRD.Plural,

			ServiceClusterKind:    c.Obj.Spec.ServiceClusterCRD.Kind,
			ServiceClusterVersion: c.Obj.Spec.ServiceClusterCRD.Version,
			ServiceClusterGroup:   c.Obj.Spec.ServiceClusterCRD.Group,
			ServiceClusterPlural:  c.Obj.Spec.ServiceClusterCRD.Plural,

			ServiceClusterName:   c.Obj.Spec.ServiceCluster.Name,
			ServiceClusterSecret: ferry.Spec.KubeconfigSecret.Name,
			WebhookStrategy:      string(c.Obj.Spec.WebhookStrategy),
		})
}

// CatapultReconciler reconciles a Catapult object
type CatapultReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=catapults,verbs=get;list;watch;create;update;patch;delete
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

// Reconcile function reconciles the Catapult object which specified by the request. Currently, it does the following:
// 1. Fetch the Catapult object.
// 2. Handle the deletion of the Catapult object (Remove the objects that the Catapult owns, and remove the finalizer).
// 3. Reconcile the objects that owned by Catapult object.
// 4. Update the status of the Catapult object.
func (r *CatapultReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("catapult", req.NamespacedName)

	br := BaseReconciler{
		Client:    r.Client,
		Log:       log,
		Scheme:    r.Scheme,
		Finalizer: "catapult.kubecarrier.io/controller",
		Name:      "Catapult",
	}

	catapult := &operatorv1alpha1.Catapult{}
	ctr := &catapultController{Obj: catapult, Client: r.Client}
	return br.Reconcile(ctx, req, ctr)

}

func (r *CatapultReconciler) SetupWithManager(mgr ctrl.Manager) error {
	enqueuer := owner.EnqueueRequestForOwner(&operatorv1alpha1.Catapult{}, r.Scheme)

	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.Catapult{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&certv1alpha2.Issuer{}).
		Owns(&certv1alpha2.Certificate{}).
		Watches(&source.Kind{Type: &rbacv1.ClusterRole{}}, enqueuer).
		Watches(&source.Kind{Type: &rbacv1.ClusterRoleBinding{}}, enqueuer).
		Watches(&source.Kind{Type: &adminv1beta1.MutatingWebhookConfiguration{}}, enqueuer).
		Watches(&source.Kind{Type: &adminv1beta1.ValidatingWebhookConfiguration{}}, enqueuer).
		Complete(r)
}

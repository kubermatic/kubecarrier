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

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	resourceferry "github.com/kubermatic/kubecarrier/pkg/internal/resources/ferry"
)

type ferryController struct {
	Obj *operatorv1alpha1.Ferry
}

func (c *ferryController) GetObj() Component {
	return c.Obj
}

func (c *ferryController) GetManifests(ctx context.Context) ([]unstructured.Unstructured, error) {
	return resourceferry.Manifests(
		resourceferry.Config{
			ProviderNamespace:    c.Obj.Namespace,
			Name:                 c.Obj.Name,
			KubeconfigSecretName: c.Obj.Spec.KubeconfigSecret.Name,
		})
}

// FerryReconciler reconciles a Ferry object
type FerryReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete;escalate;bind
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=ferries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=ferries/status,verbs=get;update;patch

// Reconcile function reconciles the Ferry object which specified by the request. Currently, it does the following:
// * fetches the Ferry object
// * handles object deletion if neccessary
// * create all necessary objects from resource
// * Updates the ferry status
func (r *FerryReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("ferry", req.NamespacedName)

	br := BaseReconciler{
		Client:    r.Client,
		Log:       log,
		Scheme:    r.Scheme,
		Finalizer: "ferry.kubecarrier.io/controller",
		Name:      "Ferry",
	}

	ferry := &operatorv1alpha1.Ferry{}
	ctr := &ferryController{Obj: ferry}

	return br.Reconcile(ctx, req, ctr)
}

var ferryControllerObjects = []runtime.Object{
	&corev1.Service{},
	&corev1.ServiceAccount{},
	&rbacv1.Role{},
	&rbacv1.RoleBinding{},
	&appsv1.Deployment{},
}

func (r *FerryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	cm := ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.Ferry{})

	for _, obj := range ferryControllerObjects {
		cm = cm.Watches(&source.Kind{Type: obj}, &handler.EnqueueRequestForOwner{OwnerType: &operatorv1alpha1.Ferry{}})
	}

	return cm.Complete(r)
}

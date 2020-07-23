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
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/kubermatic/utils/pkg/owner"
	"github.com/kubermatic/utils/pkg/util"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	resourcescatapult "github.com/kubermatic/kubecarrier/pkg/internal/resources/catapult"
)

const (
	catapultControllerFinalizer = "catapult.kubecarrier.io/controller"
)

// CatapultReconciler reconciles a Catapult object
type CatapultReconciler struct {
	client.Client
	Log        logr.Logger
	Scheme     *runtime.Scheme
	RESTMapper meta.RESTMapper
}

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

func (r *CatapultReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("catapult", req.NamespacedName)

	catapult := &operatorv1alpha1.Catapult{}
	if err := r.Get(ctx, req.NamespacedName, catapult); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if catapult.Spec.Paused.IsPaused() {
		if catapult.SetPausedCondition() {
			if err := r.Client.Status().Update(ctx, catapult); err != nil {
				return ctrl.Result{}, fmt.Errorf("updating %s status: %w", catapult.Name, err)
			}
		}
		// reconciliation paused, skip all other handlers
		return ctrl.Result{}, nil
	}

	if !catapult.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, catapult); err != nil {
			return ctrl.Result{}, fmt.Errorf("handle deletion: %w", err)
		}
		return ctrl.Result{}, nil
	}

	if util.AddFinalizer(catapult, catapultControllerFinalizer) {
		if err := r.Update(ctx, catapult); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating Catapult finalizers: %w", err)
		}
	}

	// Lookup Ferry to get name of secret.
	ferry := &operatorv1alpha1.Ferry{}
	if err := r.Client.Get(ctx, types.NamespacedName{
		Name:      catapult.Spec.ServiceCluster.Name,
		Namespace: catapult.Namespace,
	}, ferry); err != nil {
		return ctrl.Result{}, fmt.Errorf("getting Ferry: %w", err)
	}

	objects, err := resourcescatapult.Manifests(
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
			LogLevel:             catapult.Spec.LogLevel,
		})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("creating Catapult manifests: %w", err)
	}

	deploymentIsReady, err := reconcileOwnedObjectsForNamespacedOwner(ctx, log, r.Scheme, r.RESTMapper, r.Client, catapult, objects)
	if err != nil {
		return ctrl.Result{}, err
	}

	// 4. Update the status of the Catapult object.
	if err := r.updateStatus(ctx, catapult, deploymentIsReady); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *CatapultReconciler) SetupWithManager(mgr ctrl.Manager) error {
	enqueuer := owner.EnqueueRequestForOwner(&operatorv1alpha1.Catapult{}, r.Scheme)

	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.Catapult{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&certv1alpha2.Issuer{}).
		Owns(&certv1alpha2.Certificate{}).
		Watches(&source.Kind{Type: &rbacv1.ClusterRole{}}, enqueuer).
		Watches(&source.Kind{Type: &rbacv1.ClusterRoleBinding{}}, enqueuer).
		Watches(&source.Kind{Type: &adminv1beta1.MutatingWebhookConfiguration{}}, enqueuer).
		Complete(r)
}

func (r *CatapultReconciler) handleDeletion(ctx context.Context, catapult *operatorv1alpha1.Catapult) error {
	if catapult.SetTerminatingCondition() {
		if err := r.Client.Status().Update(ctx, catapult); err != nil {
			return fmt.Errorf("updating %s status: %w", catapult.Name, err)
		}
	}

	cleanedUp, err := util.DeleteObjects(ctx, r.Client, r.Scheme, []runtime.Object{
		&rbacv1.ClusterRole{},
		&rbacv1.ClusterRoleBinding{},
		&adminv1beta1.MutatingWebhookConfiguration{},
	}, owner.OwnedBy(catapult, r.Scheme))
	if err != nil {
		return fmt.Errorf("DeleteObjects: %w", err)
	}

	if cleanedUp && util.RemoveFinalizer(catapult, catapultControllerFinalizer) {
		if err := r.Update(ctx, catapult); err != nil {
			return fmt.Errorf("updating Catapult finalizers: %w", err)
		}
	}
	return nil
}

func (r *CatapultReconciler) updateStatus(ctx context.Context, catapult *operatorv1alpha1.Catapult, deploymentIsReady bool) error {
	var pausedChanged, readyChanged bool

	if !catapult.Spec.Paused.IsPaused() {
		pausedChanged = catapult.SetUnPausedCondition()
	}

	if deploymentIsReady {
		readyChanged = catapult.SetReadyCondition()
	} else {
		readyChanged = catapult.SetUnReadyCondition()
	}

	if readyChanged || pausedChanged {
		if err := r.Client.Status().Update(ctx, catapult); err != nil {
			return fmt.Errorf("updating %s status: %w", catapult.Name, err)
		}
	}
	return nil
}

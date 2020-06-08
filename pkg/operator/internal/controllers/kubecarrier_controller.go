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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/constants"
	"github.com/kubermatic/kubecarrier/pkg/internal/reconcile"
	"github.com/kubermatic/kubecarrier/pkg/internal/resources/manager"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

// KubeCarrierReconciler reconciles a KubeCarrier object
type KubeCarrierReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=operator.kubecarrier.io,resources=apiservers,verbs=get;list;watch;create;update;patch
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
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=issuers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete

func (r *KubeCarrierReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("kubecarrier", req.NamespacedName)

	kubeCarrier := &operatorv1alpha1.KubeCarrier{}
	if err := r.Get(ctx, req.NamespacedName, kubeCarrier); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !kubeCarrier.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, kubeCarrier); err != nil {
			return ctrl.Result{}, fmt.Errorf("handle deletion: %w", err)
		}
		return ctrl.Result{}, nil
	}

	objects, err := manager.Manifests(
		manager.Config{
			Name:      kubeCarrier.Name,
			Namespace: constants.KubeCarrierDefaultNamespace,
		})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("creating manager manifests: %w", err)
	}

	var deploymentIsReady bool
	for _, object := range objects {
		if err := controllerutil.SetControllerReference(kubeCarrier, &object, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		curObj, err := reconcile.Unstructured(ctx, log, r.Client, &object)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("reconcile kind: %s, err: %w", object.GroupVersionKind().Kind, err)
		}
		switch ctr := curObj.(type) {
		case *appsv1.Deployment:
			deploymentIsReady = util.DeploymentIsAvailable(ctr)
		}

	}

	if !deploymentIsReady {
		if err := r.updateStatus(ctx, kubeCarrier, operatorv1alpha1.KubeCarrierCondition{
			Type:    operatorv1alpha1.KubeCarrierDeploymentReady,
			Status:  operatorv1alpha1.ConditionFalse,
			Reason:  "DeploymentUnReady",
			Message: "the deployment of KubeCarrier is not ready",
		}); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	if err := r.updateStatus(ctx, kubeCarrier, operatorv1alpha1.KubeCarrierCondition{
		Type:    operatorv1alpha1.KubeCarrierDeploymentReady,
		Status:  operatorv1alpha1.ConditionTrue,
		Reason:  "DeploymentReady",
		Message: "the deployment of KubeCarrier is ready",
	}); err != nil {
		return ctrl.Result{}, err
	}

	apiServerIsReady, err := r.reconcileAPIServer(ctx, kubeCarrier)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("reconcile API Server: %w", err)
	}
	if !apiServerIsReady {
		if err := r.updateStatus(ctx, kubeCarrier, operatorv1alpha1.KubeCarrierCondition{
			Type:    operatorv1alpha1.KubeCarrierAPIServerReady,
			Status:  operatorv1alpha1.ConditionFalse,
			Reason:  "APIServerUnReady",
			Message: "APIServer is not ready",
		}); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	if err := r.updateStatus(ctx, kubeCarrier, operatorv1alpha1.KubeCarrierCondition{
		Type:    operatorv1alpha1.KubeCarrierAPIServerReady,
		Status:  operatorv1alpha1.ConditionTrue,
		Reason:  "APIServerReady",
		Message: "APIServer is ready",
	}); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *KubeCarrierReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.KubeCarrier{}).
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
		Owns(&adminv1beta1.ValidatingWebhookConfiguration{}).
		Owns(&operatorv1alpha1.APIServer{}).
		Complete(r)
}

func (r *KubeCarrierReconciler) handleDeletion(ctx context.Context, kubeCarrier *operatorv1alpha1.KubeCarrier) error {
	// Update the object Status to Terminating.
	if kubeCarrier.SetTerminatingCondition() {
		if err := r.Client.Status().Update(ctx, kubeCarrier); err != nil {
			return fmt.Errorf("updating %s status: %w", kubeCarrier.Name, err)
		}
	}
	return nil
}

// updateStatus - update the status of the object
func (r *KubeCarrierReconciler) updateStatus(ctx context.Context, kubeCarrier *operatorv1alpha1.KubeCarrier, condition operatorv1alpha1.KubeCarrierCondition) error {
	kubeCarrier.Status.ObservedGeneration = kubeCarrier.Generation
	kubeCarrier.Status.SetCondition(condition)
	kubeCarrierDeploymentReady, _ := kubeCarrier.Status.GetCondition(operatorv1alpha1.KubeCarrierDeploymentReady)
	apiServerReady, _ := kubeCarrier.Status.GetCondition(operatorv1alpha1.KubeCarrierAPIServerReady)
	if kubeCarrierDeploymentReady.True() && apiServerReady.True() {
		kubeCarrier.Status.SetCondition(operatorv1alpha1.KubeCarrierCondition{
			Type:    operatorv1alpha1.KubeCarrierReady,
			Status:  operatorv1alpha1.ConditionTrue,
			Reason:  "KubeCarrierReady",
			Message: "the deployment of KubeCarrier and API Server are ready.",
		})
	} else if !kubeCarrierDeploymentReady.True() {
		kubeCarrier.Status.SetCondition(operatorv1alpha1.KubeCarrierCondition{
			Type:    operatorv1alpha1.KubeCarrierReady,
			Status:  operatorv1alpha1.ConditionFalse,
			Reason:  "KubeCarrierDeploymentUnReady",
			Message: "the deployment of KubeCarrier is not ready.",
		})
	} else if !apiServerReady.True() {
		kubeCarrier.Status.SetCondition(operatorv1alpha1.KubeCarrierCondition{
			Type:    operatorv1alpha1.KubeCarrierReady,
			Status:  operatorv1alpha1.ConditionFalse,
			Reason:  "APIServerUnReady",
			Message: "API Server is not ready.",
		})
	}

	if err := r.Client.Status().Update(ctx, kubeCarrier); err != nil {
		return fmt.Errorf("updating %s status: %w", kubeCarrier.Name, err)
	}
	return nil
}

func (r *KubeCarrierReconciler) reconcileAPIServer(ctx context.Context, kubeCarrier *operatorv1alpha1.KubeCarrier) (apiServerIsReady bool, err error) {
	desiredAPIServer := &operatorv1alpha1.APIServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeCarrier.Name,
			Namespace: constants.KubeCarrierDefaultNamespace,
		},
		Spec: operatorv1alpha1.APIServerSpec{
			API: kubeCarrier.Spec.API,
		},
	}

	if err := controllerutil.SetControllerReference(kubeCarrier, desiredAPIServer, r.Scheme); err != nil {
		return false, fmt.Errorf("set controller reference for APIServer object: %w", err)
	}

	currentAPIServer := &operatorv1alpha1.APIServer{}
	err = r.Get(ctx, types.NamespacedName{
		Name:      desiredAPIServer.Name,
		Namespace: desiredAPIServer.Namespace,
	}, currentAPIServer)
	if err != nil && !errors.IsNotFound(err) {
		return false, fmt.Errorf("getting APIServer: %w", err)
	}
	if errors.IsNotFound(err) {
		if err = r.Create(ctx, desiredAPIServer); err != nil {
			return false, fmt.Errorf("creating APIServer: %w", err)
		}
		return false, nil
	}
	// Update APIServer
	currentAPIServer.Spec = desiredAPIServer.Spec
	if err = r.Update(ctx, currentAPIServer); err != nil {
		return false, fmt.Errorf("updating APIServer: %w", err)
	}
	return currentAPIServer.IsReady(), nil
}

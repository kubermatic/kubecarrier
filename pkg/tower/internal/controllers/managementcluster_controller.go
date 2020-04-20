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
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	masterv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/master/v1alpha1"
	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/constants"
)

var (
	managementClusterScheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(operatorv1alpha1.AddToScheme(managementClusterScheme))
}

// ManagementClusterReconciler reconciles a ManagementCluster object
type ManagementClusterReconciler struct {
	client.Client
	Log                                logr.Logger
	Scheme                             *runtime.Scheme
	ManagementClusterHealthCheckPeriod time.Duration

	managementClusterClients map[string]client.Client
}

// +kubebuilder:rbac:groups=master.kubecarrier.io,resources=managementclusters,verbs=get;list;watch;update;
// +kubebuilder:rbac:groups=master.kubecarrier.io,resources=managementclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

func (r ManagementClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	managementCluster := &masterv1alpha1.ManagementCluster{}
	if err := r.Get(ctx, req.NamespacedName, managementCluster); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !managementCluster.DeletionTimestamp.IsZero() {
		if err := r.handleDeletion(ctx, managementCluster); err != nil {
			return ctrl.Result{}, fmt.Errorf("handle deletion: %w", err)
		}
		return ctrl.Result{}, nil
	}

	if managementCluster.Name == constants.LocalManagementClusterName {
		if err := r.updateStatus(ctx, managementCluster, masterv1alpha1.ManagementClusterCondition{
			Type:    masterv1alpha1.ManagementClusterReady,
			Status:  masterv1alpha1.ConditionTrue,
			Reason:  "MasterManagementClusterIsReady",
			Message: "Master ManagementCluster is ready",
		}); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: r.ManagementClusterHealthCheckPeriod}, nil
	}

	kubeconfigSecret, err := r.getKubeconfigSecret(ctx, managementCluster)
	if err != nil {
		return ctrl.Result{}, r.updateStatus(ctx, managementCluster, masterv1alpha1.ManagementClusterCondition{
			Type:    masterv1alpha1.ManagementClusterReady,
			Status:  masterv1alpha1.ConditionFalse,
			Reason:  "KubeconfigIsNotReady",
			Message: fmt.Sprintf("Master ManagementCluster kubeconfig is not ready: %s\n", err.Error()),
		})
	}

	if r.managementClusterClients == nil {
		r.managementClusterClients = make(map[string]client.Client)
	}
	if _, present := r.managementClusterClients[managementCluster.Name]; !present {
		cli, err := r.getManagementClusterClient(kubeconfigSecret)
		if err != nil {
			return ctrl.Result{}, r.updateStatus(ctx, managementCluster, masterv1alpha1.ManagementClusterCondition{
				Type:    masterv1alpha1.ManagementClusterReady,
				Status:  masterv1alpha1.ConditionFalse,
				Reason:  "ManagementClusterClientIsNotReady",
				Message: fmt.Sprintf("Master ManagementCluster client is not ready: %s\n", err.Error()),
			})
		}
		r.managementClusterClients[managementCluster.Name] = cli
	}
	managementClusterClient := r.managementClusterClients[managementCluster.Name]

	kubeCarrier := &operatorv1alpha1.KubeCarrier{}
	if err := managementClusterClient.Get(ctx, types.NamespacedName{
		Name: constants.KubeCarrierDefaultName,
	}, kubeCarrier); err != nil {
		if errors.IsNotFound(err) {
			if err := r.updateStatus(ctx, managementCluster, masterv1alpha1.ManagementClusterCondition{
				Type:    masterv1alpha1.ManagementClusterReady,
				Status:  masterv1alpha1.ConditionFalse,
				Reason:  "KubeCarrierNotFound",
				Message: "KubeCarrier installation is not found on the management cluster",
			}); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, err
	}

	if kubeCarrier.IsReady() {
		if err := r.updateStatus(ctx, managementCluster, masterv1alpha1.ManagementClusterCondition{
			Type:    masterv1alpha1.ManagementClusterReady,
			Status:  masterv1alpha1.ConditionTrue,
			Reason:  "KubeCarrierIsReady",
			Message: "KubeCarrier installation is ready on the management cluster",
		}); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		if err := r.updateStatus(ctx, managementCluster, masterv1alpha1.ManagementClusterCondition{
			Type:    masterv1alpha1.ManagementClusterReady,
			Status:  masterv1alpha1.ConditionFalse,
			Reason:  "KubeCarrierIsNotReady",
			Message: "KubeCarrier installation is not ready on the management cluster",
		}); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{RequeueAfter: r.ManagementClusterHealthCheckPeriod}, nil
}

func (r *ManagementClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&masterv1alpha1.ManagementCluster{}).
		Complete(r)
}

func (r *ManagementClusterReconciler) handleDeletion(ctx context.Context, managementCluster *masterv1alpha1.ManagementCluster) error {
	if managementCluster.SetTerminatingCondition() {
		if err := r.Client.Status().Update(ctx, managementCluster); err != nil {
			return fmt.Errorf("updating %s status: %w", managementCluster.Name, err)
		}
	}
	delete(r.managementClusterClients, managementCluster.Name)
	return nil
}

func (r ManagementClusterReconciler) updateStatus(
	ctx context.Context,
	managementCluster *masterv1alpha1.ManagementCluster,
	condition masterv1alpha1.ManagementClusterCondition,
) error {
	managementCluster.Status.ObservedGeneration = managementCluster.Generation
	managementCluster.Status.SetCondition(condition)
	if err := r.Status().Update(ctx, managementCluster); err != nil {
		return fmt.Errorf("updating ManagementCluster status: %w", err)
	}
	return nil
}

func (r ManagementClusterReconciler) getKubeconfigSecret(ctx context.Context, managementCluster *masterv1alpha1.ManagementCluster) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      managementCluster.Spec.KubeconfigSecret.Name,
		Namespace: constants.KubeCarrierDefaultNamespace,
	}, secret); err != nil {
		return nil, err
	}
	return secret, nil
}

func (r ManagementClusterReconciler) getManagementClusterClient(kubeconfigSecret *corev1.Secret) (client.Client, error) {
	clientConfig, err := clientcmd.NewClientConfigFromBytes(kubeconfigSecret.Data[constants.KubeconfigSecretKey])
	if err != nil {
		return nil, err
	}

	cliConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	cl, err := client.New(cliConfig, client.Options{
		Scheme: managementClusterScheme,
	})
	if err != nil {
		return nil, err
	}
	return cl, nil
}

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
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
)

const crddsLabel = "crdds.kubecarrier.io/controlled-by"

// CustomResourceDefinitionDiscoverySetReconciler reconciles a CustomResourceDefinitionDiscovery object
type CustomResourceDefinitionDiscoverySetReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=kubecarrier.io,resources=customresourcedefinitiondiscoverysets,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=kubecarrier.io,resources=customresourcedefinitiondiscoverysets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kubecarrier.io,resources=customresourcedefinitiondiscoveries,verbs=get;list;watch;update;create;delete
// +kubebuilder:rbac:groups=kubecarrier.io,resources=customresourcedefinitiondiscoveries/status,verbs=get

func (r *CustomResourceDefinitionDiscoverySetReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var (
		ctx    = context.Background()
		result ctrl.Result
	)

	crdDiscoverySet := &corev1alpha1.CustomResourceDefinitionDiscoverySet{}
	if err := r.Get(ctx, req.NamespacedName, crdDiscoverySet); err != nil {
		return result, client.IgnoreNotFound(err)
	}

	// List ServiceClusters
	serviceClusterSelector, err := metav1.LabelSelectorAsSelector(&crdDiscoverySet.Spec.ServiceClusterSelector)
	if err != nil {
		return result, fmt.Errorf("parsing ServiceCluster selector: %w", err)
	}
	serviceClusterList := &corev1alpha1.ServiceClusterList{}
	if err := r.List(ctx, serviceClusterList,
		client.InNamespace(crdDiscoverySet.Namespace),
		client.MatchingLabelsSelector{Selector: serviceClusterSelector},
	); err != nil {
		return result, fmt.Errorf("listing ServiceClusters: %w", err)
	}

	// Reconcile CRDDs
	var unreadyCRDDNames []string
	existingCRDDNames := map[string]struct{}{}
	for _, serviceCluster := range serviceClusterList.Items {
		currentCRDD, err := r.reconcileCRDD(ctx, &serviceCluster, crdDiscoverySet)
		if err != nil {
			return result, fmt.Errorf(
				"reconciling CustomResourceDefinitionDiscovery for ServiceCluster %s: %w", serviceCluster.Name, err)
		}
		existingCRDDNames[currentCRDD.Name] = struct{}{}

		ready, _ := currentCRDD.Status.GetCondition(corev1alpha1.CustomResourceDefinitionDiscoveryReady)
		if ready.Status != corev1alpha1.ConditionTrue {
			unreadyCRDDNames = append(unreadyCRDDNames, currentCRDD.Name)
		}
	}

	// Cleanup uncontrolled CRDDs
	crddList := &corev1alpha1.CustomResourceDefinitionDiscoveryList{}
	if err := r.List(ctx, crddList, client.MatchingLabels{
		crddsLabel: crdDiscoverySet.Name,
	}); err != nil {
		return result, fmt.Errorf(
			"listing all CustomResourceDefinitionDiscovery for this Set: %w", err)
	}
	for _, crdd := range crddList.Items {
		_, ok := existingCRDDNames[crdd.Name]
		if ok {
			continue
		}

		// delete crdd that should no longer exist
		if err := r.Delete(ctx, &crdd); err != nil {
			return result, fmt.Errorf("deleting CustomResourceDefinitionDiscovery: %w", err)
		}
	}

	// Report status
	crdDiscoverySet.Status.ObservedGeneration = crdDiscoverySet.Generation
	if len(unreadyCRDDNames) > 0 {
		// Unready
		crdDiscoverySet.Status.SetCondition(corev1alpha1.CustomResourceDefinitionDiscoverySetCondition{
			Type:   corev1alpha1.CustomResourceDefinitionDiscoverySetReady,
			Status: corev1alpha1.ConditionFalse,
			Reason: "ComponentsUnready",
			Message: fmt.Sprintf(
				"Some CustomResourceDefinitionDiscovery objects are unready [%s]", strings.Join(unreadyCRDDNames, ",")),
		})
	} else {
		// Ready
		crdDiscoverySet.Status.SetCondition(corev1alpha1.CustomResourceDefinitionDiscoverySetCondition{
			Type:    corev1alpha1.CustomResourceDefinitionDiscoverySetReady,
			Status:  corev1alpha1.ConditionTrue,
			Reason:  "ComponentsReady",
			Message: "All CustomResourceDefinitionDiscovery objects are ready.",
		})
	}

	if err := r.Status().Update(ctx, crdDiscoverySet); err != nil {
		return result, fmt.Errorf("updating Status: %w", err)
	}
	return result, nil
}

func (r *CustomResourceDefinitionDiscoverySetReconciler) reconcileCRDD(
	ctx context.Context, serviceCluster *corev1alpha1.ServiceCluster,
	crdDiscoverySet *corev1alpha1.CustomResourceDefinitionDiscoverySet,
) (*corev1alpha1.CustomResourceDefinitionDiscovery, error) {
	desiredCRDD := &corev1alpha1.CustomResourceDefinitionDiscovery{
		ObjectMeta: metav1.ObjectMeta{
			Name:      crdDiscoverySet.Name + "." + serviceCluster.Name,
			Namespace: crdDiscoverySet.Namespace,
			Labels: map[string]string{
				crddsLabel: crdDiscoverySet.Name,
			},
		},
		Spec: corev1alpha1.CustomResourceDefinitionDiscoverySpec{
			CRD: crdDiscoverySet.Spec.CRD,
			ServiceCluster: corev1alpha1.ObjectReference{
				Name: serviceCluster.Name,
			},
			KindOverride: crdDiscoverySet.Spec.KindOverride,
		},
	}
	err := controllerutil.SetControllerReference(
		crdDiscoverySet, desiredCRDD, r.Scheme)
	if err != nil {
		return nil, fmt.Errorf("set controller reference: %w", err)
	}

	currentCRDD := &corev1alpha1.CustomResourceDefinitionDiscovery{}
	err = r.Get(ctx, types.NamespacedName{
		Name:      desiredCRDD.Name,
		Namespace: desiredCRDD.Namespace,
	}, currentCRDD)
	if err != nil && !errors.IsNotFound(err) {
		return nil, fmt.Errorf("getting CustomResourceDefinitionDiscovery: %w", err)
	}

	if errors.IsNotFound(err) {
		// create CRDD instance
		if err := r.Create(ctx, desiredCRDD); err != nil {
			return nil, fmt.Errorf("creating CustomResourceDefinitionDiscovery: %w", err)
		}
		return desiredCRDD, nil
	}

	// update existing
	currentCRDD.Spec = desiredCRDD.Spec
	if err := r.Update(ctx, currentCRDD); err != nil {
		return nil, fmt.Errorf("updating CustomResourceDefinitionDiscovery: %w", err)
	}
	return currentCRDD, nil
}

func (r *CustomResourceDefinitionDiscoverySetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.CustomResourceDefinitionDiscoverySet{}).
		Owns(&corev1alpha1.CustomResourceDefinitionDiscovery{}).
		Complete(r)
}

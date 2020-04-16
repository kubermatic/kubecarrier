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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	fakev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/fake/v1alpha1"
)

// BackupReconciler reconciles a Backup object
type BackupReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=fake.kubecarrier.io,resources=backups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=fake.kubecarrier.io,resources=backups/status,verbs=get;update;patch

func (r *BackupReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	backup := &fakev1alpha1.Backup{}
	if err := r.Get(ctx, req.NamespacedName, backup); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("cannot fetch database: %w", err)
	}

	if !backup.GetDeletionTimestamp().IsZero() {
		if backup.SetTerminatingCondition() {
			if err := r.Client.Status().Update(ctx, backup); err != nil {
				return ctrl.Result{}, fmt.Errorf("updating %s status: %w", "FakeBackup controller", err)
			}
		}
		return ctrl.Result{}, nil
	}

	currentSnapshot, err := r.reconcileSnapshot(ctx, backup)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("reconciling Snapshot: %w", err)
	}

	if !currentSnapshot.IsReady() {
		if backup.SetUnReadyCondition() {
			if err := r.Status().Update(ctx, backup); err != nil {
				return ctrl.Result{}, fmt.Errorf("cannot update s: %w", err)
			}
		}
		return ctrl.Result{}, nil
	}

	if backup.SetReadyCondition() {
		if err := r.Status().Update(ctx, backup); err != nil {
			return ctrl.Result{}, fmt.Errorf("cannot update s: %w", err)
		}
	}
	return ctrl.Result{}, nil
}

func (r *BackupReconciler) reconcileSnapshot(ctx context.Context, backup *fakev1alpha1.Backup) (*fakev1alpha1.Snapshot, error) {
	desiredSnapshot := &fakev1alpha1.Snapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      backup.Name,
			Namespace: backup.Namespace,
		},
		Spec: fakev1alpha1.SnapshotSpec{
			DBName: backup.Spec.DBName,
		},
	}
	if err := controllerutil.SetControllerReference(
		backup, desiredSnapshot, r.Scheme); err != nil {
		return nil, fmt.Errorf("set controller reference: %w", err)
	}
	currentSnapshot := &fakev1alpha1.Snapshot{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      desiredSnapshot.Name,
		Namespace: desiredSnapshot.Namespace,
	}, currentSnapshot)
	if err != nil && !errors.IsNotFound(err) {
		return nil, fmt.Errorf("getting Snapshot: %w", err)
	}

	if errors.IsNotFound(err) {
		// Create Snapshot
		if err = r.Create(ctx, desiredSnapshot); err != nil {
			return nil, fmt.Errorf("creating Snapshot: %w", err)
		}
		return desiredSnapshot, nil
	}
	// Update Snapshot
	currentSnapshot.Spec = desiredSnapshot.Spec
	if err = r.Update(ctx, currentSnapshot); err != nil {
		return nil, fmt.Errorf("updating Snapshot: %w", err)
	}

	return currentSnapshot, nil
}
func (r *BackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fakev1alpha1.Backup{}).
		Complete(r)
}

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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ServiceClusterAssignmentMasterClusterNamespaceFieldIndex = "sca.kubecarrier.io/master-cluster-namespace"
	ServiceClusterAssignmentServiceClusterFieldIndex         = "sca.kubecarrier.io/service-cluster"
)

func RegisterIndexes(mgr ctrl.Manager) error {
	err := mgr.GetFieldIndexer().IndexField(
		&ServiceClusterAssignment{},
		ServiceClusterAssignmentMasterClusterNamespaceFieldIndex,
		client.IndexerFunc(func(obj runtime.Object) []string {
			sca := obj.(*ServiceClusterAssignment)
			return []string{sca.Spec.MasterClusterNamespace.Name}
		}))
	if err != nil {
		return err
	}

	err = mgr.GetFieldIndexer().IndexField(
		&ServiceClusterAssignment{},
		ServiceClusterAssignmentServiceClusterFieldIndex,
		client.IndexerFunc(func(obj runtime.Object) []string {
			sca := obj.(*ServiceClusterAssignment)
			return []string{sca.Spec.ServiceCluster.Name}
		}))
	if err != nil {
		return err
	}
	return nil
}

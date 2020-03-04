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
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ServiceClusterAssignmentNamespaceFieldIndex = "sca.kubecarrier.io/namespace"
)

// RegisterServiceClusterAssignmentNamespaceFieldIndex adds a field index for ServiceClusterAssignment.Status.ServiceClusterNamespace.Name.
func RegisterServiceClusterAssignmentNamespaceFieldIndex(indexer client.FieldIndexer) error {
	return indexer.IndexField(
		&ServiceClusterAssignment{}, ServiceClusterAssignmentNamespaceFieldIndex,
		client.IndexerFunc(func(obj runtime.Object) []string {
			sca := obj.(*ServiceClusterAssignment)
			return []string{sca.Status.ServiceClusterNamespace.Name}
		}))
}

func GetServiceClusterAssignmentByServiceClusterNamespace(ctx context.Context, c client.Client, serviceClusterNamespace string) (*ServiceClusterAssignment, error) {
	scaList := &ServiceClusterAssignmentList{}
	if err := c.List(ctx, scaList,
		client.MatchingFields{
			ServiceClusterAssignmentNamespaceFieldIndex: serviceClusterNamespace,
		},
	); err != nil {
		return nil, err
	}
	switch len(scaList.Items) {
	case 0:
		// not found
		return nil, fmt.Errorf("serviceclusterassignments.kubecarrier.io with index %q not found", ServiceClusterAssignmentNamespaceFieldIndex)
	case 1:
		// found!
		return &scaList.Items[0], nil
	default:
		// found too many
		return nil, fmt.Errorf("multiple serviceclusterassignments.kubecarrier.io with index %q found", ServiceClusterAssignmentNamespaceFieldIndex)
	}
}

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
)

// ServiceClusterReferenceSpec defines the desired state of ServiceClusterReference
type ServiceClusterReferenceSpec struct {
	// Metadata contains the metadata (display name, description, etc) of the ServiceCluster.
	Metadata corev1alpha1.ServiceClusterMetadata `json:"metadata,omitempty"`

	// Provider references the Provider that this ServiceCluster belongs to.
	Provider ObjectReference `json:"provider"`
}

// ServiceClusterReference exposes information about a Providers Cluster.
//
// ServiceClusterReference objects are created automatically by KubeCarrier in Account namespaces, that have a service offered to them via a Catalog.
// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Provider",type="string",JSONPath=".spec.provider.name"
// +kubebuilder:printcolumn:name="Display Name",type="string",JSONPath=".spec.metadata.displayName"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:categories=kubecarrier-tenant,shortName=scr
type ServiceClusterReference struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ServiceClusterReferenceSpec `json:"spec,omitempty"`
}

// ServiceClusterReferenceList contains a list of ServiceClusterReference.
// +kubebuilder:object:root=true
type ServiceClusterReferenceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceClusterReference `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceClusterReference{}, &ServiceClusterReferenceList{})
}

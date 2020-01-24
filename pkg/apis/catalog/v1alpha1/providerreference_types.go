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
)

// ProviderReferenceSpec defines the desired state of ProviderReference
type ProviderReferenceSpec struct {
	// Metadata contains the metadata (display name, description, etc) of the Provider.
	Metadata ProviderMetadata `json:"metadata,omitempty"`
}

// ProviderReference exposes information of the Provider(displayName, description).
// This object lives in the tenant namespace for each provider the tenant is allowed utilizing (e.g. there's catalog
// selecting this tenant as its user)
// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Display Name",type="string",JSONPath=".spec.metadata.displayName"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:shortName=pr
type ProviderReference struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ProviderReferenceSpec `json:"spec,omitempty"`
}

// ProviderReferenceList contains a list of ProviderReference
// +kubebuilder:object:root=true
type ProviderReferenceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProviderReference `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProviderReference{}, &ProviderReferenceList{})
}

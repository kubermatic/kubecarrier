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

// OfferingData defines the data (metadata, provider, crds, etc.) of Offering.
type OfferingData struct {
	Metadata OfferingMetadata `json:"metadata,omitempty"`
	// Provider references a ProviderReference of this Offering.
	Provider ObjectReference `json:"provider"`
	// CRDs holds the information about the underlying CRDs that are offered by this offering.
	CRDs []CRDInformation `json:"crds,omitempty"`
}

// OfferingMetadata contains the metadata (display name, description, etc) of the Offering.
type OfferingMetadata struct {
	// DisplayName shows the human-readable name of this Offering.
	DisplayName string `json:"displayName,omitempty"`
	// Description shows the human-readable description of this Offering.
	Description string `json:"description,omitempty"`
}

// Offering is used for Tenants to discover services that have been made available to them.
// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Display Name",type="string",JSONPath=".offering.displayName"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Offering struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Offering OfferingData `json:"offering,omitempty"`
}

// OfferingList contains a list of Offering.
// +kubebuilder:object:root=true
type OfferingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Offering `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Offering{}, &OfferingList{})
}

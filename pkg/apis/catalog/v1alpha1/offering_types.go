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
	// Provider references the Provider managing this Offering.
	Provider ObjectReference `json:"provider"`
	// CRD holds the information about the underlying CRD that is offered by this offering.
	CRD CRDInformation `json:"crd,omitempty"`
}

// OfferingMetadata contains the metadata (display name, description, etc) of the Offering.
type OfferingMetadata struct {
	// DisplayName shows the human-readable name of this Offering.
	// +kubebuilder:validation:MinLength=1
	DisplayName string `json:"displayName,omitempty"`
	// Description shows the human-readable description of this Offering.
	// +kubebuilder:validation:MinLength=1
	Description string `json:"description,omitempty"`
}

// Offering is used for Tenants to discover services that have been made available to them.
//
// Offering objects are created automatically by KubeCarrier in Account namespaces, that have a service offered to them via a Catalog.
// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Display Name",type="string",JSONPath=".offering.metadata.displayName"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:categories=kubecarrier-tenant,shortName=off
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

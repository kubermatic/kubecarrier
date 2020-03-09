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

// CatalogEntrySpec describes the desired state of CatalogEntry.
type CatalogEntrySpec struct {
	// Metadata contains the metadata of the CatalogEntry for the Service Catalog.
	Metadata CatalogEntryMetadata `json:"metadata,omitempty"`
	// BaseCRD is the underlying BaseCRD objects that this CatalogEntry refers to.
	BaseCRD ObjectReference `json:"baseCRD,omitempty"`
	// Derive contains the configuration to generate DerivedCustomResource from the BaseCRD of this CatalogEntry.
	Derive *DerivedConfig `json:"derive,omitempty"`
}

// DerivedConfig can be used to limit fields that should be exposed to a Tenant.
type DerivedConfig struct {
	// overrides the kind of the derived CRD.
	KindOverride string `json:"kindOverride,omitempty"`
	// controls which fields will be present in the derived CRD.
	Expose []VersionExposeConfig `json:"expose"`
}

// CatalogEntryMetadata contains metadata of the CatalogEntry.
type CatalogEntryMetadata struct {
	// DisplayName shows the human-readable name of this CatalogEntry.
	DisplayName string `json:"displayName,omitempty"`
	// Description shows the human-readable description of this CatalogEntry.
	Description string `json:"description,omitempty"`
}

// CatalogEntryStatus represents the observed state of CatalogEntry.
type CatalogEntryStatus struct {
	// TenantCRD holds the information about the Tenant facing CRD that is offered by this CatalogEntry.
	TenantCRD *CRDInformation `json:"tenantCRD,omitempty"`
	// ProviderCRD holds the information about the Provider facing CRD that is offered by this CatalogEntry.
	ProviderCRD *CRDInformation `json:"providerCRD,omitempty"`

	// ObservedGeneration is the most recent generation observed for this CatalogEntry by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions represents the latest available observations of a CatalogEntry's current state.
	Conditions []CatalogEntryCondition `json:"conditions,omitempty"`
	// DEPRECATED.
	// Phase represents the current lifecycle state of this object.
	// Consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to strings when printing the property.
	// This is only for display purpose, for everything else use conditions.
	Phase CatalogEntryPhaseType `json:"phase,omitempty"`
}

// CatalogEntryPhaseType represents all conditions as a single string for printing by using kubectl commands.
type CatalogEntryPhaseType string

// Values of CatalogEntryPhaseType.
const (
	CatalogEntryPhaseReady       CatalogEntryPhaseType = "Ready"
	CatalogEntryPhaseNotReady    CatalogEntryPhaseType = "NotReady"
	CatalogEntryPhaseUnknown     CatalogEntryPhaseType = "Unknown"
	CatalogEntryPhaseTerminating CatalogEntryPhaseType = "Terminating"
)

const (
	CatalogEntryTerminatingReason = "Deleting"
)

// updatePhase updates the phase property based on the current conditions.
// this method should be called every time the conditions are updated.
func (s *CatalogEntryStatus) updatePhase() {
	for _, condition := range s.Conditions {
		if condition.Type != CatalogEntryReady {
			continue
		}

		switch condition.Status {
		case ConditionTrue:
			s.Phase = CatalogEntryPhaseReady
		case ConditionFalse:
			if condition.Reason == CatalogEntryTerminatingReason {
				s.Phase = CatalogEntryPhaseTerminating
			} else {
				s.Phase = CatalogEntryPhaseNotReady
			}
		case ConditionUnknown:
			s.Phase = CatalogEntryPhaseUnknown
		}
		return
	}

	s.Phase = CatalogEntryPhaseUnknown
}

// CatalogEntryConditionType represents a CatalogEntryCondition value.
type CatalogEntryConditionType string

const (
	// CatalogEntryReady represents a CatalogEntry condition is in ready state.
	CatalogEntryReady CatalogEntryConditionType = "Ready"
)

// CatalogEntryCondition contains details for the current condition of this CatalogEntry.
type CatalogEntryCondition struct {
	// Type is the type of the CatalogEntry condition, currently ('Ready').
	Type CatalogEntryConditionType `json:"type"`
	// Status is the status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus `json:"status"`
	// LastTransitionTime is the last time the condition transits from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Reason is the (brief) reason for the condition's last transition.
	Reason string `json:"reason"`
	// Message is the human readable message indicating details about last transition.
	Message string `json:"message"`
}

// True returns whether .Status == "True"
func (c CatalogEntryCondition) True() bool {
	return c.Status == ConditionTrue
}

// GetCondition returns the Condition of the given condition type, if it exists.
func (s *CatalogEntryStatus) GetCondition(t CatalogEntryConditionType) (condition CatalogEntryCondition, exists bool) {
	for _, cond := range s.Conditions {
		if cond.Type == t {
			condition = cond
			exists = true
			return
		}
	}
	return
}

// SetCondition replaces or adds the given condition.
func (s *CatalogEntryStatus) SetCondition(condition CatalogEntryCondition) {
	defer s.updatePhase()

	if condition.LastTransitionTime.IsZero() {
		condition.LastTransitionTime = metav1.Now()
	}

	for i := range s.Conditions {
		if s.Conditions[i].Type == condition.Type {

			// Only update the LastTransitionTime when the Status is changed.
			if s.Conditions[i].Status != condition.Status {
				s.Conditions[i].LastTransitionTime = condition.LastTransitionTime
			}

			s.Conditions[i].Status = condition.Status
			s.Conditions[i].Reason = condition.Reason
			s.Conditions[i].Message = condition.Message

			return
		}
	}

	s.Conditions = append(s.Conditions, condition)
}

// CatalogEntry controls how to offer a CRD to other Tenants.
//
// A CatalogEntry references a single CRD, adds metadata to it and allows to limit field access for Tenants.
//
// **Simple Example**
// ```yaml
// apiVersion: catalog.kubecarrier.io/v1alpha1
// kind: CatalogEntry
// metadata:
//   name: couchdbs.eu-west-1
// spec:
//   metadata:
//     displayName: CouchDB
//     description: The compfy database
//   baseCRD:
//     name: couchdbs.eu-west-1.loodse
// ```
//
// **Example with limited fields**
// ```yaml
// apiVersion: catalog.kubecarrier.io/v1alpha1
// kind: CatalogEntry
// metadata:
//   name: couchdbs.eu-west-1
// spec:
//   metadata:
//     displayName: CouchDB
//     description: The compfy database
//   baseCRD:
//     name: couchdbs.eu-west-1.loodse
//   derive:
//     kindOverride: CouchDBPublic
//     expose:
//     - versions:
//       - v1alpha1
//       fields:
//       - jsonPath: .spec.username
//       - jsonPath: .spec.password
//       - jsonPath: .status.phase
//       - jsonPath: .status.fauxtonAddress
//       - jsonPath: .status.address
//       - jsonPath: .status.observedGeneration
// ```
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:categories=kubecarrier-provider,shortName=ce
type CatalogEntry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CatalogEntrySpec   `json:"spec,omitempty"`
	Status CatalogEntryStatus `json:"status,omitempty"`
}

// IsReady returns if the CatalogEntry is ready.
func (s *CatalogEntry) IsReady() bool {
	if s.Generation != s.Status.ObservedGeneration {
		return false
	}

	for _, condition := range s.Status.Conditions {
		if condition.Type == CatalogEntryReady &&
			condition.Status == ConditionTrue {
			return true
		}
	}
	return false
}

// CatalogEntryList contains a list of CatalogEntry.
// +kubebuilder:object:root=true
type CatalogEntryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CatalogEntry `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CatalogEntry{}, &CatalogEntryList{})
}

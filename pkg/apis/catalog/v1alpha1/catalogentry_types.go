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

// CatalogEntrySpec defines the desired state of CatalogEntry
type CatalogEntrySpec struct {
	// Metadata contains the metadata (display name, description, etc) of the CatalogEntry.
	Metadata CatalogEntryMetadata `json:"metadata,omitempty"`
	// CRDSelector selects CRD objects that should be part of this catalog.
	CRDSelector *metav1.LabelSelector `json:"crdSelector,omitempty"`
}

// CatalogEntryMetadata contains the metadata (display name, description, etc) of the CatalogEntry.
type CatalogEntryMetadata struct {
	// DisplayName shows the human-readable name of this CatalogEntry.
	DisplayName string `json:"displayName,omitempty"`
	// Description shows the human-readable description of this CatalogEntry.
	Description string `json:"description,omitempty"`
}

// CatalogEntryStatus defines the observed state of CatalogEntry.
type CatalogEntryStatus struct {
	// CRDs holds the information about the underlying CRDs that are offered by this CatalogEntry.
	CRDs []CRDInformation `json:"crds,omitempty"`

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

// updatePhase updates the phase property based on the current conditions
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

// CatalogEntry adds additional metadata to a set of CRD's and groups the same Kind for multiple ServiceClusters
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:shortName=ce
type CatalogEntry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CatalogEntrySpec   `json:"spec,omitempty"`
	Status CatalogEntryStatus `json:"status,omitempty"`
}

// CatalogEntryList contains a list of CatalogEntry
// +kubebuilder:object:root=true
type CatalogEntryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CatalogEntry `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CatalogEntry{}, &CatalogEntryList{})
}

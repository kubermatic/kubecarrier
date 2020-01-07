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

// CatalogSpec defines the desired state of Catalog
type CatalogSpec struct {

	// CatalogEntrySelector selects CatalogEntry objects that should be part of this catalog.
	// If this is not specified, it will match all CatalogEntries.
	CatalogEntrySelector *metav1.LabelSelector `json:"catalogEntrySelector,omitempty"`

	// TenantReferenceSelector selects TenantReference objects that the catalog should be published to.
	// If this is not specified, it will match all TenantReferences.
	TenantReferenceSelector *metav1.LabelSelector `json:"tenantReferenceSelector,omitempty"`
}

// CatalogStatus defines the observed state of Catalog.
type CatalogStatus struct {
	// Tenants is the list of the Tenants(TenantReference) that selected by this Catalog.
	Tenants []ObjectReference `json:"tenants,omitempty"`
	// Entries is the list of the CatalogEntries that selected by this Catalog.
	Entries []ObjectReference `json:"entries,omitempty"`
	// ObservedGeneration is the most recent generation observed for this Catalog by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions represents the latest available observations of a Catalog's current state.
	Conditions []CatalogCondition `json:"conditions,omitempty"`
	// DEPRECATED.
	// Phase represents the current lifecycle state of this object.
	// Consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to strings when printing the property.
	// This is only for display purpose, for everything else use conditions.
	Phase CatalogPhaseType `json:"phase,omitempty"`
}

// CatalogPhaseType represents all conditions as a single string for printing by using kubectl commands.
type CatalogPhaseType string

// Values of CatalogPhaseType.
const (
	CatalogPhaseReady       CatalogPhaseType = "Ready"
	CatalogPhaseNotReady    CatalogPhaseType = "NotReady"
	CatalogPhaseUnknown     CatalogPhaseType = "Unknown"
	CatalogPhaseTerminating CatalogPhaseType = "Terminating"
)

const (
	CatalogTerminatingReason = "Deleting"
)

// updatePhase updates the phase property based on the current conditions
// this method should be called every time the conditions are updated.
func (s *CatalogStatus) updatePhase() {
	for _, condition := range s.Conditions {
		if condition.Type != CatalogReady {
			continue
		}

		switch condition.Status {
		case ConditionTrue:
			s.Phase = CatalogPhaseReady
		case ConditionFalse:
			if condition.Reason == CatalogTerminatingReason {
				s.Phase = CatalogPhaseTerminating
			} else {
				s.Phase = CatalogPhaseNotReady
			}
		case ConditionUnknown:
			s.Phase = CatalogPhaseUnknown
		}
		return
	}

	s.Phase = CatalogPhaseUnknown
}

// CatalogConditionType represents a CatalogCondition value.
type CatalogConditionType string

const (
	// CatalogReady represents a Catalog condition is in ready state.
	CatalogReady CatalogConditionType = "Ready"
)

// CatalogCondition contains details for the current condition of this Catalog.
type CatalogCondition struct {
	// Type is the type of the Catalog condition, currently ('Ready').
	Type CatalogConditionType `json:"type"`
	// Status is the status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus `json:"status"`
	// LastTransitionTime is the last time the condition transits from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Reason is the (brief) reason for the condition's last transition.
	Reason string `json:"reason"`
	// Message is the human readable message indicating details about last transition.
	Message string `json:"message"`
}

// GetCondition returns the Condition of the given condition type, if it exists.
func (s *CatalogStatus) GetCondition(t CatalogConditionType) (condition CatalogCondition, exists bool) {
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
func (s *CatalogStatus) SetCondition(condition CatalogCondition) {
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

// Catalog publishes a selection of CatalogEntries to a selection of Tenants.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:shortName=cl
type Catalog struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CatalogSpec   `json:"spec,omitempty"`
	Status CatalogStatus `json:"status,omitempty"`
}

// CatalogList contains a list of Catalog
// +kubebuilder:object:root=true
type CatalogList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Catalog `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Catalog{}, &CatalogList{})
}

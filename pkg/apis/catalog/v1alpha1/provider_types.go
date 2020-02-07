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

// ProviderSpec defines the desired state of Provider.
type ProviderSpec struct {
	Metadata ProviderMetadata `json:"metadata,omitempty"`
}

// ProviderMetadata contains the metadata (display name, description, etc) of the Provider.
type ProviderMetadata struct {
	// DisplayName shows the human-readable name of this Provider.
	DisplayName string `json:"displayName,omitempty"`
	// Description shows the human-readable description of this Provider.
	Description string `json:"description,omitempty"`
}

// ProviderStatus defines the observed state of Provider.
type ProviderStatus struct {
	// NamespaceName is the name of the namespace that the Provider manages.
	NamespaceName string `json:"namespaceName,omitempty"`
	// ObservedGeneration is the most recent generation observed for this Provider by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions represents the latest available observations of a Provider's current state.
	Conditions []ProviderCondition `json:"conditions,omitempty"`
	// DEPRECATED.
	// Phase represents the current lifecycle state of this object.
	// Consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to strings when printing the property.
	// This is only for display purpose, for everything else use conditions.
	Phase ProviderPhaseType `json:"phase,omitempty"`
}

// ProviderPhaseType represents all conditions as a single string for printing by using kubectl commands.
type ProviderPhaseType string

// Values of ProviderPhaseType.
const (
	ProviderPhaseReady       ProviderPhaseType = "Ready"
	ProviderPhaseNotReady    ProviderPhaseType = "NotReady"
	ProviderPhaseUnknown     ProviderPhaseType = "Unknown"
	ProviderPhaseTerminating ProviderPhaseType = "Terminating"
)

const (
	ProviderTerminatingReason = "Deleting"
)

// updatePhase updates the phase property based on the current conditions
// this method should be called every time the conditions are updated.
func (s *ProviderStatus) updatePhase() {
	for _, condition := range s.Conditions {
		if condition.Type != ProviderReady {
			continue
		}

		switch condition.Status {
		case ConditionTrue:
			s.Phase = ProviderPhaseReady
		case ConditionFalse:
			if condition.Reason == ProviderTerminatingReason {
				s.Phase = ProviderPhaseTerminating
			} else {
				s.Phase = ProviderPhaseNotReady
			}
		case ConditionUnknown:
			s.Phase = ProviderPhaseUnknown
		}
		return
	}

	s.Phase = ProviderPhaseUnknown
}

// ProviderConditionType represents a ProviderCondition value.
type ProviderConditionType string

const (
	// ProviderReady represents a Provider condition is in ready state.
	ProviderReady ProviderConditionType = "Ready"
)

// ProviderCondition contains details for the current condition of this Provider.
type ProviderCondition struct {
	// Type is the type of the Provider condition, currently ('Ready').
	Type ProviderConditionType `json:"type"`
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
func (s *ProviderStatus) GetCondition(t ProviderConditionType) (condition ProviderCondition, exists bool) {
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
func (s *ProviderStatus) SetCondition(condition ProviderCondition) {
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

// Provider is the service provider representation in the KubeCarrier control-plane.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Provider Namespace",type="string",JSONPath=".status.namespaceName"
// +kubebuilder:printcolumn:name="Display Name",type="string",JSONPath=".spec.displayName"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:categories=kubecarrier-admin,shortName=pdr
type Provider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProviderSpec   `json:"spec,omitempty"`
	Status ProviderStatus `json:"status,omitempty"`
}

// ProviderList contains a list of Provider.
// +kubebuilder:object:root=true
type ProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Provider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Provider{}, &ProviderList{})
}

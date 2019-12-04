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

// TenantAssignmentSpec defines the desired state of TenantAssignment
type TenantAssignmentSpec struct {
	Tenant         ObjectReference `json:"tenant"`
	ServiceCluster ObjectReference `json:"serviceCluster"`
}

// TenantAssignmentStatus defines the observed state of TenantAssignment
type TenantAssignmentStatus struct {
	// DEPRECATED.
	// Phase represents the current lifecycle state of this object
	// consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to a string when printing the property
	// is only present for display purposes, for everything else use conditions
	Phase TenantAssignmentPhaseType `json:"phase,omitempty"`
	// Conditions is a list of all conditions this TenantAssignment is in.
	Conditions []TenantAssignmentCondition `json:"conditions,omitempty"`
	// The most recent generation observed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// NamespaceName references the Namespace on the ServiceCluster that was assigned.
	NamespaceName string `json:"namespaceName,omitempty"`
}

// TenantAssignmentPhaseType represents all conditions as a single string for printing in kubectl.
type TenantAssignmentPhaseType string

// Values of TenantAssignmentPhaseType
const (
	TenantAssignmentPhaseReady    TenantAssignmentPhaseType = "Ready"
	TenantAssignmentPhaseNotReady TenantAssignmentPhaseType = "NotReady"
	TenantAssignmentPhaseUnknown  TenantAssignmentPhaseType = "Unknown"
)

// TenantAssignmentConditionType represents a TenantAssignmentCondition value.
type TenantAssignmentConditionType string

const (
	// TenantAssignmentReady represents a TenantAssignment condition is in ready state.
	TenantAssignmentReady         TenantAssignmentConditionType = "Ready"
	TenantAssignmentNamesAccepted TenantAssignmentConditionType = "NamesAccepted"
)

// TenantAssignmentCondition contains details for the current condition of this TenantAssignment.
type TenantAssignmentCondition struct {
	// LastTransitionTime is the last time the condition transit from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Message is the human readable message indicating details about last transition.
	Message string `json:"message"`
	// Reason is the (brief) reason for the condition's last transition.
	Reason string `json:"reason"`
	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus `json:"status"`
	// Type of the condition, currently ('Ready').
	Type TenantAssignmentConditionType `json:"type"`
}

// updatePhase updates the phase property based on the current conditions
// this method should be called every time the conditions are updated
func (s *TenantAssignmentStatus) updatePhase() {
	for _, condition := range s.Conditions {
		if condition.Type != TenantAssignmentReady {
			continue
		}

		switch condition.Status {
		case ConditionTrue:
			s.Phase = TenantAssignmentPhaseReady
		case ConditionFalse:
			s.Phase = TenantAssignmentPhaseNotReady
		case ConditionUnknown:
			s.Phase = TenantAssignmentPhaseUnknown
		}
		return
	}
	s.Phase = TenantAssignmentPhaseUnknown
}

// SetCondition replaces or adds the given condition
func (s *TenantAssignmentStatus) SetCondition(condition TenantAssignmentCondition) {
	defer s.updatePhase()

	if condition.LastTransitionTime.IsZero() {
		condition.LastTransitionTime = metav1.Now()
	}

	for i := range s.Conditions {
		if s.Conditions[i].Type == condition.Type {
			// Only change the LastTransitionTime when Status changed
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

// GetCondition returns the Condition of the given type, if it exists
func (s *TenantAssignmentStatus) GetCondition(conditionType TenantAssignmentConditionType) (condition TenantAssignmentCondition, exists bool) {
	for _, cond := range s.Conditions {
		if cond.Type == conditionType {
			condition = cond
			exists = true
			return
		}
	}
	return
}

// TenantAssignment represents the assignment of a Tenant to a ServiceCluster.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type TenantAssignment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TenantAssignmentSpec   `json:"spec,omitempty"`
	Status TenantAssignmentStatus `json:"status,omitempty"`
}

// TenantAssignmentList contains a list of TenantAssignment
// +kubebuilder:object:root=true
type TenantAssignmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TenantAssignment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TenantAssignment{}, &TenantAssignmentList{})
}

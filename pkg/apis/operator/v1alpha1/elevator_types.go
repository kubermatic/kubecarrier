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

// ElevatorSpec defines the desired state of Elevator.
type ElevatorSpec struct {
	// References the provider or internal CRD, that should be created in the provider namespace.
	ProviderCRD CRDReference `json:"providerCRD"`
	// References the public CRD that will be synced into the provider namespace.
	TenantCRD CRDReference `json:"tenantCRD"`
	// References the DerivedCustomResource controlling the Tenant-side CRD.
	DerivedCR ObjectReference `json:"derivedCR"`
}

// ElevatorStatus defines the observed state of Elevator.
type ElevatorStatus struct {
	// ObservedGeneration is the most recent generation observed for this Elevator by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions represents the latest available observations of a Elevator's current state.
	Conditions []ElevatorCondition `json:"conditions,omitempty"`
	// DEPRECATED.
	// Phase represents the current lifecycle state of this object.
	// Consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to strings when printing the property.
	// This is only for display purpose, for everything else use conditions.
	Phase ElevatorPhaseType `json:"phase,omitempty"`
}

// ElevatorPhaseType represents all conditions as a single string for printing by using kubectl commands.
type ElevatorPhaseType string

// Values of ElevatorPhaseType.
const (
	ElevatorPhaseReady       ElevatorPhaseType = "Ready"
	ElevatorPhaseNotReady    ElevatorPhaseType = "NotReady"
	ElevatorPhaseUnknown     ElevatorPhaseType = "Unknown"
	ElevatorPhaseTerminating ElevatorPhaseType = "Terminating"
)

const (
	ElevatorTerminatingReason = "Deleting"
)

// updatePhase updates the phase property based on the current conditions.
// this method should be called every time the conditions are updated.
func (s *ElevatorStatus) updatePhase() {

	for _, condition := range s.Conditions {
		if condition.Type != ElevatorReady {
			continue
		}

		switch condition.Status {
		case ConditionTrue:
			s.Phase = ElevatorPhaseReady
		case ConditionFalse:
			if condition.Reason == ElevatorTerminatingReason {
				s.Phase = ElevatorPhaseTerminating
			} else {
				s.Phase = ElevatorPhaseNotReady
			}
		case ConditionUnknown:
			s.Phase = ElevatorPhaseUnknown
		}
		return
	}

	s.Phase = ElevatorPhaseUnknown
}

// ElevatorConditionType represents a ElevatorCondition value.
type ElevatorConditionType string

const (
	// ElevatorReady represents a Elevator condition is in ready state.
	ElevatorReady ElevatorConditionType = "Ready"
)

// ElevatorCondition contains details for the current condition of this Elevator.
type ElevatorCondition struct {
	// Type is the type of the Elevator condition, currently ('Ready').
	Type ElevatorConditionType `json:"type"`
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
func (c ElevatorCondition) True() bool {
	return c.Status == ConditionTrue
}

// GetCondition returns the Condition of the given condition type, if it exists.
func (s *ElevatorStatus) GetCondition(t ElevatorConditionType) (condition ElevatorCondition, exists bool) {
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
func (s *ElevatorStatus) SetCondition(condition ElevatorCondition) {
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

// Elevator manages the deployment of the Elevator controller manager.
//
// For each `DerivedCustomResource` a Elevator instance is launched to propagate the derived CRD instance into the Namespace of it's provider.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Elevator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ElevatorSpec   `json:"spec,omitempty"`
	Status ElevatorStatus `json:"status,omitempty"`
}

// IsReady returns if the Elevator is ready.
func (s *Elevator) IsReady() bool {
	if s.Generation != s.Status.ObservedGeneration {
		return false
	}

	for _, condition := range s.Status.Conditions {
		if condition.Type == ElevatorReady &&
			condition.Status == ConditionTrue {
			return true
		}
	}
	return false
}

func (s *Elevator) SetReadyCondition() bool {
	readyCondition, _ := s.Status.GetCondition(ElevatorReady)
	if readyCondition.Status != ConditionTrue {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(ElevatorCondition{
			Type:    ElevatorReady,
			Status:  ConditionTrue,
			Reason:  "DeploymentReady",
			Message: "the deployment of the Elevator controller manager is ready",
		})
		return true
	}
	return false
}
func (s *Elevator) SetUnReadyCondition() bool {
	readyCondition, _ := s.Status.GetCondition(ElevatorReady)
	if readyCondition.Status != ConditionFalse {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(ElevatorCondition{
			Type:    ElevatorReady,
			Status:  ConditionFalse,
			Reason:  "DeploymentUnready",
			Message: "the deployment of the Elevator controller manager is not ready",
		})
		return true
	}
	return false
}

func (s *Elevator) SetTerminatingCondition() bool {
	readyCondition, _ := s.Status.GetCondition(ElevatorReady)
	if readyCondition.Status != ConditionFalse ||
		readyCondition.Status == ConditionFalse && readyCondition.Reason != ElevatorTerminatingReason {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(ElevatorCondition{
			Type:    ElevatorReady,
			Status:  ConditionFalse,
			Reason:  ElevatorTerminatingReason,
			Message: "Elevator is being deleted",
		})
		return true
	}
	return false
}

// ElevatorList contains a list of Elevator.
// +kubebuilder:object:root=true
type ElevatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Elevator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Elevator{}, &ElevatorList{})
}

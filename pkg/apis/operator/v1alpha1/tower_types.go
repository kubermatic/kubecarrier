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

// TowerSpec defines the desired state of Tower
type TowerSpec struct {
}

// TowerStatus defines the observed state of Tower
type TowerStatus struct {
	// ObservedGeneration is the most recent generation observed for this Tower by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions represents the latest available observations of a Tower's current state.
	Conditions []TowerCondition `json:"conditions,omitempty"`
	// DEPRECATED.
	// Phase represents the current lifecycle state of this object.
	// Consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to strings when printing the property.
	// This is only for display purpose, for everything else use conditions.
	Phase TowerPhaseType `json:"phase,omitempty"`
}

// TowerPhaseType represents all conditions as a single string for printing by using kubectl commands.
type TowerPhaseType string

// Values of TowerPhaseType.
const (
	TowerPhaseReady       TowerPhaseType = "Ready"
	TowerPhaseNotReady    TowerPhaseType = "NotReady"
	TowerPhaseUnknown     TowerPhaseType = "Unknown"
	TowerPhaseTerminating TowerPhaseType = "Terminating"
)

const (
	TowerTerminatingReason = "Deleting"
)

// updatePhase updates the phase property based on the current conditions
// this method should be called every time the conditions are updated.
func (s *TowerStatus) updatePhase() {

	for _, condition := range s.Conditions {
		if condition.Type != TowerReady {
			continue
		}

		switch condition.Status {
		case ConditionTrue:
			s.Phase = TowerPhaseReady
		case ConditionFalse:
			if condition.Reason == TowerTerminatingReason {
				s.Phase = TowerPhaseTerminating
			} else {
				s.Phase = TowerPhaseNotReady
			}
		case ConditionUnknown:
			s.Phase = TowerPhaseUnknown
		}
		return
	}

	s.Phase = TowerPhaseUnknown
}

// TowerConditionType represents a TowerCondition value.
type TowerConditionType string

const (
	// TowerReady represents a Tower condition is in ready state.
	TowerReady TowerConditionType = "Ready"
)

// TowerCondition contains details for the current condition of this Tower.
type TowerCondition struct {
	// Type is the type of the Tower condition, currently ('Ready').
	Type TowerConditionType `json:"type"`
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
func (c TowerCondition) True() bool {
	return c.Status == ConditionTrue
}

// GetCondition returns the Condition of the given condition type, if it exists.
func (s *TowerStatus) GetCondition(t TowerConditionType) (condition TowerCondition, exists bool) {
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
func (s *TowerStatus) SetCondition(condition TowerCondition) {
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

// Tower manages the deployment of the KubeCarrier master controller manager.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Tower struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TowerSpec   `json:"spec,omitempty"`
	Status TowerStatus `json:"status,omitempty"`
}

// IsReady returns if the Tower is ready.
func (s *Tower) IsReady() bool {
	if s.Generation != s.Status.ObservedGeneration {
		return false
	}

	for _, condition := range s.Status.Conditions {
		if condition.Type == TowerReady &&
			condition.Status == ConditionTrue {
			return true
		}
	}
	return false
}

func (s *Tower) SetReadyCondition() bool {
	if !s.IsReady() {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(TowerCondition{
			Type:    TowerReady,
			Status:  ConditionTrue,
			Reason:  "DeploymentReady",
			Message: "the deployment of the Tower controller manager is ready",
		})
		return true
	}
	return false
}
func (s *Tower) SetUnReadyCondition() bool {
	readyCondition, _ := s.Status.GetCondition(TowerReady)
	if readyCondition.Status != ConditionFalse {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(TowerCondition{
			Type:    TowerReady,
			Status:  ConditionFalse,
			Reason:  "DeploymentUnready",
			Message: "the deployment of the Tower controller manager is not ready",
		})
		return true
	}
	return false
}

func (s *Tower) SetTerminatingCondition() bool {
	readyCondition, _ := s.Status.GetCondition(TowerReady)
	if readyCondition.Status != ConditionFalse ||
		readyCondition.Status == ConditionFalse && readyCondition.Reason != TowerTerminatingReason {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(TowerCondition{
			Type:    TowerReady,
			Status:  ConditionFalse,
			Reason:  TowerTerminatingReason,
			Message: "Tower is being deleted",
		})
		return true
	}
	return false
}

// +kubebuilder:object:root=true

// TowerList contains a list of Tower
type TowerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tower `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tower{}, &TowerList{})
}

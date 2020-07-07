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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SnapshotSpec defines the desired state of Snapshot
type SnapshotSpec struct {
	// DBName is the name of the source DB
	DBName string `json:"dbName,omitempty"`
}

// SnapshotStatus defines the observed state of Snapshot
type SnapshotStatus struct {
	// ObservedGeneration is the most recent generation observed for this Snapshot by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions represents the latest available observations of a Snapshot's current state.
	Conditions []SnapshotCondition `json:"conditions,omitempty"`
	// DEPRECATED.
	// Phase represents the current lifecycle state of this object.
	// Consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to strings when printing the property.
	// This is only for display purpose, for everything else use conditions.
	Phase SnapshotPhaseType `json:"phase,omitempty"`
	// Date when snapshot was taken
	Date metav1.Time `json:"date,omitempty"`
}

// SnapshotPhaseType represents all conditions as a single string for printing by using kubectl commands.
type SnapshotPhaseType string

// Values of SnapshotPhaseType.
const (
	SnapshotPhaseReady       SnapshotPhaseType = "Ready"
	SnapshotPhaseNotReady    SnapshotPhaseType = "NotReady"
	SnapshotPhaseUnknown     SnapshotPhaseType = "Unknown"
	SnapshotPhaseTerminating SnapshotPhaseType = "Terminating"
)

const (
	SnapshotTerminatingReason = "Deleting"
)

// SnapshotConditionType represents a SnapshotCondition value.
type SnapshotConditionType string

const (
	// SnapshotReady represents a Snapshot condition is in ready state.
	SnapshotReady SnapshotConditionType = "Ready"
)

// SnapshotCondition contains details for the current condition of this Snapshot.
type SnapshotCondition struct {
	// Type is the type of the Snapshot condition, currently ('Ready').
	Type SnapshotConditionType `json:"type"`
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
func (c SnapshotCondition) True() bool {
	return c.Status == ConditionTrue
}

// updatePhase updates the phase property based on the current conditions.
// this method should be called every time the conditions are updated.
func (s *SnapshotStatus) updatePhase() {

	for _, condition := range s.Conditions {
		if condition.Type != SnapshotReady {
			continue
		}

		switch condition.Status {
		case ConditionTrue:
			s.Phase = SnapshotPhaseReady
		case ConditionFalse:
			if condition.Reason == SnapshotTerminatingReason {
				s.Phase = SnapshotPhaseTerminating
			} else {
				s.Phase = SnapshotPhaseNotReady
			}
		case ConditionUnknown:
			s.Phase = SnapshotPhaseUnknown
		}
		return
	}

	s.Phase = SnapshotPhaseUnknown
}

// GetCondition returns the Condition of the given condition type, if it exists.
func (s *SnapshotStatus) GetCondition(t SnapshotConditionType) (condition SnapshotCondition, exists bool) {
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
func (s *SnapshotStatus) SetCondition(condition SnapshotCondition) {
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

// Snapshot is snapshot of the DB element for e2e operator
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:categories=all
type Snapshot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SnapshotSpec   `json:"spec,omitempty"`
	Status SnapshotStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// SnapshotList contains a list of Snapshot
type SnapshotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Snapshot `json:"items"`
}

// IsReady returns if the Snapshot is ready.
func (s *Snapshot) IsReady() bool {
	if s.Generation != s.Status.ObservedGeneration {
		return false
	}

	for _, condition := range s.Status.Conditions {
		if condition.Type == SnapshotReady &&
			condition.Status == ConditionTrue {
			return true
		}
	}
	return false
}

func (s *Snapshot) SetReadyCondition() bool {
	if !s.IsReady() {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(SnapshotCondition{
			Type:    SnapshotReady,
			Status:  ConditionTrue,
			Reason:  "DeploymentReady",
			Message: "the Snapshot is ready",
		})
		return true
	}
	return false
}
func (s *Snapshot) SetUnReadyCondition() bool {
	readyCondition, _ := s.Status.GetCondition(SnapshotReady)
	if readyCondition.Status != ConditionFalse {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(SnapshotCondition{
			Type:    SnapshotReady,
			Status:  ConditionFalse,
			Reason:  "SnapshotUnready",
			Message: "the Snapshot is not ready",
		})
		return true
	}
	return false
}

func (s *Snapshot) SetTerminatingCondition() bool {
	readyCondition, _ := s.Status.GetCondition(SnapshotReady)
	if readyCondition.Status != ConditionFalse ||
		readyCondition.Status == ConditionFalse && readyCondition.Reason != SnapshotTerminatingReason {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(SnapshotCondition{
			Type:    SnapshotReady,
			Status:  ConditionFalse,
			Reason:  SnapshotTerminatingReason,
			Message: "Snapshot is being deleted",
		})
		return true
	}
	return false
}

func init() {
	SchemeBuilder.Register(&Snapshot{}, &SnapshotList{})
}

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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// BackupSpec defines the desired state of Backup
type BackupSpec struct {
	DBName string `json:"dbName"`
}

// BackupStatus defines the observed state of Backup
type BackupStatus struct {
	// ObservedGeneration is the most recent generation observed for this Backup by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions represents the latest available observations of a Backup's current state.
	Conditions []BackupCondition `json:"conditions,omitempty"`
	// DEPRECATED.
	// Phase represents the current lifecycle state of this object.
	// Consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to strings when printing the property.
	// This is only for display purpose, for everything else use conditions.
	Phase BackupPhaseType `json:"phase,omitempty"`
}

// BackupPhaseType represents all conditions as a single string for printing by using kubectl commands.
type BackupPhaseType string

// Values of BackupPhaseType.
const (
	BackupPhaseReady       BackupPhaseType = "Ready"
	BackupPhaseNotReady    BackupPhaseType = "NotReady"
	BackupPhaseUnknown     BackupPhaseType = "Unknown"
	BackupPhaseTerminating BackupPhaseType = "Terminating"
)

const (
	BackupTerminatingReason = "Deleting"
)

// BackupConditionType represents a BackupCondition value.
type BackupConditionType string

const (
	// BackupReady represents a Backup condition is in ready state.
	BackupReady BackupConditionType = "Ready"
)

// BackupCondition contains details for the current condition of this Backup.
type BackupCondition struct {
	// Type is the type of the Backup condition, currently ('Ready').
	Type BackupConditionType `json:"type"`
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
func (c BackupCondition) True() bool {
	return c.Status == ConditionTrue
}

// updatePhase updates the phase property based on the current conditions.
// this method should be called every time the conditions are updated.
func (s *BackupStatus) updatePhase() {

	for _, condition := range s.Conditions {
		if condition.Type != BackupReady {
			continue
		}

		switch condition.Status {
		case ConditionTrue:
			s.Phase = BackupPhaseReady
		case ConditionFalse:
			if condition.Reason == BackupTerminatingReason {
				s.Phase = BackupPhaseTerminating
			} else {
				s.Phase = BackupPhaseNotReady
			}
		case ConditionUnknown:
			s.Phase = BackupPhaseUnknown
		}
		return
	}

	s.Phase = BackupPhaseUnknown
}

// GetCondition returns the Condition of the given condition type, if it exists.
func (s *BackupStatus) GetCondition(t BackupConditionType) (condition BackupCondition, exists bool) {
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
func (s *BackupStatus) SetCondition(condition BackupCondition) {
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

// Backup is backup of the DB element for e2e operator
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:categories=all
type Backup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackupSpec   `json:"spec,omitempty"`
	Status BackupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// BackupList contains a list of Backup
type BackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Backup `json:"items"`
}

// IsReady returns if the Backup is ready.
func (s *Backup) IsReady() bool {
	if s.Generation != s.Status.ObservedGeneration {
		return false
	}

	for _, condition := range s.Status.Conditions {
		if condition.Type == BackupReady &&
			condition.Status == ConditionTrue {
			return true
		}
	}
	return false
}

func (s *Backup) SetReadyCondition() bool {
	if !s.IsReady() {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(BackupCondition{
			Type:    BackupReady,
			Status:  ConditionTrue,
			Reason:  "DeploymentReady",
			Message: "the Backup is ready",
		})
		return true
	}
	return false
}
func (s *Backup) SetUnReadyCondition() bool {
	readyCondition, _ := s.Status.GetCondition(BackupReady)
	if readyCondition.Status != ConditionFalse {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(BackupCondition{
			Type:    BackupReady,
			Status:  ConditionFalse,
			Reason:  "BackupUnready",
			Message: "the Backup is not ready",
		})
		return true
	}
	return false
}

func (s *Backup) SetTerminatingCondition() bool {
	readyCondition, _ := s.Status.GetCondition(BackupReady)
	if readyCondition.Status != ConditionFalse ||
		readyCondition.Status == ConditionFalse && readyCondition.Reason != BackupTerminatingReason {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(BackupCondition{
			Type:    BackupReady,
			Status:  ConditionFalse,
			Reason:  BackupTerminatingReason,
			Message: "Backup is being deleted",
		})
		return true
	}
	return false
}

func init() {
	SchemeBuilder.Register(&Backup{}, &BackupList{})
}

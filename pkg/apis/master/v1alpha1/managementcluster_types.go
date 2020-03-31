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

// ManagementClusterSpec defines the desired state of ManagementCluster.
type ManagementClusterSpec struct {
}

// ManagementClusterStatus defines the observed state of ManagementCluster.
type ManagementClusterStatus struct {
	// ObservedGeneration is the most recent generation observed for this ManagementCluster by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions represents the latest available observations of a ManagementCluster's current state.
	Conditions []ManagementClusterCondition `json:"conditions,omitempty"`
	// DEPRECATED.
	// Phase represents the current lifecycle state of this object.
	// Consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to strings when printing the property.
	// This is only for display purpose, for everything else use conditions.
	Phase ManagementClusterPhaseType `json:"phase,omitempty"`
}

// ManagementClusterPhaseType represents all conditions as a single string for printing by using kubectl commands.
type ManagementClusterPhaseType string

// Values of ManagementClusterPhaseType.
const (
	ManagementClusterPhaseReady       ManagementClusterPhaseType = "Ready"
	ManagementClusterPhaseNotReady    ManagementClusterPhaseType = "NotReady"
	ManagementClusterPhaseUnknown     ManagementClusterPhaseType = "Unknown"
	ManagementClusterPhaseTerminating ManagementClusterPhaseType = "Terminating"
)

const (
	ManagementClusterTerminatingReason = "Deleting"
)

// updatePhase updates the phase property based on the current conditions.
// this method should be called every time the conditions are updated.
func (s *ManagementClusterStatus) updatePhase() {

	for _, condition := range s.Conditions {
		if condition.Type != ManagementClusterReady {
			continue
		}

		switch condition.Status {
		case ConditionTrue:
			s.Phase = ManagementClusterPhaseReady
		case ConditionFalse:
			if condition.Reason == ManagementClusterTerminatingReason {
				s.Phase = ManagementClusterPhaseTerminating
			} else {
				s.Phase = ManagementClusterPhaseNotReady
			}
		case ConditionUnknown:
			s.Phase = ManagementClusterPhaseUnknown
		}
		return
	}

	s.Phase = ManagementClusterPhaseUnknown
}

// ManagementClusterConditionType represents a ManagementClusterCondition value.
type ManagementClusterConditionType string

const (
	// ManagementClusterReady represents a ManagementCluster condition is in ready state.
	ManagementClusterReady ManagementClusterConditionType = "Ready"
)

// ManagementClusterCondition contains details for the current condition of this ManagementCluster.
type ManagementClusterCondition struct {
	// Type is the type of the ManagementCluster condition, currently ('Ready').
	Type ManagementClusterConditionType `json:"type"`
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
func (c ManagementClusterCondition) True() bool {
	return c.Status == ConditionTrue
}

// GetCondition returns the Condition of the given condition type, if it exists.
func (s *ManagementClusterStatus) GetCondition(t ManagementClusterConditionType) (condition ManagementClusterCondition, exists bool) {
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
func (s *ManagementClusterStatus) SetCondition(condition ManagementClusterCondition) {
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

// ManagementCluster represents a management cluster which has KubeCarrier installation running on it.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
type ManagementCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManagementClusterSpec   `json:"spec,omitempty"`
	Status ManagementClusterStatus `json:"status,omitempty"`
}

// IsReady returns if the ManagementCluster is ready.
func (s *ManagementCluster) IsReady() bool {
	if s.Generation != s.Status.ObservedGeneration {
		return false
	}

	for _, condition := range s.Status.Conditions {
		if condition.Type == ManagementClusterReady &&
			condition.Status == ConditionTrue {
			return true
		}
	}
	return false
}

// ManagementClusterList contains a list of ManagementCluster.
// +kubebuilder:object:root=true
type ManagementClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManagementCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ManagementCluster{}, &ManagementClusterList{})
}

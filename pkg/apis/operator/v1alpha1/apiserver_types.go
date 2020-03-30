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

// APIServerSpec defines the desired state of APIServer
type APIServerSpec struct {
}

// APIServerStatus defines the observed state of APIServer
type APIServerStatus struct {
	// ObservedGeneration is the most recent generation observed for this APIServer by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions represents the latest available observations of a APIServer's current state.
	Conditions []APIServerCondition `json:"conditions,omitempty"`
	// DEPRECATED.
	// Phase represents the current lifecycle state of this object.
	// Consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to strings when printing the property.
	// This is only for display purpose, for everything else use conditions.
	Phase APIServerPhaseType `json:"phase,omitempty"`
}

// APIServerPhaseType represents all conditions as a single string for printing by using kubectl commands.
type APIServerPhaseType string

// Values of APIServerPhaseType.
const (
	APIServerPhaseReady       APIServerPhaseType = "Ready"
	APIServerPhaseNotReady    APIServerPhaseType = "NotReady"
	APIServerPhaseUnknown     APIServerPhaseType = "Unknown"
	APIServerPhaseTerminating APIServerPhaseType = "Terminating"
)

const (
	APIServerTerminatingReason = "Deleting"
)

// updatePhase updates the phase property based on the current conditions
// this method should be called every time the conditions are updated.
func (s *APIServerStatus) updatePhase() {

	for _, condition := range s.Conditions {
		if condition.Type != APIServerReady {
			continue
		}

		switch condition.Status {
		case ConditionTrue:
			s.Phase = APIServerPhaseReady
		case ConditionFalse:
			if condition.Reason == APIServerTerminatingReason {
				s.Phase = APIServerPhaseTerminating
			} else {
				s.Phase = APIServerPhaseNotReady
			}
		case ConditionUnknown:
			s.Phase = APIServerPhaseUnknown
		}
		return
	}

	s.Phase = APIServerPhaseUnknown
}

// APIServerConditionType represents a APIServerCondition value.
type APIServerConditionType string

const (
	// APIServerReady represents a APIServer condition is in ready state.
	APIServerReady APIServerConditionType = "Ready"
)

// APIServerCondition contains details for the current condition of this APIServer.
type APIServerCondition struct {
	// Type is the type of the APIServer condition, currently ('Ready').
	Type APIServerConditionType `json:"type"`
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
func (c APIServerCondition) True() bool {
	return c.Status == ConditionTrue
}

// GetCondition returns the Condition of the given condition type, if it exists.
func (s *APIServerStatus) GetCondition(t APIServerConditionType) (condition APIServerCondition, exists bool) {
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
func (s *APIServerStatus) SetCondition(condition APIServerCondition) {
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

// APIServer manages the deployment of the KubeCarrier master controller manager.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type APIServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   APIServerSpec   `json:"spec,omitempty"`
	Status APIServerStatus `json:"status,omitempty"`
}

// IsReady returns if the APIServer is ready.
func (s *APIServer) IsReady() bool {
	if s.Generation != s.Status.ObservedGeneration {
		return false
	}

	for _, condition := range s.Status.Conditions {
		if condition.Type == APIServerReady &&
			condition.Status == ConditionTrue {
			return true
		}
	}
	return false
}

func (s *APIServer) SetReadyCondition() bool {
	if !s.IsReady() {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(APIServerCondition{
			Type:    APIServerReady,
			Status:  ConditionTrue,
			Reason:  "DeploymentReady",
			Message: "the deployment of the APIServer controller manager is ready",
		})
		return true
	}
	return false
}
func (s *APIServer) SetUnReadyCondition() bool {
	readyCondition, _ := s.Status.GetCondition(APIServerReady)
	if readyCondition.Status != ConditionFalse {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(APIServerCondition{
			Type:    APIServerReady,
			Status:  ConditionFalse,
			Reason:  "DeploymentUnready",
			Message: "the deployment of the APIServer controller manager is not ready",
		})
		return true
	}
	return false
}

func (s *APIServer) SetTerminatingCondition() bool {
	readyCondition, _ := s.Status.GetCondition(APIServerReady)
	if readyCondition.Status != ConditionFalse ||
		readyCondition.Status == ConditionFalse && readyCondition.Reason != APIServerTerminatingReason {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(APIServerCondition{
			Type:    APIServerReady,
			Status:  ConditionFalse,
			Reason:  APIServerTerminatingReason,
			Message: "APIServer is being deleted",
		})
		return true
	}
	return false
}

// +kubebuilder:object:root=true

// APIServerList contains a list of APIServer
type APIServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []APIServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&APIServer{}, &APIServerList{})
}

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

// KubeCarrierSpec defines the desired state of KubeCarrier
type KubeCarrierSpec struct {
}

// KubeCarrierStatus defines the observed state of KubeCarrier
type KubeCarrierStatus struct {
	// ObservedGeneration is the most recent generation observed for this KubeCarrier by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions represents the latest available observations of a KubeCarrier's current state.
	Conditions []KubeCarrierCondition `json:"conditions,omitempty"`
	// DEPRECATED.
	// Phase represents the current lifecycle state of this object.
	// Consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to strings when printing the property.
	// This is only for display purpose, for everything else use conditions.
	Phase KubeCarrierPhaseType `json:"phase,omitempty"`
}

// KubeCarrierPhaseType represents all conditions as a single string for printing by using kubectl commands.
type KubeCarrierPhaseType string

// Values of KubeCarrierPhaseType.
const (
	KubeCarrierPhaseReady       KubeCarrierPhaseType = "Ready"
	KubeCarrierPhaseNotReady    KubeCarrierPhaseType = "NotReady"
	KubeCarrierPhaseUnknown     KubeCarrierPhaseType = "Unknown"
	KubeCarrierPhaseTerminating KubeCarrierPhaseType = "Terminating"
)

const (
	KubeCarrierTerminatingReason = "Deleting"
)

// updatePhase updates the phase property based on the current conditions
// this method should be called every time the conditions are updated.
func (s *KubeCarrierStatus) updatePhase() {

	for _, condition := range s.Conditions {
		if condition.Type != KubeCarrierReady {
			continue
		}

		switch condition.Status {
		case ConditionTrue:
			s.Phase = KubeCarrierPhaseReady
		case ConditionFalse:
			if condition.Reason == KubeCarrierTerminatingReason {
				s.Phase = KubeCarrierPhaseTerminating
			} else {
				s.Phase = KubeCarrierPhaseNotReady
			}
		case ConditionUnknown:
			s.Phase = KubeCarrierPhaseUnknown
		}
		return
	}

	s.Phase = KubeCarrierPhaseUnknown
}

// KubeCarrierConditionType represents a KubeCarrierCondition value.
type KubeCarrierConditionType string

const (
	// KubeCarrierReady represents a KubeCarrier condition is in ready state.
	KubeCarrierReady KubeCarrierConditionType = "Ready"
)

// KubeCarrierCondition contains details for the current condition of this KubeCarrier.
type KubeCarrierCondition struct {
	// Type is the type of the KubeCarrier condition, currently ('Ready').
	Type KubeCarrierConditionType `json:"type"`
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
func (c KubeCarrierCondition) True() bool {
	return c.Status == ConditionTrue
}

// GetCondition returns the Condition of the given condition type, if it exists.
func (s *KubeCarrierStatus) GetCondition(t KubeCarrierConditionType) (condition KubeCarrierCondition, exists bool) {
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
func (s *KubeCarrierStatus) SetCondition(condition KubeCarrierCondition) {
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

// KubeCarrier manages the deployment of the KubeCarrier controller manager.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories=all
type KubeCarrier struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubeCarrierSpec   `json:"spec,omitempty"`
	Status KubeCarrierStatus `json:"status,omitempty"`
}

// IsReady returns if the KubeCarrier is ready.
func (s *KubeCarrier) IsReady() bool {
	if s.Generation != s.Status.ObservedGeneration {
		return false
	}

	for _, condition := range s.Status.Conditions {
		if condition.Type == KubeCarrierReady &&
			condition.Status == ConditionTrue {
			return true
		}
	}
	return false
}

func (s *KubeCarrier) SetReadyCondition() bool {
	if !s.IsReady() {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(KubeCarrierCondition{
			Type:    KubeCarrierReady,
			Status:  ConditionTrue,
			Reason:  "DeploymentReady",
			Message: "the deployment of the KubeCarrier controller manager is ready",
		})
		return true
	}
	return false
}
func (s *KubeCarrier) SetUnReadyCondition() bool {
	readyCondition, _ := s.Status.GetCondition(KubeCarrierReady)
	if readyCondition.Status != ConditionFalse {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(KubeCarrierCondition{
			Type:    KubeCarrierReady,
			Status:  ConditionFalse,
			Reason:  "DeploymentUnready",
			Message: "the deployment of the KubeCarrier controller manager is not ready",
		})
		return true
	}
	return false
}

func (s *KubeCarrier) SetTerminatingCondition() bool {
	readyCondition, _ := s.Status.GetCondition(KubeCarrierReady)
	if readyCondition.Status != ConditionFalse ||
		readyCondition.Status == ConditionFalse && readyCondition.Reason != KubeCarrierTerminatingReason {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(KubeCarrierCondition{
			Type:    KubeCarrierReady,
			Status:  ConditionFalse,
			Reason:  KubeCarrierTerminatingReason,
			Message: "KubeCarrier is being deleted",
		})
		return true
	}
	return false
}

// +kubebuilder:object:root=true

// KubeCarrierList contains a list of KubeCarrier
type KubeCarrierList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubeCarrier `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubeCarrier{}, &KubeCarrierList{})
}

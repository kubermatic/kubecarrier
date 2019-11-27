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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// JokeSpec defines the desired state of Joke
type JokeSpec struct {
	// JokeDatabase specifies all known jokes
	JokeDatabase []JokeItem `json:"jokeDatabase,omitempty"`

	// JokeType to select
	JokeType string `json:"jokeType"`
}

// JokeCRDConfiguration holds static type information for the Joke instance.
type JokeItem struct {
	// Text of the joke
	Text string `json:"text"`

	// Type of the Joke
	Type string `json:"type"`
}

// JokeStatus defines the observed state of Joke
type JokeStatus struct {
	// Phase represents the current lifecycle state of this object
	// consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to a string when printing the property
	// is only present for display purposes, for everything else use conditions
	Phase JokePhaseType `json:"phase,omitempty"`
	// Conditions is a list of all conditions this Joke is in.
	Conditions []JokeCondition `json:"conditions,omitempty"`
	// The most recent generation observed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Selected joke represent currently selected joke
	SelectedJoke *JokeItem `json:"selectedJoke,omitempty"`
}

// JokePhaseType represents all conditions as a single string for printing in kubectl
type JokePhaseType string

// Values of JokePhaseType
const (
	JokePhaseReady    JokePhaseType = "Ready"
	JokePhaseNotReady JokePhaseType = "NotReady"
	JokePhaseUnknown  JokePhaseType = "Unknown"
)

// JokeConditionType represents a JokeCondition value.
type JokeConditionType string

const (
	// JokeReady represents a Joke condition is in ready state.
	JokeReady JokeConditionType = "Ready"
)

// JokeCondition contains details for the current condition of this Joke.
type JokeCondition struct {
	// LastTransitionTime is the last time the condition transit from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Message is the human readable message indicating details about last transition.
	Message string `json:"message"`
	// Reason is the (brief) reason for the condition's last transition.
	Reason string `json:"reason"`
	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus `json:"status"`
	// Type of the condition, currently ('Ready').
	Type JokeConditionType `json:"type"`
}

// UpdatePhase updates the phase property based on the current conditions
// this method should be called everytime the conditions are updated
func (s *JokeStatus) updatePhase() {
	for _, condition := range s.Conditions {
		if condition.Type != JokeReady {
			continue
		}

		switch condition.Status {
		case ConditionTrue:
			s.Phase = JokePhaseReady
		case ConditionFalse:
			s.Phase = JokePhaseNotReady
		case ConditionUnknown:
			s.Phase = JokePhaseUnknown
		}
		return
	}
	s.Phase = JokePhaseUnknown
}

// SetCondition replaces or adds the given condition
func (s *JokeStatus) SetCondition(condition JokeCondition) {
	defer s.updatePhase()

	for i := range s.Conditions {
		if s.Conditions[i].Type == condition.Type {

			s.Conditions[i].Status = condition.Status
			s.Conditions[i].Reason = condition.Reason
			s.Conditions[i].Message = condition.Message
			s.Conditions[i].LastTransitionTime = metav1.Now()
			return
		}
	}

	condition.LastTransitionTime = metav1.Now()
	s.Conditions = append(s.Conditions, condition)
}

// GetCondition returns the Condition of the given type, if it exists
func (s *JokeStatus) GetCondition(t JokeConditionType) (condition JokeCondition, exists bool) {
	for _, cond := range s.Conditions {
		if cond.Type == t {
			condition = cond
			exists = true
			return
		}
	}
	return
}

// Joke is core element in joke generation operator
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Text",type="string",JSONPath=".status.selectedJoke.text"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Joke struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JokeSpec   `json:"spec,omitempty"`
	Status JokeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// JokeList contains a list of Joke
type JokeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Joke `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Joke{}, &JokeList{})
}

/*
Copyright 2019 The Kubecarrier Authors.

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
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	"github.com/kubermatic/kubecarrier/pkg/apis/e2e/v1alpha2"
)

// JokeSpec defines the desired state of Joke
type JokeSpec struct {
	// Jokes holds all known jokes
	Jokes []string `json:"jokes,omitempty"`
}

// JokeStatus defines the observed state of Joke
type JokeStatus struct {
	// Phase represents the current lifecycle state of this object
	// consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to a string when printing the property
	// is only present for display purposes, for everything else use conditions
	Phase v1alpha2.JokePhaseType `json:"phase,omitempty"`
	// Conditions is a list of all conditions this Joke is in.
	Conditions []v1alpha2.JokeCondition `json:"conditions,omitempty"`
	// The most recent generation observed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// JokeText represent currently selected joke's text
	JokeText string `json:"jokeText"`
}

// UpdatePhase updates the phase property based on the current conditions
// this method should be called everytime the conditions are updated
func (s *JokeStatus) updatePhase() {
	for _, condition := range s.Conditions {
		if condition.Type != v1alpha2.JokeReady {
			continue
		}

		switch condition.Status {
		case v1alpha2.ConditionTrue:
			s.Phase = v1alpha2.JokePhaseReady
		case v1alpha2.ConditionFalse:
			s.Phase = v1alpha2.JokePhaseNotReady
		case v1alpha2.ConditionUnknown:
			s.Phase = v1alpha2.JokePhaseUnknown
		}
		return
	}
	s.Phase = v1alpha2.JokePhaseUnknown
}

// SetCondition replaces or adds the given condition
func (s *JokeStatus) SetCondition(condition v1alpha2.JokeCondition) {
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
func (s *JokeStatus) GetCondition(t v1alpha2.JokeConditionType) (condition v1alpha2.JokeCondition, exists bool) {
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
// +kubebuilder:printcolumn:name="Text",type="date",JSONPath=".status.jokeText"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Joke struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JokeSpec   `json:"spec,omitempty"`
	Status JokeStatus `json:"status,omitempty"`
}

func (src *Joke) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha2.Joke)
	dst.Status.Conditions = src.Status.Conditions
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.Phase = src.Status.Phase
	dst.Status.SelectedJoke = &v1alpha2.JokeItem{
		Text: src.Status.JokeText,
		Type: "LegacyType",
	}

	dst.Spec.JokeDatabase = nil
	for _, joke := range src.Spec.Jokes {
		dst.Spec.JokeDatabase = append(dst.Spec.JokeDatabase, v1alpha2.JokeItem{
			Text: joke,
			Type: "LegacyType",
		})
	}
	return nil
}

func (dst *Joke) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha2.Joke)
	dst.Status.Conditions = src.Status.Conditions
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.Phase = src.Status.Phase
	dst.Status.JokeText = src.Status.SelectedJoke.Text

	dst.Spec.Jokes = nil
	for _, joke := range src.Spec.JokeDatabase {
		dst.Spec.Jokes = append(dst.Spec.Jokes, joke.Text)
	}
	return nil
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

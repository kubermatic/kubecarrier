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

// TenderSpec defines the desired state of Tender
type TenderSpec struct {
	// KubeconfigSecret specifies the KubeConfig to use when connecting to the ServiceCluster.
	KubeconfigSecret ObjectReference `json:"kubeconfigSecret"`
}

// TenderStatus defines the observed state of Tender
type TenderStatus struct {
	// Phase represents the current lifecycle state of this object
	// consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to a string when printing the property
	// is only present for display purposes, for everything else use conditions
	Phase TenderPhaseType `json:"phase,omitempty"`
	// Conditions is a list of all conditions this Tender is in.
	Conditions []TenderCondition `json:"conditions,omitempty"`
	// The most recent generation observed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// TenderPhaseType represents all conditions as a single string for printing in kubectl
type TenderPhaseType string

// Values of TenderPhaseType
const (
	TenderPhaseReady       TenderPhaseType = "Ready"
	TenderPhaseNotReady    TenderPhaseType = "NotReady"
	TenderPhaseTerminating TenderPhaseType = "Terminating"
	TenderPhaseUnknown     TenderPhaseType = "Unknown"
)

// TenderConditionType represents a TenderCondition value.
type TenderConditionType string

const (
	// TenderReady represents a Tender condition is in ready state.
	TenderReady TenderConditionType = "Ready"
)

const (
	TenderTerminatingReason string = "Terminating"
)

// TenderCondition contains details for the current condition of this Tender.
type TenderCondition struct {
	// LastTransitionTime is the last time the condition transit from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Message is the human readable message indicating details about last transition.
	Message string `json:"message"`
	// Reason is the (brief) reason for the condition's last transition.
	Reason string `json:"reason"`
	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus `json:"status"`
	// Type of the condition, currently ('Ready').
	Type TenderConditionType `json:"type"`
}

// UpdatePhase updates the phase property based on the current conditions
// this method should be called everytime the conditions are updated
func (s *TenderStatus) updatePhase() {
	for _, condition := range s.Conditions {
		if condition.Type != TenderReady {
			continue
		}

		switch condition.Status {
		case ConditionTrue:
			s.Phase = TenderPhaseReady
		case ConditionFalse:
			if condition.Reason == TenderTerminatingReason {
				s.Phase = TenderPhaseTerminating
			} else {
				s.Phase = TenderPhaseNotReady
			}
		case ConditionUnknown:
			s.Phase = TenderPhaseUnknown
		}
		return
	}
	s.Phase = TenderPhaseUnknown
}

// SetCondition replaces or adds the given condition
func (s *TenderStatus) SetCondition(condition TenderCondition) {
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
func (s *TenderStatus) GetCondition(t TenderConditionType) (condition TenderCondition, exists bool) {
	for _, cond := range s.Conditions {
		if cond.Type == t {
			condition = cond
			exists = true
			return
		}
	}
	return
}

// Tender represents single kubernetes cluster belonging to the provider
//
// Tender lives in the provider namespace. For each tender the kubecarrier operator spins up
// the tender controller deployment, necessary roles, service accounts, and role bindings
//
// The reason for tender controller deployment are multiples:
// * security --> kubecarrier operator has greater privileges then tender controller
// * resource isolation --> each tender controller pod operates only on a single service cluster,
// 		thus resource allocation and monitoring is separate per tenders. This allows finer grade
// 		resource tuning and monitoring
// * flexibility --> If needed different tenders could have different deployments depending on
// 		their specific need (e.g. kubecarrier image version for gradual rolling upgrade, different resource allocation, etc),
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Tender struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TenderSpec   `json:"spec,omitempty"`
	Status TenderStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TenderList contains a list of Tender
type TenderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tender `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tender{}, &TenderList{})
}

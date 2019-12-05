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

// DinghySpec defines the desired state of Dinghy
type DinghySpec struct {
	// CRDConfiguration references the CRDConfiguration object that holds type information.
	CRDConfiguration ObjectReference `json:"crdConfiguration"`

	// Kind of internal/external CRD.
	Kind string `json:"kind"`
	// Version of internal/external CRD.
	Version string `json:"version"`
	// API Group of the internal CRD.
	InternalGroup string `json:"internalGroup"`
	// API Group of the external CRD.
	ExternalGroup string `json:"externalGroup"`
	// API Group of the CRD on the ServiceCluster.
	Group string `json:"serviceClusterGroup"`

	// ServiceCluster references the ServiceCluster that this Dinghy connects to.
	ServiceCluster ObjectReference `json:"serviceCluster"`
}

// DinghyStatus defines the observed state of Dinghy
type DinghyStatus struct {
	// DEPRECATED.
	// Phase represents the current lifecycle state of this object
	// consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to a string when printing the property
	// is only present for display purposes, for everything else use conditions
	Phase DinghyPhaseType `json:"phase,omitempty"`
	// Conditions is a list of all conditions this Dinghy is in.
	Conditions []DinghyCondition `json:"conditions,omitempty"`
	// The most recent generation observed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// DinghyPhaseType represents all conditions as a single string for printing in kubectl
type DinghyPhaseType string

// Values of DinghyPhaseType
const (
	DinghyPhaseReady       DinghyPhaseType = "Ready"
	DinghyPhaseNotReady    DinghyPhaseType = "NotReady"
	DinghyPhaseTerminating DinghyPhaseType = "Terminating"
	DinghyPhaseUnknown     DinghyPhaseType = "Unknown"
)

// DinghyConditionType represents a DinghyCondition value.
type DinghyConditionType string

const (
	// DinghyReady represents a Dinghy condition is in ready state.
	DinghyReady DinghyConditionType = "Ready"
)

const (
	// DinghyTerminatingReason signals the dinghy is being terminated
	DinghyTerminatingReason string = "Terminating"
)

// DinghyCondition contains details for the current condition of this Dinghy.
type DinghyCondition struct {
	// LastTransitionTime is the last time the condition transit from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Message is the human readable message indicating details about last transition.
	Message string `json:"message"`
	// Reason is the (brief) reason for the condition's last transition.
	Reason string `json:"reason"`
	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus `json:"status"`
	// Type of the condition, currently ('Ready').
	Type DinghyConditionType `json:"type"`
}

// UpdatePhase updates the phase property based on the current conditions
// this method should be called everytime the conditions are updated
func (s *DinghyStatus) updatePhase() {
	for _, condition := range s.Conditions {
		if condition.Type != DinghyReady {
			continue
		}

		switch condition.Status {
		case ConditionTrue:
			s.Phase = DinghyPhaseReady
		case ConditionFalse:
			if condition.Reason == DinghyTerminatingReason {
				s.Phase = DinghyPhaseTerminating
			} else {
				s.Phase = DinghyPhaseNotReady
			}
		case ConditionUnknown:
			s.Phase = DinghyPhaseUnknown
		}
		return
	}
	s.Phase = DinghyPhaseUnknown
}

// SetCondition replaces or adds the given condition
func (s *DinghyStatus) SetCondition(condition DinghyCondition) {
	defer s.updatePhase()
	if condition.LastTransitionTime.IsZero() {
		condition.LastTransitionTime = metav1.Now()
	}

	for i := range s.Conditions {
		if s.Conditions[i].Type == condition.Type {
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

// GetCondition returns the Condition of the given type, if it exists
func (s *DinghyStatus) GetCondition(t DinghyConditionType) (condition DinghyCondition, exists bool) {
	for _, cond := range s.Conditions {
		if cond.Type == t {
			condition = cond
			exists = true
			return
		}
	}
	return
}

// Dinghy is a service helping the Tender of a ServiceCluster to ferry dynamic workload types between the ServiceCluster
// and the Sponson Control Plane. Every new CustomResourceDefinition that is managed by Sponson for a Provider will have
// its own Dinghy assigned.
//
// Dinghy lives in the provider namespace. For each tender the sponson operator spins up
// the tender controller deployment, necessary roles, service accounts, and role bindings
//
// The primary motivation for this separation is separating per object type watches into dedicated dinghy service are
// known and possible issues with watches memory leaks & closing. Most of the controller-runtime doesn't handle
// informers closing ideally
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Dinghy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DinghySpec   `json:"spec,omitempty"`
	Status DinghyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DinghyList contains a list of Dinghy
type DinghyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Dinghy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Dinghy{}, &DinghyList{})
}

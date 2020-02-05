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

// TenantSpec defines the desired state of Tenant.
type TenantSpec struct{}

// TenantStatus defines the observed state of Tenant.
type TenantStatus struct {
	// NamespaceName is the name of the namespace that the Tenant manages.
	NamespaceName string `json:"namespaceName,omitempty"`
	// ObservedGeneration is the most recent generation observed for this Tenant by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions represents the latest available observations of a Tenant's current state.
	Conditions []TenantCondition `json:"conditions,omitempty"`
	// DEPRECATED.
	// Phase represents the current lifecycle state of this object.
	// Consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to strings when printing the property.
	// This is only for display purpose, for everything else use conditions.
	Phase TenantPhaseType `json:"phase,omitempty"`
}

// TenantPhaseType represents all conditions as a single string for printing by using kubectl commands.
type TenantPhaseType string

// Values of TenantPhaseType.
const (
	TenantPhaseReady       TenantPhaseType = "Ready"
	TenantPhaseNotReady    TenantPhaseType = "NotReady"
	TenantPhaseUnknown     TenantPhaseType = "Unknown"
	TenantPhaseTerminating TenantPhaseType = "Terminating"
)

const (
	TenantTerminatingReason = "Deleting"
)

// updatePhase updates the phase property based on the current conditions
// this method should be called every time the conditions are updated.
func (s *TenantStatus) updatePhase() {
	for _, condition := range s.Conditions {
		if condition.Type != TenantReady {
			continue
		}

		switch condition.Status {
		case ConditionTrue:
			s.Phase = TenantPhaseReady
		case ConditionFalse:
			if condition.Reason == TenantTerminatingReason {
				s.Phase = TenantPhaseTerminating
			} else {
				s.Phase = TenantPhaseNotReady
			}
		case ConditionUnknown:
			s.Phase = TenantPhaseUnknown
		}
		return
	}

	s.Phase = TenantPhaseUnknown
}

// TenantConditionType represents a TenantCondition value.
type TenantConditionType string

const (
	// TenantReady represents a Tenant condition is in ready state.
	TenantReady TenantConditionType = "Ready"
)

// TenantCondition contains details for the current condition of this Tenant.
type TenantCondition struct {
	// Type is the type of the Tenant condition, currently ('Ready').
	Type TenantConditionType `json:"type"`
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
func (c TenantCondition) True() bool {
	return c.Status == ConditionTrue
}

// GetCondition returns the Condition of the given condition type, if it exists.
func (s *TenantStatus) GetCondition(t TenantConditionType) (condition TenantCondition, exists bool) {
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
func (s *TenantStatus) SetCondition(condition TenantCondition) {
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

// Tenant sets up permissions and references to allow a end-user group to interact with providers' services.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:shortName=tn
type Tenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TenantSpec   `json:"spec,omitempty"`
	Status TenantStatus `json:"status,omitempty"`
}

// IsReady returns if the Tenant is ready.
func (s *Tenant) IsReady() bool {
	if s.Generation != s.Status.ObservedGeneration {
		return false
	}

	for _, condition := range s.Status.Conditions {
		if condition.Type == TenantReady &&
			condition.Status == ConditionTrue {
			return true
		}
	}
	return false
}

// TenantList contains a list of Tenant.
// +kubebuilder:object:root=true
type TenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tenant `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tenant{}, &TenantList{})
}

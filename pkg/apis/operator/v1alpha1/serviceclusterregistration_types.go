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

// ServiceClusterRegistrationSpec defines the desired state of ServiceClusterRegistration
type ServiceClusterRegistrationSpec struct {
	// KubeconfigSecret specifies the Kubeconfig to use when connecting to the ServiceCluster.
	KubeconfigSecret ObjectReference `json:"kubeconfigSecret"`
}

// ServiceClusterRegistrationStatus defines the observed state of ServiceClusterRegistration
type ServiceClusterRegistrationStatus struct {
	// DEPRECATED.
	// Phase represents the current lifecycle state of this object.
	// Consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to strings when printing the property.
	// This is only for display purpose, for everything else use conditions.
	Phase ServiceClusterRegistrationPhaseType `json:"phase,omitempty"`
	// Conditions is a list of all conditions this ServiceClusterRegistration is in.
	Conditions []ServiceClusterRegistrationCondition `json:"conditions,omitempty"`
	// The most recent generation observed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// ServiceClusterRegistrationPhaseType represents all conditions as a single string for printing in kubectl
type ServiceClusterRegistrationPhaseType string

// Values of ServiceClusterRegistrationPhaseType
const (
	ServiceClusterRegistrationPhaseReady       ServiceClusterRegistrationPhaseType = "Ready"
	ServiceClusterRegistrationPhaseNotReady    ServiceClusterRegistrationPhaseType = "NotReady"
	ServiceClusterRegistrationPhaseTerminating ServiceClusterRegistrationPhaseType = "Terminating"
	ServiceClusterRegistrationPhaseUnknown     ServiceClusterRegistrationPhaseType = "Unknown"
)

// ServiceClusterRegistrationConditionType represents a ServiceClusterRegistrationCondition value.
type ServiceClusterRegistrationConditionType string

const (
	// ServiceClusterRegistrationReady represents a ServiceClusterRegistration condition is in ready state.
	ServiceClusterRegistrationReady ServiceClusterRegistrationConditionType = "Ready"
)

const (
	ServiceClusterRegistrationTerminatingReason string = "Terminating"
)

// ServiceClusterRegistrationCondition contains details for the current condition of this ServiceClusterRegistration.
type ServiceClusterRegistrationCondition struct {
	// LastTransitionTime is the last time the condition transit from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Message is the human readable message indicating details about last transition.
	Message string `json:"message"`
	// Reason is the (brief) reason for the condition's last transition.
	Reason string `json:"reason"`
	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus `json:"status"`
	// Type of the condition, currently ('Ready').
	Type ServiceClusterRegistrationConditionType `json:"type"`
}

// True returns whether .Status == "True"
func (c ServiceClusterRegistrationCondition) True() bool {
	return c.Status == ConditionTrue
}

// UpdatePhase updates the phase property based on the current conditions
// this method should be called everytime the conditions are updated
func (s *ServiceClusterRegistrationStatus) updatePhase() {
	for _, condition := range s.Conditions {
		if condition.Type != ServiceClusterRegistrationReady {
			continue
		}

		switch condition.Status {
		case ConditionTrue:
			s.Phase = ServiceClusterRegistrationPhaseReady
		case ConditionFalse:
			if condition.Reason == ServiceClusterRegistrationTerminatingReason {
				s.Phase = ServiceClusterRegistrationPhaseTerminating
			} else {
				s.Phase = ServiceClusterRegistrationPhaseNotReady
			}
		case ConditionUnknown:
			s.Phase = ServiceClusterRegistrationPhaseUnknown
		}
		return
	}
	s.Phase = ServiceClusterRegistrationPhaseUnknown
}

// SetCondition replaces or adds the given condition
func (s *ServiceClusterRegistrationStatus) SetCondition(condition ServiceClusterRegistrationCondition) {
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
func (s *ServiceClusterRegistrationStatus) GetCondition(t ServiceClusterRegistrationConditionType) (condition ServiceClusterRegistrationCondition, exists bool) {
	for _, cond := range s.Conditions {
		if cond.Type == t {
			condition = cond
			exists = true
			return
		}
	}
	return
}

// ServiceClusterRegistration represents single kubernetes cluster belonging to the provider
//
// ServiceClusterRegistration lives in the provider namespace. For each ferry the kubecarrier operator spins up
// the ferry controller deployment, necessary roles, service accounts, and role bindings
//
// The reason for ferry controller deployment are multiples:
// * security --> kubecarrier operator has greater privileges then ferry controller
// * resource isolation --> each ferry controller pod operates only on a single service cluster,
// 		thus resource allocation and monitoring is separate per ferrys. This allows finer grade
// 		resource tuning and monitoring
// * flexibility --> If needed different ferrys could have different deployments depending on
// 		their specific need (e.g. kubecarrier image version for gradual rolling upgrade, different resource allocation, etc),
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type ServiceClusterRegistration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceClusterRegistrationSpec   `json:"spec,omitempty"`
	Status ServiceClusterRegistrationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ServiceClusterRegistrationList contains a list of ServiceClusterRegistration
type ServiceClusterRegistrationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceClusterRegistration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceClusterRegistration{}, &ServiceClusterRegistrationList{})
}

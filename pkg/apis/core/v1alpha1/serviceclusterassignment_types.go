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

// ServiceClusterAssignmentSpec describes the desired state of ServiceClusterAssignment.
type ServiceClusterAssignmentSpec struct {
	// References the ServiceCluster.
	ServiceCluster ObjectReference `json:"serviceCluster"`
	// References the source namespace in the management cluster.
	ManagementClusterNamespace ObjectReference `json:"managementNamespace"`
}

// ServiceClusterAssignmentStatus represents the observed state of ServiceClusterAssignment.
type ServiceClusterAssignmentStatus struct {
	// DEPRECATED.
	// Phase represents the current lifecycle state of this object
	// consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to a string when printing the property
	// is only present for display purposes, for everything else use conditions
	Phase ServiceClusterAssignmentPhaseType `json:"phase,omitempty"`
	// Conditions is a list of all conditions this ServiceClusterAssignment is in.
	Conditions []ServiceClusterAssignmentCondition `json:"conditions,omitempty"`
	// The most recent generation observed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// ServiceClusterNamespace references the Namespace on the ServiceCluster that was assigned.
	ServiceClusterNamespace *ObjectReference `json:"serviceClusterNamespace,omitempty"`
}

// ServiceClusterAssignmentPhaseType represents all conditions as a single string for printing in kubectl.
type ServiceClusterAssignmentPhaseType string

// Values of ServiceClusterAssignmentPhaseType
const (
	ServiceClusterAssignmentPhaseReady    ServiceClusterAssignmentPhaseType = "Ready"
	ServiceClusterAssignmentPhaseNotReady ServiceClusterAssignmentPhaseType = "NotReady"
	ServiceClusterAssignmentPhaseUnknown  ServiceClusterAssignmentPhaseType = "Unknown"
)

// ServiceClusterAssignmentConditionType represents a ServiceClusterAssignmentCondition value.
type ServiceClusterAssignmentConditionType string

const (
	// ServiceClusterAssignmentReady represents a ServiceClusterAssignment condition is in ready state.
	ServiceClusterAssignmentReady ServiceClusterAssignmentConditionType = "Ready"
)

// ServiceClusterAssignmentCondition contains details for the current condition of this ServiceClusterAssignment.
type ServiceClusterAssignmentCondition struct {
	// LastTransitionTime is the last time the condition transit from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Message is the human readable message indicating details about last transition.
	Message string `json:"message"`
	// Reason is the (brief) reason for the condition's last transition.
	Reason string `json:"reason"`
	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus `json:"status"`
	// Type of the condition, currently ('Ready').
	Type ServiceClusterAssignmentConditionType `json:"type"`
}

// True returns whether .Status == "True"
func (c ServiceClusterAssignmentCondition) True() bool {
	return c.Status == ConditionTrue
}

// updatePhase updates the phase property based on the current conditions
// this method should be called every time the conditions are updated
func (s *ServiceClusterAssignmentStatus) updatePhase() {
	for _, condition := range s.Conditions {
		if condition.Type != ServiceClusterAssignmentReady {
			continue
		}

		switch condition.Status {
		case ConditionTrue:
			s.Phase = ServiceClusterAssignmentPhaseReady
		case ConditionFalse:
			s.Phase = ServiceClusterAssignmentPhaseNotReady
		case ConditionUnknown:
			s.Phase = ServiceClusterAssignmentPhaseUnknown
		}
		return
	}
	s.Phase = ServiceClusterAssignmentPhaseUnknown
}

// SetCondition replaces or adds the given condition
func (s *ServiceClusterAssignmentStatus) SetCondition(condition ServiceClusterAssignmentCondition) {
	defer s.updatePhase()

	if condition.LastTransitionTime.IsZero() {
		condition.LastTransitionTime = metav1.Now()
	}

	for i := range s.Conditions {
		if s.Conditions[i].Type == condition.Type {
			// Only change the LastTransitionTime when Status changed
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
func (s *ServiceClusterAssignmentStatus) GetCondition(conditionType ServiceClusterAssignmentConditionType) (condition ServiceClusterAssignmentCondition, exists bool) {
	for _, cond := range s.Conditions {
		if cond.Type == conditionType {
			condition = cond
			exists = true
			return
		}
	}
	return
}

// ServiceClusterAssignment is assigning a Namespace in the Management cluster with a Namespace on the ServiceCluster.
//
// The Namespace in the ServiceCluster will be created automatically and is reported in the instance status.
//
// **Example**
// ```yaml
// apiVersion: kubecarrier.io/v1alpha1
// kind: ServiceClusterAssignment
// metadata:
//   name: example1.eu-west-1
// spec:
//   serviceCluster:
//     name: eu-west-1
//   managementNamespace:
//     name: example1
// ```
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:shortName=sca
type ServiceClusterAssignment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceClusterAssignmentSpec   `json:"spec,omitempty"`
	Status ServiceClusterAssignmentStatus `json:"status,omitempty"`
}

// IsReady returns if the ServiceClusterAssignment is ready.
func (s *ServiceClusterAssignment) IsReady() bool {
	if s.Generation != s.Status.ObservedGeneration {
		return false
	}

	for _, condition := range s.Status.Conditions {
		if condition.Type == ServiceClusterAssignmentReady &&
			condition.Status == ConditionTrue {
			return true
		}
	}
	return false
}

// ServiceClusterAssignmentList contains a list of ServiceClusterAssignment.
// +kubebuilder:object:root=true
type ServiceClusterAssignmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceClusterAssignment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceClusterAssignment{}, &ServiceClusterAssignmentList{})
}

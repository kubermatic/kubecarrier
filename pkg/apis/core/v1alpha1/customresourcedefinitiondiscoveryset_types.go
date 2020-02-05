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

type CustomResourceDefinitionDiscoverySetSpec struct {
	// CRD references a CustomResourceDefinition within the ServiceCluster.
	CRD ObjectReference `json:"crd"`
	// ServiceClusterSelector references a set of ServiceClusters to search the CustomResourceDefinition on.
	ServiceClusterSelector metav1.LabelSelector `json:"serviceClusterSelector"`
	// KindOverride overrides resulting internal CRDs kind
	KindOverride string `json:"kindOverride,omitempty"`
}

type CustomResourceDefinitionDiscoverySetStatus struct {
	// DEPRECATED.
	// Phase represents the current lifecycle state of this object
	// consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to a string when printing the property
	// is only present for display purposes, for everything else use conditions
	Phase CustomResourceDefinitionDiscoverySetPhaseType `json:"phase,omitempty"`
	// Conditions is a list of all conditions this CustomResourceDefinitionDiscovery is in.
	Conditions []CustomResourceDefinitionDiscoverySetCondition `json:"conditions,omitempty"`
	// The most recent generation observed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// CustomResourceDefinitionDiscoverySetPhaseType represents all conditions as a single string for printing in kubectl
type CustomResourceDefinitionDiscoverySetPhaseType string

// Values of CustomResourceDefinitionDiscoverySetPhaseType
const (
	CustomResourceDefinitionDiscoverySetPhaseReady    CustomResourceDefinitionDiscoverySetPhaseType = "Ready"
	CustomResourceDefinitionDiscoverySetPhaseNotReady CustomResourceDefinitionDiscoverySetPhaseType = "NotReady"
	CustomResourceDefinitionDiscoverySetPhaseUnknown  CustomResourceDefinitionDiscoverySetPhaseType = "Unknown"
)

// updatePhase updates the phase property based on the current conditions
// this method should be called everytime the conditions are updated
func (s *CustomResourceDefinitionDiscoverySetStatus) updatePhase() {
	for _, condition := range s.Conditions {
		if condition.Type != CustomResourceDefinitionDiscoverySetReady {
			continue
		}
		switch condition.Status {
		case ConditionTrue:
			s.Phase = CustomResourceDefinitionDiscoverySetPhaseReady
		case ConditionFalse:
			s.Phase = CustomResourceDefinitionDiscoverySetPhaseNotReady
		default:
			s.Phase = CustomResourceDefinitionDiscoverySetPhaseUnknown
		}
		return
	}

	s.Phase = CustomResourceDefinitionDiscoverySetPhaseUnknown
}

// SetCondition replaces or adds the given condition
func (s *CustomResourceDefinitionDiscoverySetStatus) SetCondition(condition CustomResourceDefinitionDiscoverySetCondition) {
	defer s.updatePhase()
	if condition.LastTransitionTime.IsZero() {
		condition.LastTransitionTime = metav1.Now()
	}

	for i := range s.Conditions {
		if s.Conditions[i].Type == condition.Type {
			if s.Conditions[i].Status != condition.Status {
				s.Conditions[i].LastTransitionTime = metav1.Now()
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
func (s *CustomResourceDefinitionDiscoverySetStatus) GetCondition(t CustomResourceDefinitionDiscoverySetConditionType) (condition CustomResourceDefinitionDiscoverySetCondition, exists bool) {
	for _, cond := range s.Conditions {
		if cond.Type == t {
			condition = cond
			exists = true
			return
		}
	}
	return
}

// CustomResourceDefinitionDiscoverySetConditionType represents a CustomResourceDefinitionDiscoverySetCondition value.
type CustomResourceDefinitionDiscoverySetConditionType string

const (
	// CustomResourceDefinitionDiscoverySetReady is True when all CRDDs are ready.
	CustomResourceDefinitionDiscoverySetReady CustomResourceDefinitionDiscoverySetConditionType = "Ready"
)

// CustomResourceDefinitionDiscoverySetCondition contains details for the current condition of this CustomResourceDefinitionDiscoverySet.
type CustomResourceDefinitionDiscoverySetCondition struct {
	// LastTransitionTime is the last time the condition transit from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Message is the human readable message indicating details about last transition.
	Message string `json:"message"`
	// Reason is the (brief) reason for the condition's last transition.
	Reason string `json:"reason"`
	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus `json:"status"`
	// Type of the condition, currently ('Ready').
	Type CustomResourceDefinitionDiscoverySetConditionType `json:"type"`
}

// CustomResourceDefinitionDiscoverySet manages multiple CustomResourceDefinitionDiscovery objects for a set of service clusters.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="CustomResourceDefinition",type="string",JSONPath=".spec.crd.name"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
type CustomResourceDefinitionDiscoverySet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomResourceDefinitionDiscoverySetSpec   `json:"spec,omitempty"`
	Status CustomResourceDefinitionDiscoverySetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type CustomResourceDefinitionDiscoverySetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomResourceDefinitionDiscoverySet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CustomResourceDefinitionDiscoverySet{}, &CustomResourceDefinitionDiscoverySetList{})
}

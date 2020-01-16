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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CustomResourceDefinitionDiscoverySpec defines the desired state of crdreference
type CustomResourceDefinitionDiscoverySpec struct {
	// CRD references a CustomResourceDefinition within the ServiceCluster.
	CRD ObjectReference `json:"crd"`
	// ServiceCluster references a ServiceCluster to search the CustomResourceDefinition on.
	ServiceCluster ObjectReference `json:"serviceCluster"`
}

// CustomResourceDefinitionDiscoveryStatus defines the observed state of crdreference
type CustomResourceDefinitionDiscoveryStatus struct {
	// CRD defines the original CustomResourceDefinition specification from the service cluster
	// +kubebuilder:pruning:PreserveUnknownFields
	CRD *apiextensionsv1.CustomResourceDefinition `json:"crd,omitempty"`
	// DEPRECATED.
	// Phase represents the current lifecycle state of this object
	// consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to a string when printing the property
	// is only present for display purposes, for everything else use conditions
	Phase CustomResourceDefinitionDiscoveryPhaseType `json:"phase,omitempty"`
	// Conditions is a list of all conditions this CustomResourceDefinitionDiscovery is in.
	Conditions []CustomResourceDefinitionDiscoveryCondition `json:"conditions,omitempty"`
	// The most recent generation observed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// CustomResourceDefinitionDiscoveryPhaseType represents all conditions as a single string for printing in kubectl
type CustomResourceDefinitionDiscoveryPhaseType string

// Values of CustomResourceDefinitionDiscoveryPhaseType
const (
	CustomResourceDefinitionDiscoveryPhaseReady    CustomResourceDefinitionDiscoveryPhaseType = "Ready"
	CustomResourceDefinitionDiscoveryPhaseNotReady CustomResourceDefinitionDiscoveryPhaseType = "NotReady"
	CustomResourceDefinitionDiscoveryPhaseUnknown  CustomResourceDefinitionDiscoveryPhaseType = "Unknown"
)

// updatePhase updates the phase property based on the current conditions
// this method should be called everytime the conditions are updated
func (s *CustomResourceDefinitionDiscoveryStatus) updatePhase() {
	for _, condition := range s.Conditions {
		if condition.Type != CustomResourceDefinitionDiscoveryReady {
			continue
		}
		switch condition.Status {
		case ConditionTrue:
			s.Phase = CustomResourceDefinitionDiscoveryPhaseReady
		case ConditionFalse:
			s.Phase = CustomResourceDefinitionDiscoveryPhaseNotReady
		default:
			s.Phase = CustomResourceDefinitionDiscoveryPhaseUnknown
		}
		return
	}

	s.Phase = CustomResourceDefinitionDiscoveryPhaseUnknown
}

// SetCondition replaces or adds the given condition
func (s *CustomResourceDefinitionDiscoveryStatus) SetCondition(condition CustomResourceDefinitionDiscoveryCondition) {
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
func (s *CustomResourceDefinitionDiscoveryStatus) GetCondition(t CustomResourceDefinitionDiscoveryConditionType) (condition CustomResourceDefinitionDiscoveryCondition, exists bool) {
	for _, cond := range s.Conditions {
		if cond.Type == t {
			condition = cond
			exists = true
			return
		}
	}
	return
}

// CustomResourceDefinitionDiscoveryConditionType represents a CustomResourceDefinitionDiscoveryCondition value.
type CustomResourceDefinitionDiscoveryConditionType string

const (
	// CustomResourceDefinitionDiscoveryReady represents a CustomResourceDefinitionDiscovery condition is in ready state.
	CustomResourceDefinitionDiscoveryReady CustomResourceDefinitionDiscoveryConditionType = "Ready"
)

// CustomResourceDefinitionDiscoveryCondition contains details for the current condition of this CustomResourceDefinitionDiscovery.
type CustomResourceDefinitionDiscoveryCondition struct {
	// LastTransitionTime is the last time the condition transit from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Message is the human readable message indicating details about last transition.
	Message string `json:"message"`
	// Reason is the (brief) reason for the condition's last transition.
	Reason string `json:"reason"`
	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus `json:"status"`
	// Type of the condition, currently ('Ready').
	Type CustomResourceDefinitionDiscoveryConditionType `json:"type"`
}

// CustomResourceDefinitionDiscovery is used inside KubeCarrier to fetch a CustomResourceDefinition from another cluster and to offload cross cluster access to another component.
// +kubebuilder:subresource:status
// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="CustomResourceDefinition",type="string",JSONPath=".spec.crd.metadata.name"
// +kubebuilder:printcolumn:name="Service Cluster",type="string",JSONPath=".spec.serviceCluster.name"
type CustomResourceDefinitionDiscovery struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomResourceDefinitionDiscoverySpec   `json:"spec,omitempty"`
	Status CustomResourceDefinitionDiscoveryStatus `json:"status,omitempty"`
}

// CustomResourceDefinitionDiscoveryList contains a list of crdreference
// +kubebuilder:object:root=true
type CustomResourceDefinitionDiscoveryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomResourceDefinitionDiscovery `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CustomResourceDefinitionDiscovery{}, &CustomResourceDefinitionDiscoveryList{})
}

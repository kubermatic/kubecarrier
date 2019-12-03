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
	"k8s.io/apimachinery/pkg/runtime"
)

// CRDReferenceSpec defines the desired state of crdreference
type CRDReferenceSpec struct {
	// CRD references a CustomResourceDefinition within the ServiceCluster.
	CRD ObjectReference `json:"crd"`
	// ServiceCluster references a ServiceCluster to search the CRD on.
	ServiceCluster ObjectReference `json:"serviceCluster"`
}

// CRDReferenceStatus defines the observed state of crdreference
type CRDReferenceStatus struct {
	// CRDSpec defines the original CRD specification from the service cluster
	CRDSpec *runtime.RawExtension `json:"crdSpec,omitempty"`
	// Phase represents the current lifecycle state of this object
	// consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to a string when printing the property
	// is only present for display purposes, for everything else use conditions
	Phase CRDReferencePhaseType `json:"phase,omitempty"`
	// Conditions is a list of all conditions this CRDReference is in.
	Conditions []CRDReferenceCondition `json:"conditions,omitempty"`
	// The most recent generation observed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// CRDReferencePhaseType represents all conditions as a single string for printing in kubectl
type CRDReferencePhaseType string

// Values of CRDReferencePhaseType
const (
	CRDReferencePhaseReady    CRDReferencePhaseType = "Ready"
	CRDReferencePhaseNotReady CRDReferencePhaseType = "NotReady"
	CRDReferencePhaseUnknown  CRDReferencePhaseType = "Unknown"
)

// updatePhase updates the phase property based on the current conditions
// this method should be called everytime the conditions are updated
func (s *CRDReferenceStatus) updatePhase() {
	for _, condition := range s.Conditions {
		if condition.Type != CRDReferenceReady {
			continue
		}
		switch condition.Status {
		case ConditionTrue:
			s.Phase = CRDReferencePhaseReady
		case ConditionFalse:
			s.Phase = CRDReferencePhaseNotReady
		default:
			s.Phase = CRDReferencePhaseUnknown
		}
		return
	}

	s.Phase = CRDReferencePhaseUnknown
}

// SetCondition replaces or adds the given condition
func (s *CRDReferenceStatus) SetCondition(condition CRDReferenceCondition) {
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
func (s *CRDReferenceStatus) GetCondition(t CRDReferenceConditionType) (condition CRDReferenceCondition, exists bool) {
	for _, cond := range s.Conditions {
		if cond.Type == t {
			condition = cond
			exists = true
			return
		}
	}
	return
}

// CRDReferenceConditionType represents a CRDReferenceCondition value.
type CRDReferenceConditionType string

const (
	// CRDReferenceReady represents a CRDReference condition is in ready state.
	CRDReferenceReady CRDReferenceConditionType = "Ready"
)

// CRDReferenceCondition contains details for the current condition of this CRDReference.
type CRDReferenceCondition struct {
	// LastTransitionTime is the last time the condition transit from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Message is the human readable message indicating details about last transition.
	Message string `json:"message"`
	// Reason is the (brief) reason for the condition's last transition.
	Reason string `json:"reason"`
	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus `json:"status"`
	// Type of the condition, currently ('Ready').
	Type CRDReferenceConditionType `json:"type"`
}

// CRDReference is used inside KubeCarrier to fetch a CRD from another cluster and to offload cross cluster access to another component.
// +kubebuilder:subresource:status
// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="CRD",type="string",JSONPath=".spec.crd.name"
// +kubebuilder:printcolumn:name="Service Cluster",type="string",JSONPath=".spec.serviceCluster.name"
type CRDReference struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CRDReferenceSpec   `json:"spec,omitempty"`
	Status CRDReferenceStatus `json:"status,omitempty"`
}

// CRDReferenceList contains a list of crdreference
// +kubebuilder:object:root=true
type CRDReferenceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CRDReference `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CRDReference{}, &CRDReferenceList{})
}

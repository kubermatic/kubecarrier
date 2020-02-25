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

// CustomResourceDiscoverySpec defines the desired state of crdreference
type CustomResourceDiscoverySpec struct {
	// CRD references a CustomResourceDefinition within the ServiceCluster.
	CRD ObjectReference `json:"crd"`
	// ServiceCluster references a ServiceCluster to search the CustomResourceDefinition on.
	ServiceCluster ObjectReference `json:"serviceCluster"`
	// KindOverride overrides resulting internal CRDs kind
	KindOverride string `json:"kindOverride,omitempty"`
}

// CustomResourceDiscoveryStatus defines the observed state of crdreference
type CustomResourceDiscoveryStatus struct {
	// CRD defines the original CustomResourceDefinition specification from the service cluster
	// +kubebuilder:pruning:PreserveUnknownFields
	CRD *apiextensionsv1.CustomResourceDefinition `json:"crd,omitempty"`
	// DEPRECATED.
	// Phase represents the current lifecycle state of this object
	// consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to a string when printing the property
	// is only present for display purposes, for everything else use conditions
	Phase CustomResourceDiscoveryPhaseType `json:"phase,omitempty"`
	// Conditions is a list of all conditions this CustomResourceDiscovery is in.
	Conditions []CustomResourceDiscoveryCondition `json:"conditions,omitempty"`
	// The most recent generation observed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// CustomResourceDiscoveryPhaseType represents all conditions as a single string for printing in kubectl
type CustomResourceDiscoveryPhaseType string

// Values of CustomResourceDiscoveryPhaseType
const (
	CustomResourceDiscoveryPhaseReady    CustomResourceDiscoveryPhaseType = "Ready"
	CustomResourceDiscoveryPhaseNotReady CustomResourceDiscoveryPhaseType = "NotReady"
	CustomResourceDiscoveryPhaseUnknown  CustomResourceDiscoveryPhaseType = "Unknown"
)

// updatePhase updates the phase property based on the current conditions
// this method should be called everytime the conditions are updated
func (s *CustomResourceDiscoveryStatus) updatePhase() {
	for _, condition := range s.Conditions {
		if condition.Type != CustomResourceDiscoveryReady {
			continue
		}
		switch condition.Status {
		case ConditionTrue:
			s.Phase = CustomResourceDiscoveryPhaseReady
		case ConditionFalse:
			s.Phase = CustomResourceDiscoveryPhaseNotReady
		default:
			s.Phase = CustomResourceDiscoveryPhaseUnknown
		}
		return
	}

	s.Phase = CustomResourceDiscoveryPhaseUnknown
}

// SetCondition replaces or adds the given condition
func (s *CustomResourceDiscoveryStatus) SetCondition(condition CustomResourceDiscoveryCondition) {
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
func (s *CustomResourceDiscoveryStatus) GetCondition(t CustomResourceDiscoveryConditionType) (condition CustomResourceDiscoveryCondition, exists bool) {
	for _, cond := range s.Conditions {
		if cond.Type == t {
			condition = cond
			exists = true
			return
		}
	}
	return
}

// CustomResourceDiscoveryConditionType represents a CustomResourceDiscoveryCondition value.
type CustomResourceDiscoveryConditionType string

const (
	// CustomResourceDiscoveryReady represents a CustomResourceDiscovery condition is in ready state.
	CustomResourceDiscoveryReady CustomResourceDiscoveryConditionType = "Ready"
	// CustomResourceDiscoveryDiscovered represents a CustomResourceDiscovery has been discovered by the manager controller.
	CustomResourceDiscoveryDiscovered CustomResourceDiscoveryConditionType = "Discovered"
	// CustomResourceDiscoveryEstablished is True if the crd could be registered in the management cluster and is now served by the kube-apiserver.
	CustomResourceDiscoveryEstablished CustomResourceDiscoveryConditionType = "Established"
	// CustomResourceDiscoveryControllerReady is Ture if the controller to propagate the crd into the service cluster is ready.
	CustomResourceDiscoveryControllerReady CustomResourceDiscoveryConditionType = "ControllerReady"
)

// CustomResourceDiscoveryCondition contains details for the current condition of this CustomResourceDiscovery.
type CustomResourceDiscoveryCondition struct {
	// LastTransitionTime is the last time the condition transit from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Message is the human readable message indicating details about last transition.
	Message string `json:"message"`
	// Reason is the (brief) reason for the condition's last transition.
	Reason string `json:"reason"`
	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus `json:"status"`
	// Type of the condition, currently ('Ready').
	Type CustomResourceDiscoveryConditionType `json:"type"`
}

// True returns whether .Status == "True"
func (c CustomResourceDiscoveryCondition) True() bool {
	return c.Status == ConditionTrue
}

// CustomResourceDiscovery is used inside KubeCarrier to fetch a CustomResourceDefinition from another cluster and to offload cross cluster access to another component.
// +kubebuilder:subresource:status
// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="CustomResourceDefinition",type="string",JSONPath=".spec.crd.name"
// +kubebuilder:printcolumn:name="Service Cluster",type="string",JSONPath=".spec.serviceCluster.name"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:resource:shortName=crdis
type CustomResourceDiscovery struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomResourceDiscoverySpec   `json:"spec,omitempty"`
	Status CustomResourceDiscoveryStatus `json:"status,omitempty"`
}

// CustomResourceDiscoveryList contains a list of crdreference
// +kubebuilder:object:root=true
type CustomResourceDiscoveryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomResourceDiscovery `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CustomResourceDiscovery{}, &CustomResourceDiscoveryList{})
}

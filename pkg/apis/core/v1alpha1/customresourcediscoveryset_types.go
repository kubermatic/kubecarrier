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

type CustomResourceDiscoverySetSpec struct {
	// CRD references a CustomResourceDefinition within the ServiceCluster.
	CRD ObjectReference `json:"crd"`
	// ServiceClusterSelector references a set of ServiceClusters to search the CustomResourceDefinition on.
	ServiceClusterSelector metav1.LabelSelector `json:"serviceClusterSelector"`
	// KindOverride overrides resulting internal CRDs kind
	KindOverride string `json:"kindOverride,omitempty"`
	// WebhookStrategy configs the webhooks of the CRDs which are registered in the management cluster by this CustomResourceDiscoverySet.
	// There are two possible values for this configuration {None (by default), ServiceCluster}
	// None (by default): Webhook will only check if there is an available ServiceClusterAssignment in the current Namespace.
	// ServiceCluster: Webhook will call webhooks of the CRD in the ServiceCluster with dry-run flag.
	// +kubebuilder:default:=None
	WebhookStrategy WebhookStrategyType `json:"webhookStrategy,omitempty"`
}

type CustomResourceDiscoverySetStatus struct {
	// ManagementClusterCRDs contains the CRDs information that created by the CustomResourceDiscovery objects of this CustomResourceDiscoverySet.
	ManagementClusterCRDs []ObjectReference `json:"managementClusterCRDs,omitempty"`
	// DEPRECATED.
	// Phase represents the current lifecycle state of this object
	// consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to a string when printing the property
	// is only present for display purposes, for everything else use conditions
	Phase CustomResourceDiscoverySetPhaseType `json:"phase,omitempty"`
	// Conditions is a list of all conditions this CustomResourceDiscovery is in.
	Conditions []CustomResourceDiscoverySetCondition `json:"conditions,omitempty"`
	// The most recent generation observed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// CustomResourceDiscoverySetPhaseType represents all conditions as a single string for printing in kubectl
type CustomResourceDiscoverySetPhaseType string

// Values of CustomResourceDiscoverySetPhaseType
const (
	CustomResourceDiscoverySetPhaseReady    CustomResourceDiscoverySetPhaseType = "Ready"
	CustomResourceDiscoverySetPhaseNotReady CustomResourceDiscoverySetPhaseType = "NotReady"
	CustomResourceDiscoverySetPhaseUnknown  CustomResourceDiscoverySetPhaseType = "Unknown"
)

// updatePhase updates the phase property based on the current conditions
// this method should be called everytime the conditions are updated
func (s *CustomResourceDiscoverySetStatus) updatePhase() {
	for _, condition := range s.Conditions {
		if condition.Type != CustomResourceDiscoverySetReady {
			continue
		}
		switch condition.Status {
		case ConditionTrue:
			s.Phase = CustomResourceDiscoverySetPhaseReady
		case ConditionFalse:
			s.Phase = CustomResourceDiscoverySetPhaseNotReady
		default:
			s.Phase = CustomResourceDiscoverySetPhaseUnknown
		}
		return
	}

	s.Phase = CustomResourceDiscoverySetPhaseUnknown
}

// SetCondition replaces or adds the given condition
func (s *CustomResourceDiscoverySetStatus) SetCondition(condition CustomResourceDiscoverySetCondition) {
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
func (s *CustomResourceDiscoverySetStatus) GetCondition(t CustomResourceDiscoverySetConditionType) (condition CustomResourceDiscoverySetCondition, exists bool) {
	for _, cond := range s.Conditions {
		if cond.Type == t {
			condition = cond
			exists = true
			return
		}
	}
	return
}

// CustomResourceDiscoverySetConditionType represents a CustomResourceDiscoverySetCondition value.
type CustomResourceDiscoverySetConditionType string

const (
	// CustomResourceDiscoverySetReady is True when all CRDiscoveries are ready.
	CustomResourceDiscoverySetReady CustomResourceDiscoverySetConditionType = "Ready"
)

// CustomResourceDiscoverySetCondition contains details for the current condition of this CustomResourceDiscoverySet.
type CustomResourceDiscoverySetCondition struct {
	// LastTransitionTime is the last time the condition transit from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Message is the human readable message indicating details about last transition.
	Message string `json:"message"`
	// Reason is the (brief) reason for the condition's last transition.
	Reason string `json:"reason"`
	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus `json:"status"`
	// Type of the condition, currently ('Ready').
	Type CustomResourceDiscoverySetConditionType `json:"type"`
}

// CustomResourceDiscoverySet manages multiple CustomResourceDiscovery objects for a set of service clusters.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="CustomResourceDefinition",type="string",JSONPath=".spec.crd.name"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:resource:categories=kubecarrier-provider,shortName=crdisset
type CustomResourceDiscoverySet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomResourceDiscoverySetSpec   `json:"spec,omitempty"`
	Status CustomResourceDiscoverySetStatus `json:"status,omitempty"`
}

// IsReady returns if the CustomResourceDiscoverySet is ready.
func (s *CustomResourceDiscoverySet) IsReady() bool {
	if s.Generation != s.Status.ObservedGeneration {
		return false
	}

	for _, condition := range s.Status.Conditions {
		if condition.Type == CustomResourceDiscoverySetReady &&
			condition.Status == ConditionTrue {
			return true
		}
	}
	return false
}

// +kubebuilder:object:root=true
type CustomResourceDiscoverySetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomResourceDiscoverySet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CustomResourceDiscoverySet{}, &CustomResourceDiscoverySetList{})
}

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
	"k8s.io/apimachinery/pkg/version"
)

// ServiceClusterSpec describes the desired state of a ServiceCluster.
type ServiceClusterSpec struct {
	// Metadata for display in the Service Hub.
	Metadata ServiceClusterMetadata `json:"metadata,omitempty"`
	// KubeconfigSecret specifies the Kubeconfig to use when connecting to the ServiceCluster.
	KubeconfigSecret ObjectReference `json:"kubeconfigSecret"`
}

// ServiceClusterMetadata describes metadata of the ServiceCluster for the Service Hub.
type ServiceClusterMetadata struct {
	// DisplayName is the human-readable name of this ServiceCluster.
	DisplayName string `json:"displayName,omitempty"`
	// Description is the human-readable description of this ServiceCluster.
	Description string `json:"description,omitempty"`
}

// ServiceClusterStatus represents the observed state of a ServiceCluster.
type ServiceClusterStatus struct {
	// DEPRECATED.
	// Phase represents the current lifecycle state of this object
	// consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to a string when printing the property
	// is only present for display purposes, for everything else use conditions
	Phase ServiceClusterPhaseType `json:"phase,omitempty"`
	// Conditions is a list of all conditions this ServiceCluster is in.
	Conditions []ServiceClusterCondition `json:"conditions,omitempty"`
	// The most recent generation observed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// KubernetesVersion of the service cluster API Server
	KubernetesVersion *version.Info `json:"kubernetesVersion,omitempty"`
}

// ServiceClusterPhaseType represents all conditions as a single string for printing in kubectl.
type ServiceClusterPhaseType string

// Values of ServiceClusterPhaseType
const (
	ServiceClusterPhaseUnknown     ServiceClusterPhaseType = "Unknown"
	ServiceClusterPhaseReady       ServiceClusterPhaseType = "Ready"
	ServiceClusterPhaseNotReady    ServiceClusterPhaseType = "NotReady"
	ServiceClusterPhaseTerminating ServiceClusterPhaseType = "Terminating"
)

const (
	ServiceClusterTerminatingReason = "Deleting"
)

// ServiceClusterConditionType represents a ServiceClusterCondition value.
type ServiceClusterConditionType string

const (
	// ServiceClusterReady represents a ServiceCluster condition is in ready state.
	ServiceClusterReady ServiceClusterConditionType = "Ready"
	// ServiceClusterControllerReady is True if the Ferry component is ready.
	ServiceClusterControllerReady ServiceClusterConditionType = "ControllerReady"
	// ServiceClusterControllerReachable is True if the Remote Service Cluster is Reachable.
	ServiceClusterReachable ServiceClusterConditionType = "ServiceClusterReachable"
)

// ServiceClusterCondition contains details for the current condition of this ServiceCluster.
type ServiceClusterCondition struct {
	// LastHeartbeatTime is the timestamp corresponding to the last update of this condition.
	LastHeartbeatTime metav1.Time `json:"lastHeartbeatTime"`
	// LastTransitionTime is the last time the condition transit from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Message is the human readable message indicating details about last transition.
	Message string `json:"message"`
	// Reason is the (brief) reason for the condition's last transition.
	Reason string `json:"reason"`
	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus `json:"status"`
	// Type of the condition, currently ('Ready').
	Type ServiceClusterConditionType `json:"type"`
}

// True returns whether .Status == "True"
func (c ServiceClusterCondition) True() bool {
	return c.Status == ConditionTrue
}

// updatePhase updates the phase property based on the current conditions
// this method should be called everytime the conditions are updated
func (s *ServiceClusterStatus) updatePhase() {
	for _, condition := range s.Conditions {
		if condition.Type != ServiceClusterReady {
			continue
		}

		switch condition.Status {
		case ConditionTrue:
			// Cluster reports its Ready
			s.Phase = ServiceClusterPhaseReady
		case ConditionFalse:
			// Cluster reports its NotReady for a reason
			if condition.Reason == ServiceClusterTerminatingReason {
				s.Phase = ServiceClusterPhaseTerminating
			} else {
				s.Phase = ServiceClusterPhaseNotReady
			}
		case ConditionUnknown:
			s.Phase = ServiceClusterPhaseUnknown
		}
		return
	}

	s.Phase = ServiceClusterPhaseUnknown
}

// SetCondition replaces or adds the given condition
func (s *ServiceClusterStatus) SetCondition(condition ServiceClusterCondition) {
	defer s.updatePhase()

	if condition.LastTransitionTime.IsZero() ||
		condition.LastHeartbeatTime.IsZero() {
		n := metav1.Now()
		// LastTransitionTime should always be set
		if condition.LastTransitionTime.IsZero() {
			condition.LastTransitionTime = n
		}

		// LastHeartbeatTime should always be set
		if condition.LastHeartbeatTime.IsZero() {
			condition.LastHeartbeatTime = n
		}
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
			s.Conditions[i].LastHeartbeatTime = condition.LastHeartbeatTime

			return
		}
	}

	s.Conditions = append(s.Conditions, condition)
}

// GetCondition returns the Condition of the given type, if it exists
func (s *ServiceClusterStatus) GetCondition(t ServiceClusterConditionType) (condition ServiceClusterCondition, exists bool) {
	for _, cond := range s.Conditions {
		if cond.Type == t {
			condition = cond
			exists = true
			return
		}
	}
	return
}

// ServiceCluster represents a Kubernets Cluster registered into KubeCarrier.
//
// **Example**
// ```yaml
// apiVersion: kubecarrier.io/v1alpha1
// kind: ServiceCluster
// metadata:
//   name: eu-west-1
// spec:
//   metadata:
//     displayName: EU West 1
//   kubeconfigSecret:
//     name: eu-west-1-kubeconfig
// ```
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Display Name",type="string",JSONPath=".spec.metadata.displayName"
// +kubebuilder:printcolumn:name="Kubernetes Version",type="string",JSONPath=".status.kubernetesVersion.gitVersion"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:categories=all;kubecarrier-provider,shortName=sc
type ServiceCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceClusterSpec   `json:"spec,omitempty"`
	Status ServiceClusterStatus `json:"status,omitempty"`
}

// ServiceClusterList contains a list of ServiceCluster.
// +kubebuilder:object:root=true
type ServiceClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceCluster{}, &ServiceClusterList{})
}

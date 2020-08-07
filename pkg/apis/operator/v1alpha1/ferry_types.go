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

// FerrySpec defines the desired state of Ferry.
type FerrySpec struct {
	// KubeconfigSecret specifies the Kubeconfig to use when connecting to the ServiceCluster.
	KubeconfigSecret ObjectReference `json:"kubeconfigSecret"`
	// Paused tell controller to pause reconciliation process and assume that Ferry is ready
	Paused PausedFlagType `json:"paused,omitempty"`
	// LogLevel
	// +optional
	LogLevel *int `json:"logLevel,omitempty"`
}

func (a *FerrySpec) SetLogLevel(logLevel int) {
	a.LogLevel = &logLevel
}

// FerryStatus defines the observed state of Ferry.
type FerryStatus struct {
	// DEPRECATED.
	// Phase represents the current lifecycle state of this object.
	// Consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to strings when printing the property.
	// This is only for display purpose, for everything else use conditions.
	Phase FerryPhaseType `json:"phase,omitempty"`
	// Conditions is a list of all conditions this Ferry is in.
	Conditions []FerryCondition `json:"conditions,omitempty"`
	// The most recent generation observed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// FerryPhaseType represents all conditions as a single string for printing in kubectl.
type FerryPhaseType string

// Values of FerryPhaseType.
const (
	FerryPhaseReady       FerryPhaseType = "Ready"
	FerryPhasePaused      FerryPhaseType = "Paused"
	FerryPhaseNotReady    FerryPhaseType = "NotReady"
	FerryPhaseTerminating FerryPhaseType = "Terminating"
	FerryPhaseUnknown     FerryPhaseType = "Unknown"
)

// FerryConditionType represents a FerryCondition value.
type FerryConditionType string

const (
	// FerryReady represents a Ferry condition is in ready state.
	FerryReady FerryConditionType = "Ready"
	// FerryPaused represents a Ferry condition is in paused state.
	FerryPaused FerryConditionType = "Paused"
)

const (
	FerryTerminatingReason string = "Terminating"
)

// FerryCondition contains details for the current condition of this Ferry.
type FerryCondition struct {
	// LastTransitionTime is the last time the condition transit from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Message is the human readable message indicating details about last transition.
	Message string `json:"message"`
	// Reason is the (brief) reason for the condition's last transition.
	Reason string `json:"reason"`
	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus `json:"status"`
	// Type of the condition, currently ('Ready').
	Type FerryConditionType `json:"type"`
}

// True returns whether .Status == "True"
func (c FerryCondition) True() bool {
	return c.Status == ConditionTrue
}

// UpdatePhase updates the phase property based on the current conditions.
// this method should be called everytime the conditions are updated.
func (s *FerryStatus) updatePhase() {
	for _, condition := range s.Conditions {
		if condition.Type == FerryPaused && condition.Status == ConditionTrue {
			s.Phase = FerryPhasePaused
			return
		}
	}

	for _, condition := range s.Conditions {
		if condition.Type == FerryReady {

			switch condition.Status {
			case ConditionTrue:
				s.Phase = FerryPhaseReady
			case ConditionFalse:
				if condition.Reason == FerryTerminatingReason {
					s.Phase = FerryPhaseTerminating
				} else {
					s.Phase = FerryPhaseNotReady
				}
			case ConditionUnknown:
				s.Phase = FerryPhaseUnknown
			}
			return
		}
	}
	s.Phase = FerryPhaseUnknown
}

// SetCondition replaces or adds the given condition.
func (s *FerryStatus) SetCondition(condition FerryCondition) {
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

// GetCondition returns the Condition of the given type, if it exists.
func (s *FerryStatus) GetCondition(t FerryConditionType) (condition FerryCondition, exists bool) {
	for _, cond := range s.Conditions {
		if cond.Type == t {
			condition = cond
			exists = true
			return
		}
	}
	return
}

// Ferry manages the deployment of the Ferry controller manager.
//
// Ferry lives in the Provider Namespace. For each ferry the KubeCarrier operator spins up
// the ferry controller deployment, necessary roles, service accounts, and role bindings.
//
// The reason for ferry controller deployment are multiples:
// * security --> KubeCarrier operator has greater privileges then ferry controller
// * resource isolation --> each ferry controller pod operates only on a single service cluster,
// 		thus resource allocation and monitoring is separate per ferry. This allows finer grade
// 		resource tuning and monitoring
// * flexibility --> If needed different ferries could have different deployments depending on
// 		their specific need (e.g. KubeCarrier image version for gradual rolling upgrade, different resource allocation, etc),
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:categories=all
type Ferry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FerrySpec   `json:"spec,omitempty"`
	Status FerryStatus `json:"status,omitempty"`
}

// IsReady returns if the Ferry is ready.
func (s *Ferry) IsReady() bool {
	if s.Generation != s.Status.ObservedGeneration {
		return false
	}

	for _, condition := range s.Status.Conditions {
		if condition.Type == FerryReady &&
			condition.Status == ConditionTrue {
			return true
		}
	}
	return false
}

func (s *Ferry) SetReadyCondition() bool {
	if !s.IsReady() {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(FerryCondition{
			Type:    FerryReady,
			Status:  ConditionTrue,
			Reason:  "DeploymentReady",
			Message: "the deployment of the Ferry controller manager is ready",
		})
		return true
	}
	return false
}
func (s *Ferry) SetUnReadyCondition() bool {
	readyCondition, _ := s.Status.GetCondition(FerryReady)
	if readyCondition.Status != ConditionFalse {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(FerryCondition{
			Type:    FerryReady,
			Status:  ConditionFalse,
			Reason:  "DeploymentUnready",
			Message: "the deployment of the Ferry controller manager is not ready",
		})
		return true
	}
	return false
}

// IsPaused returns if the Ferry is paused.
func (s *Ferry) IsPaused() bool {
	if s.Generation != s.Status.ObservedGeneration {
		return false
	}

	for _, condition := range s.Status.Conditions {
		if condition.Type == FerryPaused &&
			condition.Status == ConditionTrue {
			return true
		}
	}
	return false
}

func (s *Ferry) SetPausedCondition() bool {
	var changed bool
	if !s.IsReady() {
		changed = true
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(FerryCondition{
			Type:    FerryReady,
			Status:  ConditionTrue,
			Reason:  "Paused",
			Message: "Reconcilation is paused, assuming component is ready.",
		})
	}
	if !s.IsPaused() {
		changed = true
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(FerryCondition{
			Type:    FerryPaused,
			Status:  ConditionTrue,
			Reason:  "Paused",
			Message: "Reconcilation is paused",
		})
	}
	return changed
}

func (s *Ferry) SetUnPausedCondition() bool {
	if s.IsPaused() {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(FerryCondition{
			Type:    FerryPaused,
			Status:  ConditionFalse,
			Reason:  "UnPaused",
			Message: "Reconcilation is resumed",
		})
		return true
	}
	return false
}

func (s *Ferry) SetTerminatingCondition() bool {
	readyCondition, _ := s.Status.GetCondition(FerryReady)
	if readyCondition.Status != ConditionFalse ||
		readyCondition.Status == ConditionFalse && readyCondition.Reason != FerryTerminatingReason {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(FerryCondition{
			Type:    FerryReady,
			Status:  ConditionFalse,
			Reason:  FerryTerminatingReason,
			Message: "Ferry is being deleted",
		})
		return true
	}
	return false
}

// FerryList contains a list of Ferry.
// +kubebuilder:object:root=true
type FerryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Ferry `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Ferry{}, &FerryList{})
}

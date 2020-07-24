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

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
)

// CatapultSpec defines the desired state of Catapult.
type CatapultSpec struct {
	// References the CRD in the Management Cluster.
	ManagementClusterCRD CRDReference `json:"managementClusterCRD"`
	// References the CRD in the ServiceCluster.
	ServiceClusterCRD CRDReference `json:"serviceClusterCRD"`
	// References the ServiceCluster object that this object belongs to.
	ServiceCluster ObjectReference `json:"serviceCluster"`
	// WebhookStrategy configs the webhook of the CRD which is registered in the management cluster by this Catapult.
	// There are two possible values for this configuration {None (by default), ServiceCluster}
	// None (by default): Webhook will only check if there is an available ServiceClusterAssignment in the current Namespace.
	// ServiceCluster: Webhook will call webhooks of the CRD in the ServiceCluster with dry-run flag.
	// +kubebuilder:default:=None
	WebhookStrategy corev1alpha1.WebhookStrategyType `json:"webhookStrategy,omitempty"`
	// Paused tell controller to pause reconciliation process and assume that Catapult is ready
	Paused PausedFlagType `json:"paused,omitempty"`
	// LogLevel
	// +optional
	LogLevel *int `json:"logLevel,omitempty"`
}

func (a *CatapultSpec) SetLogLevel(logLevel int) {
	a.LogLevel = &logLevel
}

// CatapultStatus defines the observed state of Catapult.
type CatapultStatus struct {
	// ObservedGeneration is the most recent generation observed for this Catapult by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions represents the latest available observations of a Catapult's current state.
	Conditions []CatapultCondition `json:"conditions,omitempty"`
	// DEPRECATED.
	// Phase represents the current lifecycle state of this object.
	// Consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to strings when printing the property.
	// This is only for display purpose, for everything else use conditions.
	Phase CatapultPhaseType `json:"phase,omitempty"`
}

// CatapultPhaseType represents all conditions as a single string for printing by using kubectl commands.
type CatapultPhaseType string

// Values of CatapultPhaseType.
const (
	CatapultPhaseReady       CatapultPhaseType = "Ready"
	CatapultPhasePaused      CatapultPhaseType = "Paused"
	CatapultPhaseNotReady    CatapultPhaseType = "NotReady"
	CatapultPhaseUnknown     CatapultPhaseType = "Unknown"
	CatapultPhaseTerminating CatapultPhaseType = "Terminating"
)

const (
	CatapultTerminatingReason = "Deleting"
)

// updatePhase updates the phase property based on the current conditions.
// this method should be called every time the conditions are updated.
func (s *CatapultStatus) updatePhase() {

	for _, condition := range s.Conditions {

		if condition.Type == CatapultPaused && condition.Status == ConditionTrue {
			s.Phase = CatapultPhasePaused
			return
		}

		if condition.Type == CatapultReady {

			switch condition.Status {
			case ConditionTrue:
				s.Phase = CatapultPhaseReady
			case ConditionFalse:
				if condition.Reason == CatapultTerminatingReason {
					s.Phase = CatapultPhaseTerminating
				} else {
					s.Phase = CatapultPhaseNotReady
				}
			case ConditionUnknown:
				s.Phase = CatapultPhaseUnknown
			}
		}
	}
}

// CatapultConditionType represents a CatapultCondition value.
type CatapultConditionType string

const (
	// CatapultReady represents a Catapult condition is in ready state.
	CatapultReady CatapultConditionType = "Ready"
	// CatapultPaused represents a Catapult condition is in paused state.
	CatapultPaused CatapultConditionType = "Paused"
)

// CatapultCondition contains details for the current condition of this Catapult.
type CatapultCondition struct {
	// Type is the type of the Catapult condition, currently ('Ready').
	Type CatapultConditionType `json:"type"`
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
func (c CatapultCondition) True() bool {
	return c.Status == ConditionTrue
}

// GetCondition returns the Condition of the given condition type, if it exists.
func (s *CatapultStatus) GetCondition(t CatapultConditionType) (condition CatapultCondition, exists bool) {
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
func (s *CatapultStatus) SetCondition(condition CatapultCondition) {
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

// Catapult manages the deployment of the Catapult controller manager.
//
// A Catapult instance is started for each CustomResourceDiscovery instance and responsible for reconciling CRD instances across Kubernetes Clusters.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:categories=all
type Catapult struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CatapultSpec   `json:"spec,omitempty"`
	Status CatapultStatus `json:"status,omitempty"`
}

// IsReady returns if the Catapult is ready.
func (s *Catapult) IsReady() bool {
	if s.Generation != s.Status.ObservedGeneration {
		return false
	}

	for _, condition := range s.Status.Conditions {
		if condition.Type == CatapultReady &&
			condition.Status == ConditionTrue {
			return true
		}
	}
	return false
}

// IsPaused returns if the Catapult is paused.
func (s *Catapult) IsPaused() bool {
	if s.Generation != s.Status.ObservedGeneration {
		return false
	}

	for _, condition := range s.Status.Conditions {
		if condition.Type == CatapultPaused &&
			condition.Status == ConditionTrue {
			return true
		}
	}
	return false
}

func (s *Catapult) SetReadyCondition() bool {
	if !s.IsReady() {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(CatapultCondition{
			Type:    CatapultReady,
			Status:  ConditionTrue,
			Reason:  "DeploymentReady",
			Message: "the deployment of the Catapult controller manager is ready",
		})
		return true
	}
	return false
}

func (s *Catapult) SetPausedCondition() bool {
	var changed bool
	if !s.IsReady() {
		changed = true
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(CatapultCondition{
			Type:    CatapultReady,
			Status:  ConditionTrue,
			Reason:  "Paused",
			Message: "Reconcilation is paused, assuming component is ready.",
		})
		return changed
	}
	if !s.IsPaused() {
		changed = true
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(CatapultCondition{
			Type:    CatapultPaused,
			Status:  ConditionTrue,
			Reason:  "Paused",
			Message: "Reconcilation is paused",
		})
	}
	return changed
}

func (s *Catapult) SetUnPausedCondition() bool {
	if s.IsPaused() {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(CatapultCondition{
			Type:    CatapultPaused,
			Status:  ConditionFalse,
			Reason:  "UnPaused",
			Message: "Reconcilation is resumed",
		})
		return true
	}
	return false
}

func (s *Catapult) SetUnReadyCondition() bool {
	readyCondition, _ := s.Status.GetCondition(CatapultReady)
	if readyCondition.Status != ConditionFalse {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(CatapultCondition{
			Type:    CatapultReady,
			Status:  ConditionFalse,
			Reason:  "DeploymentUnready",
			Message: "the deployment of the Catapult controller manager is not ready",
		})
		return true
	}
	return false
}

func (s *Catapult) SetTerminatingCondition() bool {
	readyCondition, _ := s.Status.GetCondition(CatapultReady)
	if readyCondition.Status != ConditionFalse ||
		readyCondition.Status == ConditionFalse && readyCondition.Reason != CatapultTerminatingReason {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(CatapultCondition{
			Type:    CatapultReady,
			Status:  ConditionFalse,
			Reason:  CatapultTerminatingReason,
			Message: "Catapult is being deleted",
		})
		return true
	}
	return false
}

// CatapultList contains a list of Catapult.
// +kubebuilder:object:root=true
type CatapultList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Catapult `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Catapult{}, &CatapultList{})
}

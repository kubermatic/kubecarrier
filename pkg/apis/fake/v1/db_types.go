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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DBSpec defines the desired state of DB
type DBSpec struct {
	// RootPassword is root account password for this data. Leave blank for auto-generation
	RootPassword string `json:"rootPassword,omitempty"`

	// DatabaseName of the created database at start up
	// +kubebuilder:validation:MinLength=1
	DatabaseName string `json:"databaseName"`

	// DatabaseUser for created database
	DatabaseUser string `json:"databaseUser"`

	// DatabasePassword for the created database. Leave blank for auto-generation
	DatabasePassword string `json:"databasePassword,omitempty"`

	// Config is E2E tests params
	Config Config `json:"config,omitempty"`
}

// OperationFlagType represents a enable/disable flag
type OperationFlagType string

func (o OperationFlagType) Enabled() bool {
	if o == "" {
		return true
	}
	return o == OperationFlagEnabled
}

// Values of OperationFlagType.
const (
	OperationFlagEnabled  OperationFlagType = "Enabled"
	OperationFlagDisabled OperationFlagType = "Disabled"
)

// Config defines the e2e tests params
type Config struct {
	// ReadyAfterSeconds represents duration after which operator will mark DB as Ready
	ReadyAfterSeconds int `json:"readyAfterSeconds,omitempty"`
	// DeletionAfterSeconds represents duration after which operator will remove finalizer
	DeletionAfterSeconds int `json:"deletionAfterSeconds,omitempty"`
	// CreateEnable control whether create operation enabled or not
	// +kubebuilder:default:=Enabled
	Create OperationFlagType `json:"create,omitempty"`
	// UpdateEnable control whether update operation enabled or not
	// +kubebuilder:default:=Enabled
	Update OperationFlagType `json:"update,omitempty"`
	// DeleteEnable control whether delete operation enabled or not
	// +kubebuilder:default:=Enabled
	Delete OperationFlagType `json:"delete,omitempty"`
}

// DBStatus defines the observed state of DB
type DBStatus struct {
	// ObservedGeneration is the most recent generation observed for this FakeDB by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions represents the latest available observations of a FakeDB's current state.
	Conditions []DBCondition `json:"conditions,omitempty"`
	// DEPRECATED.
	// Phase represents the current lifecycle state of this object.
	// Consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to strings when printing the property.
	// This is only for display purpose, for everything else use conditions.
	Phase DBPhaseType `json:"phase,omitempty"`
	// Connection is the connection string for FakeDB
	Connection string `json:"connection,omitempty"`
}

// DBPhaseType represents all conditions as a single string for printing by using kubectl commands.
type DBPhaseType string

// Values of DBPhaseType.
const (
	DBPhaseReady       DBPhaseType = "Ready"
	DBPhaseNotReady    DBPhaseType = "NotReady"
	DBPhaseUnknown     DBPhaseType = "Unknown"
	DBPhaseTerminating DBPhaseType = "Terminating"
)

const (
	DBTerminatingReason = "Deleting"
)

// DBConditionType represents a DBCondition value.
type DBConditionType string

const (
	// DBReady represents a DB condition is in ready state.
	DBReady DBConditionType = "Ready"
)

// DBCondition contains details for the current condition of this DB.
type DBCondition struct {
	// Type is the type of the DB condition, currently ('Ready').
	Type DBConditionType `json:"type"`
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
func (c DBCondition) True() bool {
	return c.Status == ConditionTrue
}

// updatePhase updates the phase property based on the current conditions.
// this method should be called every time the conditions are updated.
func (s *DBStatus) updatePhase() {

	for _, condition := range s.Conditions {
		if condition.Type != DBReady {
			continue
		}

		switch condition.Status {
		case ConditionTrue:
			s.Phase = DBPhaseReady
		case ConditionFalse:
			if condition.Reason == DBTerminatingReason {
				s.Phase = DBPhaseTerminating
			} else {
				s.Phase = DBPhaseNotReady
			}
		case ConditionUnknown:
			s.Phase = DBPhaseUnknown
		}
		return
	}

	s.Phase = DBPhaseUnknown
}

// GetCondition returns the Condition of the given condition type, if it exists.
func (s *DBStatus) GetCondition(t DBConditionType) (condition DBCondition, exists bool) {
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
func (s *DBStatus) SetCondition(condition DBCondition) {
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

// DB is core element in e2e operator
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Connection",type="string",JSONPath=".status.connection"
type DB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DBSpec   `json:"spec,omitempty"`
	Status DBStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// DBList contains a list of DB
type DBList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DB `json:"items"`
}

// IsReady returns if the DB is ready.
func (s *DB) IsReady() bool {
	if s.Generation != s.Status.ObservedGeneration {
		return false
	}

	for _, condition := range s.Status.Conditions {
		if condition.Type == DBReady &&
			condition.Status == ConditionTrue {
			return true
		}
	}
	return false
}

func (s *DB) SetReadyCondition() bool {
	if !s.IsReady() {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(DBCondition{
			Type:    DBReady,
			Status:  ConditionTrue,
			Reason:  "DeploymentReady",
			Message: "the DB is ready",
		})
		return true
	}
	return false
}
func (s *DB) SetUnReadyCondition() bool {
	readyCondition, _ := s.Status.GetCondition(DBReady)
	if readyCondition.Status != ConditionFalse {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(DBCondition{
			Type:    DBReady,
			Status:  ConditionFalse,
			Reason:  "DBUnready",
			Message: "the DB is not ready",
		})
		return true
	}
	return false
}

func (s *DB) SetTerminatingCondition() bool {
	readyCondition, _ := s.Status.GetCondition(DBReady)
	if readyCondition.Status != ConditionFalse ||
		readyCondition.Status == ConditionFalse && readyCondition.Reason != DBTerminatingReason {
		s.Status.ObservedGeneration = s.Generation
		s.Status.SetCondition(DBCondition{
			Type:    DBReady,
			Status:  ConditionFalse,
			Reason:  DBTerminatingReason,
			Message: "DB is being deleted",
		})
		return true
	}
	return false
}

func init() {
	SchemeBuilder.Register(&DB{}, &DBList{})
}

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

// DBSpec defines the desired state of DB
type DBSpec struct {
	// RootPassword is root account password for this data. Leave blank for auto-generation
	RootPassword string `json:"rootPassword"`

	// DatabaseName of the created database at start up
	DatabaseName string `json:"databaseName"`

	// DatabaseUser for created database
	DatabaseUser string `json:"databaseUser"`

	// DatabasePassword for the created database. Leave blank for auto-generation
	DatabasePassword string `json:"databasePassword"`

	// Size of this database instance
	Size string `json:"size"`
}

// DBStatus defines the observed state of DB
type DBStatus struct {
	// Phase represents the current lifecycle state of this object
	// consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to a string when printing the property
	// is only present for display purposes, for everything else use conditions
	Phase DBPhaseType `json:"phase,omitempty"`
	// Conditions is a list of all conditions this DB is in.
	Conditions []DBCondition `json:"conditions,omitempty"`
	// The most recent generation observed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// RootCredentials for this database
	RootCredentials *Connection `json:"rootCredentials,omitempty"`

	// UserCredentials for this database
	UserCredentials *Connection `json:"userCredentials,omitempty"`
}

// Connection defines necessary endpoints and credential for DB usage
type Connection struct {
	// Endpoint for this database
	Endpoint string `json:"endpoint"`

	// Database name
	Name string `json:"name"`

	// Username for this database
	Username string `json:"username"`

	// Password for this database
	Password string `json:"password"`
}

// DBPhaseType represents all conditions as a single string for printing in kubectl
type DBPhaseType string

// Values of DBPhaseType
const (
	DBPhaseReady    DBPhaseType = "Ready"
	DBPhaseNotReady DBPhaseType = "NotReady"
	DBPhaseUnknown  DBPhaseType = "Unknown"
)

// DBConditionType represents a DBCondition value.
type DBConditionType string

const (
	// DBReady represents a DB condition is in ready state.
	DBReady DBConditionType = "Ready"
)

// DBCondition contains details for the current condition of this DB.
type DBCondition struct {
	// LastTransitionTime is the last time the condition transit from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Message is the human readable message indicating details about last transition.
	Message string `json:"message"`
	// Reason is the (brief) reason for the condition's last transition.
	Reason string `json:"reason"`
	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus `json:"status"`
	// Type of the condition, currently ('Ready').
	Type DBConditionType `json:"type"`
}

// UpdatePhase updates the phase property based on the current conditions
// this method should be called everytime the conditions are updated
func (s *DBStatus) updatePhase() {
	for _, condition := range s.Conditions {
		if condition.Type != DBReady {
			continue
		}

		switch condition.Status {
		case ConditionTrue:
			s.Phase = DBPhaseReady
		case ConditionFalse:
			s.Phase = DBPhaseNotReady
		case ConditionUnknown:
			s.Phase = DBPhaseUnknown
		}
		return
	}
	s.Phase = DBPhaseUnknown
}

// SetCondition replaces or adds the given condition
func (s *DBStatus) SetCondition(condition DBCondition) {
	defer s.updatePhase()

	if condition.LastTransitionTime.IsZero() {
		condition.LastTransitionTime = metav1.Now()
	}
	for i := range s.Conditions {
		if s.Conditions[i].Type == condition.Type {
			s.Conditions[i].Status = condition.Status
			s.Conditions[i].Reason = condition.Reason
			s.Conditions[i].Message = condition.Message
			if s.Conditions[i].Status != condition.Status {
				s.Conditions[i].LastTransitionTime = metav1.Now()
			}
			return
		}
	}

	condition.LastTransitionTime = metav1.Now()
	s.Conditions = append(s.Conditions, condition)
}

// GetCondition returns the Condition of the given type, if it exists
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

// DB is core element in joke generation operator
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
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

func init() {
	SchemeBuilder.Register(&DB{}, &DBList{})
}

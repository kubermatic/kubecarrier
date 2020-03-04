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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AccountSpec describes the desired state of Account.
type AccountSpec struct {
	// Metadata	contains additional human readable account details.
	Metadata AccountMetadata `json:"metadata,omitempty"`

	// Roles this account uses.
	// +kubebuilder:validation:MinItems=1
	Roles []AccountRole `json:"roles"`
	// Subjects contains a list of references to the group identities role binding applies to.
	// +kubebuilder:validation:MinItems=1
	Subjects []rbacv1.Subject `json:"subjects,omitempty"`
}

// AccountMetadata contains the metadata of the Account.
type AccountMetadata struct {
	// DisplayName is the human-readable name of this Account.
	// +kubebuilder:validation:MinLength=1
	DisplayName string `json:"displayName,omitempty"`
	// Description is the human-readable description of this Account.
	// +kubebuilder:validation:MinLength=1
	Description string `json:"description,omitempty"`
}

// AccountRole type represents available Account roles.
// +kubebuilder:validation:Enum=Provider;Tenant
type AccountRole string

const (
	ProviderRole AccountRole = "Provider"
	TenantRole   AccountRole = "Tenant"
)

// AccountStatus represents the observed state of Account.
type AccountStatus struct {
	// NamespaceName is the name of the Namespace that the Account manages.
	Namespace ObjectReference `json:"namespace,omitempty"`
	// ObservedGeneration is the most recent generation observed for this Account by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions represents the latest available observations of a Account's current state.
	Conditions []AccountCondition `json:"conditions,omitempty"`
	// DEPRECATED.
	// Phase represents the current lifecycle state of this object.
	// Consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to strings when printing the property.
	// This is only for display purpose, for everything else use conditions.
	Phase AccountPhaseType `json:"phase,omitempty"`
}

// AccountPhaseType represents all conditions as a single string for printing by using kubectl commands.
// +kubebuilder:validation:Ready;NotReady;Unknown;Terminating
type AccountPhaseType string

// Values of AccountPhaseType.
const (
	AccountPhaseReady       AccountPhaseType = "Ready"
	AccountPhaseNotReady    AccountPhaseType = "NotReady"
	AccountPhaseUnknown     AccountPhaseType = "Unknown"
	AccountPhaseTerminating AccountPhaseType = "Terminating"
)

const (
	AccountTerminatingReason = "Deleting"
)

// updatePhase updates the phase property based on the current conditions.
// this method should be called every time the conditions are updated.
func (s *AccountStatus) updatePhase() {
	for _, condition := range s.Conditions {
		if condition.Type != AccountReady {
			continue
		}

		switch condition.Status {
		case ConditionTrue:
			s.Phase = AccountPhaseReady
		case ConditionFalse:
			if condition.Reason == AccountTerminatingReason {
				s.Phase = AccountPhaseTerminating
			} else {
				s.Phase = AccountPhaseNotReady
			}
		case ConditionUnknown:
			s.Phase = AccountPhaseUnknown
		}
		return
	}

	s.Phase = AccountPhaseUnknown
}

// AccountConditionType represents a AccountCondition value.
// +kubebuilder:validation:Ready
type AccountConditionType string

const (
	// AccountReady represents a Account condition is in ready state.
	AccountReady AccountConditionType = "Ready"
)

// AccountCondition contains details for the current condition of this Account.
type AccountCondition struct {
	// Type is the type of the Account condition, currently ('Ready').
	Type AccountConditionType `json:"type"`
	// Status is the status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus `json:"status"`
	// LastTransitionTime is the last time the condition transits from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Reason is the (brief) reason for the condition's last transition.
	Reason string `json:"reason"`
	// Message is the human readable message indicating details about last transition.
	Message string `json:"message"`
}

// GetCondition returns the Condition of the given condition type, if it exists.
func (s *AccountStatus) GetCondition(t AccountConditionType) (condition AccountCondition, exists bool) {
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
func (s *AccountStatus) SetCondition(condition AccountCondition) {
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

// Account represents an actor in KubeCarrier. Depending on it's roles, it can provide services, consume offered services or both.
//
// KubeCarrier creates a new Namespace for each Account. The Account Metadata is exposed to users that are offered services from this Account.
//
// **Example**
// ```yaml
// apiVersion: catalog.kubecarrier.io/v1alpha1
// kind: Account
// metadata:
//   name: team-a
// spec:
//   metadata:
//     displayName: The A Team
//     description: In 1972, a crack commando unit was sent to prison by a military court...
//   roles:
//   - Provider
//   - Tenant
// ```
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Account Namespace",type="string",JSONPath=".status.namespace.name"
// +kubebuilder:printcolumn:name="Display Name",type="string",JSONPath=".spec.metadata.displayName"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:categories=kubecarrier-admin,shortName=acc,scope=Cluster
type Account struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AccountSpec   `json:"spec,omitempty"`
	Status AccountStatus `json:"status,omitempty"`
}

func (account *Account) HasRole(role AccountRole) bool {
	for _, r := range account.Spec.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// IsReady returns if the Account is ready.
func (account *Account) IsReady() bool {
	if !account.DeletionTimestamp.IsZero() {
		return false
	}

	if account.Generation != account.Status.ObservedGeneration {
		return false
	}

	for _, condition := range account.Status.Conditions {
		if condition.Type == AccountReady &&
			condition.Status == ConditionTrue {
			return true
		}
	}
	return false
}

// AccountList contains a list of Account.
// +kubebuilder:object:root=true
type AccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Account `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Account{}, &AccountList{})
}

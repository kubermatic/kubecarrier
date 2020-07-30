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

// CatalogEntrySetSpec defines the desired state of CatalogEntrySet.
type CatalogEntrySetSpec struct {
	// Metadata contains the metadata of each CatalogEntry for the Service Catalog.
	Metadata CatalogEntrySetMetadata `json:"metadata"`
	// Derive contains the configuration to generate DerivedCustomResources from the BaseCRDs that are selected by this CatalogEntrySet.
	Derive *DerivedConfig `json:"derive,omitempty"`
	// Discover contains the configuration to create a CustomResourceDiscoverySet.
	Discover CustomResourceDiscoverySetConfig `json:"discover"`
}

type CustomResourceDiscoverySetConfig struct {
	// CRD references a CustomResourceDefinition within the ServiceCluster.
	CRD ObjectReference `json:"crd"`
	// ServiceClusterSelector references a set of ServiceClusters to search the CustomResourceDefinition on.
	ServiceClusterSelector metav1.LabelSelector `json:"serviceClusterSelector"`
	// KindOverride overrides resulting internal CRDs kind
	KindOverride string `json:"kindOverride,omitempty"`
	// WebhookStrategy configs the webhook of the CRD which is registered in the management cluster by CustomResourceDiscovery object.
	// There are two possible values for this configuration {None (by default), ServiceCluster}
	// None (by default): Webhook will only check if there is an available ServiceClusterAssignment in the current Namespace.
	// ServiceCluster: Webhook will call webhooks of the CRD in the ServiceCluster with dry-run flag.
	// +kubebuilder:default:=None
	WebhookStrategy corev1alpha1.WebhookStrategyType `json:"webhookStrategy,omitempty"`
}

// CatalogEntrySetMetadata contains the metadata (display name, description, etc) of the CatalogEntrySet.
type CatalogEntrySetMetadata struct {
	CommonMetadata `json:",inline"`
}

// CatalogEntrySetStatus defines the observed state of CatalogEntrySet.
type CatalogEntrySetStatus struct {
	// ObservedGeneration is the most recent generation observed for this CatalogEntrySet by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions represents the latest available observations of a CatalogEntrySet's current state.
	Conditions []CatalogEntrySetCondition `json:"conditions,omitempty"`
	// DEPRECATED.
	// Phase represents the current lifecycle state of this object.
	// Consider this field DEPRECATED, it will be removed as soon as there
	// is a mechanism to map conditions to strings when printing the property.
	// This is only for display purpose, for everything else use conditions.
	Phase CatalogEntrySetPhaseType `json:"phase,omitempty"`
}

// CatalogEntrySetPhaseType represents all conditions as a single string for printing by using kubectl commands.
type CatalogEntrySetPhaseType string

// Values of CatalogEntrySetPhaseType.
const (
	CatalogEntrySetPhaseReady       CatalogEntrySetPhaseType = "Ready"
	CatalogEntrySetPhaseNotReady    CatalogEntrySetPhaseType = "NotReady"
	CatalogEntrySetPhaseUnknown     CatalogEntrySetPhaseType = "Unknown"
	CatalogEntrySetPhaseTerminating CatalogEntrySetPhaseType = "Terminating"
)

const (
	CatalogEntrySetTerminatingReason = "Deleting"
)

// updatePhase updates the phase property based on the current conditions
// this method should be called every time the conditions are updated.
func (s *CatalogEntrySetStatus) updatePhase() {
	for _, condition := range s.Conditions {
		if condition.Type != CatalogEntrySetReady {
			continue
		}

		switch condition.Status {
		case ConditionTrue:
			s.Phase = CatalogEntrySetPhaseReady
		case ConditionFalse:
			if condition.Reason == CatalogEntrySetTerminatingReason {
				s.Phase = CatalogEntrySetPhaseTerminating
			} else {
				s.Phase = CatalogEntrySetPhaseNotReady
			}
		case ConditionUnknown:
			s.Phase = CatalogEntrySetPhaseUnknown
		}
		return
	}

	s.Phase = CatalogEntrySetPhaseUnknown
}

// CatalogEntrySetConditionType represents a CatalogEntrySetCondition value.
type CatalogEntrySetConditionType string

const (
	// CatalogEntrySetReady represents a CatalogEntrySet condition is in ready state.
	CatalogEntrySetReady CatalogEntrySetConditionType = "Ready"
	// CustomResourceDiscoverySetReady represents the CustomResourceDiscoverySet that owned by this CatalogEntrySet is in ready state.
	CustomResourceDiscoverySetReady CatalogEntrySetConditionType = "CustomResourceDiscoveryReady"
	// CatalogEntriesReady represents the CatalogEntry objects that owned by this CatalogEntrySet are in ready state.
	CatalogEntriesReady CatalogEntrySetConditionType = "CatalogEntriesReady"
)

// CatalogEntrySetCondition contains details for the current condition of this CatalogEntrySet.
type CatalogEntrySetCondition struct {
	// Type is the type of the CatalogEntrySet condition, currently ('Ready').
	Type CatalogEntrySetConditionType `json:"type"`
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
func (c CatalogEntrySetCondition) True() bool {
	return c.Status == ConditionTrue
}

// GetCondition returns the Condition of the given condition type, if it exists.
func (s *CatalogEntrySetStatus) GetCondition(t CatalogEntrySetConditionType) (condition CatalogEntrySetCondition, exists bool) {
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
func (s *CatalogEntrySetStatus) SetCondition(condition CatalogEntrySetCondition) {
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

// CatalogEntrySet provides fully automation for provider to create both CustomResourceDiscoverySet and CatalogEntry for the same CRD in multiple service clusters.

// CatalogEntrySet manages a CustomResourceDiscoverySet and creates CatalogEntries for each CRD discovered from the selected ServiceClusters.
//
// **Example**
// See CatalogEntry documentation for more configuration details.
// ```yaml
// apiVersion: catalog.kubecarrier.io/v1alpha1
// kind: CatalogEntrySet
// metadata:
//   name: couchdbs
// spec:
//   metadata:
//     displayName: CouchDB
//     description: The compfy database
//   discoverySet:
//     crd:
//       name: couchdbs.couchdb.io
//     serviceClusterSelector: {}
// ```
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="CRD",type="string",JSONPath=".spec.discover.crd.name"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:categories=all;kubecarrier-provider,shortName=ces
type CatalogEntrySet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CatalogEntrySetSpec   `json:"spec,omitempty"`
	Status CatalogEntrySetStatus `json:"status,omitempty"`
}

// IsReady returns if the CatalogEntrySet is ready.
func (s *CatalogEntrySet) IsReady() bool {
	if s.Generation != s.Status.ObservedGeneration {
		return false
	}

	for _, condition := range s.Status.Conditions {
		if condition.Type == CatalogEntrySetReady &&
			condition.Status == ConditionTrue {
			return true
		}
	}
	return false
}

// CatalogEntrySetList contains a list of CatalogEntrySet.
// +kubebuilder:object:root=true
type CatalogEntrySetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CatalogEntrySet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CatalogEntrySet{}, &CatalogEntrySetList{})
}

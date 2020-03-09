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

// TenantSpec defines the desired state of Tenant
type TenantSpec struct{}

// Tenant exposes information about available Tenants on the platform and allows a Provider to set custom labels on them.
//
// Tenant objects will be created for all Accounts with the role "Tenant" in all Account Namespaces with the role "Provider".
// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:categories=kubecarrier-provider,shortName=tr
type Tenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec TenantSpec `json:"spec,omitempty"`
}

// TenantList contains a list of Tenant.
// +kubebuilder:object:root=true
type TenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tenant `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tenant{}, &TenantList{})
}

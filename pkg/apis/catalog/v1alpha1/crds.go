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
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

// CRDInformation contains type information about the CRD.
type CRDInformation struct {
	Name     string       `json:"name"`
	APIGroup string       `json:"apiGroup"`
	Kind     string       `json:"kind"`
	Versions []CRDVersion `json:"versions"`

	// ServiceCluster references a ServiceClusterReference of this CRD.
	ServiceCluster ObjectReference `json:"serviceCluster"`
}

// CRDVersion holds CRD version specific details.
type CRDVersion struct {
	Name   string                                         `json:"name"`
	Schema *apiextensionsv1beta1.CustomResourceValidation `json:"schema,omitempty"`
}

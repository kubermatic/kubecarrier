// +build !ignore_autogenerated

/*
Copyright 2020 The KubeCarrier Authors.

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/version"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomResourceDiscovery) DeepCopyInto(out *CustomResourceDiscovery) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomResourceDiscovery.
func (in *CustomResourceDiscovery) DeepCopy() *CustomResourceDiscovery {
	if in == nil {
		return nil
	}
	out := new(CustomResourceDiscovery)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CustomResourceDiscovery) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomResourceDiscoveryCondition) DeepCopyInto(out *CustomResourceDiscoveryCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomResourceDiscoveryCondition.
func (in *CustomResourceDiscoveryCondition) DeepCopy() *CustomResourceDiscoveryCondition {
	if in == nil {
		return nil
	}
	out := new(CustomResourceDiscoveryCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomResourceDiscoveryList) DeepCopyInto(out *CustomResourceDiscoveryList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CustomResourceDiscovery, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomResourceDiscoveryList.
func (in *CustomResourceDiscoveryList) DeepCopy() *CustomResourceDiscoveryList {
	if in == nil {
		return nil
	}
	out := new(CustomResourceDiscoveryList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CustomResourceDiscoveryList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomResourceDiscoverySet) DeepCopyInto(out *CustomResourceDiscoverySet) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomResourceDiscoverySet.
func (in *CustomResourceDiscoverySet) DeepCopy() *CustomResourceDiscoverySet {
	if in == nil {
		return nil
	}
	out := new(CustomResourceDiscoverySet)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CustomResourceDiscoverySet) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomResourceDiscoverySetCondition) DeepCopyInto(out *CustomResourceDiscoverySetCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomResourceDiscoverySetCondition.
func (in *CustomResourceDiscoverySetCondition) DeepCopy() *CustomResourceDiscoverySetCondition {
	if in == nil {
		return nil
	}
	out := new(CustomResourceDiscoverySetCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomResourceDiscoverySetList) DeepCopyInto(out *CustomResourceDiscoverySetList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CustomResourceDiscoverySet, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomResourceDiscoverySetList.
func (in *CustomResourceDiscoverySetList) DeepCopy() *CustomResourceDiscoverySetList {
	if in == nil {
		return nil
	}
	out := new(CustomResourceDiscoverySetList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CustomResourceDiscoverySetList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomResourceDiscoverySetSpec) DeepCopyInto(out *CustomResourceDiscoverySetSpec) {
	*out = *in
	out.CRD = in.CRD
	in.ServiceClusterSelector.DeepCopyInto(&out.ServiceClusterSelector)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomResourceDiscoverySetSpec.
func (in *CustomResourceDiscoverySetSpec) DeepCopy() *CustomResourceDiscoverySetSpec {
	if in == nil {
		return nil
	}
	out := new(CustomResourceDiscoverySetSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomResourceDiscoverySetStatus) DeepCopyInto(out *CustomResourceDiscoverySetStatus) {
	*out = *in
	if in.ManagementClusterCRDs != nil {
		in, out := &in.ManagementClusterCRDs, &out.ManagementClusterCRDs
		*out = make([]ObjectReference, len(*in))
		copy(*out, *in)
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]CustomResourceDiscoverySetCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomResourceDiscoverySetStatus.
func (in *CustomResourceDiscoverySetStatus) DeepCopy() *CustomResourceDiscoverySetStatus {
	if in == nil {
		return nil
	}
	out := new(CustomResourceDiscoverySetStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomResourceDiscoverySpec) DeepCopyInto(out *CustomResourceDiscoverySpec) {
	*out = *in
	out.CRD = in.CRD
	out.ServiceCluster = in.ServiceCluster
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomResourceDiscoverySpec.
func (in *CustomResourceDiscoverySpec) DeepCopy() *CustomResourceDiscoverySpec {
	if in == nil {
		return nil
	}
	out := new(CustomResourceDiscoverySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomResourceDiscoveryStatus) DeepCopyInto(out *CustomResourceDiscoveryStatus) {
	*out = *in
	if in.CRD != nil {
		in, out := &in.CRD, &out.CRD
		*out = new(v1.CustomResourceDefinition)
		(*in).DeepCopyInto(*out)
	}
	out.ManagementClusterCRD = in.ManagementClusterCRD
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]CustomResourceDiscoveryCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomResourceDiscoveryStatus.
func (in *CustomResourceDiscoveryStatus) DeepCopy() *CustomResourceDiscoveryStatus {
	if in == nil {
		return nil
	}
	out := new(CustomResourceDiscoveryStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ObjectReference) DeepCopyInto(out *ObjectReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ObjectReference.
func (in *ObjectReference) DeepCopy() *ObjectReference {
	if in == nil {
		return nil
	}
	out := new(ObjectReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceCluster) DeepCopyInto(out *ServiceCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceCluster.
func (in *ServiceCluster) DeepCopy() *ServiceCluster {
	if in == nil {
		return nil
	}
	out := new(ServiceCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServiceCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceClusterAssignment) DeepCopyInto(out *ServiceClusterAssignment) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceClusterAssignment.
func (in *ServiceClusterAssignment) DeepCopy() *ServiceClusterAssignment {
	if in == nil {
		return nil
	}
	out := new(ServiceClusterAssignment)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServiceClusterAssignment) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceClusterAssignmentCondition) DeepCopyInto(out *ServiceClusterAssignmentCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceClusterAssignmentCondition.
func (in *ServiceClusterAssignmentCondition) DeepCopy() *ServiceClusterAssignmentCondition {
	if in == nil {
		return nil
	}
	out := new(ServiceClusterAssignmentCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceClusterAssignmentList) DeepCopyInto(out *ServiceClusterAssignmentList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ServiceClusterAssignment, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceClusterAssignmentList.
func (in *ServiceClusterAssignmentList) DeepCopy() *ServiceClusterAssignmentList {
	if in == nil {
		return nil
	}
	out := new(ServiceClusterAssignmentList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServiceClusterAssignmentList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceClusterAssignmentSpec) DeepCopyInto(out *ServiceClusterAssignmentSpec) {
	*out = *in
	out.ServiceCluster = in.ServiceCluster
	out.ManagementClusterNamespace = in.ManagementClusterNamespace
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceClusterAssignmentSpec.
func (in *ServiceClusterAssignmentSpec) DeepCopy() *ServiceClusterAssignmentSpec {
	if in == nil {
		return nil
	}
	out := new(ServiceClusterAssignmentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceClusterAssignmentStatus) DeepCopyInto(out *ServiceClusterAssignmentStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]ServiceClusterAssignmentCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	out.ServiceClusterNamespace = in.ServiceClusterNamespace
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceClusterAssignmentStatus.
func (in *ServiceClusterAssignmentStatus) DeepCopy() *ServiceClusterAssignmentStatus {
	if in == nil {
		return nil
	}
	out := new(ServiceClusterAssignmentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceClusterCondition) DeepCopyInto(out *ServiceClusterCondition) {
	*out = *in
	in.LastHeartbeatTime.DeepCopyInto(&out.LastHeartbeatTime)
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceClusterCondition.
func (in *ServiceClusterCondition) DeepCopy() *ServiceClusterCondition {
	if in == nil {
		return nil
	}
	out := new(ServiceClusterCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceClusterList) DeepCopyInto(out *ServiceClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ServiceCluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceClusterList.
func (in *ServiceClusterList) DeepCopy() *ServiceClusterList {
	if in == nil {
		return nil
	}
	out := new(ServiceClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServiceClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceClusterMetadata) DeepCopyInto(out *ServiceClusterMetadata) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceClusterMetadata.
func (in *ServiceClusterMetadata) DeepCopy() *ServiceClusterMetadata {
	if in == nil {
		return nil
	}
	out := new(ServiceClusterMetadata)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceClusterSpec) DeepCopyInto(out *ServiceClusterSpec) {
	*out = *in
	out.Metadata = in.Metadata
	out.KubeconfigSecret = in.KubeconfigSecret
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceClusterSpec.
func (in *ServiceClusterSpec) DeepCopy() *ServiceClusterSpec {
	if in == nil {
		return nil
	}
	out := new(ServiceClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceClusterStatus) DeepCopyInto(out *ServiceClusterStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]ServiceClusterCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.KubernetesVersion != nil {
		in, out := &in.KubernetesVersion, &out.KubernetesVersion
		*out = new(version.Info)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceClusterStatus.
func (in *ServiceClusterStatus) DeepCopy() *ServiceClusterStatus {
	if in == nil {
		return nil
	}
	out := new(ServiceClusterStatus)
	in.DeepCopyInto(out)
	return out
}

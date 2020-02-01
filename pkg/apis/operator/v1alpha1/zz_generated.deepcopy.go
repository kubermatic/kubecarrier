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
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CRDReference) DeepCopyInto(out *CRDReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CRDReference.
func (in *CRDReference) DeepCopy() *CRDReference {
	if in == nil {
		return nil
	}
	out := new(CRDReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Catapult) DeepCopyInto(out *Catapult) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Catapult.
func (in *Catapult) DeepCopy() *Catapult {
	if in == nil {
		return nil
	}
	out := new(Catapult)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Catapult) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CatapultCondition) DeepCopyInto(out *CatapultCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CatapultCondition.
func (in *CatapultCondition) DeepCopy() *CatapultCondition {
	if in == nil {
		return nil
	}
	out := new(CatapultCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CatapultList) DeepCopyInto(out *CatapultList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Catapult, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CatapultList.
func (in *CatapultList) DeepCopy() *CatapultList {
	if in == nil {
		return nil
	}
	out := new(CatapultList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CatapultList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CatapultSpec) DeepCopyInto(out *CatapultSpec) {
	*out = *in
	out.MasterClusterCRD = in.MasterClusterCRD
	out.ServiceClusterCRD = in.ServiceClusterCRD
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CatapultSpec.
func (in *CatapultSpec) DeepCopy() *CatapultSpec {
	if in == nil {
		return nil
	}
	out := new(CatapultSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CatapultStatus) DeepCopyInto(out *CatapultStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]CatapultCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CatapultStatus.
func (in *CatapultStatus) DeepCopy() *CatapultStatus {
	if in == nil {
		return nil
	}
	out := new(CatapultStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Elevator) DeepCopyInto(out *Elevator) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Elevator.
func (in *Elevator) DeepCopy() *Elevator {
	if in == nil {
		return nil
	}
	out := new(Elevator)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Elevator) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElevatorCondition) DeepCopyInto(out *ElevatorCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElevatorCondition.
func (in *ElevatorCondition) DeepCopy() *ElevatorCondition {
	if in == nil {
		return nil
	}
	out := new(ElevatorCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElevatorList) DeepCopyInto(out *ElevatorList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Elevator, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElevatorList.
func (in *ElevatorList) DeepCopy() *ElevatorList {
	if in == nil {
		return nil
	}
	out := new(ElevatorList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ElevatorList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElevatorSpec) DeepCopyInto(out *ElevatorSpec) {
	*out = *in
	out.ProviderCRD = in.ProviderCRD
	out.TenantCRD = in.TenantCRD
	out.DerivedCRD = in.DerivedCRD
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElevatorSpec.
func (in *ElevatorSpec) DeepCopy() *ElevatorSpec {
	if in == nil {
		return nil
	}
	out := new(ElevatorSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElevatorStatus) DeepCopyInto(out *ElevatorStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]ElevatorCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElevatorStatus.
func (in *ElevatorStatus) DeepCopy() *ElevatorStatus {
	if in == nil {
		return nil
	}
	out := new(ElevatorStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubeCarrier) DeepCopyInto(out *KubeCarrier) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubeCarrier.
func (in *KubeCarrier) DeepCopy() *KubeCarrier {
	if in == nil {
		return nil
	}
	out := new(KubeCarrier)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KubeCarrier) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubeCarrierCondition) DeepCopyInto(out *KubeCarrierCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubeCarrierCondition.
func (in *KubeCarrierCondition) DeepCopy() *KubeCarrierCondition {
	if in == nil {
		return nil
	}
	out := new(KubeCarrierCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubeCarrierList) DeepCopyInto(out *KubeCarrierList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]KubeCarrier, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubeCarrierList.
func (in *KubeCarrierList) DeepCopy() *KubeCarrierList {
	if in == nil {
		return nil
	}
	out := new(KubeCarrierList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KubeCarrierList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubeCarrierSpec) DeepCopyInto(out *KubeCarrierSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubeCarrierSpec.
func (in *KubeCarrierSpec) DeepCopy() *KubeCarrierSpec {
	if in == nil {
		return nil
	}
	out := new(KubeCarrierSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubeCarrierStatus) DeepCopyInto(out *KubeCarrierStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]KubeCarrierCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubeCarrierStatus.
func (in *KubeCarrierStatus) DeepCopy() *KubeCarrierStatus {
	if in == nil {
		return nil
	}
	out := new(KubeCarrierStatus)
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
func (in *ServiceClusterRegistration) DeepCopyInto(out *ServiceClusterRegistration) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceClusterRegistration.
func (in *ServiceClusterRegistration) DeepCopy() *ServiceClusterRegistration {
	if in == nil {
		return nil
	}
	out := new(ServiceClusterRegistration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServiceClusterRegistration) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceClusterRegistrationCondition) DeepCopyInto(out *ServiceClusterRegistrationCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceClusterRegistrationCondition.
func (in *ServiceClusterRegistrationCondition) DeepCopy() *ServiceClusterRegistrationCondition {
	if in == nil {
		return nil
	}
	out := new(ServiceClusterRegistrationCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceClusterRegistrationList) DeepCopyInto(out *ServiceClusterRegistrationList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ServiceClusterRegistration, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceClusterRegistrationList.
func (in *ServiceClusterRegistrationList) DeepCopy() *ServiceClusterRegistrationList {
	if in == nil {
		return nil
	}
	out := new(ServiceClusterRegistrationList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServiceClusterRegistrationList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceClusterRegistrationSpec) DeepCopyInto(out *ServiceClusterRegistrationSpec) {
	*out = *in
	out.KubeconfigSecret = in.KubeconfigSecret
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceClusterRegistrationSpec.
func (in *ServiceClusterRegistrationSpec) DeepCopy() *ServiceClusterRegistrationSpec {
	if in == nil {
		return nil
	}
	out := new(ServiceClusterRegistrationSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceClusterRegistrationStatus) DeepCopyInto(out *ServiceClusterRegistrationStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]ServiceClusterRegistrationCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceClusterRegistrationStatus.
func (in *ServiceClusterRegistrationStatus) DeepCopy() *ServiceClusterRegistrationStatus {
	if in == nil {
		return nil
	}
	out := new(ServiceClusterRegistrationStatus)
	in.DeepCopyInto(out)
	return out
}

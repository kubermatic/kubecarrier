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
// Code generated by protoc-gen-go. DO NOT EDIT.
// source: instance.proto

package v1

import (
	context "context"
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
	empty "github.com/golang/protobuf/ptypes/empty"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Instance struct {
	Metadata             *ObjectMeta `protobuf:"bytes,1,opt,name=metadata,proto3" json:"metadata,omitempty"`
	Spec                 string      `protobuf:"bytes,2,opt,name=spec,proto3" json:"spec,omitempty"`
	Status               string      `protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *Instance) Reset()         { *m = Instance{} }
func (m *Instance) String() string { return proto.CompactTextString(m) }
func (*Instance) ProtoMessage()    {}
func (*Instance) Descriptor() ([]byte, []int) {
	return fileDescriptor_fd22322185b2070b, []int{0}
}

func (m *Instance) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Instance.Unmarshal(m, b)
}
func (m *Instance) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Instance.Marshal(b, m, deterministic)
}
func (m *Instance) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Instance.Merge(m, src)
}
func (m *Instance) XXX_Size() int {
	return xxx_messageInfo_Instance.Size(m)
}
func (m *Instance) XXX_DiscardUnknown() {
	xxx_messageInfo_Instance.DiscardUnknown(m)
}

var xxx_messageInfo_Instance proto.InternalMessageInfo

func (m *Instance) GetMetadata() *ObjectMeta {
	if m != nil {
		return m.Metadata
	}
	return nil
}

func (m *Instance) GetSpec() string {
	if m != nil {
		return m.Spec
	}
	return ""
}

func (m *Instance) GetStatus() string {
	if m != nil {
		return m.Status
	}
	return ""
}

type InstanceList struct {
	Metadata             *ListMeta   `protobuf:"bytes,1,opt,name=metadata,proto3" json:"metadata,omitempty"`
	Items                []*Instance `protobuf:"bytes,2,rep,name=items,proto3" json:"items,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *InstanceList) Reset()         { *m = InstanceList{} }
func (m *InstanceList) String() string { return proto.CompactTextString(m) }
func (*InstanceList) ProtoMessage()    {}
func (*InstanceList) Descriptor() ([]byte, []int) {
	return fileDescriptor_fd22322185b2070b, []int{1}
}

func (m *InstanceList) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_InstanceList.Unmarshal(m, b)
}
func (m *InstanceList) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_InstanceList.Marshal(b, m, deterministic)
}
func (m *InstanceList) XXX_Merge(src proto.Message) {
	xxx_messageInfo_InstanceList.Merge(m, src)
}
func (m *InstanceList) XXX_Size() int {
	return xxx_messageInfo_InstanceList.Size(m)
}
func (m *InstanceList) XXX_DiscardUnknown() {
	xxx_messageInfo_InstanceList.DiscardUnknown(m)
}

var xxx_messageInfo_InstanceList proto.InternalMessageInfo

func (m *InstanceList) GetMetadata() *ListMeta {
	if m != nil {
		return m.Metadata
	}
	return nil
}

func (m *InstanceList) GetItems() []*Instance {
	if m != nil {
		return m.Items
	}
	return nil
}

type InstanceGetRequest struct {
	Instance             string   `protobuf:"bytes,1,opt,name=instance,proto3" json:"instance,omitempty"`
	Version              string   `protobuf:"bytes,2,opt,name=version,proto3" json:"version,omitempty"`
	Name                 string   `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	Account              string   `protobuf:"bytes,4,opt,name=account,proto3" json:"account,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *InstanceGetRequest) Reset()         { *m = InstanceGetRequest{} }
func (m *InstanceGetRequest) String() string { return proto.CompactTextString(m) }
func (*InstanceGetRequest) ProtoMessage()    {}
func (*InstanceGetRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_fd22322185b2070b, []int{2}
}

func (m *InstanceGetRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_InstanceGetRequest.Unmarshal(m, b)
}
func (m *InstanceGetRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_InstanceGetRequest.Marshal(b, m, deterministic)
}
func (m *InstanceGetRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_InstanceGetRequest.Merge(m, src)
}
func (m *InstanceGetRequest) XXX_Size() int {
	return xxx_messageInfo_InstanceGetRequest.Size(m)
}
func (m *InstanceGetRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_InstanceGetRequest.DiscardUnknown(m)
}

var xxx_messageInfo_InstanceGetRequest proto.InternalMessageInfo

func (m *InstanceGetRequest) GetInstance() string {
	if m != nil {
		return m.Instance
	}
	return ""
}

func (m *InstanceGetRequest) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *InstanceGetRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *InstanceGetRequest) GetAccount() string {
	if m != nil {
		return m.Account
	}
	return ""
}

type InstanceDeleteRequest struct {
	Instance             string   `protobuf:"bytes,1,opt,name=instance,proto3" json:"instance,omitempty"`
	Version              string   `protobuf:"bytes,2,opt,name=version,proto3" json:"version,omitempty"`
	Name                 string   `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	Account              string   `protobuf:"bytes,4,opt,name=account,proto3" json:"account,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *InstanceDeleteRequest) Reset()         { *m = InstanceDeleteRequest{} }
func (m *InstanceDeleteRequest) String() string { return proto.CompactTextString(m) }
func (*InstanceDeleteRequest) ProtoMessage()    {}
func (*InstanceDeleteRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_fd22322185b2070b, []int{3}
}

func (m *InstanceDeleteRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_InstanceDeleteRequest.Unmarshal(m, b)
}
func (m *InstanceDeleteRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_InstanceDeleteRequest.Marshal(b, m, deterministic)
}
func (m *InstanceDeleteRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_InstanceDeleteRequest.Merge(m, src)
}
func (m *InstanceDeleteRequest) XXX_Size() int {
	return xxx_messageInfo_InstanceDeleteRequest.Size(m)
}
func (m *InstanceDeleteRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_InstanceDeleteRequest.DiscardUnknown(m)
}

var xxx_messageInfo_InstanceDeleteRequest proto.InternalMessageInfo

func (m *InstanceDeleteRequest) GetInstance() string {
	if m != nil {
		return m.Instance
	}
	return ""
}

func (m *InstanceDeleteRequest) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *InstanceDeleteRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *InstanceDeleteRequest) GetAccount() string {
	if m != nil {
		return m.Account
	}
	return ""
}

type InstanceListRequest struct {
	Instance             string   `protobuf:"bytes,1,opt,name=instance,proto3" json:"instance,omitempty"`
	Version              string   `protobuf:"bytes,2,opt,name=version,proto3" json:"version,omitempty"`
	Account              string   `protobuf:"bytes,3,opt,name=account,proto3" json:"account,omitempty"`
	LabelSelector        string   `protobuf:"bytes,4,opt,name=labelSelector,proto3" json:"labelSelector,omitempty"`
	Limit                int64    `protobuf:"varint,5,opt,name=limit,proto3" json:"limit,omitempty"`
	Continue             string   `protobuf:"bytes,6,opt,name=continue,proto3" json:"continue,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *InstanceListRequest) Reset()         { *m = InstanceListRequest{} }
func (m *InstanceListRequest) String() string { return proto.CompactTextString(m) }
func (*InstanceListRequest) ProtoMessage()    {}
func (*InstanceListRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_fd22322185b2070b, []int{4}
}

func (m *InstanceListRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_InstanceListRequest.Unmarshal(m, b)
}
func (m *InstanceListRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_InstanceListRequest.Marshal(b, m, deterministic)
}
func (m *InstanceListRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_InstanceListRequest.Merge(m, src)
}
func (m *InstanceListRequest) XXX_Size() int {
	return xxx_messageInfo_InstanceListRequest.Size(m)
}
func (m *InstanceListRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_InstanceListRequest.DiscardUnknown(m)
}

var xxx_messageInfo_InstanceListRequest proto.InternalMessageInfo

func (m *InstanceListRequest) GetInstance() string {
	if m != nil {
		return m.Instance
	}
	return ""
}

func (m *InstanceListRequest) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *InstanceListRequest) GetAccount() string {
	if m != nil {
		return m.Account
	}
	return ""
}

func (m *InstanceListRequest) GetLabelSelector() string {
	if m != nil {
		return m.LabelSelector
	}
	return ""
}

func (m *InstanceListRequest) GetLimit() int64 {
	if m != nil {
		return m.Limit
	}
	return 0
}

func (m *InstanceListRequest) GetContinue() string {
	if m != nil {
		return m.Continue
	}
	return ""
}

type InstanceCreateRequest struct {
	Instance             string    `protobuf:"bytes,1,opt,name=instance,proto3" json:"instance,omitempty"`
	Version              string    `protobuf:"bytes,2,opt,name=version,proto3" json:"version,omitempty"`
	Spec                 *Instance `protobuf:"bytes,3,opt,name=spec,proto3" json:"spec,omitempty"`
	Account              string    `protobuf:"bytes,4,opt,name=account,proto3" json:"account,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *InstanceCreateRequest) Reset()         { *m = InstanceCreateRequest{} }
func (m *InstanceCreateRequest) String() string { return proto.CompactTextString(m) }
func (*InstanceCreateRequest) ProtoMessage()    {}
func (*InstanceCreateRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_fd22322185b2070b, []int{5}
}

func (m *InstanceCreateRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_InstanceCreateRequest.Unmarshal(m, b)
}
func (m *InstanceCreateRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_InstanceCreateRequest.Marshal(b, m, deterministic)
}
func (m *InstanceCreateRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_InstanceCreateRequest.Merge(m, src)
}
func (m *InstanceCreateRequest) XXX_Size() int {
	return xxx_messageInfo_InstanceCreateRequest.Size(m)
}
func (m *InstanceCreateRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_InstanceCreateRequest.DiscardUnknown(m)
}

var xxx_messageInfo_InstanceCreateRequest proto.InternalMessageInfo

func (m *InstanceCreateRequest) GetInstance() string {
	if m != nil {
		return m.Instance
	}
	return ""
}

func (m *InstanceCreateRequest) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *InstanceCreateRequest) GetSpec() *Instance {
	if m != nil {
		return m.Spec
	}
	return nil
}

func (m *InstanceCreateRequest) GetAccount() string {
	if m != nil {
		return m.Account
	}
	return ""
}

func init() {
	proto.RegisterType((*Instance)(nil), "kubecarrier.api.v1.Instance")
	proto.RegisterType((*InstanceList)(nil), "kubecarrier.api.v1.InstanceList")
	proto.RegisterType((*InstanceGetRequest)(nil), "kubecarrier.api.v1.InstanceGetRequest")
	proto.RegisterType((*InstanceDeleteRequest)(nil), "kubecarrier.api.v1.InstanceDeleteRequest")
	proto.RegisterType((*InstanceListRequest)(nil), "kubecarrier.api.v1.InstanceListRequest")
	proto.RegisterType((*InstanceCreateRequest)(nil), "kubecarrier.api.v1.InstanceCreateRequest")
}

func init() {
	proto.RegisterFile("instance.proto", fileDescriptor_fd22322185b2070b)
}

var fileDescriptor_fd22322185b2070b = []byte{
	// 540 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xbc, 0x94, 0xcd, 0x8a, 0x13, 0x41,
	0x10, 0xc7, 0x99, 0x4c, 0x12, 0x93, 0xf6, 0xe3, 0xd0, 0xea, 0x32, 0x8c, 0x8b, 0x84, 0x41, 0x34,
	0x7a, 0x98, 0x31, 0x51, 0x44, 0xd6, 0xf5, 0xb2, 0xab, 0x2c, 0x82, 0x22, 0x8c, 0x37, 0x6f, 0x9d,
	0xb1, 0x5c, 0x5a, 0x27, 0xdd, 0x71, 0xba, 0x26, 0x28, 0x31, 0x17, 0x5f, 0x40, 0x41, 0x3c, 0xf8,
	0x30, 0x3e, 0x85, 0x47, 0xaf, 0x3e, 0x88, 0x4c, 0x7f, 0x64, 0x13, 0xd7, 0xcc, 0xc2, 0x06, 0xbc,
	0x55, 0x75, 0xd7, 0xcc, 0xef, 0x5f, 0xd5, 0xff, 0x6e, 0x72, 0x81, 0x0b, 0x85, 0x4c, 0x64, 0x10,
	0x4f, 0x0a, 0x89, 0x92, 0xd2, 0xb7, 0xe5, 0x08, 0x32, 0x56, 0x14, 0x1c, 0x8a, 0x98, 0x4d, 0x78,
	0x3c, 0x1d, 0x84, 0xdb, 0x87, 0x52, 0x1e, 0xe6, 0x90, 0xb0, 0x09, 0x4f, 0x98, 0x10, 0x12, 0x19,
	0x72, 0x29, 0x94, 0xf9, 0x22, 0xbc, 0x62, 0x77, 0x75, 0x36, 0x2a, 0x5f, 0x27, 0x30, 0x9e, 0xe0,
	0x07, 0xbb, 0x49, 0xc6, 0x80, 0xcc, 0xc4, 0x51, 0x41, 0x3a, 0x4f, 0x2c, 0x8c, 0xee, 0x90, 0x4e,
	0xb5, 0xf3, 0x8a, 0x21, 0x0b, 0xbc, 0x9e, 0xd7, 0x3f, 0x3b, 0xbc, 0x1a, 0x1f, 0x27, 0xc7, 0xcf,
	0x47, 0x6f, 0x20, 0xc3, 0x67, 0x80, 0x2c, 0x5d, 0xd4, 0x53, 0x4a, 0x9a, 0x6a, 0x02, 0x59, 0xd0,
	0xe8, 0x79, 0xfd, 0x6e, 0xaa, 0x63, 0xba, 0x45, 0xda, 0x0a, 0x19, 0x96, 0x2a, 0xf0, 0xf5, 0xaa,
	0xcd, 0xa2, 0x8f, 0xe4, 0x9c, 0x63, 0x3e, 0xe5, 0x0a, 0xe9, 0xfd, 0x63, 0xdc, 0xed, 0x7f, 0x71,
	0xab, 0xda, 0xbf, 0xa8, 0x43, 0xd2, 0xe2, 0x08, 0x63, 0x15, 0x34, 0x7a, 0xfe, 0xba, 0xcf, 0x1c,
	0x2a, 0x35, 0xa5, 0xd1, 0x7b, 0x42, 0xdd, 0xd2, 0x01, 0x60, 0x0a, 0xef, 0x4a, 0x50, 0x48, 0x43,
	0xd2, 0x71, 0x43, 0xd7, 0x1a, 0xba, 0xe9, 0x22, 0xa7, 0x01, 0x39, 0x33, 0x85, 0x42, 0x71, 0x29,
	0x6c, 0x7b, 0x2e, 0xad, 0xba, 0x16, 0x6c, 0x0c, 0xb6, 0x3f, 0x1d, 0x57, 0xd5, 0x2c, 0xcb, 0x64,
	0x29, 0x30, 0x68, 0x9a, 0x6a, 0x9b, 0x46, 0x33, 0x72, 0xd9, 0x91, 0x1f, 0x41, 0x0e, 0x08, 0xff,
	0x13, 0xfe, 0xc3, 0x23, 0x17, 0x97, 0xa7, 0xbe, 0x19, 0x7b, 0x89, 0xe3, 0xaf, 0x70, 0xe8, 0x35,
	0x72, 0x3e, 0x67, 0x23, 0xc8, 0x5f, 0x40, 0x0e, 0x19, 0xca, 0xc2, 0xea, 0x58, 0x5d, 0xa4, 0x97,
	0x48, 0x2b, 0xe7, 0x63, 0x8e, 0x41, 0xab, 0xe7, 0xf5, 0xfd, 0xd4, 0x24, 0x95, 0x96, 0x4c, 0x0a,
	0xe4, 0xa2, 0x84, 0xa0, 0x6d, 0xb4, 0xb8, 0x3c, 0xfa, 0xee, 0x1d, 0x4d, 0x6f, 0xbf, 0x00, 0xb6,
	0xe9, 0xf4, 0x6e, 0x5b, 0xc3, 0xfa, 0xeb, 0x0d, 0xb7, 0x70, 0x8e, 0xb1, 0xf3, 0xda, 0xd9, 0x0e,
	0x7f, 0x35, 0x49, 0xd7, 0x15, 0x2b, 0xfa, 0xd9, 0x23, 0x4d, 0xed, 0xeb, 0x1b, 0x75, 0x3f, 0x5d,
	0x3a, 0x83, 0xb0, 0x77, 0x52, 0x61, 0xb4, 0xfb, 0xe9, 0xe7, 0xef, 0xaf, 0x8d, 0x7b, 0xf4, 0x6e,
	0x32, 0x1d, 0x24, 0x16, 0xab, 0x92, 0x99, 0x8d, 0xe6, 0x89, 0x6b, 0x38, 0x99, 0xb9, 0x68, 0x9e,
	0xcc, 0x6c, 0xab, 0x73, 0xfa, 0xc5, 0x23, 0xfe, 0x01, 0x20, 0xbd, 0x5e, 0xc7, 0x39, 0xba, 0x0c,
	0x61, 0xed, 0x34, 0xa2, 0x7d, 0xad, 0xe5, 0x21, 0x7d, 0x70, 0x1a, 0x2d, 0xc9, 0xac, 0xf2, 0xa9,
	0x96, 0xd4, 0x36, 0x97, 0x80, 0xde, 0xac, 0xa3, 0xad, 0x5c, 0x94, 0x70, 0x2b, 0x36, 0xef, 0x5a,
	0xec, 0xde, 0xb5, 0xf8, 0x71, 0xf5, 0xae, 0x39, 0x49, 0xb7, 0x36, 0x92, 0xf4, 0xcd, 0x23, 0x6d,
	0xe3, 0xac, 0x7a, 0x49, 0x2b, 0xee, 0x3b, 0x61, 0x56, 0x7b, 0x5a, 0xd8, 0x6e, 0x74, 0xaa, 0x73,
	0xdb, 0xd1, 0xbe, 0xdb, 0x6b, 0xbe, 0x6c, 0x4c, 0x07, 0xa3, 0xb6, 0x6e, 0xf9, 0xce, 0x9f, 0x00,
	0x00, 0x00, 0xff, 0xff, 0x6f, 0xd0, 0x74, 0x8b, 0x1c, 0x06, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// InstancesClient is the client API for Instances service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type InstancesClient interface {
	List(ctx context.Context, in *InstanceListRequest, opts ...grpc.CallOption) (*InstanceList, error)
	Get(ctx context.Context, in *InstanceGetRequest, opts ...grpc.CallOption) (*Instance, error)
	Delete(ctx context.Context, in *InstanceDeleteRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	Create(ctx context.Context, in *InstanceCreateRequest, opts ...grpc.CallOption) (*Instance, error)
}

type instancesClient struct {
	cc grpc.ClientConnInterface
}

func NewInstancesClient(cc grpc.ClientConnInterface) InstancesClient {
	return &instancesClient{cc}
}

func (c *instancesClient) List(ctx context.Context, in *InstanceListRequest, opts ...grpc.CallOption) (*InstanceList, error) {
	out := new(InstanceList)
	err := c.cc.Invoke(ctx, "/kubecarrier.api.v1.Instances/List", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *instancesClient) Get(ctx context.Context, in *InstanceGetRequest, opts ...grpc.CallOption) (*Instance, error) {
	out := new(Instance)
	err := c.cc.Invoke(ctx, "/kubecarrier.api.v1.Instances/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *instancesClient) Delete(ctx context.Context, in *InstanceDeleteRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/kubecarrier.api.v1.Instances/Delete", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *instancesClient) Create(ctx context.Context, in *InstanceCreateRequest, opts ...grpc.CallOption) (*Instance, error) {
	out := new(Instance)
	err := c.cc.Invoke(ctx, "/kubecarrier.api.v1.Instances/Create", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// InstancesServer is the server API for Instances service.
type InstancesServer interface {
	List(context.Context, *InstanceListRequest) (*InstanceList, error)
	Get(context.Context, *InstanceGetRequest) (*Instance, error)
	Delete(context.Context, *InstanceDeleteRequest) (*empty.Empty, error)
	Create(context.Context, *InstanceCreateRequest) (*Instance, error)
}

// UnimplementedInstancesServer can be embedded to have forward compatible implementations.
type UnimplementedInstancesServer struct {
}

func (*UnimplementedInstancesServer) List(ctx context.Context, req *InstanceListRequest) (*InstanceList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
func (*UnimplementedInstancesServer) Get(ctx context.Context, req *InstanceGetRequest) (*Instance, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (*UnimplementedInstancesServer) Delete(ctx context.Context, req *InstanceDeleteRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}
func (*UnimplementedInstancesServer) Create(ctx context.Context, req *InstanceCreateRequest) (*Instance, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}

func RegisterInstancesServer(s *grpc.Server, srv InstancesServer) {
	s.RegisterService(&_Instances_serviceDesc, srv)
}

func _Instances_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InstanceListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InstancesServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubecarrier.api.v1.Instances/List",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InstancesServer).List(ctx, req.(*InstanceListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Instances_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InstanceGetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InstancesServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubecarrier.api.v1.Instances/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InstancesServer).Get(ctx, req.(*InstanceGetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Instances_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InstanceDeleteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InstancesServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubecarrier.api.v1.Instances/Delete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InstancesServer).Delete(ctx, req.(*InstanceDeleteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Instances_Create_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InstanceCreateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InstancesServer).Create(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubecarrier.api.v1.Instances/Create",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InstancesServer).Create(ctx, req.(*InstanceCreateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Instances_serviceDesc = grpc.ServiceDesc{
	ServiceName: "kubecarrier.api.v1.Instances",
	HandlerType: (*InstancesServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "List",
			Handler:    _Instances_List_Handler,
		},
		{
			MethodName: "Get",
			Handler:    _Instances_Get_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _Instances_Delete_Handler,
		},
		{
			MethodName: "Create",
			Handler:    _Instances_Create_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "instance.proto",
}

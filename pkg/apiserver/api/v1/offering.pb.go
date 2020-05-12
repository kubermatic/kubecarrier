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
// source: offering.proto

package v1

import (
	context "context"
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
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

type Offering struct {
	ObjectMeta           *ObjectMeta   `protobuf:"bytes,1,opt,name=objectMeta,proto3" json:"objectMeta,omitempty"`
	Spec                 *OfferingSpec `protobuf:"bytes,2,opt,name=spec,proto3" json:"spec,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *Offering) Reset()         { *m = Offering{} }
func (m *Offering) String() string { return proto.CompactTextString(m) }
func (*Offering) ProtoMessage()    {}
func (*Offering) Descriptor() ([]byte, []int) {
	return fileDescriptor_d876d3a4bab4c43a, []int{0}
}

func (m *Offering) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Offering.Unmarshal(m, b)
}
func (m *Offering) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Offering.Marshal(b, m, deterministic)
}
func (m *Offering) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Offering.Merge(m, src)
}
func (m *Offering) XXX_Size() int {
	return xxx_messageInfo_Offering.Size(m)
}
func (m *Offering) XXX_DiscardUnknown() {
	xxx_messageInfo_Offering.DiscardUnknown(m)
}

var xxx_messageInfo_Offering proto.InternalMessageInfo

func (m *Offering) GetObjectMeta() *ObjectMeta {
	if m != nil {
		return m.ObjectMeta
	}
	return nil
}

func (m *Offering) GetSpec() *OfferingSpec {
	if m != nil {
		return m.Spec
	}
	return nil
}

type OfferingSpec struct {
	Metadata             *OfferingMetadata `protobuf:"bytes,1,opt,name=metadata,proto3" json:"metadata,omitempty"`
	Provider             *ObjectReference  `protobuf:"bytes,2,opt,name=provider,proto3" json:"provider,omitempty"`
	Crd                  *CRDInformation   `protobuf:"bytes,3,opt,name=crd,proto3" json:"crd,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *OfferingSpec) Reset()         { *m = OfferingSpec{} }
func (m *OfferingSpec) String() string { return proto.CompactTextString(m) }
func (*OfferingSpec) ProtoMessage()    {}
func (*OfferingSpec) Descriptor() ([]byte, []int) {
	return fileDescriptor_d876d3a4bab4c43a, []int{1}
}

func (m *OfferingSpec) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_OfferingSpec.Unmarshal(m, b)
}
func (m *OfferingSpec) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_OfferingSpec.Marshal(b, m, deterministic)
}
func (m *OfferingSpec) XXX_Merge(src proto.Message) {
	xxx_messageInfo_OfferingSpec.Merge(m, src)
}
func (m *OfferingSpec) XXX_Size() int {
	return xxx_messageInfo_OfferingSpec.Size(m)
}
func (m *OfferingSpec) XXX_DiscardUnknown() {
	xxx_messageInfo_OfferingSpec.DiscardUnknown(m)
}

var xxx_messageInfo_OfferingSpec proto.InternalMessageInfo

func (m *OfferingSpec) GetMetadata() *OfferingMetadata {
	if m != nil {
		return m.Metadata
	}
	return nil
}

func (m *OfferingSpec) GetProvider() *ObjectReference {
	if m != nil {
		return m.Provider
	}
	return nil
}

func (m *OfferingSpec) GetCrd() *CRDInformation {
	if m != nil {
		return m.Crd
	}
	return nil
}

type OfferingMetadata struct {
	DisplayName          string   `protobuf:"bytes,1,opt,name=displayName,proto3" json:"displayName,omitempty"`
	Description          string   `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *OfferingMetadata) Reset()         { *m = OfferingMetadata{} }
func (m *OfferingMetadata) String() string { return proto.CompactTextString(m) }
func (*OfferingMetadata) ProtoMessage()    {}
func (*OfferingMetadata) Descriptor() ([]byte, []int) {
	return fileDescriptor_d876d3a4bab4c43a, []int{2}
}

func (m *OfferingMetadata) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_OfferingMetadata.Unmarshal(m, b)
}
func (m *OfferingMetadata) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_OfferingMetadata.Marshal(b, m, deterministic)
}
func (m *OfferingMetadata) XXX_Merge(src proto.Message) {
	xxx_messageInfo_OfferingMetadata.Merge(m, src)
}
func (m *OfferingMetadata) XXX_Size() int {
	return xxx_messageInfo_OfferingMetadata.Size(m)
}
func (m *OfferingMetadata) XXX_DiscardUnknown() {
	xxx_messageInfo_OfferingMetadata.DiscardUnknown(m)
}

var xxx_messageInfo_OfferingMetadata proto.InternalMessageInfo

func (m *OfferingMetadata) GetDisplayName() string {
	if m != nil {
		return m.DisplayName
	}
	return ""
}

func (m *OfferingMetadata) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

type OfferingList struct {
	ListMeta             *ListMeta   `protobuf:"bytes,1,opt,name=listMeta,proto3" json:"listMeta,omitempty"`
	Items                []*Offering `protobuf:"bytes,2,rep,name=items,proto3" json:"items,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *OfferingList) Reset()         { *m = OfferingList{} }
func (m *OfferingList) String() string { return proto.CompactTextString(m) }
func (*OfferingList) ProtoMessage()    {}
func (*OfferingList) Descriptor() ([]byte, []int) {
	return fileDescriptor_d876d3a4bab4c43a, []int{3}
}

func (m *OfferingList) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_OfferingList.Unmarshal(m, b)
}
func (m *OfferingList) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_OfferingList.Marshal(b, m, deterministic)
}
func (m *OfferingList) XXX_Merge(src proto.Message) {
	xxx_messageInfo_OfferingList.Merge(m, src)
}
func (m *OfferingList) XXX_Size() int {
	return xxx_messageInfo_OfferingList.Size(m)
}
func (m *OfferingList) XXX_DiscardUnknown() {
	xxx_messageInfo_OfferingList.DiscardUnknown(m)
}

var xxx_messageInfo_OfferingList proto.InternalMessageInfo

func (m *OfferingList) GetListMeta() *ListMeta {
	if m != nil {
		return m.ListMeta
	}
	return nil
}

func (m *OfferingList) GetItems() []*Offering {
	if m != nil {
		return m.Items
	}
	return nil
}

type OfferingGetRequest struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Account              string   `protobuf:"bytes,2,opt,name=account,proto3" json:"account,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *OfferingGetRequest) Reset()         { *m = OfferingGetRequest{} }
func (m *OfferingGetRequest) String() string { return proto.CompactTextString(m) }
func (*OfferingGetRequest) ProtoMessage()    {}
func (*OfferingGetRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_d876d3a4bab4c43a, []int{4}
}

func (m *OfferingGetRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_OfferingGetRequest.Unmarshal(m, b)
}
func (m *OfferingGetRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_OfferingGetRequest.Marshal(b, m, deterministic)
}
func (m *OfferingGetRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_OfferingGetRequest.Merge(m, src)
}
func (m *OfferingGetRequest) XXX_Size() int {
	return xxx_messageInfo_OfferingGetRequest.Size(m)
}
func (m *OfferingGetRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_OfferingGetRequest.DiscardUnknown(m)
}

var xxx_messageInfo_OfferingGetRequest proto.InternalMessageInfo

func (m *OfferingGetRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *OfferingGetRequest) GetAccount() string {
	if m != nil {
		return m.Account
	}
	return ""
}

type OfferingListRequest struct {
	Account              string   `protobuf:"bytes,1,opt,name=account,proto3" json:"account,omitempty"`
	LabelSelector        string   `protobuf:"bytes,2,opt,name=labelSelector,proto3" json:"labelSelector,omitempty"`
	Limit                int64    `protobuf:"varint,3,opt,name=limit,proto3" json:"limit,omitempty"`
	Continue             string   `protobuf:"bytes,4,opt,name=continue,proto3" json:"continue,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *OfferingListRequest) Reset()         { *m = OfferingListRequest{} }
func (m *OfferingListRequest) String() string { return proto.CompactTextString(m) }
func (*OfferingListRequest) ProtoMessage()    {}
func (*OfferingListRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_d876d3a4bab4c43a, []int{5}
}

func (m *OfferingListRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_OfferingListRequest.Unmarshal(m, b)
}
func (m *OfferingListRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_OfferingListRequest.Marshal(b, m, deterministic)
}
func (m *OfferingListRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_OfferingListRequest.Merge(m, src)
}
func (m *OfferingListRequest) XXX_Size() int {
	return xxx_messageInfo_OfferingListRequest.Size(m)
}
func (m *OfferingListRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_OfferingListRequest.DiscardUnknown(m)
}

var xxx_messageInfo_OfferingListRequest proto.InternalMessageInfo

func (m *OfferingListRequest) GetAccount() string {
	if m != nil {
		return m.Account
	}
	return ""
}

func (m *OfferingListRequest) GetLabelSelector() string {
	if m != nil {
		return m.LabelSelector
	}
	return ""
}

func (m *OfferingListRequest) GetLimit() int64 {
	if m != nil {
		return m.Limit
	}
	return 0
}

func (m *OfferingListRequest) GetContinue() string {
	if m != nil {
		return m.Continue
	}
	return ""
}

func init() {
	proto.RegisterType((*Offering)(nil), "kubecarrier.api.v1.Offering")
	proto.RegisterType((*OfferingSpec)(nil), "kubecarrier.api.v1.OfferingSpec")
	proto.RegisterType((*OfferingMetadata)(nil), "kubecarrier.api.v1.OfferingMetadata")
	proto.RegisterType((*OfferingList)(nil), "kubecarrier.api.v1.OfferingList")
	proto.RegisterType((*OfferingGetRequest)(nil), "kubecarrier.api.v1.OfferingGetRequest")
	proto.RegisterType((*OfferingListRequest)(nil), "kubecarrier.api.v1.OfferingListRequest")
}

func init() {
	proto.RegisterFile("offering.proto", fileDescriptor_d876d3a4bab4c43a)
}

var fileDescriptor_d876d3a4bab4c43a = []byte{
	// 512 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x94, 0xcf, 0x6e, 0xd3, 0x40,
	0x10, 0xc6, 0xe5, 0x24, 0x85, 0x30, 0xe1, 0x9f, 0x06, 0x0e, 0x56, 0x54, 0x21, 0xcb, 0x54, 0xb4,
	0x5c, 0x62, 0x25, 0xf4, 0xc0, 0x09, 0x50, 0x41, 0xaa, 0x90, 0x5a, 0x2a, 0x6d, 0x25, 0x0e, 0xdc,
	0x36, 0x9b, 0x49, 0xb4, 0xe0, 0xec, 0x9a, 0xdd, 0x4d, 0xa4, 0x2a, 0xad, 0x84, 0xb8, 0x70, 0xe1,
	0xc6, 0x43, 0xf1, 0x00, 0xbc, 0x02, 0x0f, 0x82, 0xbc, 0xb1, 0x1d, 0x97, 0xa6, 0xe9, 0x6d, 0x67,
	0xf4, 0xfd, 0x66, 0xbe, 0x99, 0x49, 0x0c, 0xf7, 0xf5, 0x78, 0x4c, 0x46, 0xaa, 0x49, 0x2f, 0x33,
	0xda, 0x69, 0xc4, 0x2f, 0xb3, 0x21, 0x09, 0x6e, 0x8c, 0x24, 0xd3, 0xe3, 0x99, 0xec, 0xcd, 0xfb,
	0xdd, 0xed, 0x89, 0xd6, 0x93, 0x94, 0x12, 0x9e, 0xc9, 0x84, 0x2b, 0xa5, 0x1d, 0x77, 0x52, 0x2b,
	0xbb, 0x24, 0xba, 0x1d, 0x77, 0x96, 0x51, 0x19, 0xc0, 0x94, 0x1c, 0x5f, 0xbe, 0xe3, 0x6f, 0x01,
	0xb4, 0x4f, 0x8a, 0xea, 0xf8, 0x0a, 0x40, 0x0f, 0x3f, 0x93, 0x70, 0xc7, 0xe4, 0x78, 0x18, 0x44,
	0xc1, 0x5e, 0x67, 0xf0, 0xa4, 0x77, 0xb5, 0x59, 0xef, 0xa4, 0x52, 0xb1, 0x1a, 0x81, 0xfb, 0xd0,
	0xb2, 0x19, 0x89, 0xb0, 0xe1, 0xc9, 0x68, 0x2d, 0x59, 0xf4, 0x3a, 0xcd, 0x48, 0x30, 0xaf, 0x8e,
	0x7f, 0x07, 0x70, 0xb7, 0x9e, 0xc6, 0x37, 0xd0, 0xce, 0x1d, 0x8e, 0x78, 0x65, 0x62, 0x67, 0x53,
	0xa9, 0xe3, 0x42, 0xcb, 0x2a, 0x0a, 0x5f, 0x43, 0x3b, 0x33, 0x7a, 0x2e, 0x47, 0x64, 0x0a, 0x33,
	0x4f, 0xaf, 0x1f, 0x83, 0xd1, 0x98, 0x0c, 0x29, 0x41, 0xac, 0x82, 0x70, 0x1f, 0x9a, 0xc2, 0x8c,
	0xc2, 0xa6, 0x67, 0xe3, 0x75, 0xec, 0x5b, 0xf6, 0xee, 0xbd, 0x1a, 0x6b, 0x33, 0xf5, 0x7b, 0x66,
	0xb9, 0x3c, 0xfe, 0x08, 0x0f, 0xff, 0x37, 0x85, 0x11, 0x74, 0x46, 0xd2, 0x66, 0x29, 0x3f, 0xfb,
	0xc0, 0xa7, 0xe4, 0xe7, 0xb9, 0xc3, 0xea, 0x29, 0xaf, 0x20, 0x2b, 0x8c, 0xcc, 0xf2, 0x4a, 0xde,
	0x6f, 0xae, 0x58, 0xa5, 0xe2, 0xf3, 0xd5, 0x82, 0x8e, 0xa4, 0x75, 0xf8, 0x12, 0xda, 0xa9, 0xb4,
	0xf5, 0x2b, 0x6d, 0xaf, 0xb3, 0x78, 0x54, 0x68, 0x58, 0xa5, 0xc6, 0x01, 0x6c, 0x49, 0x47, 0x53,
	0x1b, 0x36, 0xa2, 0xe6, 0x75, 0x58, 0xd9, 0x8a, 0x2d, 0xa5, 0xf1, 0x01, 0x60, 0x99, 0x3a, 0x24,
	0xc7, 0xe8, 0xeb, 0x8c, 0xac, 0x43, 0x84, 0x96, 0x5a, 0x0d, 0xe4, 0xdf, 0x18, 0xc2, 0x6d, 0x2e,
	0x84, 0x9e, 0x29, 0x57, 0x4c, 0x51, 0x86, 0xf1, 0x8f, 0x00, 0x1e, 0xd5, 0x47, 0x28, 0xab, 0xd4,
	0x88, 0xe0, 0x12, 0x81, 0x3b, 0x70, 0x2f, 0xe5, 0x43, 0x4a, 0x4f, 0x29, 0x25, 0xe1, 0xb4, 0x29,
	0x2a, 0x5e, 0x4e, 0xe2, 0x63, 0xd8, 0x4a, 0xe5, 0x54, 0x3a, 0x7f, 0xa9, 0x26, 0x5b, 0x06, 0xd8,
	0x85, 0xb6, 0xd0, 0xca, 0x49, 0x35, 0xa3, 0xb0, 0xe5, 0xb1, 0x2a, 0x1e, 0xfc, 0x6c, 0xc0, 0x83,
	0xea, 0xd7, 0x46, 0x66, 0x2e, 0x05, 0xe1, 0x02, 0x5a, 0x7e, 0xaf, 0xbb, 0x9b, 0xd6, 0x51, 0xb3,
	0xdd, 0x8d, 0x6e, 0x12, 0xc6, 0x7b, 0xdf, 0xff, 0xfc, 0xfd, 0xd5, 0x88, 0x31, 0x4a, 0xe6, 0xfd,
	0xa4, 0x98, 0xc9, 0x26, 0x8b, 0xe2, 0x75, 0x91, 0x94, 0xff, 0x68, 0x8b, 0xe7, 0xd0, 0x3c, 0x24,
	0x87, 0xcf, 0x36, 0x95, 0x5c, 0xed, 0xbd, 0xbb, 0xf1, 0x64, 0x71, 0xe2, 0xdb, 0x3e, 0xc7, 0xdd,
	0x9b, 0xda, 0x26, 0x8b, 0xfc, 0x62, 0x17, 0x07, 0xad, 0x4f, 0x8d, 0x79, 0x7f, 0x78, 0xcb, 0x7f,
	0x0c, 0x5e, 0xfc, 0x0b, 0x00, 0x00, 0xff, 0xff, 0x74, 0x8c, 0x91, 0xb1, 0x69, 0x04, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// OfferingServiceClient is the client API for OfferingService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type OfferingServiceClient interface {
	List(ctx context.Context, in *OfferingListRequest, opts ...grpc.CallOption) (*OfferingList, error)
	Get(ctx context.Context, in *OfferingGetRequest, opts ...grpc.CallOption) (*Offering, error)
}

type offeringServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewOfferingServiceClient(cc grpc.ClientConnInterface) OfferingServiceClient {
	return &offeringServiceClient{cc}
}

func (c *offeringServiceClient) List(ctx context.Context, in *OfferingListRequest, opts ...grpc.CallOption) (*OfferingList, error) {
	out := new(OfferingList)
	err := c.cc.Invoke(ctx, "/kubecarrier.api.v1.OfferingService/List", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *offeringServiceClient) Get(ctx context.Context, in *OfferingGetRequest, opts ...grpc.CallOption) (*Offering, error) {
	out := new(Offering)
	err := c.cc.Invoke(ctx, "/kubecarrier.api.v1.OfferingService/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// OfferingServiceServer is the server API for OfferingService service.
type OfferingServiceServer interface {
	List(context.Context, *OfferingListRequest) (*OfferingList, error)
	Get(context.Context, *OfferingGetRequest) (*Offering, error)
}

// UnimplementedOfferingServiceServer can be embedded to have forward compatible implementations.
type UnimplementedOfferingServiceServer struct {
}

func (*UnimplementedOfferingServiceServer) List(ctx context.Context, req *OfferingListRequest) (*OfferingList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
func (*UnimplementedOfferingServiceServer) Get(ctx context.Context, req *OfferingGetRequest) (*Offering, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}

func RegisterOfferingServiceServer(s *grpc.Server, srv OfferingServiceServer) {
	s.RegisterService(&_OfferingService_serviceDesc, srv)
}

func _OfferingService_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OfferingListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OfferingServiceServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubecarrier.api.v1.OfferingService/List",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OfferingServiceServer).List(ctx, req.(*OfferingListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OfferingService_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OfferingGetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OfferingServiceServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubecarrier.api.v1.OfferingService/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OfferingServiceServer).Get(ctx, req.(*OfferingGetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _OfferingService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "kubecarrier.api.v1.OfferingService",
	HandlerType: (*OfferingServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "List",
			Handler:    _OfferingService_List_Handler,
		},
		{
			MethodName: "Get",
			Handler:    _OfferingService_Get_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "offering.proto",
}

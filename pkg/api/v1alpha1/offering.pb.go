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

package v1alpha1

import (
	context "context"
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
	_ "github.com/golang/protobuf/ptypes/timestamp"
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
	Name                 string            `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Namespace            string            `protobuf:"bytes,2,opt,name=namespace,proto3" json:"namespace,omitempty"`
	Metadata             *OfferingMetadata `protobuf:"bytes,3,opt,name=metadata,proto3" json:"metadata,omitempty"`
	Provider             *ObjectReference  `protobuf:"bytes,4,opt,name=provider,proto3" json:"provider,omitempty"`
	Crd                  *CRDInformation   `protobuf:"bytes,5,opt,name=crd,proto3" json:"crd,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
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

func (m *Offering) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Offering) GetNamespace() string {
	if m != nil {
		return m.Namespace
	}
	return ""
}

func (m *Offering) GetMetadata() *OfferingMetadata {
	if m != nil {
		return m.Metadata
	}
	return nil
}

func (m *Offering) GetProvider() *ObjectReference {
	if m != nil {
		return m.Provider
	}
	return nil
}

func (m *Offering) GetCrd() *CRDInformation {
	if m != nil {
		return m.Crd
	}
	return nil
}

type OfferingList struct {
	Items                []*Offering `protobuf:"bytes,1,rep,name=items,proto3" json:"items,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *OfferingList) Reset()         { *m = OfferingList{} }
func (m *OfferingList) String() string { return proto.CompactTextString(m) }
func (*OfferingList) ProtoMessage()    {}
func (*OfferingList) Descriptor() ([]byte, []int) {
	return fileDescriptor_d876d3a4bab4c43a, []int{1}
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

func (m *OfferingList) GetItems() []*Offering {
	if m != nil {
		return m.Items
	}
	return nil
}

type OfferingListRequest struct {
	Selector             string   `protobuf:"bytes,1,opt,name=selector,proto3" json:"selector,omitempty"`
	Namespace            string   `protobuf:"bytes,2,opt,name=namespace,proto3" json:"namespace,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *OfferingListRequest) Reset()         { *m = OfferingListRequest{} }
func (m *OfferingListRequest) String() string { return proto.CompactTextString(m) }
func (*OfferingListRequest) ProtoMessage()    {}
func (*OfferingListRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_d876d3a4bab4c43a, []int{2}
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

func (m *OfferingListRequest) GetSelector() string {
	if m != nil {
		return m.Selector
	}
	return ""
}

func (m *OfferingListRequest) GetNamespace() string {
	if m != nil {
		return m.Namespace
	}
	return ""
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
	return fileDescriptor_d876d3a4bab4c43a, []int{3}
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

func init() {
	proto.RegisterType((*Offering)(nil), "kubecarrier.api.v1alpha1.Offering")
	proto.RegisterType((*OfferingList)(nil), "kubecarrier.api.v1alpha1.OfferingList")
	proto.RegisterType((*OfferingListRequest)(nil), "kubecarrier.api.v1alpha1.OfferingListRequest")
	proto.RegisterType((*OfferingMetadata)(nil), "kubecarrier.api.v1alpha1.OfferingMetadata")
}

func init() {
	proto.RegisterFile("offering.proto", fileDescriptor_d876d3a4bab4c43a)
}

var fileDescriptor_d876d3a4bab4c43a = []byte{
	// 425 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x53, 0x5d, 0x8b, 0xd4, 0x30,
	0x14, 0x65, 0x3e, 0x56, 0xba, 0xb7, 0xa2, 0x92, 0x45, 0x28, 0x75, 0xc0, 0x92, 0x07, 0xa9, 0x82,
	0x2d, 0x3b, 0xbe, 0x88, 0x8f, 0x7e, 0x0b, 0xea, 0x42, 0x04, 0x1f, 0x7c, 0x4b, 0xd3, 0xdb, 0x31,
	0xda, 0x36, 0xd9, 0x24, 0x33, 0xb0, 0x8a, 0x20, 0xbe, 0xf8, 0x03, 0xfc, 0x69, 0xfe, 0x05, 0x7f,
	0x88, 0x34, 0xdb, 0xce, 0x16, 0x71, 0x77, 0xe7, 0xa9, 0xcd, 0xbd, 0xe7, 0x9c, 0x7b, 0xcf, 0x69,
	0x03, 0xd7, 0x54, 0x55, 0xa1, 0x91, 0xed, 0x2a, 0xd3, 0x46, 0x39, 0x45, 0xa2, 0xcf, 0xeb, 0x02,
	0x05, 0x37, 0x46, 0xa2, 0xc9, 0xb8, 0x96, 0xd9, 0xe6, 0x90, 0xd7, 0xfa, 0x23, 0x3f, 0x8c, 0x6f,
	0xaf, 0x94, 0x5a, 0xd5, 0x98, 0x7b, 0x5c, 0xb1, 0xae, 0x72, 0x27, 0x1b, 0xb4, 0x8e, 0x37, 0xfa,
	0x94, 0x1a, 0x2f, 0x7a, 0x00, 0xd7, 0x32, 0xe7, 0x6d, 0xab, 0x1c, 0x77, 0x52, 0xb5, 0xb6, 0xef,
	0x86, 0xee, 0x44, 0x63, 0x7f, 0xa0, 0x3f, 0xa7, 0x10, 0x1c, 0xf5, 0x83, 0x09, 0x81, 0x79, 0xcb,
	0x1b, 0x8c, 0x26, 0xc9, 0x24, 0xdd, 0x67, 0xfe, 0x9d, 0x2c, 0x60, 0xbf, 0x7b, 0x5a, 0xcd, 0x05,
	0x46, 0x53, 0xdf, 0x38, 0x2b, 0x90, 0xe7, 0x10, 0x34, 0xe8, 0x78, 0xc9, 0x1d, 0x8f, 0x66, 0xc9,
	0x24, 0x0d, 0x97, 0xf7, 0xb2, 0xf3, 0xf6, 0xce, 0x86, 0x39, 0x6f, 0x7a, 0x06, 0xdb, 0x72, 0xc9,
	0x33, 0x08, 0xb4, 0x51, 0x1b, 0x59, 0xa2, 0x89, 0xe6, 0x5e, 0xe7, 0xee, 0x05, 0x3a, 0xc5, 0x27,
	0x14, 0x8e, 0x61, 0x85, 0x06, 0x5b, 0x81, 0x6c, 0x4b, 0x25, 0x8f, 0x60, 0x26, 0x4c, 0x19, 0xed,
	0x79, 0x85, 0xf4, 0x7c, 0x85, 0x27, 0xec, 0xe9, 0xab, 0xb6, 0x52, 0xa6, 0xf1, 0xc1, 0xb0, 0x8e,
	0x44, 0x5f, 0xc2, 0xd5, 0x61, 0xc1, 0xd7, 0xd2, 0x3a, 0xf2, 0x10, 0xf6, 0xa4, 0xc3, 0xc6, 0x46,
	0x93, 0x64, 0x96, 0x86, 0x4b, 0x7a, 0xb9, 0x2f, 0x76, 0x4a, 0xa0, 0x47, 0x70, 0x30, 0x56, 0x62,
	0x78, 0xbc, 0x46, 0xeb, 0x48, 0x0c, 0x81, 0xc5, 0x1a, 0x85, 0x53, 0xa6, 0x4f, 0x78, 0x7b, 0xbe,
	0x38, 0x65, 0xfa, 0x1e, 0x6e, 0xfc, 0x9b, 0x1d, 0x49, 0x20, 0x2c, 0xa5, 0xd5, 0x35, 0x3f, 0x79,
	0x7b, 0xf6, 0xc9, 0xc6, 0x25, 0x8f, 0x40, 0x2b, 0x8c, 0xd4, 0x9d, 0xc9, 0x5e, 0x75, 0x5c, 0x5a,
	0x7e, 0x9f, 0xc2, 0xf5, 0x41, 0xf8, 0x1d, 0x9a, 0x8d, 0x14, 0x48, 0xbe, 0xc0, 0xdc, 0xdb, 0xbf,
	0x7f, 0xb9, 0xdf, 0x91, 0xb9, 0xf8, 0xce, 0x6e, 0x70, 0x7a, 0xeb, 0xc7, 0xef, 0x3f, 0xbf, 0xa6,
	0x37, 0xc9, 0x41, 0x3e, 0xf4, 0xf3, 0xe1, 0xbf, 0xb7, 0xe4, 0x18, 0x66, 0x2f, 0xd0, 0x91, 0x1d,
	0xa2, 0x8e, 0x77, 0xc0, 0x50, 0xea, 0x67, 0x2d, 0x48, 0xfc, 0x9f, 0x59, 0xf9, 0xd7, 0x2e, 0xde,
	0x6f, 0x8f, 0xe1, 0x43, 0x30, 0x34, 0x8b, 0x2b, 0xfe, 0x4a, 0x3c, 0xf8, 0x1b, 0x00, 0x00, 0xff,
	0xff, 0x0e, 0x81, 0xb7, 0x79, 0x8a, 0x03, 0x00, 0x00,
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
	Get(ctx context.Context, in *Offering, opts ...grpc.CallOption) (*Offering, error)
}

type offeringServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewOfferingServiceClient(cc grpc.ClientConnInterface) OfferingServiceClient {
	return &offeringServiceClient{cc}
}

func (c *offeringServiceClient) List(ctx context.Context, in *OfferingListRequest, opts ...grpc.CallOption) (*OfferingList, error) {
	out := new(OfferingList)
	err := c.cc.Invoke(ctx, "/kubecarrier.api.v1alpha1.OfferingService/List", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *offeringServiceClient) Get(ctx context.Context, in *Offering, opts ...grpc.CallOption) (*Offering, error) {
	out := new(Offering)
	err := c.cc.Invoke(ctx, "/kubecarrier.api.v1alpha1.OfferingService/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// OfferingServiceServer is the server API for OfferingService service.
type OfferingServiceServer interface {
	List(context.Context, *OfferingListRequest) (*OfferingList, error)
	Get(context.Context, *Offering) (*Offering, error)
}

// UnimplementedOfferingServiceServer can be embedded to have forward compatible implementations.
type UnimplementedOfferingServiceServer struct {
}

func (*UnimplementedOfferingServiceServer) List(ctx context.Context, req *OfferingListRequest) (*OfferingList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
func (*UnimplementedOfferingServiceServer) Get(ctx context.Context, req *Offering) (*Offering, error) {
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
		FullMethod: "/kubecarrier.api.v1alpha1.OfferingService/List",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OfferingServiceServer).List(ctx, req.(*OfferingListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OfferingService_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Offering)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OfferingServiceServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubecarrier.api.v1alpha1.OfferingService/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OfferingServiceServer).Get(ctx, req.(*Offering))
	}
	return interceptor(ctx, in, info, handler)
}

var _OfferingService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "kubecarrier.api.v1alpha1.OfferingService",
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

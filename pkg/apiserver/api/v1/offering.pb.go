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
	Name                 string            `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Tenant               string            `protobuf:"bytes,2,opt,name=tenant,proto3" json:"tenant,omitempty"`
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

func (m *Offering) GetTenant() string {
	if m != nil {
		return m.Tenant
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
	Continue             string      `protobuf:"bytes,2,opt,name=continue,proto3" json:"continue,omitempty"`
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

func (m *OfferingList) GetContinue() string {
	if m != nil {
		return m.Continue
	}
	return ""
}

type OfferingRequest struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Tenant               string   `protobuf:"bytes,2,opt,name=tenant,proto3" json:"tenant,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *OfferingRequest) Reset()         { *m = OfferingRequest{} }
func (m *OfferingRequest) String() string { return proto.CompactTextString(m) }
func (*OfferingRequest) ProtoMessage()    {}
func (*OfferingRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_d876d3a4bab4c43a, []int{2}
}

func (m *OfferingRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_OfferingRequest.Unmarshal(m, b)
}
func (m *OfferingRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_OfferingRequest.Marshal(b, m, deterministic)
}
func (m *OfferingRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_OfferingRequest.Merge(m, src)
}
func (m *OfferingRequest) XXX_Size() int {
	return xxx_messageInfo_OfferingRequest.Size(m)
}
func (m *OfferingRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_OfferingRequest.DiscardUnknown(m)
}

var xxx_messageInfo_OfferingRequest proto.InternalMessageInfo

func (m *OfferingRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *OfferingRequest) GetTenant() string {
	if m != nil {
		return m.Tenant
	}
	return ""
}

type OfferingListRequest struct {
	Tenant               string   `protobuf:"bytes,1,opt,name=tenant,proto3" json:"tenant,omitempty"`
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
	return fileDescriptor_d876d3a4bab4c43a, []int{3}
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

func (m *OfferingListRequest) GetTenant() string {
	if m != nil {
		return m.Tenant
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
	return fileDescriptor_d876d3a4bab4c43a, []int{4}
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
	proto.RegisterType((*Offering)(nil), "kubecarrier.api.v1.Offering")
	proto.RegisterType((*OfferingList)(nil), "kubecarrier.api.v1.OfferingList")
	proto.RegisterType((*OfferingRequest)(nil), "kubecarrier.api.v1.OfferingRequest")
	proto.RegisterType((*OfferingListRequest)(nil), "kubecarrier.api.v1.OfferingListRequest")
	proto.RegisterType((*OfferingMetadata)(nil), "kubecarrier.api.v1.OfferingMetadata")
}

func init() {
	proto.RegisterFile("offering.proto", fileDescriptor_d876d3a4bab4c43a)
}

var fileDescriptor_d876d3a4bab4c43a = []byte{
	// 460 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x53, 0x5d, 0x8b, 0xd3, 0x40,
	0x14, 0x25, 0x4d, 0xba, 0xd4, 0x5b, 0xbf, 0xb8, 0x8a, 0x84, 0xb2, 0x48, 0xc8, 0xae, 0x6e, 0x5f,
	0x4c, 0x68, 0xf5, 0x55, 0x14, 0x15, 0x44, 0xf0, 0x03, 0xb2, 0xe0, 0x83, 0x0f, 0xc2, 0x34, 0xbd,
	0x2d, 0xa3, 0xc9, 0x4c, 0x9c, 0x4c, 0x23, 0x65, 0xd9, 0x17, 0x61, 0x7f, 0x81, 0x3f, 0xcd, 0xbf,
	0xe0, 0xbb, 0x7f, 0x41, 0x3a, 0x99, 0xb6, 0xe9, 0x5a, 0x22, 0xfb, 0x96, 0x7b, 0x39, 0xe7, 0xdc,
	0x7b, 0xcf, 0xc9, 0xc0, 0x4d, 0x39, 0x9b, 0x91, 0xe2, 0x62, 0x1e, 0x15, 0x4a, 0x6a, 0x89, 0xf8,
	0x75, 0x31, 0xa1, 0x94, 0x29, 0xc5, 0x49, 0x45, 0xac, 0xe0, 0x51, 0x35, 0x1a, 0x1c, 0xce, 0xa5,
	0x9c, 0x67, 0x14, 0xb3, 0x82, 0xc7, 0x4c, 0x08, 0xa9, 0x99, 0xe6, 0x52, 0x94, 0x35, 0x63, 0xd0,
	0xd7, 0xcb, 0x82, 0x6c, 0x11, 0xfe, 0x71, 0xa0, 0xf7, 0xc1, 0x2a, 0x22, 0x82, 0x27, 0x58, 0x4e,
	0xbe, 0x13, 0x38, 0xc3, 0x6b, 0x89, 0xf9, 0xc6, 0x7b, 0x70, 0xa0, 0x49, 0x30, 0xa1, 0xfd, 0x8e,
	0xe9, 0xda, 0x0a, 0x9f, 0x43, 0x2f, 0x27, 0xcd, 0xa6, 0x4c, 0x33, 0xdf, 0x0d, 0x9c, 0x61, 0x7f,
	0x7c, 0x1c, 0xfd, 0xbb, 0x4a, 0xb4, 0xd6, 0x7e, 0x67, 0xb1, 0xc9, 0x86, 0x85, 0xcf, 0xa0, 0x57,
	0x28, 0x59, 0xf1, 0x29, 0x29, 0xdf, 0x33, 0x0a, 0x47, 0x7b, 0x15, 0x26, 0x5f, 0x28, 0xd5, 0x09,
	0xcd, 0x48, 0x91, 0x48, 0x29, 0xd9, 0x90, 0xf0, 0x09, 0xb8, 0xa9, 0x9a, 0xfa, 0x5d, 0xc3, 0x0d,
	0xf7, 0x71, 0x5f, 0x26, 0xaf, 0xde, 0x88, 0x99, 0x54, 0xb9, 0x31, 0x20, 0x59, 0xc1, 0xc3, 0xcf,
	0x70, 0x7d, 0xbd, 0xd4, 0x5b, 0x5e, 0x6a, 0x1c, 0x43, 0x97, 0x6b, 0xca, 0x4b, 0xdf, 0x09, 0xdc,
	0x61, 0x7f, 0x7c, 0xd8, 0x76, 0x45, 0x52, 0x43, 0x71, 0x00, 0xbd, 0x54, 0x0a, 0xcd, 0xc5, 0x82,
	0xac, 0x2d, 0x9b, 0x3a, 0x7c, 0x0a, 0xb7, 0x36, 0x70, 0xfa, 0xb6, 0xa0, 0x52, 0x5f, 0xc5, 0xd7,
	0xf0, 0xc2, 0x81, 0x3b, 0xcd, 0xfd, 0xd6, 0x1a, 0x5b, 0xbc, 0xb3, 0x93, 0xc3, 0x31, 0xdc, 0xc8,
	0xd8, 0x84, 0xb2, 0x53, 0xca, 0x28, 0xd5, 0x52, 0x59, 0xb9, 0xdd, 0x26, 0xde, 0x85, 0x6e, 0xc6,
	0x73, 0xae, 0x4d, 0x54, 0x6e, 0x52, 0x17, 0x3b, 0x67, 0x78, 0x97, 0xce, 0xf8, 0x08, 0xb7, 0x2f,
	0x67, 0x87, 0x01, 0xf4, 0xa7, 0xbc, 0x2c, 0x32, 0xb6, 0x7c, 0xbf, 0x3d, 0xa7, 0xd9, 0x32, 0x08,
	0x2a, 0x53, 0xc5, 0x8b, 0x95, 0xe1, 0x76, 0x97, 0x66, 0x6b, 0x7c, 0xd1, 0xd9, 0xfa, 0x73, 0x4a,
	0xaa, 0xe2, 0x29, 0xe1, 0x12, 0x3c, 0x13, 0xc5, 0x49, 0x9b, 0xf7, 0x0d, 0x33, 0x06, 0xc1, 0xff,
	0x80, 0xe1, 0xc3, 0x1f, 0xbf, 0x7e, 0xff, 0xec, 0x04, 0x78, 0x3f, 0xae, 0x46, 0x71, 0x6d, 0x55,
	0x19, 0x9f, 0xd5, 0x1f, 0xe7, 0xf1, 0xfa, 0x0d, 0x95, 0xf8, 0x1d, 0xdc, 0xd7, 0xa4, 0xf1, 0xa8,
	0x35, 0x75, 0x3b, 0xb5, 0xf5, 0xd7, 0x08, 0x1f, 0x99, 0x89, 0x27, 0xf8, 0xa0, 0x7d, 0x62, 0x7c,
	0xb6, 0x8a, 0xff, 0xfc, 0x85, 0xf7, 0xa9, 0x53, 0x8d, 0x26, 0x07, 0xe6, 0x15, 0x3e, 0xfe, 0x1b,
	0x00, 0x00, 0xff, 0xff, 0xf5, 0xab, 0x12, 0x7f, 0xd6, 0x03, 0x00, 0x00,
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
	Get(ctx context.Context, in *OfferingRequest, opts ...grpc.CallOption) (*Offering, error)
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

func (c *offeringServiceClient) Get(ctx context.Context, in *OfferingRequest, opts ...grpc.CallOption) (*Offering, error) {
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
	Get(context.Context, *OfferingRequest) (*Offering, error)
}

// UnimplementedOfferingServiceServer can be embedded to have forward compatible implementations.
type UnimplementedOfferingServiceServer struct {
}

func (*UnimplementedOfferingServiceServer) List(ctx context.Context, req *OfferingListRequest) (*OfferingList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
func (*UnimplementedOfferingServiceServer) Get(ctx context.Context, req *OfferingRequest) (*Offering, error) {
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
	in := new(OfferingRequest)
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
		return srv.(OfferingServiceServer).Get(ctx, req.(*OfferingRequest))
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

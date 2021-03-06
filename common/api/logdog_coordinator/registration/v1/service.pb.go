// Code generated by protoc-gen-go.
// source: service.proto
// DO NOT EDIT!

/*
Package logdog is a generated protocol buffer package.

It is generated from these files:
	service.proto

It has these top-level messages:
	RegisterPrefixRequest
	RegisterPrefixResponse
*/
package logdog

import prpccommon "github.com/luci/luci-go/common/prpc"
import prpc "github.com/luci/luci-go/server/prpc"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/luci/luci-go/common/proto/google"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
const _ = proto.ProtoPackageIsVersion1

// RegisterPrefixRequest registers a new Prefix with the Coordinator.
type RegisterPrefixRequest struct {
	// The log stream's project.
	Project string `protobuf:"bytes,1,opt,name=project" json:"project,omitempty"`
	// The log stream prefix to register.
	Prefix string `protobuf:"bytes,2,opt,name=prefix" json:"prefix,omitempty"`
	// Optional information about the registering agent.
	SourceInfo []string `protobuf:"bytes,3,rep,name=source_info,json=sourceInfo" json:"source_info,omitempty"`
	// The prefix expiration time. If <= 0, the project's default prefix
	// expiration period will be applied.
	//
	// The prefix will be closed by the Coordinator after its expiration period.
	// Once closed, new stream registration requests will no longer be accepted.
	//
	// If supplied, this value should exceed the timeout of the local task, else
	// some of the task's streams may be dropped due to failing registration.
	Expiration *google_protobuf.Duration `protobuf:"bytes,10,opt,name=expiration" json:"expiration,omitempty"`
}

func (m *RegisterPrefixRequest) Reset()                    { *m = RegisterPrefixRequest{} }
func (m *RegisterPrefixRequest) String() string            { return proto.CompactTextString(m) }
func (*RegisterPrefixRequest) ProtoMessage()               {}
func (*RegisterPrefixRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *RegisterPrefixRequest) GetExpiration() *google_protobuf.Duration {
	if m != nil {
		return m.Expiration
	}
	return nil
}

// The response message for the RegisterPrefix RPC.
type RegisterPrefixResponse struct {
	// Secret is the prefix's secret. This must be included verbatim in Butler
	// bundles to assert ownership of this prefix.
	Secret []byte `protobuf:"bytes,1,opt,name=secret,proto3" json:"secret,omitempty"`
	// The name of the Pub/Sub topic to publish butlerproto-formatted Butler log
	// bundles to.
	LogBundleTopic string `protobuf:"bytes,2,opt,name=log_bundle_topic,json=logBundleTopic" json:"log_bundle_topic,omitempty"`
}

func (m *RegisterPrefixResponse) Reset()                    { *m = RegisterPrefixResponse{} }
func (m *RegisterPrefixResponse) String() string            { return proto.CompactTextString(m) }
func (*RegisterPrefixResponse) ProtoMessage()               {}
func (*RegisterPrefixResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func init() {
	proto.RegisterType((*RegisterPrefixRequest)(nil), "logdog.RegisterPrefixRequest")
	proto.RegisterType((*RegisterPrefixResponse)(nil), "logdog.RegisterPrefixResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion2

// Client API for Registration service

type RegistrationClient interface {
	// RegisterStream allows a Butler instance to register a log stream with the
	// Coordinator. Upon success, the Coordinator will return registration
	// information and streaming parameters to the Butler.
	//
	// This should be called by a Butler instance to gain the ability to publish
	// to a prefix space. The caller must have WRITE access to its project's
	// stream space. If WRITE access is not present, this will fail with the
	// "PermissionDenied" gRPC code.
	//
	// A stream prefix may be registered at most once. Additional registration
	// requests will fail with the "AlreadyExists" gRPC code.
	RegisterPrefix(ctx context.Context, in *RegisterPrefixRequest, opts ...grpc.CallOption) (*RegisterPrefixResponse, error)
}
type registrationPRPCClient struct {
	client *prpccommon.Client
}

func NewRegistrationPRPCClient(client *prpccommon.Client) RegistrationClient {
	return &registrationPRPCClient{client}
}

func (c *registrationPRPCClient) RegisterPrefix(ctx context.Context, in *RegisterPrefixRequest, opts ...grpc.CallOption) (*RegisterPrefixResponse, error) {
	out := new(RegisterPrefixResponse)
	err := c.client.Call(ctx, "logdog.Registration", "RegisterPrefix", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type registrationClient struct {
	cc *grpc.ClientConn
}

func NewRegistrationClient(cc *grpc.ClientConn) RegistrationClient {
	return &registrationClient{cc}
}

func (c *registrationClient) RegisterPrefix(ctx context.Context, in *RegisterPrefixRequest, opts ...grpc.CallOption) (*RegisterPrefixResponse, error) {
	out := new(RegisterPrefixResponse)
	err := grpc.Invoke(ctx, "/logdog.Registration/RegisterPrefix", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Registration service

type RegistrationServer interface {
	// RegisterStream allows a Butler instance to register a log stream with the
	// Coordinator. Upon success, the Coordinator will return registration
	// information and streaming parameters to the Butler.
	//
	// This should be called by a Butler instance to gain the ability to publish
	// to a prefix space. The caller must have WRITE access to its project's
	// stream space. If WRITE access is not present, this will fail with the
	// "PermissionDenied" gRPC code.
	//
	// A stream prefix may be registered at most once. Additional registration
	// requests will fail with the "AlreadyExists" gRPC code.
	RegisterPrefix(context.Context, *RegisterPrefixRequest) (*RegisterPrefixResponse, error)
}

func RegisterRegistrationServer(s prpc.Registrar, srv RegistrationServer) {
	s.RegisterService(&_Registration_serviceDesc, srv)
}

func _Registration_RegisterPrefix_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterPrefixRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RegistrationServer).RegisterPrefix(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/logdog.Registration/RegisterPrefix",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RegistrationServer).RegisterPrefix(ctx, req.(*RegisterPrefixRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Registration_serviceDesc = grpc.ServiceDesc{
	ServiceName: "logdog.Registration",
	HandlerType: (*RegistrationServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RegisterPrefix",
			Handler:    _Registration_RegisterPrefix_Handler,
		},
	},
	Streams: []grpc.StreamDesc{},
}

var fileDescriptor0 = []byte{
	// 266 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x74, 0x90, 0x4d, 0x4b, 0xc3, 0x40,
	0x10, 0x86, 0x89, 0x85, 0x88, 0xd3, 0x5a, 0x64, 0xc1, 0x12, 0x0b, 0xd6, 0xd2, 0x53, 0x4f, 0x5b,
	0xa8, 0x27, 0xaf, 0xe2, 0xc5, 0x93, 0x12, 0x3c, 0x79, 0x09, 0x66, 0x33, 0x59, 0x56, 0x42, 0x66,
	0xdd, 0x0f, 0xe9, 0x1f, 0xf2, 0x7f, 0xba, 0xdd, 0x4d, 0x41, 0xc5, 0x1e, 0xe7, 0x99, 0x37, 0x99,
	0x67, 0x5f, 0x38, 0xb7, 0x68, 0x3e, 0x95, 0x40, 0xae, 0x0d, 0x39, 0x62, 0x79, 0x47, 0xb2, 0x21,
	0x39, 0x5f, 0x48, 0x22, 0xd9, 0xe1, 0x26, 0xd2, 0xda, 0xb7, 0x9b, 0xc6, 0x9b, 0x37, 0xa7, 0xa8,
	0x4f, 0xb9, 0xd5, 0x57, 0x06, 0x97, 0x25, 0x4a, 0x65, 0x1d, 0x9a, 0x67, 0x83, 0xad, 0xda, 0x95,
	0xf8, 0xe1, 0xd1, 0x3a, 0x56, 0xc0, 0x69, 0x88, 0xbc, 0xa3, 0x70, 0x45, 0xb6, 0xcc, 0xd6, 0x67,
	0xe5, 0x61, 0x64, 0x33, 0xc8, 0x75, 0x8c, 0x16, 0x27, 0x71, 0x31, 0x4c, 0xec, 0x06, 0xc6, 0x96,
	0xbc, 0x11, 0x58, 0xa9, 0xbe, 0xa5, 0x62, 0xb4, 0x1c, 0x85, 0x25, 0x24, 0xf4, 0x18, 0x08, 0xbb,
	0x03, 0xc0, 0x9d, 0x56, 0x49, 0xa0, 0x80, 0xf0, 0xf1, 0x78, 0x7b, 0xc5, 0x93, 0x21, 0x3f, 0x18,
	0xf2, 0x87, 0xc1, 0xb0, 0xfc, 0x11, 0x5e, 0xbd, 0xc2, 0xec, 0xaf, 0xa6, 0xd5, 0xd4, 0x5b, 0xdc,
	0xdb, 0x58, 0x14, 0x06, 0x93, 0xe6, 0xa4, 0x1c, 0x26, 0xb6, 0x86, 0x8b, 0xd0, 0x41, 0x55, 0xfb,
	0xbe, 0xe9, 0xb0, 0x72, 0xa4, 0x95, 0x18, 0x7c, 0xa7, 0x81, 0xdf, 0x47, 0xfc, 0xb2, 0xa7, 0xdb,
	0x0a, 0x26, 0xe9, 0xdf, 0xe9, 0x16, 0x7b, 0x82, 0xe9, 0xef, 0x5b, 0xec, 0x9a, 0xa7, 0x3a, 0xf9,
	0xbf, 0x55, 0xcd, 0x17, 0xc7, 0xd6, 0x49, 0xb1, 0xce, 0xe3, 0xdb, 0x6e, 0xbf, 0x03, 0x00, 0x00,
	0xff, 0xff, 0x70, 0xdd, 0xcb, 0x0b, 0xa4, 0x01, 0x00, 0x00,
}

// Code generated by protoc-gen-go.
// source: github.com/luci/luci-go/logdog/api/endpoints/coordinator/services/v1/service.proto
// DO NOT EDIT!

/*
Package logdog is a generated protocol buffer package.

It is generated from these files:
	github.com/luci/luci-go/logdog/api/endpoints/coordinator/services/v1/service.proto
	github.com/luci/luci-go/logdog/api/endpoints/coordinator/services/v1/state.proto
	github.com/luci/luci-go/logdog/api/endpoints/coordinator/services/v1/tasks.proto

It has these top-level messages:
	GetConfigResponse
	RegisterStreamRequest
	RegisterStreamResponse
	LoadStreamRequest
	LoadStreamResponse
	TerminateStreamRequest
	ArchiveStreamRequest
	LogStreamState
	ArchiveTask
*/
package logdog

import prpccommon "github.com/luci/luci-go/common/prpc"
import prpc "github.com/luci/luci-go/server/prpc"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/luci/luci-go/common/proto/google"
import google_protobuf1 "github.com/luci/luci-go/common/proto/google"

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
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// GetConfigResponse is the response structure for the user
// "GetConfig" endpoint.
type GetConfigResponse struct {
	// The API URL of the base "luci-config" service. If empty, the default
	// service URL will be used.
	ConfigServiceUrl string `protobuf:"bytes,1,opt,name=config_service_url,json=configServiceUrl" json:"config_service_url,omitempty"`
	// The name of the configuration set to load from.
	ConfigSet string `protobuf:"bytes,2,opt,name=config_set,json=configSet" json:"config_set,omitempty"`
	// The path of the text-serialized service configuration protobuf.
	ServiceConfigPath string `protobuf:"bytes,3,opt,name=service_config_path,json=serviceConfigPath" json:"service_config_path,omitempty"`
}

func (m *GetConfigResponse) Reset()                    { *m = GetConfigResponse{} }
func (m *GetConfigResponse) String() string            { return proto.CompactTextString(m) }
func (*GetConfigResponse) ProtoMessage()               {}
func (*GetConfigResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

// RegisterStreamRequest is the set of caller-supplied data for the
// RegisterStream Coordinator service endpoint.
type RegisterStreamRequest struct {
	// The log stream's project.
	Project string `protobuf:"bytes,1,opt,name=project" json:"project,omitempty"`
	// The log stream's secret.
	Secret []byte `protobuf:"bytes,2,opt,name=secret,proto3" json:"secret,omitempty"`
	// The protobuf version string for this stream.
	ProtoVersion string `protobuf:"bytes,3,opt,name=proto_version,json=protoVersion" json:"proto_version,omitempty"`
	// The serialized LogStreamDescriptor protobuf for this stream.
	Desc []byte `protobuf:"bytes,4,opt,name=desc,proto3" json:"desc,omitempty"`
	// The stream's terminal index. If >= 0, the terminal index will be set
	// in the registration request, avoiding the need for an additional
	// termination RPC.
	TerminalIndex int64 `protobuf:"varint,5,opt,name=terminal_index,json=terminalIndex" json:"terminal_index,omitempty"`
}

func (m *RegisterStreamRequest) Reset()                    { *m = RegisterStreamRequest{} }
func (m *RegisterStreamRequest) String() string            { return proto.CompactTextString(m) }
func (*RegisterStreamRequest) ProtoMessage()               {}
func (*RegisterStreamRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

// The response message for the RegisterStream RPC.
type RegisterStreamResponse struct {
	// The Coordinator ID of the log stream.
	Id string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	// The state of the requested log stream.
	State *LogStreamState `protobuf:"bytes,2,opt,name=state" json:"state,omitempty"`
}

func (m *RegisterStreamResponse) Reset()                    { *m = RegisterStreamResponse{} }
func (m *RegisterStreamResponse) String() string            { return proto.CompactTextString(m) }
func (*RegisterStreamResponse) ProtoMessage()               {}
func (*RegisterStreamResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *RegisterStreamResponse) GetState() *LogStreamState {
	if m != nil {
		return m.State
	}
	return nil
}

// LoadStreamRequest loads the current state of a log stream.
type LoadStreamRequest struct {
	// The log stream's project.
	Project string `protobuf:"bytes,1,opt,name=project" json:"project,omitempty"`
	// The log stream's path Coordinator ID.
	Id string `protobuf:"bytes,2,opt,name=id" json:"id,omitempty"`
	// If true, include the log stream descriptor.
	Desc bool `protobuf:"varint,3,opt,name=desc" json:"desc,omitempty"`
}

func (m *LoadStreamRequest) Reset()                    { *m = LoadStreamRequest{} }
func (m *LoadStreamRequest) String() string            { return proto.CompactTextString(m) }
func (*LoadStreamRequest) ProtoMessage()               {}
func (*LoadStreamRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

// The response message for the LoadStream RPC.
type LoadStreamResponse struct {
	// The state of the requested log stream.
	State *LogStreamState `protobuf:"bytes,1,opt,name=state" json:"state,omitempty"`
	// If requested, the serialized log stream descriptor. The protobuf version
	// of this descriptor will match the "proto_version" field in "state".
	Desc []byte `protobuf:"bytes,2,opt,name=desc,proto3" json:"desc,omitempty"`
	// The age of the log stream.
	Age *google_protobuf.Duration `protobuf:"bytes,3,opt,name=age" json:"age,omitempty"`
	// The archival key of the log stream. If this key doesn't match the key in
	// the archival request, the request is superfluous and should be deleted.
	ArchivalKey []byte `protobuf:"bytes,4,opt,name=archival_key,json=archivalKey,proto3" json:"archival_key,omitempty"`
}

func (m *LoadStreamResponse) Reset()                    { *m = LoadStreamResponse{} }
func (m *LoadStreamResponse) String() string            { return proto.CompactTextString(m) }
func (*LoadStreamResponse) ProtoMessage()               {}
func (*LoadStreamResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *LoadStreamResponse) GetState() *LogStreamState {
	if m != nil {
		return m.State
	}
	return nil
}

func (m *LoadStreamResponse) GetAge() *google_protobuf.Duration {
	if m != nil {
		return m.Age
	}
	return nil
}

// TerminateStreamRequest is the set of caller-supplied data for the
// TerminateStream service endpoint.
type TerminateStreamRequest struct {
	// The log stream's project.
	Project string `protobuf:"bytes,1,opt,name=project" json:"project,omitempty"`
	// The log stream's path Coordinator ID.
	Id string `protobuf:"bytes,2,opt,name=id" json:"id,omitempty"`
	// The log stream's secret.
	Secret []byte `protobuf:"bytes,3,opt,name=secret,proto3" json:"secret,omitempty"`
	// The terminal index of the stream.
	TerminalIndex int64 `protobuf:"varint,4,opt,name=terminal_index,json=terminalIndex" json:"terminal_index,omitempty"`
}

func (m *TerminateStreamRequest) Reset()                    { *m = TerminateStreamRequest{} }
func (m *TerminateStreamRequest) String() string            { return proto.CompactTextString(m) }
func (*TerminateStreamRequest) ProtoMessage()               {}
func (*TerminateStreamRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

// ArchiveStreamRequest is the set of caller-supplied data for the ArchiveStream
// service endpoint.
type ArchiveStreamRequest struct {
	// The log stream's project.
	Project string `protobuf:"bytes,1,opt,name=project" json:"project,omitempty"`
	// The Coordinator ID of the log stream that was archived.
	Id string `protobuf:"bytes,2,opt,name=id" json:"id,omitempty"`
	// The number of log entries that were archived.
	LogEntryCount int64 `protobuf:"varint,3,opt,name=log_entry_count,json=logEntryCount" json:"log_entry_count,omitempty"`
	// The highest log stream index that was archived.
	TerminalIndex int64 `protobuf:"varint,4,opt,name=terminal_index,json=terminalIndex" json:"terminal_index,omitempty"`
	// If not empty, there was an archival error.
	//
	// This field serves to indicate that an error occured (being non-empty) and
	// to supply an value that will show up in the Coordinator ArchiveStream
	// endpoint logs.
	Error string `protobuf:"bytes,5,opt,name=error" json:"error,omitempty"`
	// The archive URL of the log stream's stream data.
	StreamUrl string `protobuf:"bytes,10,opt,name=stream_url,json=streamUrl" json:"stream_url,omitempty"`
	// The size of the log stream's stream data.
	StreamSize int64 `protobuf:"varint,11,opt,name=stream_size,json=streamSize" json:"stream_size,omitempty"`
	// The archive URL of the log stream's index data.
	IndexUrl string `protobuf:"bytes,20,opt,name=index_url,json=indexUrl" json:"index_url,omitempty"`
	// The size of the log stream's index data.
	IndexSize int64 `protobuf:"varint,21,opt,name=index_size,json=indexSize" json:"index_size,omitempty"`
	// The archive URL of the log stream's binary data.
	DataUrl string `protobuf:"bytes,30,opt,name=data_url,json=dataUrl" json:"data_url,omitempty"`
	// The size of the log stream's binary data.
	DataSize int64 `protobuf:"varint,31,opt,name=data_size,json=dataSize" json:"data_size,omitempty"`
}

func (m *ArchiveStreamRequest) Reset()                    { *m = ArchiveStreamRequest{} }
func (m *ArchiveStreamRequest) String() string            { return proto.CompactTextString(m) }
func (*ArchiveStreamRequest) ProtoMessage()               {}
func (*ArchiveStreamRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func init() {
	proto.RegisterType((*GetConfigResponse)(nil), "logdog.GetConfigResponse")
	proto.RegisterType((*RegisterStreamRequest)(nil), "logdog.RegisterStreamRequest")
	proto.RegisterType((*RegisterStreamResponse)(nil), "logdog.RegisterStreamResponse")
	proto.RegisterType((*LoadStreamRequest)(nil), "logdog.LoadStreamRequest")
	proto.RegisterType((*LoadStreamResponse)(nil), "logdog.LoadStreamResponse")
	proto.RegisterType((*TerminateStreamRequest)(nil), "logdog.TerminateStreamRequest")
	proto.RegisterType((*ArchiveStreamRequest)(nil), "logdog.ArchiveStreamRequest")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion3

// Client API for Services service

type ServicesClient interface {
	// GetConfig allows a service to retrieve the current service configuration
	// parameters.
	GetConfig(ctx context.Context, in *google_protobuf1.Empty, opts ...grpc.CallOption) (*GetConfigResponse, error)
	// RegisterStream is an idempotent stream state register operation.
	RegisterStream(ctx context.Context, in *RegisterStreamRequest, opts ...grpc.CallOption) (*RegisterStreamResponse, error)
	// LoadStream loads the current state of a log stream.
	LoadStream(ctx context.Context, in *LoadStreamRequest, opts ...grpc.CallOption) (*LoadStreamResponse, error)
	// TerminateStream is an idempotent operation to update the stream's terminal
	// index.
	TerminateStream(ctx context.Context, in *TerminateStreamRequest, opts ...grpc.CallOption) (*google_protobuf1.Empty, error)
	// ArchiveStream is an idempotent operation to record a log stream's archival
	// parameters. It is used by the Archivist service upon successful stream
	// archival.
	ArchiveStream(ctx context.Context, in *ArchiveStreamRequest, opts ...grpc.CallOption) (*google_protobuf1.Empty, error)
}
type servicesPRPCClient struct {
	client *prpccommon.Client
}

func NewServicesPRPCClient(client *prpccommon.Client) ServicesClient {
	return &servicesPRPCClient{client}
}

func (c *servicesPRPCClient) GetConfig(ctx context.Context, in *google_protobuf1.Empty, opts ...grpc.CallOption) (*GetConfigResponse, error) {
	out := new(GetConfigResponse)
	err := c.client.Call(ctx, "logdog.Services", "GetConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *servicesPRPCClient) RegisterStream(ctx context.Context, in *RegisterStreamRequest, opts ...grpc.CallOption) (*RegisterStreamResponse, error) {
	out := new(RegisterStreamResponse)
	err := c.client.Call(ctx, "logdog.Services", "RegisterStream", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *servicesPRPCClient) LoadStream(ctx context.Context, in *LoadStreamRequest, opts ...grpc.CallOption) (*LoadStreamResponse, error) {
	out := new(LoadStreamResponse)
	err := c.client.Call(ctx, "logdog.Services", "LoadStream", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *servicesPRPCClient) TerminateStream(ctx context.Context, in *TerminateStreamRequest, opts ...grpc.CallOption) (*google_protobuf1.Empty, error) {
	out := new(google_protobuf1.Empty)
	err := c.client.Call(ctx, "logdog.Services", "TerminateStream", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *servicesPRPCClient) ArchiveStream(ctx context.Context, in *ArchiveStreamRequest, opts ...grpc.CallOption) (*google_protobuf1.Empty, error) {
	out := new(google_protobuf1.Empty)
	err := c.client.Call(ctx, "logdog.Services", "ArchiveStream", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type servicesClient struct {
	cc *grpc.ClientConn
}

func NewServicesClient(cc *grpc.ClientConn) ServicesClient {
	return &servicesClient{cc}
}

func (c *servicesClient) GetConfig(ctx context.Context, in *google_protobuf1.Empty, opts ...grpc.CallOption) (*GetConfigResponse, error) {
	out := new(GetConfigResponse)
	err := grpc.Invoke(ctx, "/logdog.Services/GetConfig", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *servicesClient) RegisterStream(ctx context.Context, in *RegisterStreamRequest, opts ...grpc.CallOption) (*RegisterStreamResponse, error) {
	out := new(RegisterStreamResponse)
	err := grpc.Invoke(ctx, "/logdog.Services/RegisterStream", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *servicesClient) LoadStream(ctx context.Context, in *LoadStreamRequest, opts ...grpc.CallOption) (*LoadStreamResponse, error) {
	out := new(LoadStreamResponse)
	err := grpc.Invoke(ctx, "/logdog.Services/LoadStream", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *servicesClient) TerminateStream(ctx context.Context, in *TerminateStreamRequest, opts ...grpc.CallOption) (*google_protobuf1.Empty, error) {
	out := new(google_protobuf1.Empty)
	err := grpc.Invoke(ctx, "/logdog.Services/TerminateStream", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *servicesClient) ArchiveStream(ctx context.Context, in *ArchiveStreamRequest, opts ...grpc.CallOption) (*google_protobuf1.Empty, error) {
	out := new(google_protobuf1.Empty)
	err := grpc.Invoke(ctx, "/logdog.Services/ArchiveStream", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Services service

type ServicesServer interface {
	// GetConfig allows a service to retrieve the current service configuration
	// parameters.
	GetConfig(context.Context, *google_protobuf1.Empty) (*GetConfigResponse, error)
	// RegisterStream is an idempotent stream state register operation.
	RegisterStream(context.Context, *RegisterStreamRequest) (*RegisterStreamResponse, error)
	// LoadStream loads the current state of a log stream.
	LoadStream(context.Context, *LoadStreamRequest) (*LoadStreamResponse, error)
	// TerminateStream is an idempotent operation to update the stream's terminal
	// index.
	TerminateStream(context.Context, *TerminateStreamRequest) (*google_protobuf1.Empty, error)
	// ArchiveStream is an idempotent operation to record a log stream's archival
	// parameters. It is used by the Archivist service upon successful stream
	// archival.
	ArchiveStream(context.Context, *ArchiveStreamRequest) (*google_protobuf1.Empty, error)
}

func RegisterServicesServer(s prpc.Registrar, srv ServicesServer) {
	s.RegisterService(&_Services_serviceDesc, srv)
}

func _Services_GetConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(google_protobuf1.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServicesServer).GetConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/logdog.Services/GetConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServicesServer).GetConfig(ctx, req.(*google_protobuf1.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Services_RegisterStream_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterStreamRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServicesServer).RegisterStream(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/logdog.Services/RegisterStream",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServicesServer).RegisterStream(ctx, req.(*RegisterStreamRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Services_LoadStream_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoadStreamRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServicesServer).LoadStream(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/logdog.Services/LoadStream",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServicesServer).LoadStream(ctx, req.(*LoadStreamRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Services_TerminateStream_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TerminateStreamRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServicesServer).TerminateStream(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/logdog.Services/TerminateStream",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServicesServer).TerminateStream(ctx, req.(*TerminateStreamRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Services_ArchiveStream_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ArchiveStreamRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServicesServer).ArchiveStream(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/logdog.Services/ArchiveStream",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServicesServer).ArchiveStream(ctx, req.(*ArchiveStreamRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Services_serviceDesc = grpc.ServiceDesc{
	ServiceName: "logdog.Services",
	HandlerType: (*ServicesServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetConfig",
			Handler:    _Services_GetConfig_Handler,
		},
		{
			MethodName: "RegisterStream",
			Handler:    _Services_RegisterStream_Handler,
		},
		{
			MethodName: "LoadStream",
			Handler:    _Services_LoadStream_Handler,
		},
		{
			MethodName: "TerminateStream",
			Handler:    _Services_TerminateStream_Handler,
		},
		{
			MethodName: "ArchiveStream",
			Handler:    _Services_ArchiveStream_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptor0,
}

func init() {
	proto.RegisterFile("github.com/luci/luci-go/logdog/api/endpoints/coordinator/services/v1/service.proto", fileDescriptor0)
}

var fileDescriptor0 = []byte{
	// 703 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xac, 0x54, 0xdf, 0x6e, 0xd3, 0x3e,
	0x14, 0x56, 0xda, 0x6d, 0xbf, 0xf6, 0x74, 0x7f, 0x7e, 0x33, 0x5b, 0x95, 0x66, 0x6c, 0x83, 0x22,
	0x10, 0x12, 0x90, 0x88, 0x71, 0x8f, 0x84, 0xc6, 0x84, 0x26, 0x26, 0x31, 0x52, 0xd8, 0x6d, 0x95,
	0x25, 0x5e, 0x66, 0x68, 0xe3, 0x62, 0x3b, 0x15, 0xe5, 0x8e, 0x37, 0xe0, 0x29, 0x90, 0x78, 0x24,
	0x5e, 0x83, 0x27, 0xc0, 0x39, 0x76, 0xb2, 0xad, 0xeb, 0xa4, 0x6a, 0xe2, 0x26, 0x8a, 0xcf, 0x77,
	0xbe, 0xf3, 0xcf, 0xdf, 0x31, 0x84, 0x29, 0x53, 0xe7, 0xf9, 0xa9, 0x1f, 0xf3, 0x61, 0x30, 0xc8,
	0x63, 0x86, 0x9f, 0x67, 0x29, 0x0f, 0x06, 0x3c, 0x4d, 0x78, 0x1a, 0x44, 0x23, 0x16, 0xd0, 0x2c,
	0x19, 0x71, 0x96, 0x29, 0x19, 0xc4, 0x9c, 0x8b, 0x84, 0x65, 0x91, 0xe2, 0x22, 0x90, 0x54, 0x8c,
	0x59, 0x4c, 0x65, 0x30, 0x7e, 0x5e, 0xfe, 0xfb, 0x23, 0xc1, 0x15, 0x27, 0x4b, 0x86, 0xeb, 0x1d,
	0xff, 0x9b, 0xd8, 0x2a, 0x52, 0x36, 0xb2, 0xb7, 0x93, 0x72, 0x9e, 0x0e, 0x68, 0x80, 0xa7, 0xd3,
	0xfc, 0x2c, 0x48, 0x72, 0x11, 0x29, 0xc6, 0x33, 0x8b, 0x6f, 0x4d, 0xe3, 0x74, 0x38, 0x52, 0x13,
	0x03, 0x76, 0x7f, 0x38, 0xb0, 0xfe, 0x86, 0xaa, 0x7d, 0x9e, 0x9d, 0xb1, 0x34, 0xa4, 0x72, 0xc4,
	0x33, 0x49, 0xc9, 0x53, 0x20, 0x31, 0x5a, 0xfa, 0x36, 0x69, 0x3f, 0x17, 0x03, 0xd7, 0xb9, 0xe7,
	0x3c, 0x6e, 0x86, 0xff, 0x1b, 0xa4, 0x67, 0x80, 0x8f, 0x62, 0x40, 0xb6, 0x01, 0x2a, 0x6f, 0xe5,
	0xd6, 0xd0, 0xab, 0x59, 0x7a, 0x29, 0xe2, 0xc3, 0x9d, 0x32, 0x8a, 0x75, 0x1b, 0x45, 0xea, 0xdc,
	0xad, 0xa3, 0xdf, 0xba, 0x85, 0x4c, 0x01, 0xc7, 0x1a, 0xe8, 0xfe, 0x72, 0x60, 0x33, 0xa4, 0x29,
	0x93, 0x8a, 0x8a, 0x9e, 0x12, 0x34, 0x1a, 0x86, 0xf4, 0x4b, 0x4e, 0xa5, 0x22, 0x2e, 0xfc, 0xa7,
	0xab, 0xfe, 0x44, 0x63, 0x65, 0x6b, 0x29, 0x8f, 0xa4, 0x0d, 0x4b, 0x92, 0xc6, 0xc2, 0xa6, 0x5f,
	0x0e, 0xed, 0x89, 0x3c, 0x80, 0x15, 0xec, 0xb3, 0x3f, 0xa6, 0x42, 0xea, 0x91, 0xd8, 0xac, 0xcb,
	0x68, 0x3c, 0x31, 0x36, 0x42, 0x60, 0x21, 0xa1, 0x32, 0x76, 0x17, 0x90, 0x8a, 0xff, 0xe4, 0x21,
	0xac, 0xea, 0xf4, 0x43, 0x3d, 0xfa, 0x41, 0x9f, 0x65, 0x09, 0xfd, 0xea, 0x2e, 0x6a, 0xb4, 0x1e,
	0xae, 0x94, 0xd6, 0xc3, 0xc2, 0xd8, 0x3d, 0x81, 0xf6, 0x74, 0xa9, 0x76, 0x84, 0xab, 0x50, 0x63,
	0x89, 0x2d, 0x53, 0xff, 0xe9, 0x91, 0x2e, 0xe2, 0xa5, 0x61, 0x81, 0xad, 0xbd, 0xb6, 0x6f, 0xee,
	0xdb, 0x3f, 0xe2, 0xa9, 0x61, 0xf6, 0x0a, 0x34, 0x34, 0x4e, 0xdd, 0xf7, 0xb0, 0x7e, 0xc4, 0xa3,
	0x64, 0xde, 0xf6, 0x4d, 0xb2, 0x5a, 0x95, 0xac, 0xec, 0xa8, 0xe8, 0xb6, 0x61, 0x3a, 0xea, 0xfe,
	0x74, 0x80, 0x5c, 0x8e, 0x59, 0x5d, 0xb5, 0xad, 0xcb, 0x99, 0xa3, 0xae, 0x2a, 0x70, 0xed, 0xd2,
	0xa8, 0x9e, 0x40, 0x3d, 0x4a, 0x29, 0xe6, 0x6a, 0xed, 0x75, 0x7c, 0xa3, 0x36, 0xbf, 0x54, 0x9b,
	0xff, 0xda, 0xaa, 0x31, 0x2c, 0xbc, 0xc8, 0x7d, 0x58, 0x8e, 0x44, 0x7c, 0xce, 0xc6, 0x7a, 0xae,
	0x9f, 0xe9, 0xc4, 0xce, 0xbc, 0x55, 0xda, 0xde, 0xd2, 0x49, 0xf7, 0xbb, 0x03, 0xed, 0x0f, 0x66,
	0xca, 0x8a, 0xde, 0x76, 0x02, 0x17, 0x82, 0xa8, 0x5f, 0x11, 0xc4, 0xf5, 0x7b, 0x5d, 0x98, 0x75,
	0xaf, 0xbf, 0x6b, 0xb0, 0xf1, 0x0a, 0x6b, 0xba, 0x75, 0x05, 0x8f, 0x60, 0x4d, 0x8f, 0xb2, 0x4f,
	0x33, 0x25, 0x26, 0x5a, 0xf8, 0x79, 0x66, 0x4a, 0xd1, 0xa9, 0xb4, 0xf9, 0xa0, 0xb0, 0xee, 0x17,
	0xc6, 0x39, 0x2b, 0x22, 0x1b, 0xb0, 0x48, 0x85, 0xe0, 0x02, 0x75, 0xd8, 0x0c, 0xcd, 0xa1, 0x58,
	0x3d, 0x89, 0xf5, 0xe1, 0x82, 0x82, 0x59, 0x3d, 0x63, 0x29, 0x36, 0x73, 0x17, 0x5a, 0x16, 0x96,
	0xec, 0x1b, 0x75, 0x5b, 0x18, 0xd8, 0x32, 0x7a, 0xda, 0x42, 0xb6, 0xa0, 0x89, 0x39, 0x91, 0xbe,
	0x81, 0xf4, 0x06, 0x1a, 0xec, 0x5e, 0x1b, 0x10, 0xc9, 0x9b, 0x48, 0x36, 0xee, 0xc8, 0xed, 0x40,
	0x23, 0x89, 0x54, 0x84, 0xd4, 0x1d, 0x33, 0x8b, 0xe2, 0x5c, 0x30, 0x75, 0x58, 0x84, 0x90, 0xb8,
	0x8b, 0x44, 0xf4, 0x2d, 0x78, 0x7b, 0x7f, 0x6a, 0xd0, 0xb0, 0xaf, 0x87, 0x24, 0x2f, 0xa1, 0x59,
	0x3d, 0x3f, 0xa4, 0x7d, 0x4d, 0x3c, 0x07, 0xc5, 0x53, 0xe5, 0x75, 0x4a, 0x51, 0x5e, 0x7f, 0xa9,
	0xde, 0xc1, 0xea, 0xd5, 0x05, 0x24, 0xdb, 0xa5, 0xf3, 0xcc, 0x37, 0xc4, 0xdb, 0xb9, 0x09, 0xb6,
	0x01, 0xf7, 0x01, 0x2e, 0xb6, 0x84, 0x74, 0x2e, 0xd6, 0x61, 0x6a, 0x1b, 0x3d, 0x6f, 0x16, 0x64,
	0x83, 0x1c, 0xc2, 0xda, 0x94, 0x82, 0x49, 0x95, 0x77, 0xb6, 0xb4, 0xbd, 0x1b, 0x7a, 0x27, 0x07,
	0xb0, 0x72, 0x45, 0x88, 0xe4, 0x6e, 0x19, 0x68, 0x96, 0x3e, 0x6f, 0x0a, 0x73, 0xba, 0x84, 0xe7,
	0x17, 0x7f, 0x03, 0x00, 0x00, 0xff, 0xff, 0x8b, 0x1b, 0xf5, 0xd1, 0xdb, 0x06, 0x00, 0x00,
}
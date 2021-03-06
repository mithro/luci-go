// Code generated by protoc-gen-go.
// source: acquisition_network_device.proto
// DO NOT EDIT!

/*
Package ts_mon_proto is a generated protocol buffer package.

It is generated from these files:
	acquisition_network_device.proto
	acquisition_task.proto
	metrics.proto

It has these top-level messages:
	NetworkDevice
	Task
	MetricsCollection
	MetricsField
	PrecomputedDistribution
	MetricsData
*/
package ts_mon_proto

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type NetworkDevice_TypeId int32

const (
	NetworkDevice_MESSAGE_TYPE_ID NetworkDevice_TypeId = 34049749
)

var NetworkDevice_TypeId_name = map[int32]string{
	34049749: "MESSAGE_TYPE_ID",
}
var NetworkDevice_TypeId_value = map[string]int32{
	"MESSAGE_TYPE_ID": 34049749,
}

func (x NetworkDevice_TypeId) Enum() *NetworkDevice_TypeId {
	p := new(NetworkDevice_TypeId)
	*p = x
	return p
}
func (x NetworkDevice_TypeId) String() string {
	return proto.EnumName(NetworkDevice_TypeId_name, int32(x))
}
func (x *NetworkDevice_TypeId) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(NetworkDevice_TypeId_value, data, "NetworkDevice_TypeId")
	if err != nil {
		return err
	}
	*x = NetworkDevice_TypeId(value)
	return nil
}
func (NetworkDevice_TypeId) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 0} }

type NetworkDevice struct {
	Alertable        *bool   `protobuf:"varint,101,opt,name=alertable" json:"alertable,omitempty"`
	Realm            *string `protobuf:"bytes,102,opt,name=realm" json:"realm,omitempty"`
	Metro            *string `protobuf:"bytes,104,opt,name=metro" json:"metro,omitempty"`
	Role             *string `protobuf:"bytes,105,opt,name=role" json:"role,omitempty"`
	Hostname         *string `protobuf:"bytes,106,opt,name=hostname" json:"hostname,omitempty"`
	Hostgroup        *string `protobuf:"bytes,108,opt,name=hostgroup" json:"hostgroup,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *NetworkDevice) Reset()                    { *m = NetworkDevice{} }
func (m *NetworkDevice) String() string            { return proto.CompactTextString(m) }
func (*NetworkDevice) ProtoMessage()               {}
func (*NetworkDevice) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *NetworkDevice) GetAlertable() bool {
	if m != nil && m.Alertable != nil {
		return *m.Alertable
	}
	return false
}

func (m *NetworkDevice) GetRealm() string {
	if m != nil && m.Realm != nil {
		return *m.Realm
	}
	return ""
}

func (m *NetworkDevice) GetMetro() string {
	if m != nil && m.Metro != nil {
		return *m.Metro
	}
	return ""
}

func (m *NetworkDevice) GetRole() string {
	if m != nil && m.Role != nil {
		return *m.Role
	}
	return ""
}

func (m *NetworkDevice) GetHostname() string {
	if m != nil && m.Hostname != nil {
		return *m.Hostname
	}
	return ""
}

func (m *NetworkDevice) GetHostgroup() string {
	if m != nil && m.Hostgroup != nil {
		return *m.Hostgroup
	}
	return ""
}

func init() {
	proto.RegisterType((*NetworkDevice)(nil), "ts_mon.proto.NetworkDevice")
	proto.RegisterEnum("ts_mon.proto.NetworkDevice_TypeId", NetworkDevice_TypeId_name, NetworkDevice_TypeId_value)
}

var fileDescriptor0 = []byte{
	// 209 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0x52, 0x48, 0x4c, 0x2e, 0x2c,
	0xcd, 0x2c, 0xce, 0x2c, 0xc9, 0xcc, 0xcf, 0x8b, 0xcf, 0x4b, 0x2d, 0x29, 0xcf, 0x2f, 0xca, 0x8e,
	0x4f, 0x49, 0x2d, 0xcb, 0x4c, 0x4e, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x29, 0x29,
	0x8e, 0xcf, 0xcd, 0xcf, 0x83, 0xf0, 0x94, 0x4e, 0x32, 0x72, 0xf1, 0xfa, 0x41, 0x94, 0xb9, 0x80,
	0x55, 0x09, 0xc9, 0x70, 0x71, 0x26, 0xe6, 0xa4, 0x16, 0x95, 0x24, 0x26, 0xe5, 0xa4, 0x4a, 0xa4,
	0x2a, 0x30, 0x6a, 0x70, 0x04, 0x21, 0x04, 0x84, 0x44, 0xb8, 0x58, 0x8b, 0x52, 0x13, 0x73, 0x72,
	0x25, 0xd2, 0x80, 0x32, 0x9c, 0x41, 0x10, 0x0e, 0x48, 0x34, 0x37, 0xb5, 0xa4, 0x28, 0x5f, 0x22,
	0x03, 0x22, 0x0a, 0xe6, 0x08, 0x09, 0x71, 0xb1, 0x14, 0xe5, 0x03, 0x0d, 0xc9, 0x04, 0x0b, 0x82,
	0xd9, 0x42, 0x52, 0x5c, 0x1c, 0x19, 0xf9, 0xc5, 0x25, 0x79, 0x89, 0xb9, 0xa9, 0x12, 0x59, 0x60,
	0x71, 0x38, 0x1f, 0x64, 0x33, 0x88, 0x9d, 0x5e, 0x94, 0x5f, 0x5a, 0x20, 0x91, 0x03, 0x96, 0x44,
	0x08, 0x28, 0x29, 0x70, 0xb1, 0x85, 0x54, 0x16, 0xa4, 0x7a, 0xa6, 0x08, 0x89, 0x71, 0xf1, 0xfb,
	0xba, 0x06, 0x07, 0x3b, 0xba, 0xbb, 0xc6, 0x87, 0x44, 0x06, 0xb8, 0xc6, 0x7b, 0xba, 0x08, 0x5c,
	0x9d, 0x3b, 0x4f, 0x00, 0x10, 0x00, 0x00, 0xff, 0xff, 0xa8, 0x5e, 0x04, 0xaa, 0xfc, 0x00, 0x00,
	0x00,
}

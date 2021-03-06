// Code generated by protoc-gen-go. DO NOT EDIT.
// source: rpc_header.proto

package raft

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
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

type ProtocolVersion int32

const (
	ProtocolVersion_ProtocolVersionMin ProtocolVersion = 0
	ProtocolVersion_ProtocolVersionOne ProtocolVersion = 1
	ProtocolVersion_ProtocolVersionTwo ProtocolVersion = 2
	ProtocolVersion_ProtocolVersionMax ProtocolVersion = 3
)

var ProtocolVersion_name = map[int32]string{
	0: "ProtocolVersionMin",
	1: "ProtocolVersionOne",
	2: "ProtocolVersionTwo",
	3: "ProtocolVersionMax",
}

var ProtocolVersion_value = map[string]int32{
	"ProtocolVersionMin": 0,
	"ProtocolVersionOne": 1,
	"ProtocolVersionTwo": 2,
	"ProtocolVersionMax": 3,
}

func (x ProtocolVersion) String() string {
	return proto.EnumName(ProtocolVersion_name, int32(x))
}

func (ProtocolVersion) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_5d4f639ec9fd9c94, []int{0}
}

type RPCHeader struct {
	ProtocolVersion      ProtocolVersion `protobuf:"varint,1,opt,name=ProtocolVersion,proto3,enum=raft.ProtocolVersion" json:"ProtocolVersion,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *RPCHeader) Reset()         { *m = RPCHeader{} }
func (m *RPCHeader) String() string { return proto.CompactTextString(m) }
func (*RPCHeader) ProtoMessage()    {}
func (*RPCHeader) Descriptor() ([]byte, []int) {
	return fileDescriptor_5d4f639ec9fd9c94, []int{0}
}

func (m *RPCHeader) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RPCHeader.Unmarshal(m, b)
}
func (m *RPCHeader) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RPCHeader.Marshal(b, m, deterministic)
}
func (m *RPCHeader) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RPCHeader.Merge(m, src)
}
func (m *RPCHeader) XXX_Size() int {
	return xxx_messageInfo_RPCHeader.Size(m)
}
func (m *RPCHeader) XXX_DiscardUnknown() {
	xxx_messageInfo_RPCHeader.DiscardUnknown(m)
}

var xxx_messageInfo_RPCHeader proto.InternalMessageInfo

func (m *RPCHeader) GetProtocolVersion() ProtocolVersion {
	if m != nil {
		return m.ProtocolVersion
	}
	return ProtocolVersion_ProtocolVersionMin
}

func init() {
	proto.RegisterEnum("raft.ProtocolVersion", ProtocolVersion_name, ProtocolVersion_value)
	proto.RegisterType((*RPCHeader)(nil), "raft.RPCHeader")
}

func init() { proto.RegisterFile("rpc_header.proto", fileDescriptor_5d4f639ec9fd9c94) }

var fileDescriptor_5d4f639ec9fd9c94 = []byte{
	// 133 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x28, 0x2a, 0x48, 0x8e,
	0xcf, 0x48, 0x4d, 0x4c, 0x49, 0x2d, 0xd2, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x29, 0x4a,
	0x4c, 0x2b, 0x51, 0xf2, 0xe1, 0xe2, 0x0c, 0x0a, 0x70, 0xf6, 0x00, 0x4b, 0x08, 0xd9, 0x73, 0xf1,
	0x07, 0x80, 0xe4, 0x92, 0xf3, 0x73, 0xc2, 0x52, 0x8b, 0x8a, 0x33, 0xf3, 0xf3, 0x24, 0x18, 0x15,
	0x18, 0x35, 0xf8, 0x8c, 0x44, 0xf5, 0x40, 0x8a, 0xf5, 0xd0, 0x24, 0x83, 0xd0, 0x55, 0x6b, 0x15,
	0x62, 0x18, 0x20, 0x24, 0xc6, 0x25, 0x84, 0x26, 0xe4, 0x9b, 0x99, 0x27, 0xc0, 0x80, 0x45, 0xdc,
	0x3f, 0x2f, 0x55, 0x80, 0x11, 0x8b, 0x78, 0x48, 0x79, 0xbe, 0x00, 0x13, 0x36, 0x73, 0x12, 0x2b,
	0x04, 0x98, 0x93, 0xd8, 0xc0, 0xbe, 0x31, 0x06, 0x04, 0x00, 0x00, 0xff, 0xff, 0x20, 0xb8, 0xbf,
	0x69, 0xe1, 0x00, 0x00, 0x00,
}

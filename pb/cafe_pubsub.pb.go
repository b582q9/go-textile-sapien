// Code generated by protoc-gen-go. DO NOT EDIT.
// source: cafe_pubsub.proto

package pb

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

type CafePubSubContactRequest struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	FindId               string   `protobuf:"bytes,2,opt,name=findId,proto3" json:"findId,omitempty"`
	FindAddress          string   `protobuf:"bytes,3,opt,name=findAddress,proto3" json:"findAddress,omitempty"`
	FindUsername         string   `protobuf:"bytes,4,opt,name=findUsername,proto3" json:"findUsername,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CafePubSubContactRequest) Reset()         { *m = CafePubSubContactRequest{} }
func (m *CafePubSubContactRequest) String() string { return proto.CompactTextString(m) }
func (*CafePubSubContactRequest) ProtoMessage()    {}
func (*CafePubSubContactRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_cafe_pubsub_cccd3d3fbe353225, []int{0}
}
func (m *CafePubSubContactRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CafePubSubContactRequest.Unmarshal(m, b)
}
func (m *CafePubSubContactRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CafePubSubContactRequest.Marshal(b, m, deterministic)
}
func (dst *CafePubSubContactRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CafePubSubContactRequest.Merge(dst, src)
}
func (m *CafePubSubContactRequest) XXX_Size() int {
	return xxx_messageInfo_CafePubSubContactRequest.Size(m)
}
func (m *CafePubSubContactRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_CafePubSubContactRequest.DiscardUnknown(m)
}

var xxx_messageInfo_CafePubSubContactRequest proto.InternalMessageInfo

func (m *CafePubSubContactRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *CafePubSubContactRequest) GetFindId() string {
	if m != nil {
		return m.FindId
	}
	return ""
}

func (m *CafePubSubContactRequest) GetFindAddress() string {
	if m != nil {
		return m.FindAddress
	}
	return ""
}

func (m *CafePubSubContactRequest) GetFindUsername() string {
	if m != nil {
		return m.FindUsername
	}
	return ""
}

type CafePubSubContactResult struct {
	Id                   string     `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Contacts             []*Contact `protobuf:"bytes,2,rep,name=contacts,proto3" json:"contacts,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *CafePubSubContactResult) Reset()         { *m = CafePubSubContactResult{} }
func (m *CafePubSubContactResult) String() string { return proto.CompactTextString(m) }
func (*CafePubSubContactResult) ProtoMessage()    {}
func (*CafePubSubContactResult) Descriptor() ([]byte, []int) {
	return fileDescriptor_cafe_pubsub_cccd3d3fbe353225, []int{1}
}
func (m *CafePubSubContactResult) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CafePubSubContactResult.Unmarshal(m, b)
}
func (m *CafePubSubContactResult) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CafePubSubContactResult.Marshal(b, m, deterministic)
}
func (dst *CafePubSubContactResult) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CafePubSubContactResult.Merge(dst, src)
}
func (m *CafePubSubContactResult) XXX_Size() int {
	return xxx_messageInfo_CafePubSubContactResult.Size(m)
}
func (m *CafePubSubContactResult) XXX_DiscardUnknown() {
	xxx_messageInfo_CafePubSubContactResult.DiscardUnknown(m)
}

var xxx_messageInfo_CafePubSubContactResult proto.InternalMessageInfo

func (m *CafePubSubContactResult) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *CafePubSubContactResult) GetContacts() []*Contact {
	if m != nil {
		return m.Contacts
	}
	return nil
}

func init() {
	proto.RegisterType((*CafePubSubContactRequest)(nil), "CafePubSubContactRequest")
	proto.RegisterType((*CafePubSubContactResult)(nil), "CafePubSubContactResult")
}

func init() { proto.RegisterFile("cafe_pubsub.proto", fileDescriptor_cafe_pubsub_cccd3d3fbe353225) }

var fileDescriptor_cafe_pubsub_cccd3d3fbe353225 = []byte{
	// 194 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x4c, 0x4e, 0x4c, 0x4b,
	0x8d, 0x2f, 0x28, 0x4d, 0x2a, 0x2e, 0x4d, 0xd2, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x97, 0xe2, 0xce,
	0xcd, 0x4f, 0x49, 0xcd, 0x81, 0x70, 0x94, 0x3a, 0x18, 0xb9, 0x24, 0x9c, 0x13, 0xd3, 0x52, 0x03,
	0x4a, 0x93, 0x82, 0x4b, 0x93, 0x9c, 0xf3, 0xf3, 0x4a, 0x12, 0x93, 0x4b, 0x82, 0x52, 0x0b, 0x4b,
	0x53, 0x8b, 0x4b, 0x84, 0xf8, 0xb8, 0x98, 0x32, 0x53, 0x24, 0x18, 0x15, 0x18, 0x35, 0x38, 0x83,
	0x98, 0x32, 0x53, 0x84, 0xc4, 0xb8, 0xd8, 0xd2, 0x32, 0xf3, 0x52, 0x3c, 0x53, 0x24, 0x98, 0xc0,
	0x62, 0x50, 0x9e, 0x90, 0x02, 0x17, 0x37, 0x88, 0xe5, 0x98, 0x92, 0x52, 0x94, 0x5a, 0x5c, 0x2c,
	0xc1, 0x0c, 0x96, 0x44, 0x16, 0x12, 0x52, 0xe2, 0xe2, 0x01, 0x71, 0x43, 0x8b, 0x53, 0x8b, 0xf2,
	0x12, 0x73, 0x53, 0x25, 0x58, 0xc0, 0x4a, 0x50, 0xc4, 0x94, 0xfc, 0xb9, 0xc4, 0xb1, 0xb8, 0xa4,
	0xb8, 0x34, 0x07, 0xd3, 0x21, 0x2a, 0x5c, 0x1c, 0xc9, 0x10, 0x05, 0xc5, 0x12, 0x4c, 0x0a, 0xcc,
	0x1a, 0xdc, 0x46, 0x1c, 0x7a, 0x30, 0x1d, 0x70, 0x19, 0x27, 0x96, 0x28, 0xa6, 0x82, 0xa4, 0x24,
	0x36, 0xb0, 0x47, 0x8d, 0x01, 0x01, 0x00, 0x00, 0xff, 0xff, 0x4e, 0x1b, 0x45, 0xce, 0x0a, 0x01,
	0x00, 0x00,
}

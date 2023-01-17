// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: auto-ibc-tx/v1beta1/genesis.proto

package types

import (
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// GenesisState - genesis state of x/auto-ibc-tx
type GenesisState struct {
	Params                     Params       `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	InterchainAccountAddresses []string     `protobuf:"bytes,2,rep,name=interchain_account_addresses,json=interchainAccountAddresses,proto3" json:"interchain_account_addresses,omitempty"`
	AutoTxInfos                []AutoTxInfo `protobuf:"bytes,3,rep,name=auto_tx_infos,json=autoTxInfos,proto3" json:"auto_tx_infos,omitempty"`
	Sequences                  []Sequence   `protobuf:"bytes,4,rep,name=sequences,proto3" json:"sequences,omitempty"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_663b4df31d9d6f5b, []int{0}
}
func (m *GenesisState) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GenesisState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GenesisState.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GenesisState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GenesisState.Merge(m, src)
}
func (m *GenesisState) XXX_Size() int {
	return m.Size()
}
func (m *GenesisState) XXX_DiscardUnknown() {
	xxx_messageInfo_GenesisState.DiscardUnknown(m)
}

var xxx_messageInfo_GenesisState proto.InternalMessageInfo

func (m *GenesisState) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

func (m *GenesisState) GetInterchainAccountAddresses() []string {
	if m != nil {
		return m.InterchainAccountAddresses
	}
	return nil
}

func (m *GenesisState) GetAutoTxInfos() []AutoTxInfo {
	if m != nil {
		return m.AutoTxInfos
	}
	return nil
}

func (m *GenesisState) GetSequences() []Sequence {
	if m != nil {
		return m.Sequences
	}
	return nil
}

// Sequence id and value of a counter
type Sequence struct {
	IDKey []byte `protobuf:"bytes,1,opt,name=id_key,json=idKey,proto3" json:"id_key,omitempty"`
	Value uint64 `protobuf:"varint,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (m *Sequence) Reset()         { *m = Sequence{} }
func (m *Sequence) String() string { return proto.CompactTextString(m) }
func (*Sequence) ProtoMessage()    {}
func (*Sequence) Descriptor() ([]byte, []int) {
	return fileDescriptor_663b4df31d9d6f5b, []int{1}
}
func (m *Sequence) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Sequence) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Sequence.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Sequence) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Sequence.Merge(m, src)
}
func (m *Sequence) XXX_Size() int {
	return m.Size()
}
func (m *Sequence) XXX_DiscardUnknown() {
	xxx_messageInfo_Sequence.DiscardUnknown(m)
}

var xxx_messageInfo_Sequence proto.InternalMessageInfo

func (m *Sequence) GetIDKey() []byte {
	if m != nil {
		return m.IDKey
	}
	return nil
}

func (m *Sequence) GetValue() uint64 {
	if m != nil {
		return m.Value
	}
	return 0
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "trst.x.autoibctx.v1beta1.GenesisState")
	proto.RegisterType((*Sequence)(nil), "trst.x.autoibctx.v1beta1.Sequence")
}

func init() { proto.RegisterFile("auto-ibc-tx/v1beta1/genesis.proto", fileDescriptor_663b4df31d9d6f5b) }

var fileDescriptor_663b4df31d9d6f5b = []byte{
	// 405 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x92, 0x4f, 0x6b, 0xd4, 0x40,
	0x18, 0xc6, 0x93, 0x66, 0x77, 0x71, 0x67, 0xeb, 0x25, 0x16, 0x0c, 0xab, 0x24, 0x71, 0x11, 0xc9,
	0xa1, 0x9d, 0xd0, 0x7a, 0x17, 0x36, 0x88, 0x52, 0x7a, 0x91, 0xd4, 0x93, 0x20, 0x61, 0x92, 0xbc,
	0x4d, 0x07, 0x9b, 0x4c, 0xcc, 0xbc, 0x29, 0xc9, 0x57, 0xf0, 0xe4, 0xc7, 0xea, 0xb1, 0x47, 0x4f,
	0x8b, 0x64, 0x6f, 0x7e, 0x0a, 0xc9, 0x9f, 0xee, 0xae, 0xe0, 0xf6, 0x36, 0x81, 0xdf, 0xf3, 0xfc,
	0x1e, 0xc8, 0x4b, 0x5e, 0xb1, 0x12, 0xc5, 0x09, 0x0f, 0xa3, 0x13, 0xac, 0xdc, 0xdb, 0xd3, 0x10,
	0x90, 0x9d, 0xba, 0x09, 0x64, 0x20, 0xb9, 0xa4, 0x79, 0x21, 0x50, 0xe8, 0x06, 0x16, 0x12, 0x69,
	0x45, 0x5b, 0x92, 0x87, 0x11, 0x56, 0x74, 0xe0, 0xe6, 0x47, 0x89, 0x48, 0x44, 0x07, 0xb9, 0xed,
	0xab, 0xe7, 0xe7, 0xd6, 0xff, 0x2a, 0xb1, 0xce, 0x61, 0x28, 0x5c, 0xfc, 0xd0, 0xc8, 0xe1, 0xc7,
	0x5e, 0x71, 0x89, 0x0c, 0x41, 0x7f, 0x47, 0x26, 0x39, 0x2b, 0x58, 0x2a, 0x0d, 0xd5, 0x56, 0x9d,
	0xd9, 0x99, 0x4d, 0xf7, 0x29, 0xe9, 0xa7, 0x8e, 0xf3, 0x46, 0x77, 0x2b, 0x4b, 0xf1, 0x87, 0x94,
	0x9e, 0x93, 0x97, 0x3c, 0x43, 0x28, 0xa2, 0x6b, 0xc6, 0xb3, 0x80, 0x45, 0x91, 0x28, 0x33, 0x0c,
	0x58, 0x1c, 0x17, 0x20, 0x25, 0x48, 0xe3, 0xc0, 0xd6, 0x9c, 0xa9, 0x47, 0xdb, 0xcc, 0x9f, 0x95,
	0xf5, 0xe6, 0x31, 0xf6, 0x58, 0xa4, 0x1c, 0x21, 0xcd, 0xb1, 0xf6, 0xe7, 0x5b, 0x6e, 0xd9, 0x63,
	0xcb, 0x07, 0x4a, 0xe7, 0xe4, 0x69, 0xbb, 0x2d, 0xc0, 0x2a, 0xe0, 0xd9, 0x95, 0x90, 0x86, 0x66,
	0x6b, 0xce, 0xec, 0xec, 0xf5, 0xfe, 0xe1, 0xcb, 0x12, 0xc5, 0xe7, 0xea, 0x3c, 0xbb, 0x12, 0x9e,
	0x35, 0x0c, 0x79, 0xfe, 0x4f, 0xc5, 0x8e, 0x79, 0xc6, 0x36, 0xb0, 0xd4, 0xbf, 0x92, 0xa9, 0x84,
	0xef, 0x25, 0x64, 0x11, 0x48, 0x63, 0xd4, 0x69, 0x16, 0xfb, 0x35, 0x97, 0x03, 0xea, 0xbd, 0x18,
	0x24, 0xcf, 0x36, 0xe1, 0x1d, 0xc1, 0xb6, 0x71, 0xe1, 0x91, 0x27, 0x0f, 0x19, 0xdd, 0x26, 0x13,
	0x1e, 0x07, 0xdf, 0xa0, 0xee, 0xfe, 0xc3, 0xa1, 0x37, 0x6d, 0x56, 0xd6, 0xf8, 0xfc, 0xfd, 0x05,
	0xd4, 0xfe, 0x98, 0xc7, 0x17, 0x50, 0xeb, 0x47, 0x64, 0x7c, 0xcb, 0x6e, 0x4a, 0x30, 0x0e, 0x6c,
	0xd5, 0x19, 0xf9, 0xfd, 0x87, 0xf7, 0xe1, 0xae, 0x31, 0xd5, 0xfb, 0xc6, 0x54, 0x7f, 0x37, 0xa6,
	0xfa, 0x73, 0x6d, 0x2a, 0xf7, 0x6b, 0x53, 0xf9, 0xb5, 0x36, 0x95, 0x2f, 0xc7, 0x09, 0xc7, 0xeb,
	0x32, 0xa4, 0x91, 0x48, 0xdd, 0x76, 0xf3, 0x0d, 0x0b, 0x65, 0xf7, 0x70, 0x2b, 0x77, 0xf7, 0x4c,
	0xba, 0xf3, 0x08, 0x27, 0xdd, 0x7d, 0xbc, 0xfd, 0x1b, 0x00, 0x00, 0xff, 0xff, 0xd1, 0xc1, 0xc5,
	0xf6, 0x95, 0x02, 0x00, 0x00,
}

func (m *GenesisState) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GenesisState) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GenesisState) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Sequences) > 0 {
		for iNdEx := len(m.Sequences) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Sequences[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x22
		}
	}
	if len(m.AutoTxInfos) > 0 {
		for iNdEx := len(m.AutoTxInfos) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.AutoTxInfos[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.InterchainAccountAddresses) > 0 {
		for iNdEx := len(m.InterchainAccountAddresses) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.InterchainAccountAddresses[iNdEx])
			copy(dAtA[i:], m.InterchainAccountAddresses[iNdEx])
			i = encodeVarintGenesis(dAtA, i, uint64(len(m.InterchainAccountAddresses[iNdEx])))
			i--
			dAtA[i] = 0x12
		}
	}
	{
		size, err := m.Params.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *Sequence) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Sequence) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Sequence) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Value != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.Value))
		i--
		dAtA[i] = 0x10
	}
	if len(m.IDKey) > 0 {
		i -= len(m.IDKey)
		copy(dAtA[i:], m.IDKey)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.IDKey)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintGenesis(dAtA []byte, offset int, v uint64) int {
	offset -= sovGenesis(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *GenesisState) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Params.Size()
	n += 1 + l + sovGenesis(uint64(l))
	if len(m.InterchainAccountAddresses) > 0 {
		for _, s := range m.InterchainAccountAddresses {
			l = len(s)
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.AutoTxInfos) > 0 {
		for _, e := range m.AutoTxInfos {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.Sequences) > 0 {
		for _, e := range m.Sequences {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	return n
}

func (m *Sequence) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.IDKey)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.Value != 0 {
		n += 1 + sovGenesis(uint64(m.Value))
	}
	return n
}

func sovGenesis(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozGenesis(x uint64) (n int) {
	return sovGenesis(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *GenesisState) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: GenesisState: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GenesisState: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Params", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Params.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field InterchainAccountAddresses", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.InterchainAccountAddresses = append(m.InterchainAccountAddresses, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AutoTxInfos", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AutoTxInfos = append(m.AutoTxInfos, AutoTxInfo{})
			if err := m.AutoTxInfos[len(m.AutoTxInfos)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Sequences", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Sequences = append(m.Sequences, Sequence{})
			if err := m.Sequences[len(m.Sequences)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Sequence) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Sequence: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Sequence: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IDKey", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.IDKey = append(m.IDKey[:0], dAtA[iNdEx:postIndex]...)
			if m.IDKey == nil {
				m.IDKey = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Value", wireType)
			}
			m.Value = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Value |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipGenesis(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowGenesis
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthGenesis
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupGenesis
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthGenesis
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthGenesis        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowGenesis          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupGenesis = fmt.Errorf("proto: unexpected end of group")
)

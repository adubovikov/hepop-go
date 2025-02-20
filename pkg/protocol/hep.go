package protocol

import (
	"encoding/binary"
	"errors"
	"net"
)

const (
	HEPv1 = 1
	HEPv2 = 2
	HEPv3 = 3
)

// HEP chunk types
const (
	TypeIPProtocolFamily  = 0x0001
	TypeIPProtocolID      = 0x0002
	TypeIPSourceIP        = 0x0003
	TypeIPDestinationIP   = 0x0004
	TypeIPSourcePort      = 0x0007
	TypeIPDestinationPort = 0x0008
	TypeTimestamp         = 0x0009
	TypeProtocolType      = 0x000a
	TypeCaptureAgentID    = 0x000b
	TypeKeepAliveTimer    = 0x000c
	TypeAuthKey           = 0x000e
	TypePayload           = 0x000f
	TypeCorrelationID     = 0x0011
	TypeVLAN              = 0x0012
)

var (
	ErrInvalidVersion = errors.New("invalid HEP version")
	ErrPacketTooShort = errors.New("packet too short")
	ErrInvalidChunk   = errors.New("invalid chunk")
)

type HEPPacket struct {
	Version   uint8
	Protocol  uint8
	SrcIP     string
	DstIP     string
	SrcPort   uint16
	DstPort   uint16
	Timestamp uint64
	ProtoType uint8
	NodeID    uint32
	NodeName  string
	Payload   []byte
	CID       string
	Vlan      uint16
}

type hepChunk struct {
	VendorID  uint16
	ChunkType uint16
	Length    uint16
	Data      []byte
}

// HEPv2 header structure
type hepv2Header struct {
	Family    byte
	Protocol  byte
	SrcPort   uint16
	DstPort   uint16
	SrcIP     [4]byte
	DstIP     [4]byte
	Timestamp uint32
	ID        uint32
	Type      byte
}

func DecodeHEP(data []byte) (*HEPPacket, error) {
	if len(data) < 4 {
		return nil, ErrPacketTooShort
	}

	// Check HEP version
	switch data[0] {
	case HEPv3:
		return decodeHEPv3(data)
	case HEPv2:
		return decodeHEPv2(data)
	case HEPv1:
		return decodeHEPv1(data)
	default:
		return nil, ErrInvalidVersion
	}
}

func decodeHEPv3(data []byte) (*HEPPacket, error) {
	if len(data) < 6 {
		return nil, ErrPacketTooShort
	}

	packet := &HEPPacket{
		Version: data[0],
	}

	length := binary.BigEndian.Uint16(data[4:6])
	if int(length) > len(data) {
		return nil, ErrPacketTooShort
	}

	cursor := 6
	for cursor < len(data) {
		if cursor+6 > len(data) {
			return nil, ErrInvalidChunk
		}

		chunk := hepChunk{
			VendorID:  binary.BigEndian.Uint16(data[cursor : cursor+2]),
			ChunkType: binary.BigEndian.Uint16(data[cursor+2 : cursor+4]),
			Length:    binary.BigEndian.Uint16(data[cursor+4 : cursor+6]),
		}

		cursor += 6
		if cursor+int(chunk.Length-6) > len(data) {
			return nil, ErrInvalidChunk
		}

		chunk.Data = data[cursor : cursor+int(chunk.Length-6)]
		cursor += int(chunk.Length - 6)

		switch chunk.ChunkType {
		case TypeIPProtocolFamily:
			packet.Protocol = chunk.Data[0]
		case TypeIPSourceIP:
			packet.SrcIP = net.IP(chunk.Data).String()
		case TypeIPDestinationIP:
			packet.DstIP = net.IP(chunk.Data).String()
		case TypeIPSourcePort:
			packet.SrcPort = binary.BigEndian.Uint16(chunk.Data)
		case TypeIPDestinationPort:
			packet.DstPort = binary.BigEndian.Uint16(chunk.Data)
		case TypeTimestamp:
			packet.Timestamp = binary.BigEndian.Uint64(chunk.Data)
		case TypeProtocolType:
			packet.ProtoType = chunk.Data[0]
		case TypeCaptureAgentID:
			packet.NodeID = binary.BigEndian.Uint32(chunk.Data)
		case TypePayload:
			packet.Payload = make([]byte, len(chunk.Data))
			copy(packet.Payload, chunk.Data)
		case TypeCorrelationID:
			packet.CID = string(chunk.Data)
		case TypeVLAN:
			packet.Vlan = binary.BigEndian.Uint16(chunk.Data)
		}
	}

	return packet, nil
}

func decodeHEPv2(data []byte) (*HEPPacket, error) {
	if len(data) < 31 { // minimum HEPv2 packet size
		return nil, ErrPacketTooShort
	}

	var header hepv2Header
	header.Family = data[1]
	header.Protocol = data[2]
	header.SrcPort = binary.BigEndian.Uint16(data[3:5])
	header.DstPort = binary.BigEndian.Uint16(data[5:7])
	copy(header.SrcIP[:], data[7:11])
	copy(header.DstIP[:], data[11:15])
	header.Timestamp = binary.BigEndian.Uint32(data[15:19])
	header.ID = binary.BigEndian.Uint32(data[19:23])
	header.Type = data[23]

	packet := &HEPPacket{
		Version:   HEPv2,
		Protocol:  header.Protocol,
		SrcIP:     net.IP(header.SrcIP[:]).String(),
		DstIP:     net.IP(header.DstIP[:]).String(),
		SrcPort:   header.SrcPort,
		DstPort:   header.DstPort,
		Timestamp: uint64(header.Timestamp),
		ProtoType: header.Type,
		NodeID:    header.ID,
		Payload:   data[24:],
	}

	return packet, nil
}

func decodeHEPv1(data []byte) (*HEPPacket, error) {
	if len(data) < 21 { // minimum HEPv1 packet size
		return nil, ErrPacketTooShort
	}

	packet := &HEPPacket{
		Version:   HEPv1,
		Protocol:  data[1],
		SrcPort:   binary.BigEndian.Uint16(data[2:4]),
		DstPort:   binary.BigEndian.Uint16(data[4:6]),
		SrcIP:     net.IP(data[6:10]).String(),
		DstIP:     net.IP(data[10:14]).String(),
		Timestamp: uint64(binary.BigEndian.Uint32(data[14:18])),
		NodeID:    binary.BigEndian.Uint32(data[18:22]),
		Payload:   data[22:],
	}

	return packet, nil
}

// Add a helper function to determine the HEP version
func GetHEPVersion(data []byte) (uint8, error) {
	if len(data) < 4 {
		return 0, ErrPacketTooShort
	}
	version := data[0]
	if version < HEPv1 || version > HEPv3 {
		return 0, ErrInvalidVersion
	}
	return version, nil
}

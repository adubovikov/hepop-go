package protocol

import (
	"encoding/binary"
	"net"
	"testing"
	"time"
)

func TestHEPv1Decode(t *testing.T) {
	// Создаем тестовый HEPv1 пакет
	packet := make([]byte, 30)
	packet[0] = HEPv1                                                    // Version
	packet[1] = 17                                                       // UDP protocol
	binary.BigEndian.PutUint16(packet[2:4], 5060)                        // SrcPort
	binary.BigEndian.PutUint16(packet[4:6], 5060)                        // DstPort
	copy(packet[6:10], net.ParseIP("192.168.1.1").To4())                 // SrcIP
	copy(packet[10:14], net.ParseIP("192.168.1.2").To4())                // DstIP
	binary.BigEndian.PutUint32(packet[14:18], uint32(time.Now().Unix())) // Timestamp
	binary.BigEndian.PutUint32(packet[18:22], 2001)                      // NodeID
	copy(packet[22:], []byte("TEST"))                                    // Payload

	hep, err := DecodeHEP(packet)
	if err != nil {
		t.Fatalf("Failed to decode HEPv1: %v", err)
	}

	if hep.Version != HEPv1 {
		t.Errorf("Expected version %d, got %d", HEPv1, hep.Version)
	}
	if hep.Protocol != 17 {
		t.Errorf("Expected protocol 17, got %d", hep.Protocol)
	}
	if hep.SrcPort != 5060 {
		t.Errorf("Expected src port 5060, got %d", hep.SrcPort)
	}
	if hep.SrcIP != "192.168.1.1" {
		t.Errorf("Expected src IP 192.168.1.1, got %s", hep.SrcIP)
	}
}

func TestHEPv2Decode(t *testing.T) {
	// Создаем тестовый HEPv2 пакет
	packet := make([]byte, 35)
	packet[0] = HEPv2                                                    // Version
	packet[1] = 2                                                        // Family (AF_INET)
	packet[2] = 17                                                       // UDP protocol
	binary.BigEndian.PutUint16(packet[3:5], 5060)                        // SrcPort
	binary.BigEndian.PutUint16(packet[5:7], 5060)                        // DstPort
	copy(packet[7:11], net.ParseIP("192.168.1.1").To4())                 // SrcIP
	copy(packet[11:15], net.ParseIP("192.168.1.2").To4())                // DstIP
	binary.BigEndian.PutUint32(packet[15:19], uint32(time.Now().Unix())) // Timestamp
	binary.BigEndian.PutUint32(packet[19:23], 2001)                      // NodeID
	packet[23] = 1                                                       // Type
	copy(packet[24:], []byte("TEST"))                                    // Payload

	hep, err := DecodeHEP(packet)
	if err != nil {
		t.Fatalf("Failed to decode HEPv2: %v", err)
	}

	if hep.Version != HEPv2 {
		t.Errorf("Expected version %d, got %d", HEPv2, hep.Version)
	}
	if hep.Protocol != 17 {
		t.Errorf("Expected protocol 17, got %d", hep.Protocol)
	}
	if hep.SrcPort != 5060 {
		t.Errorf("Expected src port 5060, got %d", hep.SrcPort)
	}
	if hep.SrcIP != "192.168.1.1" {
		t.Errorf("Expected src IP 192.168.1.1, got %s", hep.SrcIP)
	}
}

func TestHEPv3Decode(t *testing.T) {
	// Create a test HEPv3 packet with chunks
	packet := createHEPv3TestPacket()

	hep, err := DecodeHEP(packet)
	if err != nil {
		t.Fatalf("Failed to decode HEPv3: %v", err)
	}

	if hep.Version != HEPv3 {
		t.Errorf("Expected version %d, got %d", HEPv3, hep.Version)
	}
	if hep.Protocol != 17 {
		t.Errorf("Expected protocol 17, got %d", hep.Protocol)
	}
	if hep.SrcPort != 5060 {
		t.Errorf("Expected src port 5060, got %d", hep.SrcPort)
	}
	if hep.SrcIP != "192.168.1.1" {
		t.Errorf("Expected src IP 192.168.1.1, got %s", hep.SrcIP)
	}
}

func createHEPv3TestPacket() []byte {
	chunks := []struct {
		chunkType uint16
		data      []byte
	}{
		{TypeIPProtocolFamily, []byte{0x02}}, // AF_INET
		{TypeIPProtocolID, []byte{17}},       // UDP
		{TypeIPSourceIP, net.ParseIP("192.168.1.1").To4()},
		{TypeIPDestinationIP, net.ParseIP("192.168.1.2").To4()},
		{TypeIPSourcePort, []byte{0x13, 0xC4}},               // 5060
		{TypeIPDestinationPort, []byte{0x13, 0xC4}},          // 5060
		{TypeTimestamp, make([]byte, 8)},                     // Current timestamp
		{TypeProtocolType, []byte{0x01}},                     // SIP
		{TypeCaptureAgentID, []byte{0x00, 0x00, 0x07, 0xD1}}, // 2001
		{TypePayload, []byte("TEST")},
	}

	// Calculate total size
	totalSize := 6 // HEP v3 header size
	for _, chunk := range chunks {
		totalSize += 6 + len(chunk.data) // chunk header + data
	}

	packet := make([]byte, totalSize)
	packet[0] = HEPv3
	binary.BigEndian.PutUint16(packet[4:6], uint16(totalSize))

	offset := 6
	for _, chunk := range chunks {
		// Vendor ID = 0x0000
		binary.BigEndian.PutUint16(packet[offset:offset+2], chunk.chunkType)
		binary.BigEndian.PutUint16(packet[offset+2:offset+4], uint16(6+len(chunk.data)))
		copy(packet[offset+6:], chunk.data)
		offset += 6 + len(chunk.data)
	}

	return packet
}

func TestInvalidPackets(t *testing.T) {
	tests := []struct {
		name    string
		packet  []byte
		wantErr error
	}{
		{
			name:    "Empty packet",
			packet:  []byte{},
			wantErr: ErrPacketTooShort,
		},
		{
			name:    "Invalid version",
			packet:  []byte{0x04, 0x00, 0x00, 0x00},
			wantErr: ErrInvalidVersion,
		},
		{
			name:    "Short HEPv1",
			packet:  []byte{0x01, 0x00, 0x00},
			wantErr: ErrPacketTooShort,
		},
		{
			name:    "Short HEPv2",
			packet:  []byte{0x02, 0x00, 0x00},
			wantErr: ErrPacketTooShort,
		},
		{
			name:    "Short HEPv3",
			packet:  []byte{0x03, 0x00, 0x00},
			wantErr: ErrPacketTooShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeHEP(tt.packet)
			if err != tt.wantErr {
				t.Errorf("DecodeHEP() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

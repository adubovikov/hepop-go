package protocol

import (
	"encoding/binary"
	"net"
	"testing"
	"time"
)

func BenchmarkHEPv1Decode(b *testing.B) {
	packet := makeHEPv1Packet()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := DecodeHEP(packet)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHEPv2Decode(b *testing.B) {
	packet := makeHEPv2Packet()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := DecodeHEP(packet)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHEPv3Decode(b *testing.B) {
	packet := makeHEPv3Packet()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := DecodeHEP(packet)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func makeHEPv1Packet() []byte {
	packet := make([]byte, 30)
	packet[0] = HEPv1
	packet[1] = 17 // UDP
	binary.BigEndian.PutUint16(packet[2:4], 5060)
	binary.BigEndian.PutUint16(packet[4:6], 5060)
	copy(packet[6:10], net.ParseIP("192.168.1.1").To4())
	copy(packet[10:14], net.ParseIP("192.168.1.2").To4())
	binary.BigEndian.PutUint32(packet[14:18], uint32(time.Now().Unix()))
	binary.BigEndian.PutUint32(packet[18:22], 2001)
	copy(packet[22:], []byte("BENCHMARK-PAYLOAD"))
	return packet
}

func makeHEPv2Packet() []byte {
	packet := make([]byte, 35)
	packet[0] = HEPv2
	packet[1] = 2  // AF_INET
	packet[2] = 17 // UDP
	binary.BigEndian.PutUint16(packet[3:5], 5060)
	binary.BigEndian.PutUint16(packet[5:7], 5060)
	copy(packet[7:11], net.ParseIP("192.168.1.1").To4())
	copy(packet[11:15], net.ParseIP("192.168.1.2").To4())
	binary.BigEndian.PutUint32(packet[15:19], uint32(time.Now().Unix()))
	binary.BigEndian.PutUint32(packet[19:23], 2001)
	packet[23] = 1 // SIP
	copy(packet[24:], []byte("BENCHMARK-PAYLOAD"))
	return packet
}

func makeHEPv3Packet() []byte {
	chunks := []struct {
		chunkType uint16
		data      []byte
	}{
		{TypeIPProtocolFamily, []byte{0x02}},
		{TypeIPProtocolID, []byte{17}},
		{TypeIPSourceIP, net.ParseIP("192.168.1.1").To4()},
		{TypeIPDestinationIP, net.ParseIP("192.168.1.2").To4()},
		{TypeIPSourcePort, []byte{0x13, 0xC4}},
		{TypeIPDestinationPort, []byte{0x13, 0xC4}},
		{TypeTimestamp, make([]byte, 8)},
		{TypeProtocolType, []byte{0x01}},
		{TypeCaptureAgentID, []byte{0x00, 0x00, 0x07, 0xD1}},
		{TypePayload, []byte("BENCHMARK-PAYLOAD")},
	}

	totalSize := 6
	for _, chunk := range chunks {
		totalSize += 6 + len(chunk.data)
	}

	packet := make([]byte, totalSize)
	packet[0] = HEPv3
	binary.BigEndian.PutUint16(packet[4:6], uint16(totalSize))

	offset := 6
	for _, chunk := range chunks {
		binary.BigEndian.PutUint16(packet[offset:offset+2], chunk.chunkType)
		binary.BigEndian.PutUint16(packet[offset+2:offset+4], uint16(6+len(chunk.data)))
		copy(packet[offset+6:], chunk.data)
		offset += 6 + len(chunk.data)
	}

	return packet
}

// Benchmark for comparing all versions with the same payload size
func BenchmarkHEPCompare(b *testing.B) {
	v1 := makeHEPv1Packet()
	v2 := makeHEPv2Packet()
	v3 := makeHEPv3Packet()

	b.Run("HEPv1", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			DecodeHEP(v1)
		}
	})

	b.Run("HEPv2", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			DecodeHEP(v2)
		}
	})

	b.Run("HEPv3", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			DecodeHEP(v3)
		}
	})
}

// Benchmark with different payload sizes
func BenchmarkHEPPayloadSizes(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		payload := make([]byte, size)
		b.Run("Size-"+string(size), func(b *testing.B) {
			v1 := makeHEPv1PacketWithPayload(payload)
			v2 := makeHEPv2PacketWithPayload(payload)
			v3 := makeHEPv3PacketWithPayload(payload)

			b.Run("HEPv1", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					DecodeHEP(v1)
				}
			})

			b.Run("HEPv2", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					DecodeHEP(v2)
				}
			})

			b.Run("HEPv3", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					DecodeHEP(v3)
				}
			})
		})
	}
}

func makeHEPv1PacketWithPayload(payload []byte) []byte {
	packet := make([]byte, 22+len(payload))
	// ... fill as in makeHEPv1Packet
	copy(packet[22:], payload)
	return packet
}

func makeHEPv2PacketWithPayload(payload []byte) []byte {
	packet := make([]byte, 24+len(payload))
	// ... fill as in makeHEPv2Packet
	copy(packet[24:], payload)
	return packet
}

func makeHEPv3PacketWithPayload(payload []byte) []byte {
	// ... similar to makeHEPv3Packet, but with the passed payload
	return nil // TODO: implement
}

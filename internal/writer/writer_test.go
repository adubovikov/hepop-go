package writer

import (
	"context"
	"testing"
	"time"

	"github.com/sipcapture/hepop-go/pkg/protocol"
)

// Performance tests
func BenchmarkElasticWriter(b *testing.B) {
	config := ElasticConfig{
		URLs:      []string{"http://localhost:9200"},
		IndexName: "hep_bench",
		BatchSize: 1000,
	}

	writer, err := NewElasticWriter(config)
	if err != nil {
		b.Skip("Elasticsearch is not available:", err)
	}
	defer writer.Close()

	packet := createTestPacket()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer.Write(packet)
	}
}

func createTestPacket() *protocol.HEPPacket {
	return &protocol.HEPPacket{
		Version:   3,
		Protocol:  17,
		SrcIP:     "192.168.1.1",
		DstIP:     "192.168.1.2",
		SrcPort:   5060,
		DstPort:   5060,
		Timestamp: uint64(time.Now().Unix()),
		NodeID:    2001,
		Payload:   []byte("TEST PAYLOAD"),
		CID:       "test-call-id",
	}
}

// MockWriter for testing
type MockWriter struct {
	*BatchWriter
	written []*protocol.HEPPacket
}

func NewMockWriter(batchSize int) *MockWriter {
	return &MockWriter{
		BatchWriter: NewBatchWriter(batchSize),
		written:     make([]*protocol.HEPPacket, 0),
	}
}

func (w *MockWriter) Search(ctx context.Context, params SearchParams) (SearchResult, error) {
	result := SearchResult{
		Total:   int64(len(w.written)),
		Results: w.written,
	}
	return result, nil
}

func (w *MockWriter) flush() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.written = append(w.written, w.buffer...)
	w.updateStats(true, uint64(len(w.buffer)), nil)
	w.buffer = w.buffer[:0]
}

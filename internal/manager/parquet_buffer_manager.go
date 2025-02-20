package manager

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/sipcapture/hepop-go/pkg/protocol"
	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/writer"
)

type ParquetBufferManager struct {
	buffers       map[string][]*protocol.HEPPacket
	bufferSize    int
	flushInterval time.Duration
	baseDir       string
	mu            sync.Mutex
}

func NewParquetBufferManager(baseDir string, bufferSize int, flushInterval time.Duration) *ParquetBufferManager {
	manager := &ParquetBufferManager{
		buffers:       make(map[string][]*protocol.HEPPacket),
		bufferSize:    bufferSize,
		flushInterval: flushInterval,
		baseDir:       baseDir,
	}
	go manager.startFlushInterval()
	return manager
}

func (m *ParquetBufferManager) AddPacket(packet *protocol.HEPPacket) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := packet.SrcIP // or any other key based on your logic
	m.buffers[key] = append(m.buffers[key], packet)

	if len(m.buffers[key]) >= m.bufferSize {
		m.flush(key)
	}
}

func (m *ParquetBufferManager) flush(key string) {
	if len(m.buffers[key]) == 0 {
		return
	}

	filePath := filepath.Join(m.baseDir, fmt.Sprintf("%s.parquet", key))
	fw, err := local.NewLocalFileWriter(filePath)
	if err != nil {
		fmt.Printf("Error creating file writer: %v\n", err)
		return
	}
	defer fw.Close()

	pw, err := writer.NewParquetWriter(fw, new(protocol.HEPPacket), 4)
	if err != nil {
		fmt.Printf("Error creating parquet writer: %v\n", err)
		return
	}
	defer pw.WriteStop()

	for _, packet := range m.buffers[key] {
		if err := pw.Write(packet); err != nil {
			fmt.Printf("Error writing packet: %v\n", err)
		}
	}

	m.buffers[key] = nil
}

func (m *ParquetBufferManager) startFlushInterval() {
	ticker := time.NewTicker(m.flushInterval)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		for key := range m.buffers {
			m.flush(key)
		}
		m.mu.Unlock()
	}
}

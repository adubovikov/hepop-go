package writer

import (
	"context"
	"sync"
	"time"

	"github.com/sipcapture/hepop-go/pkg/protocol"
)

// Writer интерфейс определяет методы для записи и поиска HEP пакетов
type Writer interface {
	Write(*protocol.HEPPacket) error
	Search(ctx context.Context, params SearchParams) (SearchResult, error)
	Close() error
	Stats() WriterStats
}

// SearchParams определяет параметры поиска
type SearchParams struct {
	Query     string
	FromTime  time.Time
	ToTime    time.Time
	Limit     int
	Offset    int
	OrderBy   string
	OrderDesc bool
}

// SearchResult contains search results
type SearchResult struct {
	Total   int64                 `json:"total"`
	Results []*protocol.HEPPacket `json:"results"`
}

// WriterStats contains writer statistics
type WriterStats struct {
	PacketsReceived uint64
	PacketsWritten  uint64
	BytesWritten    uint64
	Errors          uint64
	LastError       error
	LastWrite       time.Time
}

// BaseWriter provides a base implementation for all writers
type BaseWriter struct {
	stats WriterStats
	mu    sync.RWMutex
}

func (w *BaseWriter) updateStats(written bool, bytes uint64, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.stats.PacketsReceived++
	if written {
		w.stats.PacketsWritten++
		w.stats.BytesWritten += bytes
		w.stats.LastWrite = time.Now()
	}
	if err != nil {
		w.stats.Errors++
		w.stats.LastError = err
	}
}

func (w *BaseWriter) Stats() WriterStats {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.stats
}

// BatchWriter adds batch writing functionality
type BatchWriter struct {
	BaseWriter
	batchSize int
	buffer    []*protocol.HEPPacket
	flushChan chan struct{}
	done      chan struct{}
}

func NewBatchWriter(batchSize int) *BatchWriter {
	w := &BatchWriter{
		batchSize: batchSize,
		buffer:    make([]*protocol.HEPPacket, 0, batchSize),
		flushChan: make(chan struct{}),
		done:      make(chan struct{}),
	}
	go w.flushLoop()
	return w
}

func (w *BatchWriter) Write(packet *protocol.HEPPacket) error {
	w.mu.Lock()
	w.buffer = append(w.buffer, packet)
	shouldFlush := len(w.buffer) >= w.batchSize
	w.mu.Unlock()

	if shouldFlush {
		w.flushChan <- struct{}{}
	}
	return nil
}

func (w *BatchWriter) Close() error {
	close(w.done)
	w.flush() // Final flush
	return nil
}

func (w *BatchWriter) flush() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.buffer) == 0 {
		return
	}

	// Specific writers should implement this logic
	w.buffer = w.buffer[:0]
}

func (w *BatchWriter) flushLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.done:
			return
		case <-w.flushChan:
			w.flush()
		case <-ticker.C:
			w.flush()
		}
	}
}

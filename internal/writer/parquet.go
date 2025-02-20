package writer

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sipcapture/hepop-go/pkg/protocol"
	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/reader"
	"github.com/xitongsys/parquet-go/writer"
)

type ParquetWriter struct {
	filePath string
	pw       *writer.ParquetWriter
	stats    WriterStats
}

type ParquetConfig struct {
	FilePath string `yaml:"file_path"`
}

func NewParquetWriter(config ParquetConfig) (*ParquetWriter, error) {
	fw, err := local.NewLocalFileWriter(config.FilePath)
	if err != nil {
		return nil, fmt.Errorf("can't create local file writer: %w", err)
	}

	pw, err := writer.NewParquetWriter(fw, new(protocol.HEPPacket), 4)
	if err != nil {
		return nil, fmt.Errorf("can't create parquet writer: %w", err)
	}

	return &ParquetWriter{
		filePath: config.FilePath,
		pw:       pw,
		stats:    WriterStats{},
	}, nil
}

func (w *ParquetWriter) Write(packet *protocol.HEPPacket) error {
	if err := w.pw.Write(packet); err != nil {
		return fmt.Errorf("can't write packet to parquet: %w", err)
	}
	return nil
}

func (w *ParquetWriter) Close() error {
	if err := w.pw.WriteStop(); err != nil {
		return fmt.Errorf("can't stop parquet writer: %w", err)
	}
	return nil
}

func (w *ParquetWriter) Search(ctx context.Context, params SearchParams) (SearchResult, error) {
	fr, err := local.NewLocalFileReader(w.filePath)
	if err != nil {
		return SearchResult{}, fmt.Errorf("can't open parquet file: %w", err)
	}
	defer fr.Close()

	pr, err := reader.NewParquetReader(fr, new(protocol.HEPPacket), 4)
	if err != nil {
		return SearchResult{}, fmt.Errorf("can't create parquet reader: %w", err)
	}
	defer pr.ReadStop()

	var results []*protocol.HEPPacket
	num := int(pr.GetNumRows())
	for i := 0; i < num; i++ {
		packet := make([]protocol.HEPPacket, 1)
		if err := pr.Read(&packet); err != nil {
			return SearchResult{}, fmt.Errorf("can't read from parquet file: %w", err)
		}

		// Apply search filters
		if matchesSearchCriteria(packet[0], params) {
			results = append(results, &packet[0])
		}
	}

	return SearchResult{
		Total:   int64(len(results)),
		Results: results,
	}, nil
}

func (w *ParquetWriter) Stats() WriterStats {
	fileInfo, err := os.Stat(w.filePath)
	if err != nil {
		w.stats.Errors++
		w.stats.LastError = err
		return w.stats
	}

	fr, err := local.NewLocalFileReader(w.filePath)
	if err != nil {
		w.stats.Errors++
		w.stats.LastError = err
		return w.stats
	}
	defer fr.Close()

	pr, err := reader.NewParquetReader(fr, new(protocol.HEPPacket), 4)
	if err != nil {
		w.stats.Errors++
		w.stats.LastError = err
		return w.stats
	}
	defer pr.ReadStop()

	w.stats.FileSize = fileInfo.Size()
	w.stats.NumRecords = pr.GetNumRows()
	w.stats.FilePath = w.filePath
	w.stats.LastModified = fileInfo.ModTime().Format(time.RFC3339)

	return w.stats
}

func matchesSearchCriteria(packet protocol.HEPPacket, params SearchParams) bool {
	// Implement your search criteria here
	// Example: filter by source IP
	if params.Query != "" && packet.SrcIP != params.Query {
		return false
	}
	// Add more conditions as needed
	return true
}

package writer

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/sipcapture/hepop-go/pkg/protocol"
)

type ClickHouseWriter struct {
	*BatchWriter
	conn      clickhouse.Conn
	tableName string
}

type ClickHouseConfig struct {
	Host      string
	Port      int
	Database  string
	Table     string
	Username  string
	Password  string
	BatchSize int
}

func NewClickHouseWriter(config ClickHouseConfig) (*ClickHouseWriter, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", config.Host, config.Port)},
		Auth: clickhouse.Auth{
			Database: config.Database,
			Username: config.Username,
			Password: config.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: 5 * time.Second,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
	})
	if err != nil {
		return nil, err
	}

	w := &ClickHouseWriter{
		BatchWriter: NewBatchWriter(config.BatchSize),
		conn:        conn,
		tableName:   config.Table,
	}

	return w, nil
}

func (w *ClickHouseWriter) flush() {
	w.mu.Lock()
	packets := make([]*protocol.HEPPacket, len(w.buffer))
	copy(packets, w.buffer)
	w.buffer = w.buffer[:0]
	w.mu.Unlock()

	if len(packets) == 0 {
		return
	}

	batch, err := w.conn.PrepareBatch(context.Background(), fmt.Sprintf(`
		INSERT INTO %s (
			version, protocol_family, protocol, 
			src_ip, dst_ip, src_port, dst_port,
			timestamp, payload, cid, vlan
		)`, w.tableName))
	if err != nil {
		w.updateStats(false, 0, err)
		return
	}

	var totalBytes uint64
	for _, packet := range packets {
		err := batch.Append(
			packet.Version,
			packet.Protocol,
			packet.ProtoType,
			packet.SrcIP,
			packet.DstIP,
			packet.SrcPort,
			packet.DstPort,
			time.Unix(int64(packet.Timestamp), 0),
			packet.Payload,
			packet.CID,
			packet.Vlan,
		)
		if err != nil {
			w.updateStats(false, 0, err)
			continue
		}
		totalBytes += uint64(len(packet.Payload))
	}

	if err := batch.Send(); err != nil {
		w.updateStats(false, 0, err)
		return
	}

	w.updateStats(true, totalBytes, nil)
}

func (w *ClickHouseWriter) Search(ctx context.Context, params SearchParams) (SearchResult, error) {
	query := `
		SELECT version, protocol, src_ip, dst_ip, src_port, dst_port, timestamp, node_id, payload, cid
		FROM hep_packets
		WHERE timestamp BETWEEN ? AND ?
		%s
		LIMIT ?
	`
	condition := ""
	if params.Query != "" {
		condition = fmt.Sprintf("AND (%s)", params.Query)
	}
	query = fmt.Sprintf(query, condition)

	rows, err := w.conn.Query(ctx, query, params.FromTime, params.ToTime, params.Limit)
	if err != nil {
		return SearchResult{}, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var results []*protocol.HEPPacket
	for rows.Next() {
		var packet protocol.HEPPacket
		if err := rows.Scan(
			&packet.Version,
			&packet.Protocol,
			&packet.SrcIP,
			&packet.DstIP,
			&packet.SrcPort,
			&packet.DstPort,
			&packet.Timestamp,
			&packet.NodeID,
			&packet.Payload,
			&packet.CID,
		); err != nil {
			return SearchResult{}, fmt.Errorf("scan failed: %w", err)
		}
		results = append(results, &packet)
	}

	return SearchResult{
		Total:   int64(len(results)),
		Results: results,
	}, nil
}

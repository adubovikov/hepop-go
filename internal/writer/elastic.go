package writer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/sipcapture/hepop-go/pkg/protocol"
)

type ElasticWriter struct {
	*BatchWriter
	client    *elasticsearch.Client
	indexName string
}

type ElasticConfig struct {
	URLs      []string
	Username  string
	Password  string
	IndexName string
	BatchSize int
}

func NewElasticWriter(config ElasticConfig) (*ElasticWriter, error) {
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: config.URLs,
		Username:  config.Username,
		Password:  config.Password,
	})
	if err != nil {
		return nil, err
	}

	return &ElasticWriter{
		BatchWriter: NewBatchWriter(config.BatchSize),
		client:      client,
		indexName:   config.IndexName,
	}, nil
}

func (w *ElasticWriter) flush() {
	w.mu.Lock()
	packets := make([]*protocol.HEPPacket, len(w.buffer))
	copy(packets, w.buffer)
	w.buffer = w.buffer[:0]
	w.mu.Unlock()

	if len(packets) == 0 {
		return
	}

	var buf bytes.Buffer
	for _, packet := range packets {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_index" : "%s" } }%s`,
			w.indexName, "\n"))
		data, err := json.Marshal(packet)
		if err != nil {
			w.updateStats(false, 0, err)
			continue
		}
		buf.Grow(len(meta) + len(data) + 1)
		buf.Write(meta)
		buf.Write(data)
		buf.WriteByte('\n')
	}

	res, err := w.client.Bulk(bytes.NewReader(buf.Bytes()))
	if err != nil {
		w.updateStats(false, 0, err)
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		w.updateStats(false, 0, fmt.Errorf("bulk write failed: %s", res.String()))
		return
	}

	w.updateStats(true, uint64(buf.Len()), nil)
}

func (w *ElasticWriter) Search(ctx context.Context, params SearchParams) (SearchResult, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"range": map[string]interface{}{
							"timestamp": map[string]interface{}{
								"gte": params.FromTime.Format(time.RFC3339),
								"lte": params.ToTime.Format(time.RFC3339),
							},
						},
					},
				},
			},
		},
		"size": params.Limit,
	}

	if params.Query != "" {
		query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = append(
			query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]interface{}),
			map[string]interface{}{
				"query_string": map[string]interface{}{
					"query": params.Query,
				},
			},
		)
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return SearchResult{}, fmt.Errorf("encoding query failed: %w", err)
	}

	res, err := w.client.Search(
		w.client.Search.WithContext(ctx),
		w.client.Search.WithIndex(w.indexName),
		w.client.Search.WithBody(&buf),
	)
	if err != nil {
		return SearchResult{}, fmt.Errorf("search query failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return SearchResult{}, fmt.Errorf("search query error: %s", res.String())
	}

	var esResponse struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source protocol.HEPPacket `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&esResponse); err != nil {
		return SearchResult{}, fmt.Errorf("decoding response failed: %w", err)
	}

	results := make([]*protocol.HEPPacket, len(esResponse.Hits.Hits))
	for i, hit := range esResponse.Hits.Hits {
		results[i] = &hit.Source
	}

	return SearchResult{
		Total:   esResponse.Hits.Total.Value,
		Results: results,
	}, nil
}

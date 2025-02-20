# API Documentation

## Packet Search (Search API)

### Endpoint

```
GET /api/v1/search

POST /api/v1/search

```

### Parameters

- `query` - query string
- `from_time` - start time
- `to_time` - end time
- `limit` - number of results
- `offset` - offset
- `order_by` - field for sorting
- `order_desc` - sort order

### Search Parameters (SearchParams)

| Parameter  | Type     | Description | Example |
|------------|----------|-------------|---------|
| query      | string   | Search query in Lucene format | `src_ip:192.168.1.1 AND protocol:17` |
| from_time  | datetime | Start of the time range (RFC3339) | `2024-01-01T00:00:00Z` |
| to_time    | datetime | End of the time range (RFC3339) | `2024-01-02T00:00:00Z` |
| limit      | integer  | Maximum number of results | `100` |
| offset     | integer  | Offset for pagination | `0` |
| order_by   | string   | Field for sorting | `timestamp` |
| order_desc | boolean  | Sort in descending order | `true` |

### Supported Search Fields

| Field     | Type   | Description | Example Query |
|-----------|--------|-------------|---------------|
| version   | int    | HEP protocol version | `version:3` |
| protocol  | int    | Protocol (UDP=17, TCP=6) | `protocol:17` |
| src_ip    | string | Source IP address | `src_ip:192.168.1.1` |
| dst_ip    | string | Destination IP address | `dst_ip:10.0.0.*` |
| src_port  | int    | Source port | `src_port:5060` |
| dst_port  | int    | Destination port | `dst_port:5060` |
| timestamp | date   | Timestamp | `timestamp:[2024-01-01 TO 2024-01-02]` |
| node_id   | int    | Node ID | `node_id:2001` |
| cid       | string | Correlation ID | `cid:*test*` |

### Request Examples

#### GET Request

```
GET /api/v1/search?query=src_ip:192.168.1.1
```

#### POST Request

```
POST /api/v1/search
```

#### Time-based Search

### Search Features for Different Storages

#### ClickHouse
- Supports full-text search on payload
- Efficient time range filtering
- Fast data aggregation

#### Elasticsearch
- Full support for Lucene syntax
- Search within nested fields in JSON payload
- Supports wildcard and regex queries

#### Loki
- Search by labels
- LogQL syntax for filtering
- Limited support for complex queries

### Response

```
json:hepop-go/docs/api.md
{
  "total": 1234,
  "results": [
    {
      "version": 3,
      "protocol": 17,
      "src_ip": "192.168.1.1",
      "dst_ip": "192.168.1.2",
      "src_port": 5060,
      "dst_port": 5060,
      "timestamp": 1704067200,
      "node_id": 2001,
      "payload": "base64encoded...",
      "cid": "test-call-id"
    }
  ]
}
```


# HEPop-Go Configuration

## General Structure

The configuration for HEPop-Go is stored in a YAML file and consists of the following main sections:
- Server - HEP server settings
- Writers - storage system settings
- API - HTTP API settings
- Metrics - Prometheus metrics settings

## Configuration Parameters

### Server

- `host` - IP address for listening
- `port` - port for listening
- `protocol` - protocol (udp, tcp, both)
- `max_packet_size` - maximum packet size
- `read_timeout` - read timeout
- `write_timeout` - write timeout
- `workers` - number of worker threads

### Writers

- `type` - type of storage system (clickhouse, elastic, loki, multi)
- `batch_size` - batch size for writing
- `flush_interval` - buffer flush interval

#### ClickHouse

- `host` - ClickHouse IP address
- `port` - ClickHouse port
- `database` - database name
- `table` - table name
- `username` - username
- `password` - user password
- `debug` - enable debug mode

#### Elasticsearch

- `urls` - list of Elasticsearch URLs
- `index_name` - index name
- `username` - username
- `password` - user password
- `debug` - enable debug mode


### API

- `host` - IP address for listening
- `port` - port for listening
- `enable_metrics` - enable Prometheus metrics
- `enable_pprof` - enable pprof profiling
- `auth_token` - authentication token
- `cors_origins` - list of allowed CORS origins
- `read_timeout` - read timeout
- `write_timeout` - write timeout

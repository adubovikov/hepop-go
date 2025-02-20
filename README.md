# HEPop-Go

HEPop-Go is a high-performance, scalable HEP (Homer Encapsulation Protocol) server written in Go. It is designed to efficiently process and store HEP packets in various storage backends, including ClickHouse, Elasticsearch, Parquet, and DuckDB.

## Features

- Supports multiple storage backends: ClickHouse, Elasticsearch, Loki, Parquet, and DuckDB.
- Provides a RESTful API for searching and retrieving HEP packets.
- Configurable via a YAML configuration file.
- Supports Prometheus metrics for monitoring.
- High-performance and scalable architecture.

## Installation

### Prerequisites

- Go 1.18 or later
- Access to your chosen storage backend (e.g., ClickHouse, Elasticsearch)

### Building from Source

1. Clone the repository:

   ```bash
   git clone https://github.com/yourusername/hepop-go.git
   cd hepop-go
   ```

2. Build the project:

   ```bash
   go build -o hepop-go cmd/hepop/main.go
   ```

3. Run the server:

   ```bash
   ./hepop-go
   ```

## Configuration

HEPop-Go is configured via a YAML file. Below is an example configuration:

```yaml
server:
  host: "0.0.0.0"
  port: 9060
  protocol: "udp"
  max_packet_size: 4096
  read_timeout: 10s
  write_timeout: 10s
  workers: 4

writers:
  type: "parquet"
  parquet:
    file_path: "/path/to/output.parquet"

api:
  host: "0.0.0.0"
  port: 8080
  enable_metrics: true
  enable_pprof: false
  auth_token: "your_auth_token"
  cors_origins:
    - "*"
  read_timeout: 10s
  write_timeout: 10s
```

## Parquet Support

HEPop-Go supports writing HEP packets to Parquet files, which are efficient for storage and analytics. The `ParquetBufferManager` handles buffering and writing of packets to Parquet files, ensuring efficient data management.

### Configuration

To use Parquet as a storage backend, configure the `writers` section in your YAML file as shown above.

## DuckDB Integration

HEPop-Go integrates with DuckDB to allow SQL-like querying of Parquet files. This enables powerful data analysis capabilities directly on the stored data.

### Usage

The `DuckDBManager` provides an interface to execute SQL queries on Parquet files. This can be used to perform complex queries and data analysis.

## Usage

### Starting the Server

Run the server with the configuration file:

```bash
./hepop-go -config config/config.yaml
```

### API Endpoints

- **Search API**: `/api/v1/search`
  - Supports GET and POST methods for searching HEP packets.

## Contributing

Contributions are welcome! Please fork the repository and submit a pull request for any improvements or bug fixes.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contact

For questions or support, please contact [@sipcapture.org](mailto:info@sipcapture.org).
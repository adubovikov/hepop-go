package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Writers WritersConfig `yaml:"writers"`
	API     APIConfig     `yaml:"api"`
	Metrics MetricsConfig `yaml:"metrics"`
}

type ServerConfig struct {
	Host          string        `yaml:"host"`
	Port          int           `yaml:"port"`
	Protocol      string        `yaml:"protocol"` // udp, tcp, both
	MaxPacketSize int           `yaml:"max_packet_size"`
	ReadTimeout   time.Duration `yaml:"read_timeout"`
	WriteTimeout  time.Duration `yaml:"write_timeout"`
	Workers       int           `yaml:"workers"`
}

type WritersConfig struct {
	Type          string        `yaml:"type"` // clickhouse, elastic, loki, multi
	BatchSize     int           `yaml:"batch_size"`
	FlushInterval time.Duration `yaml:"flush_interval"`

	// Specific writer configs
	ClickHouse *ClickHouseConfig `yaml:"clickhouse,omitempty"`
	Elastic    *ElasticConfig    `yaml:"elastic,omitempty"`
	Parquet    *ParquetConfig    `yaml:"parquet,omitempty"`
}

type ParquetConfig struct {
	FilePath string `yaml:"file_path"`
}

type ClickHouseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Table    string `yaml:"table"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Debug    bool   `yaml:"debug"`
}

type ElasticConfig struct {
	URLs      []string `yaml:"urls"`
	IndexName string   `yaml:"index_name"`
	Username  string   `yaml:"username"`
	Password  string   `yaml:"password"`
	Debug     bool     `yaml:"debug"`
}

type APIConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	EnablePprof  bool          `yaml:"enable_pprof"`
	AuthToken    string        `yaml:"auth_token"`
	CorsOrigins  []string      `yaml:"cors_origins"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type MetricsConfig struct {
	Enable bool   `yaml:"enable"`
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
	Path   string `yaml:"path"`
}

// LoadConfig loads the configuration from the file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return config, nil
}

// Validate checks the configuration
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Server.Workers <= 0 {
		c.Server.Workers = 1
	}

	if c.Writers.BatchSize <= 0 {
		c.Writers.BatchSize = 1000
	}

	if c.Writers.FlushInterval <= 0 {
		c.Writers.FlushInterval = time.Second
	}

	switch c.Writers.Type {
	case "clickhouse":
		if c.Writers.ClickHouse == nil {
			return fmt.Errorf("clickhouse config required")
		}
	case "elastic":
		if c.Writers.Elastic == nil {
			return fmt.Errorf("elastic config required")
		}

	case "multi":
		// At least one writer should be configured
		if c.Writers.ClickHouse == nil &&
			c.Writers.Elastic == nil {
			return fmt.Errorf("at least one writer required")
		}
	default:
		return fmt.Errorf("unknown writer type: %s", c.Writers.Type)
	}

	return nil
}

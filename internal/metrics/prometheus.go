package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusExporter struct {
	hepPacketsTotal   *prometheus.CounterVec
	hepPacketSize     *prometheus.HistogramVec
	writeLatency      *prometheus.HistogramVec
	writeErrors       *prometheus.CounterVec
	activeConnections prometheus.Gauge
}

func NewPrometheusExporter() *PrometheusExporter {
	e := &PrometheusExporter{
		hepPacketsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "hep_packets_total",
				Help: "Total number of HEP packets received",
			},
			[]string{"version", "type"},
		),
		hepPacketSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "hep_packet_size_bytes",
				Help:    "Size of HEP packets",
				Buckets: prometheus.ExponentialBuckets(64, 2, 10),
			},
			[]string{"version"},
		),
		writeLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "write_latency_seconds",
				Help:    "Latency of write operations",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"storage"},
		),
		writeErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "write_errors_total",
				Help: "Total number of write errors",
			},
			[]string{"storage", "error"},
		),
		activeConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "active_connections",
				Help: "Number of active connections",
			},
		),
	}

	prometheus.MustRegister(
		e.hepPacketsTotal,
		e.hepPacketSize,
		e.writeLatency,
		e.writeErrors,
		e.activeConnections,
	)

	return e
}

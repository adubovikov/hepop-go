package api

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	responseSizes   *prometheus.HistogramVec
	activeRequests  prometheus.Gauge
}

func NewMetrics() *Metrics {
	m := &Metrics{
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "http_request_duration_seconds",
				Help: "HTTP request duration in seconds",
			},
			[]string{"method", "path"},
		),
		responseSizes: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "http_response_size_bytes",
				Help: "HTTP response sizes in bytes",
			},
			[]string{"method", "path"},
		),
		activeRequests: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_active_requests",
				Help: "Number of active HTTP requests",
			},
		),
	}

	prometheus.MustRegister(
		m.requestsTotal,
		m.requestDuration,
		m.responseSizes,
		m.activeRequests,
	)

	return m
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.Handler()
}

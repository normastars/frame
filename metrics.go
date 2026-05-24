package frame

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// Prometheus metrics
var (
	prometheusRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "request_duration_seconds",
		Help:    "HTTP request duration in seconds.",
		Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 20},
	}, []string{"url", "code", "method"})

	prometheusRequestBusCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "request_buss_count",
		Help: "HTTP request business code count.",
	}, []string{"url", "bus_code", "method"})
)

var (
	sendHTTPRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "send_http_requests_total",
			Help: "Number of the http requests sent since the server started",
		},
		[]string{"method", "host", "path", "code"},
	)
	sendHTTPRequestsDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "send_http_requests_duration_seconds",
			Help:    "Duration in seconds to send http requests",
			Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 20},
		},
		[]string{"method", "host", "path", "code"},
	)
)

var registerMetricsOnce sync.Once

// registerMetrics explicitly registers Prometheus metrics, replacing init() auto-registration.
// Safe to call multiple times; registration only happens once.
func registerMetrics() {
	registerMetricsOnce.Do(func() {
		prometheus.MustRegister(prometheusRequestDuration)
		prometheus.MustRegister(prometheusRequestBusCounter)
		prometheus.MustRegister(sendHTTPRequests, sendHTTPRequestsDuration)
	})
}

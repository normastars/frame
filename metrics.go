package frame

import (
	"github.com/prometheus/client_golang/prometheus"
)

// var r = prometheus.NewRegistry()

func init() {
	prometheus.MustRegister(prometheusRequestDuration)
	prometheus.MustRegister(prometheusRequestBusCounter)
	prometheus.MustRegister(sendHTTPRequests, sendHTTPRequestsDuration)
}

// Prometheus metrics
var (
	prometheusRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "request_duration_seconds",
		Help:    "HTTP request duration in seconds.",
		Buckets: []float64{50, 100, 250, 500, 1000, 2500, 5000, 10000, 20000}, // ms
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
			Buckets: []float64{50, 100, 250, 500, 1000, 2500, 5000, 10000, 20000}, // ms
		},
		[]string{"method", "host", "path", "code"},
	)
)

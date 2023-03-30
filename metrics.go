package frame

import (
	"github.com/prometheus/client_golang/prometheus"
)

var r = prometheus.NewRegistry()

func init() {
	prometheus.MustRegister(prometheusRequestDuration)
	prometheus.MustRegister(prometheusRequestBusCounter)
}

// Prometheus metrics
var (
	prometheusRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "request_duration_seconds",
		Help:    "HTTP request duration in seconds.",
		Buckets: []float64{50, 100, 250, 500, 1000, 2500, 5000, 10000}, // ms
	}, []string{"url", "code", "method"})

	prometheusRequestBusCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "request_buss_count",
		Help: "HTTP request business code count.",
	}, []string{"url", "bus_code", "method"})
)

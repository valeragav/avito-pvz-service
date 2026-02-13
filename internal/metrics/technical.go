package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	httpResponsesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_responses_total",
			Help: "Total HTTP responses by status code.",
		},
		[]string{"method", "path", "status"},
	)
)

func RestRequestInc(method, path string) {
	httpRequestsTotal.WithLabelValues(method, path).Inc()
}

func RestRequestDurationObserve(method, path string, d time.Duration) {
	httpRequestDuration.WithLabelValues(method, path).Observe(d.Seconds())
}

func RestResponseInc(method, path string, status int) {
	httpResponsesTotal.WithLabelValues(method, path, strconv.Itoa(status)).Inc()
}

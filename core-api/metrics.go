package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	requestCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "core_api_requests_total",
			Help: "Total HTTP requests by path, method, and status code.",
		},
		[]string{"path", "method", "status"},
	)

	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "core_api_request_duration_seconds",
			Help:    "Request duration in seconds by path.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path"},
	)
)

// statusRecorder wraps http.ResponseWriter so we can read back the status
// code a handler wrote — net/http doesn't expose that directly otherwise.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// withMetrics wraps any handler so every request gets timed and counted,
// without repeating that logic inside each handler.
func withMetrics(path string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next(rec, r)

		requestCount.WithLabelValues(path, r.Method, strconv.Itoa(rec.status)).Inc()
		requestDuration.WithLabelValues(path).Observe(time.Since(start).Seconds())
	}
}
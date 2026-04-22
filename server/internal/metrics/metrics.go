package metrics

import (
	"runtime"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // Count HTTP-requests
    HttpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    // Duration HTTP-requests
    HttpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "Duration of HTTP requests",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )

    // Number of goroutines
    goroutines = promauto.NewGaugeFunc(
        prometheus.GaugeOpts{
            Name: "goroutines_count",
            Help: "Number of goroutines",
        },
        func() float64 { return float64(runtime.NumGoroutine()) },
    )

    // Number of documents currently being processed
    documentsProcessing = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "documents_processing",
            Help: "Number of documents currently being processed",
        },
    )
)

func Init() {
}


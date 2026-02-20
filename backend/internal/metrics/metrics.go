package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Counter for total classifications
	ClassificationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "log_classifications_total",
			Help: "Total number of log classifications",
		},
		[]string{"classifier", "label"},
	)

	// Histogram for classification duration
	ClassificationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "log_classification_duration_seconds",
			Help:    "Duration of log classifications",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"classifier"},
	)

	// Counter for classification errors
	ClassificationErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "log_classification_errors_total",
			Help: "Total number of classification errors",
		},
		[]string{"classifier", "error_type"},
	)

	// Gauge for active workers
	ActiveWorkers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "log_classifier_active_workers",
			Help: "Number of active worker goroutines",
		},
	)

	// Gauge for circuit breaker state
	CircuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "log_classifier_circuit_breaker_open",
			Help: "Circuit breaker state (1 = open, 0 = closed)",
		},
		[]string{"classifier"},
	)

	// Counter for HTTP requests
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "log_classifier_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"endpoint", "method", "status"},
	)

	// Histogram for HTTP request duration
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "log_classifier_http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint", "method"},
	)
)

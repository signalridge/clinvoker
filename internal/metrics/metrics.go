// Package metrics provides Prometheus metrics for observability.
package metrics

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metric names use the prefix "clinvk_" for namespacing.
const (
	namespace = "clinvk"
)

// HTTP request metrics
var (
	// RequestsTotal counts total HTTP requests by method, path, and status.
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// RequestDuration tracks HTTP request latency in seconds.
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   prometheus.ExponentialBuckets(0.001, 2, 15), // 1ms to ~16s
		},
		[]string{"method", "path"},
	)
)

// Backend execution metrics
var (
	// BackendExecutions counts backend executions by backend name and status.
	BackendExecutions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "backend_executions_total",
			Help:      "Total number of backend executions",
		},
		[]string{"backend", "status"},
	)

	// BackendExecutionDuration tracks backend execution latency in seconds.
	BackendExecutionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "backend_execution_duration_seconds",
			Help:      "Backend execution duration in seconds",
			Buckets:   prometheus.ExponentialBuckets(0.1, 2, 12), // 100ms to ~200s
		},
		[]string{"backend"},
	)
)

// Execution-mode metrics (chain/compare/parallel)
var (
	ChainStepExecutions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "chain_step_executions_total",
			Help:      "Total number of chain step executions",
		},
		[]string{"backend", "status"},
	)

	ChainStepDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "chain_step_duration_seconds",
			Help:      "Chain step execution duration in seconds",
			Buckets:   prometheus.ExponentialBuckets(0.01, 2, 15), // 10ms to ~163s
		},
		[]string{"backend", "status"},
	)

	CompareBackendExecutions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "compare_backend_executions_total",
			Help:      "Total number of compare backend executions",
		},
		[]string{"backend", "status"},
	)

	CompareBackendDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "compare_backend_duration_seconds",
			Help:      "Compare backend execution duration in seconds",
			Buckets:   prometheus.ExponentialBuckets(0.01, 2, 15), // 10ms to ~163s
		},
		[]string{"backend", "status"},
	)

	ParallelTaskExecutions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "parallel_task_executions_total",
			Help:      "Total number of parallel task executions",
		},
		[]string{"backend", "status"},
	)

	ParallelTaskDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "parallel_task_duration_seconds",
			Help:      "Parallel task execution duration in seconds",
			Buckets:   prometheus.ExponentialBuckets(0.01, 2, 15), // 10ms to ~163s
		},
		[]string{"backend", "status"},
	)
)

// Session metrics
var (
	// ActiveSessions tracks the number of active sessions.
	ActiveSessions = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "active_sessions",
			Help:      "Number of active sessions",
		},
	)

	// SessionsCreated counts total sessions created.
	SessionsCreated = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "sessions_created_total",
			Help:      "Total number of sessions created",
		},
	)
)

// RecordRequest records an HTTP request metric.
func RecordRequest(method, path, status string) {
	RequestsTotal.WithLabelValues(method, path, status).Inc()
}

// RecordRequestDuration records an HTTP request duration.
func RecordRequestDuration(method, path string, durationSeconds float64) {
	RequestDuration.WithLabelValues(method, path).Observe(durationSeconds)
}

// RecordBackendExecution records a backend execution.
func RecordBackendExecution(backend, status string) {
	BackendExecutions.WithLabelValues(backend, status).Inc()
}

// RecordBackendExecutionDuration records a backend execution duration.
func RecordBackendExecutionDuration(backend string, durationSeconds float64) {
	BackendExecutionDuration.WithLabelValues(backend).Observe(durationSeconds)
}

func normalizeExecutionBackend(backend string) string {
	backend = strings.TrimSpace(strings.ToLower(backend))
	if backend == "" {
		return "unknown"
	}
	return backend
}

func normalizeExecutionStatus(status string) string {
	status = strings.TrimSpace(strings.ToLower(status))
	switch status {
	case "ok", "failed", "timeout", "canceled":
		return status
	default:
		return "failed"
	}
}

// RecordChainStepExecution records chain step metrics with bounded labels.
func RecordChainStepExecution(backend, status string, durationSeconds float64) {
	backend = normalizeExecutionBackend(backend)
	status = normalizeExecutionStatus(status)
	ChainStepExecutions.WithLabelValues(backend, status).Inc()
	ChainStepDuration.WithLabelValues(backend, status).Observe(durationSeconds)
}

// RecordCompareBackendExecution records compare backend metrics with bounded labels.
func RecordCompareBackendExecution(backend, status string, durationSeconds float64) {
	backend = normalizeExecutionBackend(backend)
	status = normalizeExecutionStatus(status)
	CompareBackendExecutions.WithLabelValues(backend, status).Inc()
	CompareBackendDuration.WithLabelValues(backend, status).Observe(durationSeconds)
}

// RecordParallelTaskExecution records parallel task metrics with bounded labels.
func RecordParallelTaskExecution(backend, status string, durationSeconds float64) {
	backend = normalizeExecutionBackend(backend)
	status = normalizeExecutionStatus(status)
	ParallelTaskExecutions.WithLabelValues(backend, status).Inc()
	ParallelTaskDuration.WithLabelValues(backend, status).Observe(durationSeconds)
}

// SetActiveSessions sets the number of active sessions.
func SetActiveSessions(count float64) {
	ActiveSessions.Set(count)
}

// IncrementSessionsCreated increments the sessions created counter.
func IncrementSessionsCreated() {
	SessionsCreated.Inc()
}

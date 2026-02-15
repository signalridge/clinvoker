package app

import (
	"strings"

	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/metrics"
)

const (
	execStatusCanceled = "canceled"
	execStatusTimeout  = "timeout"
	execStatusOK       = "ok"
	execStatusFailed   = "failed"
)

func executionStatus(exitCode int, errMsg string) string {
	msg := strings.ToLower(errMsg)
	switch {
	case strings.Contains(msg, "canceled"):
		return execStatusCanceled
	case strings.Contains(msg, "timeout") || strings.Contains(msg, "timed out") || exitCode == 124:
		return execStatusTimeout
	case exitCode == 0 && errMsg == "":
		return execStatusOK
	default:
		return execStatusFailed
	}
}

func metricsEnabled(cfg *config.Config) bool {
	if cfg == nil {
		cfg = config.Get()
	}
	return cfg != nil && cfg.Server.MetricsEnabled
}

func recordChainStepMetrics(cfg *config.Config, backendName string, exitCode int, errMsg string, durationSeconds float64) {
	if !metricsEnabled(cfg) {
		return
	}
	metrics.RecordChainStepExecution(backendName, executionStatus(exitCode, errMsg), durationSeconds)
}

func recordCompareBackendMetrics(cfg *config.Config, backendName string, exitCode int, errMsg string, durationSeconds float64) {
	if !metricsEnabled(cfg) {
		return
	}
	metrics.RecordCompareBackendExecution(backendName, executionStatus(exitCode, errMsg), durationSeconds)
}

func recordParallelTaskMetrics(cfg *config.Config, backendName string, exitCode int, errMsg string, durationSeconds float64) {
	if !metricsEnabled(cfg) {
		return
	}
	metrics.RecordParallelTaskExecution(backendName, executionStatus(exitCode, errMsg), durationSeconds)
}

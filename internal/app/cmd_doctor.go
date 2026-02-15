package app

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/signalridge/clinvoker/internal/auth"
	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
)

const doctorSchemaVersion = "v1"

type doctorStatus string

const (
	doctorStatusPass doctorStatus = "pass"
	doctorStatusWarn doctorStatus = "warn"
	doctorStatusFail doctorStatus = "fail"
)

type doctorCheck struct {
	ID         string       `json:"id"`
	Status     doctorStatus `json:"status"`
	Message    string       `json:"message"`
	Suggestion string       `json:"suggestion,omitempty"`
}

type doctorSummary struct {
	Pass int `json:"pass"`
	Warn int `json:"warn"`
	Fail int `json:"fail"`
}

type doctorReport struct {
	SchemaVersion string        `json:"schema_version"`
	Summary       doctorSummary `json:"summary"`
	Checks        []doctorCheck `json:"checks"`
}

var doctorJSON bool

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run environment and configuration diagnostics",
	RunE:  runDoctor,
}

func init() {
	doctorCmd.Flags().BoolVar(&doctorJSON, "json", false, "output diagnostics as JSON")
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	report := runDoctorChecks(cfg)

	if doctorJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			return err
		}
	} else {
		fmt.Println("Doctor checks:")
		for _, check := range report.Checks {
			fmt.Printf("- [%s] %s: %s\n", check.Status, check.ID, check.Message)
			if check.Suggestion != "" {
				fmt.Printf("  suggestion: %s\n", check.Suggestion)
			}
		}
		fmt.Printf("\nSummary: pass=%d warn=%d fail=%d\n", report.Summary.Pass, report.Summary.Warn, report.Summary.Fail)
	}

	if report.Summary.Fail > 0 {
		return fmt.Errorf("%w: doctor found %d failing check(s)", ErrValidationFailed, report.Summary.Fail)
	}
	return nil
}

func runDoctorChecks(cfg *config.Config) *doctorReport {
	report := &doctorReport{
		SchemaVersion: doctorSchemaVersion,
		Checks:        make([]doctorCheck, 0, 8),
	}

	// Config load is implicit if cfg is non-nil.
	if cfg != nil {
		report.Checks = append(report.Checks, doctorCheck{
			ID:      "config.load",
			Status:  doctorStatusPass,
			Message: "configuration loaded",
		})
	} else {
		report.Checks = append(report.Checks, doctorCheck{
			ID:         "config.load",
			Status:     doctorStatusFail,
			Message:    "configuration unavailable",
			Suggestion: "check config initialization",
		})
	}

	validationErrors := []error{fmt.Errorf("configuration unavailable")}
	if cfg != nil {
		validationErrors = config.Validate(cfg)
	}
	if len(validationErrors) == 0 {
		report.Checks = append(report.Checks, doctorCheck{
			ID:      "config.validate",
			Status:  doctorStatusPass,
			Message: "configuration is valid",
		})
	} else {
		report.Checks = append(report.Checks, doctorCheck{
			ID:         "config.validate",
			Status:     doctorStatusFail,
			Message:    fmt.Sprintf("configuration has %d validation error(s)", len(validationErrors)),
			Suggestion: "run `clinvk config lint` for detailed validation output",
		})
	}

	if err := config.EnsureConfigDir(); err != nil {
		report.Checks = append(report.Checks, doctorCheck{
			ID:         "sessions.store_access",
			Status:     doctorStatusFail,
			Message:    fmt.Sprintf("session store inaccessible: %v", err),
			Suggestion: "check permissions for ~/.clinvk and ~/.clinvk/sessions",
		})
	} else {
		report.Checks = append(report.Checks, doctorCheck{
			ID:      "sessions.store_access",
			Status:  doctorStatusPass,
			Message: fmt.Sprintf("session store accessible: %s", config.SessionsDir()),
		})
	}

	enabled := config.EnabledBackends()
	for _, name := range enabled {
		b, err := backend.Get(name)
		if err != nil {
			report.Checks = append(report.Checks, doctorCheck{
				ID:         "backends." + name + ".available",
				Status:     doctorStatusFail,
				Message:    fmt.Sprintf("backend registry lookup failed: %v", err),
				Suggestion: "check backend registration",
			})
			continue
		}

		if b.IsAvailable() {
			report.Checks = append(report.Checks, doctorCheck{
				ID:      "backends." + name + ".available",
				Status:  doctorStatusPass,
				Message: "backend CLI is available",
			})
		} else {
			report.Checks = append(report.Checks, doctorCheck{
				ID:         "backends." + name + ".available",
				Status:     doctorStatusWarn,
				Message:    "backend CLI is not available in PATH",
				Suggestion: fmt.Sprintf("install %s CLI or disable it in config", name),
			})
		}
	}

	keys := auth.LoadAPIKeys()
	if len(keys) == 0 {
		report.Checks = append(report.Checks, doctorCheck{
			ID:         "server.auth",
			Status:     doctorStatusWarn,
			Message:    "API key authentication is disabled (no keys configured)",
			Suggestion: "configure CLINVK_API_KEYS or server.api_keys_gopass_path for production",
		})
	} else {
		report.Checks = append(report.Checks, doctorCheck{
			ID:      "server.auth",
			Status:  doctorStatusPass,
			Message: fmt.Sprintf("API key authentication is enabled with %d key(s)", len(keys)),
		})
	}

	for _, check := range report.Checks {
		switch check.Status {
		case doctorStatusPass:
			report.Summary.Pass++
		case doctorStatusWarn:
			report.Summary.Warn++
		case doctorStatusFail:
			report.Summary.Fail++
		}
	}

	return report
}

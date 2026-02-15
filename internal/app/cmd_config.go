package app

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
)

// configCmd manages configuration.
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		fmt.Printf("Default Backend: %s\n", cfg.DefaultBackend)
		fmt.Printf("\nBackends:\n")
		for name, bcfg := range cfg.Backends {
			fmt.Printf("  %s:\n", name)
			if bcfg.Model != "" {
				fmt.Printf("    model: %s\n", bcfg.Model)
			}
			if bcfg.AllowedTools != "" {
				fmt.Printf("    allowed_tools: %s\n", bcfg.AllowedTools)
			}
		}
		fmt.Printf("\nSession:\n")
		fmt.Printf("  retention_days: %d\n", cfg.Session.RetentionDays)

		fmt.Printf("\nAvailable backends:\n")
		for _, name := range backend.List() {
			b, _ := backend.Get(name)
			status := "not installed"
			if b.IsAvailable() {
				status = "available"
			}
			fmt.Printf("  %s: %s\n", name, status)
		}

		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		if err := config.Set(key, value); err != nil {
			return fmt.Errorf("failed to set config: %w", err)
		}

		fmt.Printf("Set %s = %s\n", key, value)
		return nil
	},
}

var (
	configLintFile string
	configLintJSON bool
)

type configLintReport struct {
	SchemaVersion string   `json:"schema_version"`
	Valid         bool     `json:"valid"`
	ErrorCount    int      `json:"error_count"`
	Errors        []string `json:"errors,omitempty"`
}

var configLintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Validate configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		if configLintFile != "" {
			loadedCfg, err := config.LoadFromPath(configLintFile)
			if err != nil {
				return fmt.Errorf("%w: failed to load config file %q: %v", ErrValidationFailed, configLintFile, err)
			}
			cfg = loadedCfg
		}

		validationErrors := config.Validate(cfg)
		errorMessages := make([]string, 0, len(validationErrors))
		for _, err := range validationErrors {
			errorMessages = append(errorMessages, err.Error())
		}

		report := configLintReport{
			SchemaVersion: "v1",
			Valid:         len(validationErrors) == 0,
			ErrorCount:    len(validationErrors),
			Errors:        errorMessages,
		}

		if configLintJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(report); err != nil {
				return err
			}
		} else {
			if report.Valid {
				fmt.Println("Configuration is valid.")
			} else {
				fmt.Printf("Configuration has %d error(s):\n", report.ErrorCount)
				for _, msg := range report.Errors {
					fmt.Printf("- %s\n", msg)
				}
				fmt.Printf("\nHint: use `%s` for machine-readable output.\n", strings.TrimSpace("clinvk config lint --json"))
			}
		}

		if !report.Valid {
			return fmt.Errorf("%w: configuration has %d validation error(s)", ErrValidationFailed, report.ErrorCount)
		}

		return nil
	},
}

func init() {
	configLintCmd.Flags().StringVar(&configLintFile, "config", "", "path to config file to validate")
	configLintCmd.Flags().BoolVar(&configLintJSON, "json", false, "output validation result as JSON")
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configLintCmd)
}

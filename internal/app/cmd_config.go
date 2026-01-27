package app

import (
	"fmt"

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
		fmt.Printf("  auto_resume: %v\n", cfg.Session.AutoResume)
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

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
}

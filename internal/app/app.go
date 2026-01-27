// Package app provides the CLI application using cobra.
package app

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/executor"
	"github.com/signalridge/clinvoker/internal/session"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var (
	cfgFile     string
	backendName string
	modelName   string
	workDir     string
	dryRun      bool
)

var rootCmd = &cobra.Command{
	Use:   "clinvk [prompt]",
	Short: "Unified AI CLI wrapper for multiple backends",
	Long: `clinvk is a unified CLI wrapper that orchestrates multiple AI CLI backends
(Claude Code, Codex CLI, Gemini CLI) with session persistence and parallel task execution.

Examples:
  clinvk "fix the bug in auth.go"
  clinvk --backend codex "implement user registration"
  clinvk -b gemini "generate unit tests"`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPrompt,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.clinvk/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&backendName, "backend", "b", "", "AI backend to use (claude, codex, gemini)")
	rootCmd.PersistentFlags().StringVarP(&modelName, "model", "m", "", "model to use for the backend")
	rootCmd.PersistentFlags().StringVarP(&workDir, "workdir", "w", "", "working directory for the AI backend")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "print command without executing")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(resumeCmd)
	rootCmd.AddCommand(sessionsCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(parallelCmd)
	rootCmd.AddCommand(compareCmd)
	rootCmd.AddCommand(chainCmd)
}

func initConfig() {
	if err := config.Init(cfgFile); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
	}
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// SetVersion sets the version info for the CLI.
func SetVersion(v, c, d string) {
	version = v
	commit = c
	date = d
}

func runPrompt(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	prompt := args[0]

	// Determine backend
	cfg := config.Get()
	bn := backendName
	if bn == "" {
		bn = cfg.DefaultBackend
	}
	if bn == "" {
		bn = "claude"
	}

	// Get backend
	b, err := backend.Get(bn)
	if err != nil {
		return fmt.Errorf("backend error: %w", err)
	}

	// Skip availability check in dry-run mode
	if !dryRun && !b.IsAvailable() {
		return fmt.Errorf("backend %q is not available (CLI not found in PATH)", bn)
	}

	// Build options
	opts := &backend.Options{
		WorkDir: workDir,
		Model:   modelName,
	}

	// Get backend-specific config
	if bcfg, ok := cfg.Backends[bn]; ok {
		if opts.Model == "" {
			opts.Model = bcfg.Model
		}
		opts.AllowedTools = bcfg.AllowedTools
	}

	// Build command
	execCmd := b.BuildCommand(prompt, opts)

	if dryRun {
		fmt.Printf("Would execute: %s %v\n", execCmd.Path, execCmd.Args[1:])
		return nil
	}

	// Create session
	store := session.NewStore()
	sess, err := store.Create(bn, workDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create session: %v\n", err)
	}

	// Execute
	exec := executor.New()
	exitCode, err := exec.Run(execCmd)

	// Update session
	if sess != nil {
		sess.MarkUsed()
		if err := store.Save(sess); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save session: %v\n", err)
		}
	}

	if err != nil {
		return err
	}

	if exitCode != 0 {
		os.Exit(exitCode)
	}

	return nil
}

package app

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/mock"
	"github.com/signalridge/clinvoker/internal/session"
)

func resetAppGlobals() {
	cfgFile = ""
	backendName = ""
	modelName = ""
	workDir = ""
	dryRun = false
	outputFormat = "json"
	showUsageFlag = false
	continueLastSession = false
	ephemeralMode = false
	mcpTransport = ""
	mcpHost = ""
	mcpPort = 0
	mcpPath = ""
	mcpExpose = false
	compareBackends = ""
	compareAllBackends = false
	compareJSON = false
	compareSequential = false
	parallelFile = ""
	maxParallel = defaultMaxParallel
	parallelFailFast = false
	parallelJSON = false
	parallelQuiet = false
	chainFile = ""
	chainInputFile = ""
	chainJSONFlag = false
	listBackendFilter = ""
	listStatusFilter = ""
	listTagFilter = ""
	listQueryFilter = ""
	listLimit = 0
	listOffset = 0
	listJSON = false
	showJSON = false
	cleanOlderThan = ""
	cleanDryRun = false
	tagDryRun = false
}

func newFlagTestCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().StringVarP(&outputFormat, "output-format", "o", "json", "output format")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "dry run")
	cmd.Flags().BoolVar(&showUsageFlag, "show-usage", false, "show usage")
	return cmd
}

func setupTestConfig(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	config.Reset()
	if err := config.Init(""); err != nil {
		t.Fatalf("config init failed: %v", err)
	}
}

func TestNormalizeFlags_ConfigDefaults(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()

	cfg := config.Get()
	cfg.Output.Format = "text"
	cfg.UnifiedFlags.DryRun = true

	cmd := newFlagTestCmd()
	flags := normalizeFlags(cmd)

	if flags.outputFormat != "text" {
		t.Fatalf("outputFormat = %q, want %q", flags.outputFormat, "text")
	}
	if !flags.dryRun {
		t.Fatal("dryRun should be true when config UnifiedFlags.DryRun is true")
	}
}

func TestNormalizeFlags_ExplicitFlagsOverride(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()

	cfg := config.Get()
	cfg.Output.Format = "text"

	cmd := newFlagTestCmd()
	if err := cmd.Flags().Set("output-format", "stream-json"); err != nil {
		t.Fatalf("set output-format: %v", err)
	}
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatalf("set dry-run: %v", err)
	}

	flags := normalizeFlags(cmd)
	if flags.outputFormat != "stream-json" {
		t.Fatalf("outputFormat = %q, want %q", flags.outputFormat, "stream-json")
	}
	if !flags.dryRun {
		t.Fatal("dryRun should be true when flag set")
	}
}

func TestNormalizeFlags_ShowUsageOverride(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()

	cfg := config.Get()
	cfg.Output.ShowTokens = false

	cmd := newFlagTestCmd()
	if err := cmd.Flags().Set("show-usage", "true"); err != nil {
		t.Fatalf("set show-usage: %v", err)
	}

	flags := normalizeFlags(cmd)
	if !flags.showUsage {
		t.Fatal("showUsage should be true when --show-usage is set")
	}
}

func TestPreparePromptContext_InvalidOutputFormat(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()

	cmd := newFlagTestCmd()
	if err := cmd.Flags().Set("output-format", "xml"); err != nil {
		t.Fatalf("set output-format: %v", err)
	}

	_, err := preparePromptContext(cmd, "hello")
	if err == nil {
		t.Fatal("expected error for invalid output format")
	}
	if !strings.Contains(err.Error(), "invalid output format") {
		t.Fatalf("error = %q, want to contain %q", err.Error(), "invalid output format")
	}
}

func TestPreparePromptContext_DefaultBackend(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()

	cfg := config.Get()
	cfg.DefaultBackend = "mock-default"

	mockBackend := mock.NewMockBackend("mock-default", mock.WithAvailable(true))
	t.Cleanup(mock.WithMockBackend(t, mockBackend))

	cmd := newFlagTestCmd()
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatalf("set dry-run: %v", err)
	}

	ctx, err := preparePromptContext(cmd, "hello")
	if err != nil {
		t.Fatalf("preparePromptContext failed: %v", err)
	}
	if ctx.backendName != "mock-default" {
		t.Fatalf("backendName = %q, want %q", ctx.backendName, "mock-default")
	}
	if !ctx.dryRun {
		t.Fatal("expected dryRun to be true")
	}
}

func TestRunPrompt_ContinueNoSessions(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()

	continueLastSession = true
	cmd := newFlagTestCmd()

	err := runPrompt(cmd, []string{})
	if err == nil {
		t.Fatal("expected error when no resumable sessions exist")
	}
	if !strings.Contains(err.Error(), "no resumable sessions") {
		t.Fatalf("error = %q, want to contain %q", err.Error(), "no resumable sessions")
	}
}

func TestRunContinueLastSession_DryRun(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()

	mockBackend := mock.NewMockBackend("mock", mock.WithAvailable(true))
	t.Cleanup(mock.WithMockBackend(t, mockBackend))

	store := session.NewStore()
	sess, err := store.Create("mock", "/tmp")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	sess.BackendSessionID = "session-123"
	if err := store.Save(sess); err != nil {
		t.Fatalf("save session: %v", err)
	}

	flags := &normalizedFlags{outputFormat: "json", dryRun: true}
	if err := runContinueLastSession(nil, "follow up", flags); err != nil {
		t.Fatalf("runContinueLastSession failed: %v", err)
	}
}

func TestExecute_NoArgsShowsHelp(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()

	rootCmd.SetArgs([]string{})
	defer rootCmd.SetArgs(nil)

	if err := Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
}

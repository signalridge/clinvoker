package service

import (
	"context"
	"sync"
	"testing"

	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/mock"
)

func TestExecutor_ConcurrentExecutePromptAndList(t *testing.T) {
	config.Reset()
	t.Cleanup(config.Reset)
	t.Setenv("HOME", t.TempDir())
	if err := config.Init(""); err != nil {
		t.Fatalf("config init failed: %v", err)
	}

	mockBackend := mock.NewMockBackend("mock", mock.WithAvailable(true))
	t.Cleanup(mock.WithMockBackend(t, mockBackend))

	executor := NewExecutor()
	req := &PromptRequest{
		Backend: "mock",
		Prompt:  "hello",
		DryRun:  true,
	}

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = executor.ExecutePrompt(context.Background(), req)
		}()
	}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = executor.ListSessions(context.Background())
		}()
	}

	wg.Wait()
}

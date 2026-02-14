package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func FuzzHTTPTransportHandler(f *testing.F) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
	registry := NewRegistry()
	registry.Register(&ToolDefinition{
		Tool: Tool{Name: "fuzz_tool"},
		Handler: func(_ context.Context, _ json.RawMessage) (*ToolCallResult, error) {
			return &ToolCallResult{Content: []ContentBlock{TextContent("ok")}}, nil
		},
	})

	transport := NewHTTPTransport(NewDispatcher(registry, logger), logger)
	handler := transport.Handler()

	f.Add([]byte(`{"jsonrpc":"2.0","id":1,"method":"ping"}`))
	f.Add([]byte(`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"fuzz_tool","arguments":{}}}`))
	f.Add([]byte(`{"jsonrpc":"2.0","method":"ping"}`))
	f.Add([]byte(`{"jsonrpc":"1.0","id":3,"method":"ping"}`))
	f.Add([]byte(`{bad json`))

	f.Fuzz(func(t *testing.T, body []byte) {
		if len(body) > 1<<20 {
			return
		}
		req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code == http.StatusNoContent || rec.Body.Len() == 0 {
			return
		}

		var resp Response
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	})
}

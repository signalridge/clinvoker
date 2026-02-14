package mcp

import (
	"encoding/json"
	"testing"
)

func TestRequest_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		request  Request
		wantJSON string
	}{
		{
			name: "initialize request",
			request: Request{
				JSONRPC: "2.0",
				ID:      json.RawMessage(`1`),
				Method:  "initialize",
				Params:  json.RawMessage(`{"protocolVersion":"2024-11-05"}`),
			},
			wantJSON: `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05"}}`,
		},
		{
			name: "tools/list request",
			request: Request{
				JSONRPC: "2.0",
				ID:      json.RawMessage(`"req-123"`),
				Method:  "tools/list",
				Params:  nil,
			},
			wantJSON: `{"jsonrpc":"2.0","id":"req-123","method":"tools/list"}`,
		},
		{
			name: "notification (no id)",
			request: Request{
				JSONRPC: "2.0",
				ID:      nil,
				Method:  "notifications/initialized",
				Params:  json.RawMessage(`{}`),
			},
			wantJSON: `{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("marshal error: %v", err)
			}

			// Unmarshal back
			var req Request
			if err := json.Unmarshal(data, &req); err != nil {
				t.Fatalf("unmarshal error: %v", err)
			}

			if req.JSONRPC != tt.request.JSONRPC {
				t.Errorf("JSONRPC mismatch: got %q, want %q", req.JSONRPC, tt.request.JSONRPC)
			}
			if req.Method != tt.request.Method {
				t.Errorf("Method mismatch: got %q, want %q", req.Method, tt.request.Method)
			}
		})
	}
}

func TestResponse_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		response Response
	}{
		{
			name: "success response with string result",
			response: Response{
				JSONRPC: "2.0",
				ID:      json.RawMessage(`1`),
				Result:  json.RawMessage(`{"tools":[]}`),
			},
		},
		{
			name: "success response with number id",
			response: Response{
				JSONRPC: "2.0",
				ID:      json.RawMessage(`42`),
				Result:  json.RawMessage(`{"protocolVersion":"2024-11-05","capabilities":{}}`),
			},
		},
		{
			name: "error response",
			response: Response{
				JSONRPC: "2.0",
				ID:      json.RawMessage(`"abc"`),
				Error: &RPCError{
					Code:    CodeInvalidParams,
					Message: "Missing required parameter",
					Data:    json.RawMessage(`{"field":"name"}`),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.response)
			if err != nil {
				t.Fatalf("marshal error: %v", err)
			}

			// Unmarshal back
			var resp Response
			if err := json.Unmarshal(data, &resp); err != nil {
				t.Fatalf("unmarshal error: %v", err)
			}

			if resp.JSONRPC != tt.response.JSONRPC {
				t.Errorf("JSONRPC mismatch: got %q, want %q", resp.JSONRPC, tt.response.JSONRPC)
			}
			if resp.Error != nil && tt.response.Error != nil {
				if resp.Error.Code != tt.response.Error.Code {
					t.Errorf("Error.Code mismatch: got %d, want %d", resp.Error.Code, tt.response.Error.Code)
				}
				if resp.Error.Message != tt.response.Error.Message {
					t.Errorf("Error.Message mismatch: got %q, want %q", resp.Error.Message, tt.response.Error.Message)
				}
			}
		})
	}
}

func TestNewResponse(t *testing.T) {
	id := json.RawMessage(`123`)
	result := map[string]string{"status": "ok"}

	resp := NewResponse(id, result)

	if resp.JSONRPC != jsonRPCVersion {
		t.Errorf("JSONRPC version mismatch: got %q, want %q", resp.JSONRPC, jsonRPCVersion)
	}
	if string(resp.ID) != string(id) {
		t.Errorf("ID mismatch: got %s, want %s", resp.ID, id)
	}
	if resp.Result == nil {
		t.Error("Result should not be nil")
	}
	if resp.Error != nil {
		t.Error("Error should be nil for success response")
	}
}

func TestNewErrorResponse(t *testing.T) {
	tests := []struct {
		name        string
		id          json.RawMessage
		code        int
		message     string
		data        any
		wantDataNil bool
	}{
		{
			name:        "error with data",
			id:          json.RawMessage(`"req-1"`),
			code:        CodeInvalidParams,
			message:     "Invalid parameter",
			data:        map[string]string{"field": "name"},
			wantDataNil: false,
		},
		{
			name:        "error without data",
			id:          json.RawMessage(`42`),
			code:        CodeMethodNotFound,
			message:     "Method not found",
			data:        nil,
			wantDataNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewErrorResponse(tt.id, tt.code, tt.message, tt.data)

			if resp.JSONRPC != jsonRPCVersion {
				t.Errorf("JSONRPC version mismatch: got %q, want %q", resp.JSONRPC, jsonRPCVersion)
			}
			if string(resp.ID) != string(tt.id) {
				t.Errorf("ID mismatch: got %s, want %s", resp.ID, tt.id)
			}
			if resp.Error == nil {
				t.Fatal("Error should not be nil")
			}
			if resp.Error.Code != tt.code {
				t.Errorf("Error.Code mismatch: got %d, want %d", resp.Error.Code, tt.code)
			}
			if resp.Error.Message != tt.message {
				t.Errorf("Error.Message mismatch: got %q, want %q", resp.Error.Message, tt.message)
			}
			if tt.wantDataNil {
				if resp.Error.Data != nil {
					t.Errorf("Error.Data should be nil, got %v", resp.Error.Data)
				}
			}
		})
	}
}

func TestNewNotification(t *testing.T) {
	params := map[string]string{"level": "info", "message": "test"}
	notif := NewNotification("notifications/message", params)

	if notif.JSONRPC != jsonRPCVersion {
		t.Errorf("JSONRPC = %q, want %q", notif.JSONRPC, jsonRPCVersion)
	}
	if notif.Method != "notifications/message" {
		t.Errorf("Method = %q, want %q", notif.Method, "notifications/message")
	}
	if notif.Params == nil {
		t.Error("Params should not be nil")
	}
}

func TestNotification_MarshalUnmarshal(t *testing.T) {
	notification := Notification{
		JSONRPC: "2.0",
		Method:  "notifications/message",
		Params:  map[string]string{"level": "info", "message": "test"},
	}

	data, err := json.Marshal(notification)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var notif Notification
	if err := json.Unmarshal(data, &notif); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if notif.JSONRPC != notification.JSONRPC {
		t.Errorf("JSONRPC mismatch: got %q, want %q", notif.JSONRPC, notification.JSONRPC)
	}
	if notif.Method != notification.Method {
		t.Errorf("Method mismatch: got %q, want %q", notif.Method, notification.Method)
	}
}

func TestStandardErrorCodes(t *testing.T) {
	// Verify standard JSON-RPC error codes
	if CodeParseError != -32700 {
		t.Errorf("CodeParseError = %d, want -32700", CodeParseError)
	}
	if CodeInvalidRequest != -32600 {
		t.Errorf("CodeInvalidRequest = %d, want -32600", CodeInvalidRequest)
	}
	if CodeMethodNotFound != -32601 {
		t.Errorf("CodeMethodNotFound = %d, want -32601", CodeMethodNotFound)
	}
	if CodeInvalidParams != -32602 {
		t.Errorf("CodeInvalidParams = %d, want -32602", CodeInvalidParams)
	}
	if CodeInternalError != -32603 {
		t.Errorf("CodeInternalError = %d, want -32603", CodeInternalError)
	}
}

func TestJSONRPCVersion(t *testing.T) {
	if jsonRPCVersion != "2.0" {
		t.Errorf("jsonRPCVersion = %q, want \"2.0\"", jsonRPCVersion)
	}
}

func TestRequest_IsNotification(t *testing.T) {
	tests := []struct {
		name     string
		id       json.RawMessage
		expected bool
	}{
		{
			name:     "nil id",
			id:       nil,
			expected: true,
		},
		{
			name:     "null id",
			id:       json.RawMessage(`null`),
			expected: true,
		},
		{
			name:     "number id",
			id:       json.RawMessage(`123`),
			expected: false,
		},
		{
			name:     "string id",
			id:       json.RawMessage(`"abc"`),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &Request{
				JSONRPC: "2.0",
				ID:      tt.id,
				Method:  "test",
			}
			if got := req.IsNotification(); got != tt.expected {
				t.Errorf("IsNotification() = %v, want %v", got, tt.expected)
			}
		})
	}
}

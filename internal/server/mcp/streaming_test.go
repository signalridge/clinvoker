package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestWithNotificationSender(t *testing.T) {
	sender := &mockNotificationSender{}
	ctx := WithNotificationSender(context.Background(), sender)

	retrieved := GetNotificationSender(ctx)
	if retrieved != sender {
		t.Error("GetNotificationSender did not return the same sender")
	}
}

func TestGetNotificationSender_NotSet(t *testing.T) {
	ctx := context.Background()

	sender := GetNotificationSender(ctx)
	if sender != nil {
		t.Error("expected nil sender when not set")
	}
}

func TestGetNotificationSender_WrongType(t *testing.T) {
	type wrongKey struct{}
	ctx := context.WithValue(context.Background(), wrongKey{}, "not a sender")

	sender := GetNotificationSender(ctx)
	if sender != nil {
		t.Error("expected nil sender when wrong type stored")
	}
}

func TestNotificationSender_SendError(t *testing.T) {
	sender := &mockNotificationSender{
		sendErr: errors.New("send failed"),
	}

	notif := &Notification{
		JSONRPC: "2.0",
		Method:  "test",
		Params:  map[string]string{},
	}

	err := sender.SendNotification(notif)
	if err == nil {
		t.Error("expected error from mock sender")
	}
	if err.Error() != "send failed" {
		t.Errorf("error = %v, want 'send failed'", err)
	}
}

func TestNotificationSender_MultipleCalls(t *testing.T) {
	sender := &mockNotificationSender{}
	ctx := WithNotificationSender(context.Background(), sender)

	// Send multiple notifications
	for i := 0; i < 5; i++ {
		notif := &Notification{
			JSONRPC: "2.0",
			Method:  "test",
			Params:  map[string]string{},
		}
		sender.SendNotification(notif)
	}

	if sender.callCount != 5 {
		t.Errorf("callCount = %d, want 5", sender.callCount)
	}

	// Verify context retrieval still works
	retrieved := GetNotificationSender(ctx)
	if retrieved != sender {
		t.Error("context retrieval failed after multiple sends")
	}
}

func TestTranslateEvent_NilEvent(t *testing.T) {
	notif := TranslateEvent(nil, nil)
	if notif != nil {
		t.Error("expected nil for nil event")
	}
}

func TestLogMessageParams_Marshal(t *testing.T) {
	params := LogMessageParams{
		Level:  "info",
		Logger: "clinvk",
		Data:   map[string]string{"key": "value"},
	}

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var parsed LogMessageParams
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if parsed.Level != params.Level {
		t.Errorf("level = %q, want %q", parsed.Level, params.Level)
	}
	if parsed.Logger != params.Logger {
		t.Errorf("logger = %q, want %q", parsed.Logger, params.Logger)
	}
}

func TestProgressNotificationParams_Marshal(t *testing.T) {
	params := ProgressNotificationParams{
		ProgressToken: "token-123",
		Progress:      50.5,
		Total:         100,
		Message:       "Halfway done",
	}

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var parsed ProgressNotificationParams
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if parsed.ProgressToken != params.ProgressToken {
		t.Errorf("progressToken = %v, want %v", parsed.ProgressToken, params.ProgressToken)
	}
	if parsed.Progress != params.Progress {
		t.Errorf("progress = %f, want %f", parsed.Progress, params.Progress)
	}
	if parsed.Total != params.Total {
		t.Errorf("total = %f, want %f", parsed.Total, params.Total)
	}
}

// mockNotificationSender is a test implementation of NotificationSender
type mockNotificationSender struct {
	callCount int
	sendErr   error
}

func (m *mockNotificationSender) SendNotification(notification *Notification) error {
	m.callCount++
	return m.sendErr
}

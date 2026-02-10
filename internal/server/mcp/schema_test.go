package mcp

import (
	"reflect"
	"testing"
	"time"

	"github.com/signalridge/clinvoker/internal/server/handlers"
)

func TestSchemaForType_PromptRequest(t *testing.T) {
	schema, err := schemaForType(reflect.TypeOf(handlers.PromptRequest{}))
	if err != nil {
		t.Fatalf("schemaForType error: %v", err)
	}

	if schema["type"] != "object" {
		t.Fatalf("type = %v, want object", schema["type"])
	}

	props := schema["properties"].(map[string]any)
	backend := props["backend"].(map[string]any)
	if backend["description"] == "" {
		t.Fatal("backend description missing")
	}

	required := toStringSlice(schema["required"])
	if !containsString(required, "backend") || !containsString(required, "prompt") {
		t.Fatalf("required = %v, want backend and prompt", required)
	}
	if containsString(required, "model") {
		t.Fatalf("model should not be required, required = %v", required)
	}

	metadata := props["metadata"].(map[string]any)
	if metadata["type"] != "object" {
		t.Fatalf("metadata type = %v, want object", metadata["type"])
	}
}

func TestSchemaForType_QueryAndPathTags(t *testing.T) {
	schema, err := schemaForType(reflect.TypeOf(handlers.SessionsInput{}))
	if err != nil {
		t.Fatalf("schemaForType error: %v", err)
	}

	props := schema["properties"].(map[string]any)
	for _, key := range []string{"backend", "status", "limit", "offset"} {
		if _, ok := props[key]; !ok {
			t.Fatalf("missing property %q", key)
		}
	}

	required := toStringSlice(schema["required"])
	if len(required) != 0 {
		t.Fatalf("expected no required fields, got %v", required)
	}

	pathSchema, err := schemaForType(reflect.TypeOf(handlers.GetSessionInput{}))
	if err != nil {
		t.Fatalf("schemaForType error: %v", err)
	}
	pathRequired := toStringSlice(pathSchema["required"])
	if !containsString(pathRequired, "id") {
		t.Fatalf("expected id to be required, got %v", pathRequired)
	}
}

func TestSchemaForType_TimeAndDuration(t *testing.T) {
	type sample struct {
		When time.Time     `json:"when"`
		For  time.Duration `json:"for"`
	}

	schema, err := schemaForType(reflect.TypeOf(sample{}))
	if err != nil {
		t.Fatalf("schemaForType error: %v", err)
	}

	props := schema["properties"].(map[string]any)
	when := props["when"].(map[string]any)
	if when["type"] != "string" || when["format"] != "date-time" {
		t.Fatalf("when schema = %v, want string date-time", when)
	}

	forField := props["for"].(map[string]any)
	if forField["type"] != "string" {
		t.Fatalf("for schema = %v, want string", forField)
	}
}

func toStringSlice(v any) []string {
	if v == nil {
		return nil
	}
	switch value := v.(type) {
	case []string:
		return value
	case []any:
		out := make([]string, 0, len(value))
		for _, item := range value {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func containsString(list []string, target string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}

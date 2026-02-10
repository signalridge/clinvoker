package mcp

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

var (
	timeType     = reflect.TypeOf(time.Time{})
	durationType = reflect.TypeOf(time.Duration(0))
)

type schemaGenerator struct {
	visiting map[reflect.Type]bool
}

func schemaForType(t reflect.Type) (map[string]any, error) {
	gen := &schemaGenerator{visiting: make(map[reflect.Type]bool)}
	return gen.schemaForType(t)
}

func mustSchemaJSON(v any) json.RawMessage {
	t := reflect.TypeOf(v)
	schema, err := schemaForType(t)
	if err != nil {
		panic(fmt.Sprintf("mustSchemaJSON: %v", err))
	}
	data, err := json.Marshal(schema)
	if err != nil {
		panic(fmt.Sprintf("mustSchemaJSON marshal: %v", err))
	}
	return data
}

func (g *schemaGenerator) schemaForType(t reflect.Type) (map[string]any, error) {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t == timeType {
		return map[string]any{"type": "string", "format": "date-time"}, nil
	}
	if t == durationType {
		return map[string]any{"type": "string"}, nil
	}

	switch t.Kind() {
	case reflect.Struct:
		if g.visiting[t] {
			return map[string]any{"type": "object"}, nil
		}
		g.visiting[t] = true
		defer delete(g.visiting, t)

		properties := map[string]any{}
		var required []string

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.PkgPath != "" {
				continue
			}

			name, tagSource, omitEmpty, ok := fieldSchemaName(&field)
			if !ok {
				continue
			}

			propSchema, err := g.schemaForType(field.Type)
			if err != nil {
				return nil, err
			}

			if doc := field.Tag.Get("doc"); doc != "" {
				propSchema["description"] = doc
			}

			properties[name] = propSchema

			isOptional := omitEmpty || field.Type.Kind() == reflect.Pointer || tagSource == "query"
			if !isOptional {
				required = append(required, name)
			}
		}

		schema := map[string]any{
			"type":       "object",
			"properties": properties,
		}
		if len(required) > 0 {
			schema["required"] = required
		}
		return schema, nil

	case reflect.Slice, reflect.Array:
		items, err := g.schemaForType(t.Elem())
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"type":  "array",
			"items": items,
		}, nil

	case reflect.Map:
		additional := map[string]any{"type": "string"}
		if t.Key().Kind() == reflect.String {
			valueSchema, err := g.schemaForType(t.Elem())
			if err != nil {
				return nil, err
			}
			additional = valueSchema
		}
		return map[string]any{
			"type":                 "object",
			"additionalProperties": additional,
		}, nil

	case reflect.String:
		return map[string]any{"type": "string"}, nil
	case reflect.Bool:
		return map[string]any{"type": "boolean"}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]any{"type": "integer"}, nil
	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}, nil
	case reflect.Interface:
		return map[string]any{}, nil
	case reflect.Invalid, reflect.Uintptr, reflect.Complex64, reflect.Complex128,
		reflect.Chan, reflect.Func, reflect.Pointer, reflect.UnsafePointer:
		return nil, fmt.Errorf("unsupported type: %s", t.Kind())
	default:
		return nil, fmt.Errorf("unsupported type: %s", t.Kind())
	}
}

func fieldSchemaName(field *reflect.StructField) (string, string, bool, bool) {
	if tag := field.Tag.Get("json"); tag != "" {
		name, opts := parseTag(tag)
		if name == "-" {
			return "", "", false, false
		}
		if name == "" {
			name = field.Name
		}
		return name, "json", strings.Contains(opts, "omitempty"), true
	}

	if tag := field.Tag.Get("query"); tag != "" {
		return tag, "query", false, true
	}
	if tag := field.Tag.Get("path"); tag != "" {
		return tag, "path", false, true
	}

	return field.Name, "default", false, true
}

func parseTag(tag string) (string, string) {
	if tag == "" {
		return "", ""
	}
	parts := strings.Split(tag, ",")
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], ",")
}

package policy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

// LoadRuleSource loads a rule source file from JSON or YAML.
func LoadRuleSource(path string) (*RuleSource, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("policy rules file is required when policy is enabled")
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read policy rules file %q: %w", path, err)
	}

	src := &RuleSource{}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		if err := json.Unmarshal(b, src); err != nil {
			return nil, fmt.Errorf("parse policy json %q: %w", path, err)
		}
	default:
		if err := yaml.Unmarshal(b, src); err != nil {
			return nil, fmt.Errorf("parse policy yaml %q: %w", path, err)
		}
	}

	if strings.TrimSpace(src.Version) == "" {
		src.Version = "v1"
	}
	if src.Version != "v1" {
		return nil, fmt.Errorf("unsupported policy rules version %q", src.Version)
	}

	return src, nil
}

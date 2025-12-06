package shelm

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// FormatOutput formats a value for display according to the spec:
// - string: raw string
// - []byte: interpreted as UTF-8
// - map, slice, any structured type: YAML pretty-print
// - nil: no output
func FormatOutput(v any) string {
	if v == nil {
		return ""
	}

	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	case error:
		return fmt.Sprintf("ERROR: %v", val)
	default:
		// For maps, slices, and other structured types, use YAML
		if isStructured(v) {
			data, err := yaml.Marshal(v)
			if err != nil {
				return fmt.Sprintf("%v", v)
			}
			return strings.TrimSpace(string(data))
		}
		// For other types (int, float, bool, etc.), convert to string
		return fmt.Sprintf("%v", v)
	}
}

// isStructured returns true if the value should be YAML-formatted.
func isStructured(v any) bool {
	switch v.(type) {
	case map[string]any, map[any]any, []any, []map[string]any:
		return true
	default:
		return false
	}
}

// FormatVars formats the Vars map for display.
func FormatVars(vars map[string]any) string {
	if len(vars) == 0 {
		return "(no variables set)"
	}
	data, err := yaml.Marshal(vars)
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}
	return strings.TrimSpace(string(data))
}

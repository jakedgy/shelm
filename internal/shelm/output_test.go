package shelm

import (
	"fmt"
	"strings"
	"testing"
)

func TestFormatOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "nil",
			input:    nil,
			expected: "",
		},
		{
			name:     "string",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "byte slice",
			input:    []byte("hello"),
			expected: "hello",
		},
		{
			name:     "integer",
			input:    42,
			expected: "42",
		},
		{
			name:     "float",
			input:    3.14,
			expected: "3.14",
		},
		{
			name:     "boolean true",
			input:    true,
			expected: "true",
		},
		{
			name:     "boolean false",
			input:    false,
			expected: "false",
		},
		{
			name:     "error",
			input:    fmt.Errorf("test error"),
			expected: "ERROR: test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatOutput(tt.input)
			if result != tt.expected {
				t.Errorf("FormatOutput(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatOutput_Map(t *testing.T) {
	input := map[string]any{
		"name": "myapp",
		"port": 8080,
	}

	result := FormatOutput(input)

	if !strings.Contains(result, "name: myapp") {
		t.Errorf("result should contain 'name: myapp': %s", result)
	}
	if !strings.Contains(result, "port: 8080") {
		t.Errorf("result should contain 'port: 8080': %s", result)
	}
}

func TestFormatOutput_Slice(t *testing.T) {
	input := []any{"a", "b", "c"}

	result := FormatOutput(input)

	if !strings.Contains(result, "- a") {
		t.Errorf("result should contain '- a': %s", result)
	}
	if !strings.Contains(result, "- b") {
		t.Errorf("result should contain '- b': %s", result)
	}
}

func TestFormatOutput_NestedMap(t *testing.T) {
	input := map[string]any{
		"config": map[string]any{
			"database": map[string]any{
				"host": "localhost",
				"port": 5432,
			},
		},
	}

	result := FormatOutput(input)

	if !strings.Contains(result, "config:") {
		t.Errorf("result should contain 'config:': %s", result)
	}
	if !strings.Contains(result, "database:") {
		t.Errorf("result should contain 'database:': %s", result)
	}
	if !strings.Contains(result, "host: localhost") {
		t.Errorf("result should contain 'host: localhost': %s", result)
	}
}

func TestFormatValues(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		contains []string
	}{
		{
			name:     "empty",
			input:    map[string]any{},
			contains: []string{"no values set"},
		},
		{
			name: "simple",
			input: map[string]any{
				"foo": "bar",
			},
			contains: []string{"foo: bar"},
		},
		{
			name: "multiple",
			input: map[string]any{
				"name": "myapp",
				"port": 8080,
			},
			contains: []string{"name: myapp", "port: 8080"},
		},
		{
			name: "nested",
			input: map[string]any{
				"config": map[string]any{
					"key": "value",
				},
			},
			contains: []string{"config:", "key: value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatValues(tt.input)

			for _, c := range tt.contains {
				if !strings.Contains(result, c) {
					t.Errorf("FormatValues result should contain %q: %s", c, result)
				}
			}
		})
	}
}

func TestIsStructured(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{
			name:     "map[string]any",
			input:    map[string]any{"key": "value"},
			expected: true,
		},
		{
			name:     "[]any",
			input:    []any{1, 2, 3},
			expected: true,
		},
		{
			name:     "[]map[string]any",
			input:    []map[string]any{{"a": 1}},
			expected: true,
		},
		{
			name:     "string",
			input:    "hello",
			expected: false,
		},
		{
			name:     "int",
			input:    42,
			expected: false,
		},
		{
			name:     "bool",
			input:    true,
			expected: false,
		},
		{
			name:     "nil",
			input:    nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isStructured(tt.input)
			if result != tt.expected {
				t.Errorf("isStructured(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

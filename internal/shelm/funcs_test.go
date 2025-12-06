package shelm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShelmReadFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "hello world\nline 2"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Test read
	content, err := shelmReadFile(testFile)
	if err != nil {
		t.Errorf("shelmReadFile error: %v", err)
	}
	if content != testContent {
		t.Errorf("content = %q, want %q", content, testContent)
	}

	// Test read non-existent
	_, err = shelmReadFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestShelmWriteFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "output.txt")

	content := "test content\nmultiple lines"

	// Test write
	result, err := shelmWriteFile(testFile, content)
	if err != nil {
		t.Errorf("shelmWriteFile error: %v", err)
	}
	if result != "" {
		t.Errorf("result should be empty string, got: %q", result)
	}

	// Verify content
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("read file error: %v", err)
	}
	if string(data) != content {
		t.Errorf("file content = %q, want %q", string(data), content)
	}

	// Test write to invalid path
	_, err = shelmWriteFile("/nonexistent/dir/file.txt", "test")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestFromYaml(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		checkKey string
		checkVal any
		wantErr  bool
	}{
		{
			name:     "simple map",
			input:    "foo: bar",
			checkKey: "foo",
			checkVal: "bar",
		},
		{
			name:     "nested map",
			input:    "outer:\n  inner: value",
			checkKey: "outer",
		},
		{
			name:     "integer value",
			input:    "port: 8080",
			checkKey: "port",
			checkVal: 8080,
		},
		{
			name:     "boolean value",
			input:    "enabled: true",
			checkKey: "enabled",
			checkVal: true,
		},
		{
			name:    "invalid yaml",
			input:   "foo: [invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fromYaml(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			m, ok := result.(map[string]any)
			if !ok {
				t.Errorf("result is not a map: %T", result)
				return
			}

			if tt.checkVal != nil {
				if m[tt.checkKey] != tt.checkVal {
					t.Errorf("%s = %v, want %v", tt.checkKey, m[tt.checkKey], tt.checkVal)
				}
			} else {
				if _, ok := m[tt.checkKey]; !ok {
					t.Errorf("key %q not found", tt.checkKey)
				}
			}
		})
	}
}

func TestToYaml(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		contains string
		wantErr  bool
	}{
		{
			name:     "simple map",
			input:    map[string]any{"foo": "bar"},
			contains: "foo: bar",
		},
		{
			name:     "nested map",
			input:    map[string]any{"outer": map[string]any{"inner": "value"}},
			contains: "inner: value",
		},
		{
			name:     "list",
			input:    []any{"a", "b", "c"},
			contains: "- a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toYaml(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !strings.Contains(result, tt.contains) {
				t.Errorf("result should contain %q: %s", tt.contains, result)
			}
		})
	}
}

func TestFromJson(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		checkKey string
		checkVal any
		wantErr  bool
	}{
		{
			name:     "simple object",
			input:    `{"foo": "bar"}`,
			checkKey: "foo",
			checkVal: "bar",
		},
		{
			name:     "number value",
			input:    `{"port": 8080}`,
			checkKey: "port",
			checkVal: float64(8080), // JSON numbers are float64
		},
		{
			name:     "boolean value",
			input:    `{"enabled": true}`,
			checkKey: "enabled",
			checkVal: true,
		},
		{
			name:    "invalid json",
			input:   `{"foo": }`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fromJson(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			m, ok := result.(map[string]any)
			if !ok {
				t.Errorf("result is not a map: %T", result)
				return
			}

			if m[tt.checkKey] != tt.checkVal {
				t.Errorf("%s = %v (%T), want %v (%T)", tt.checkKey, m[tt.checkKey], m[tt.checkKey], tt.checkVal, tt.checkVal)
			}
		})
	}
}

func TestToJson(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		contains string
	}{
		{
			name:     "simple map",
			input:    map[string]any{"foo": "bar"},
			contains: `"foo":"bar"`,
		},
		{
			name:     "number",
			input:    map[string]any{"num": 42},
			contains: `"num":42`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toJson(tt.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !strings.Contains(result, tt.contains) {
				t.Errorf("result should contain %q: %s", tt.contains, result)
			}
		})
	}
}

func TestToPrettyJson(t *testing.T) {
	input := map[string]any{
		"name": "myapp",
		"port": 8080,
	}

	result, err := toPrettyJson(input)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	// Should have indentation
	if !strings.Contains(result, "\n") {
		t.Error("pretty JSON should have newlines")
	}
	if !strings.Contains(result, "  ") {
		t.Error("pretty JSON should have indentation")
	}
}

func TestBuildFuncMap(t *testing.T) {
	// Without capture func
	fm := BuildFuncMap(nil)

	if fm == nil {
		t.Fatal("BuildFuncMap returned nil")
	}

	// Check Sprig functions exist
	sprigFuncs := []string{"upper", "lower", "add", "sub", "list", "dict", "default"}
	for _, name := range sprigFuncs {
		if _, ok := fm[name]; !ok {
			t.Errorf("missing Sprig function: %s", name)
		}
	}

	// Check shelm functions exist
	shelmFuncs := []string{"shelmReadFile", "shelmWriteFile", "fromYaml", "toYaml", "fromJson", "toJson", "toPrettyJson"}
	for _, name := range shelmFuncs {
		if _, ok := fm[name]; !ok {
			t.Errorf("missing shelm function: %s", name)
		}
	}

	// shelmCapture should not exist without capture func
	if _, ok := fm["shelmCapture"]; ok {
		t.Error("shelmCapture should not exist without capture func")
	}

	// With capture func
	captureFunc := func(name string, value any) any {
		return value
	}
	fm = BuildFuncMap(captureFunc)

	if _, ok := fm["shelmCapture"]; !ok {
		t.Error("shelmCapture should exist with capture func")
	}
}

func TestShelmFuncs(t *testing.T) {
	funcs := ShelmFuncs()

	expected := []string{
		"shelmReadFile",
		"shelmWriteFile",
		"fromYaml",
		"toYaml",
		"fromJson",
		"toJson",
		"toPrettyJson",
	}

	for _, name := range expected {
		if _, ok := funcs[name]; !ok {
			t.Errorf("missing function: %s", name)
		}
	}
}

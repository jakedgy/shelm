package shelm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTestHandler() (*CommandHandler, *Context) {
	ctx := NewContext()
	eval := NewEvaluator()
	handler := NewCommandHandler(ctx, eval)
	return handler, ctx
}

func TestCommandHandler_Help(t *testing.T) {
	handler, _ := newTestHandler()

	result := handler.Handle(":help")

	if result.Error != nil {
		t.Errorf("unexpected error: %v", result.Error)
	}
	if result.ShouldExit {
		t.Error("help should not exit")
	}
	if !strings.Contains(result.Output, "shelm") {
		t.Error("help output should contain 'shelm'")
	}
	if !strings.Contains(result.Output, ":values") {
		t.Error("help output should contain ':values'")
	}
	if !strings.Contains(result.Output, ":load") {
		t.Error("help output should contain ':load'")
	}
}

func TestCommandHandler_Values(t *testing.T) {
	handler, ctx := newTestHandler()

	// Empty values
	result := handler.Handle(":values")
	if result.Error != nil {
		t.Errorf("unexpected error: %v", result.Error)
	}
	if !strings.Contains(result.Output, "no values set") {
		t.Errorf("expected 'no values set', got: %s", result.Output)
	}

	// With values
	ctx.Set("foo", "bar")
	ctx.Set("num", 42)

	result = handler.Handle(":values")
	if result.Error != nil {
		t.Errorf("unexpected error: %v", result.Error)
	}
	if !strings.Contains(result.Output, "foo") {
		t.Errorf("output should contain 'foo': %s", result.Output)
	}
}

func TestCommandHandler_Set(t *testing.T) {
	handler, ctx := newTestHandler()

	tests := []struct {
		name     string
		cmd      string
		checkVal string
		checkKey string
		wantErr  bool
	}{
		{
			name:     "simple string",
			cmd:      `:set foo "bar"`,
			checkKey: "foo",
			checkVal: "bar",
		},
		{
			name:     "expression",
			cmd:      `:set result add 1 2`,
			checkKey: "result",
			checkVal: "3",
		},
		{
			name:     "upper function",
			cmd:      `:set upper upper "hello"`,
			checkKey: "upper",
			checkVal: "HELLO",
		},
		{
			name:    "missing args",
			cmd:     `:set foo`,
			wantErr: true,
		},
		{
			name:    "no args",
			cmd:     `:set`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.Handle(tt.cmd)

			if tt.wantErr {
				if result.Error == nil {
					t.Error("expected error")
				}
				return
			}

			if result.Error != nil {
				t.Errorf("unexpected error: %v", result.Error)
				return
			}

			val, ok := ctx.Get(tt.checkKey)
			if !ok {
				t.Errorf("value %q not set", tt.checkKey)
				return
			}

			// For string comparison
			if strVal, ok := val.(string); ok {
				if strVal != tt.checkVal {
					t.Errorf("value = %q, want %q", strVal, tt.checkVal)
				}
			}
		})
	}
}

func TestCommandHandler_Unset(t *testing.T) {
	handler, ctx := newTestHandler()

	ctx.Set("foo", "bar")
	ctx.Set("baz", "qux")

	// Unset existing
	result := handler.Handle(":unset foo")
	if result.Error != nil {
		t.Errorf("unexpected error: %v", result.Error)
	}
	if _, ok := ctx.Get("foo"); ok {
		t.Error("foo should be unset")
	}
	if _, ok := ctx.Get("baz"); !ok {
		t.Error("baz should still exist")
	}

	// Unset non-existent
	result = handler.Handle(":unset nonexistent")
	if result.Error == nil {
		t.Error("expected error for non-existent key")
	}

	// No args
	result = handler.Handle(":unset")
	if result.Error == nil {
		t.Error("expected error for no args")
	}
}

func TestCommandHandler_Load(t *testing.T) {
	handler, ctx := newTestHandler()

	// Create temp directory and files
	tmpDir := t.TempDir()

	// Create YAML file
	yamlFile := filepath.Join(tmpDir, "test.yaml")
	yamlContent := `name: myapp
port: 8080
nested:
  key: value
`
	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create JSON file
	jsonFile := filepath.Join(tmpDir, "test.json")
	jsonContent := `{"name": "jsonapp", "port": 3000}`
	if err := os.WriteFile(jsonFile, []byte(jsonContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Test merge YAML into values
	result := handler.Handle(":load " + yamlFile)
	if result.Error != nil {
		t.Errorf("load yaml error: %v", result.Error)
	}
	if !strings.Contains(result.Output, "merged") {
		t.Errorf("expected 'merged' in output: %s", result.Output)
	}
	if v, ok := ctx.Get("name"); !ok || v != "myapp" {
		t.Errorf("name not loaded correctly: %v", v)
	}
	if v, ok := ctx.Get("port"); !ok || v != 8080 {
		t.Errorf("port not loaded correctly: %v", v)
	}

	// Test load into specific variable
	result = handler.Handle(":load " + jsonFile + " jsondata")
	if result.Error != nil {
		t.Errorf("load json error: %v", result.Error)
	}
	if !strings.Contains(result.Output, "jsondata") {
		t.Errorf("expected 'jsondata' in output: %s", result.Output)
	}
	jsondata, ok := ctx.Get("jsondata")
	if !ok {
		t.Error("jsondata not set")
	}
	if m, ok := jsondata.(map[string]any); ok {
		if m["name"] != "jsonapp" {
			t.Errorf("jsondata.name = %v, want 'jsonapp'", m["name"])
		}
	}

	// Test load non-existent file
	result = handler.Handle(":load /nonexistent/file.yaml")
	if result.Error == nil {
		t.Error("expected error for non-existent file")
	}

	// Test no args
	result = handler.Handle(":load")
	if result.Error == nil {
		t.Error("expected error for no args")
	}
}

func TestCommandHandler_Save(t *testing.T) {
	handler, ctx := newTestHandler()

	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "output.yaml")

	ctx.Set("config", map[string]any{
		"name": "myapp",
		"port": 8080,
	})

	// Save variable
	result := handler.Handle(":save config " + outFile)
	if result.Error != nil {
		t.Errorf("save error: %v", result.Error)
	}

	// Verify file contents
	content, err := os.ReadFile(outFile)
	if err != nil {
		t.Errorf("read saved file: %v", err)
	}
	if !strings.Contains(string(content), "name: myapp") {
		t.Errorf("saved content missing 'name: myapp': %s", content)
	}

	// Save non-existent variable
	result = handler.Handle(":save nonexistent " + outFile)
	if result.Error == nil {
		t.Error("expected error for non-existent variable")
	}

	// Missing args
	result = handler.Handle(":save config")
	if result.Error == nil {
		t.Error("expected error for missing args")
	}
}

func TestCommandHandler_Template(t *testing.T) {
	handler, ctx := newTestHandler()

	tmpDir := t.TempDir()
	tmplFile := filepath.Join(tmpDir, "test.tmpl")

	// Create template file
	tmplContent := `Name: {{ .Values.name }}
Port: {{ .Values.port }}
`
	if err := os.WriteFile(tmplFile, []byte(tmplContent), 0644); err != nil {
		t.Fatal(err)
	}

	ctx.Set("name", "myapp")
	ctx.Set("port", 8080)

	result := handler.Handle(":template " + tmplFile)
	if result.Error != nil {
		t.Errorf("template error: %v", result.Error)
	}
	if !strings.Contains(result.Output, "Name: myapp") {
		t.Errorf("output missing 'Name: myapp': %s", result.Output)
	}
	if !strings.Contains(result.Output, "Port: 8080") {
		t.Errorf("output missing 'Port: 8080': %s", result.Output)
	}

	// Non-existent template
	result = handler.Handle(":template /nonexistent/file.tmpl")
	if result.Error == nil {
		t.Error("expected error for non-existent template")
	}

	// Missing args
	result = handler.Handle(":template")
	if result.Error == nil {
		t.Error("expected error for missing args")
	}
}

func TestCommandHandler_Quit(t *testing.T) {
	handler, _ := newTestHandler()

	result := handler.Handle(":quit")
	if result.Error != nil {
		t.Errorf("unexpected error: %v", result.Error)
	}
	if !result.ShouldExit {
		t.Error("quit should set ShouldExit")
	}

	result = handler.Handle(":exit")
	if !result.ShouldExit {
		t.Error("exit should set ShouldExit")
	}
}

func TestCommandHandler_UnknownCommand(t *testing.T) {
	handler, _ := newTestHandler()

	result := handler.Handle(":unknown")
	if result.Error == nil {
		t.Error("expected error for unknown command")
	}
	if !strings.Contains(result.Error.Error(), "unknown command") {
		t.Errorf("error should mention 'unknown command': %v", result.Error)
	}
}

func TestCommandHandler_EmptyCommand(t *testing.T) {
	handler, _ := newTestHandler()

	result := handler.Handle(":")
	if result.Error == nil {
		t.Error("expected error for empty command")
	}
}

func TestCommandHandler_CaseInsensitive(t *testing.T) {
	handler, _ := newTestHandler()

	// Commands should be case-insensitive
	result := handler.Handle(":HELP")
	if result.Error != nil {
		t.Errorf("HELP should work: %v", result.Error)
	}

	result = handler.Handle(":Help")
	if result.Error != nil {
		t.Errorf("Help should work: %v", result.Error)
	}

	result = handler.Handle(":VALUES")
	if result.Error != nil {
		t.Errorf("VALUES should work: %v", result.Error)
	}
}

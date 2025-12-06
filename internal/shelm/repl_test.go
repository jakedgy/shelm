package shelm

import (
	"bytes"
	"strings"
	"testing"
)

func TestAssignmentPattern(t *testing.T) {
	tests := []struct {
		input    string
		match    bool
		varName  string
		expr     string
	}{
		{
			input:   "foo = bar",
			match:   true,
			varName: "foo",
			expr:    "bar",
		},
		{
			input:   "foo=bar",
			match:   true,
			varName: "foo",
			expr:    "bar",
		},
		{
			input:   "foo  =  bar",
			match:   true,
			varName: "foo",
			expr:    "bar",
		},
		{
			input:   `result = upper "hello"`,
			match:   true,
			varName: "result",
			expr:    `upper "hello"`,
		},
		{
			input:   "foo_bar = value",
			match:   true,
			varName: "foo_bar",
			expr:    "value",
		},
		{
			input:   "_private = value",
			match:   true,
			varName: "_private",
			expr:    "value",
		},
		{
			input:   "x123 = value",
			match:   true,
			varName: "x123",
			expr:    "value",
		},
		{
			input: "123foo = value",
			match: false,
		},
		{
			input: "foo-bar = value",
			match: false,
		},
		{
			input: ":command",
			match: false,
		},
		{
			input: "just an expression",
			match: false,
		},
		{
			input: "foo",
			match: false,
		},
		{
			input: "=",
			match: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			match := assignmentPattern.FindStringSubmatch(tt.input)

			if tt.match {
				if match == nil {
					t.Errorf("expected match for %q", tt.input)
					return
				}
				if len(match) < 3 {
					t.Errorf("not enough capture groups for %q", tt.input)
					return
				}
				if match[1] != tt.varName {
					t.Errorf("varName = %q, want %q", match[1], tt.varName)
				}
				if match[2] != tt.expr {
					t.Errorf("expr = %q, want %q", match[2], tt.expr)
				}
			} else {
				if match != nil {
					t.Errorf("expected no match for %q, got %v", tt.input, match)
				}
			}
		})
	}
}

func TestREPL_ProcessLine(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		setup      func(*REPL)
		wantOutput string
		wantExit   bool
	}{
		{
			name:       "simple expression",
			input:      `upper "hello"`,
			wantOutput: "HELLO",
		},
		{
			name:  "variable assignment",
			input: `foo = upper "world"`,
			setup: func(r *REPL) {
				// No setup needed
			},
			wantOutput: "WORLD",
		},
		{
			name:  "variable lookup",
			input: "myvar",
			setup: func(r *REPL) {
				r.ctx.Set("myvar", "myvalue")
			},
			wantOutput: "myvalue",
		},
		{
			name:       "meta command help",
			input:      ":help",
			wantOutput: "shelm",
		},
		{
			name:     "meta command quit",
			input:    ":quit",
			wantExit: true,
		},
		{
			name:     "meta command exit",
			input:    ":exit",
			wantExit: true,
		},
		{
			name:       "math expression",
			input:      "add 10 20",
			wantOutput: "30",
		},
		{
			name:       "template block",
			input:      "{{ if true }}yes{{ end }}",
			wantOutput: "yes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			repl := NewREPL()
			repl.writer = &buf

			if tt.setup != nil {
				tt.setup(repl)
			}

			shouldExit := repl.processLine(tt.input)

			if shouldExit != tt.wantExit {
				t.Errorf("shouldExit = %v, want %v", shouldExit, tt.wantExit)
			}

			output := buf.String()
			if tt.wantOutput != "" && !strings.Contains(output, tt.wantOutput) {
				t.Errorf("output should contain %q, got: %s", tt.wantOutput, output)
			}
		})
	}
}

func TestREPL_HandleAssignment(t *testing.T) {
	var buf bytes.Buffer
	repl := NewREPL()
	repl.writer = &buf

	// Test simple assignment
	repl.handleAssignment("foo", `upper "bar"`)

	if _, ok := repl.ctx.Get("foo"); !ok {
		t.Error("foo not set in context")
	}

	output := buf.String()
	if !strings.Contains(output, "BAR") {
		t.Errorf("output should contain 'BAR': %s", output)
	}

	// Test assignment with Values reference
	buf.Reset()
	repl.handleAssignment("result", "add 1 2")

	val, ok := repl.ctx.Get("result")
	if !ok {
		t.Error("result not set in context")
	}
	if val != int64(3) {
		t.Errorf("result = %v, want 3", val)
	}
}

func TestREPL_HandleExpression(t *testing.T) {
	var buf bytes.Buffer
	repl := NewREPL()
	repl.writer = &buf

	// Test simple expression
	repl.handleExpression(`upper "hello"`)
	if !strings.Contains(buf.String(), "HELLO") {
		t.Errorf("output should contain 'HELLO': %s", buf.String())
	}

	// Test expression with context
	buf.Reset()
	repl.ctx.Set("name", "world")
	repl.handleExpression(".Values.name")
	if !strings.Contains(buf.String(), "world") {
		t.Errorf("output should contain 'world': %s", buf.String())
	}

	// Test error
	buf.Reset()
	repl.handleExpression("{{ invalid")
	if !strings.Contains(buf.String(), "ERROR") {
		t.Errorf("output should contain 'ERROR': %s", buf.String())
	}
}

func TestREPL_Context(t *testing.T) {
	repl := NewREPL()

	ctx := repl.Context()
	if ctx == nil {
		t.Error("Context() returned nil")
	}

	// Verify it's the same context
	ctx.Set("test", "value")
	if v, _ := repl.ctx.Get("test"); v != "value" {
		t.Error("Context() should return the same context instance")
	}
}

func TestNewREPL(t *testing.T) {
	repl := NewREPL()

	if repl == nil {
		t.Fatal("NewREPL returned nil")
	}
	if repl.ctx == nil {
		t.Error("ctx is nil")
	}
	if repl.eval == nil {
		t.Error("eval is nil")
	}
	if repl.cmdHandler == nil {
		t.Error("cmdHandler is nil")
	}
	if repl.prompt != DefaultPrompt {
		t.Errorf("prompt = %q, want %q", repl.prompt, DefaultPrompt)
	}
	if repl.completer == nil {
		t.Error("completer is nil")
	}
	if repl.writer == nil {
		t.Error("writer is nil")
	}
}

func TestREPL_Integration(t *testing.T) {
	var buf bytes.Buffer
	repl := NewREPL()
	repl.writer = &buf

	// Simulate a session
	steps := []struct {
		input    string
		contains string
	}{
		{`name = "myapp"`, "myapp"},
		{`port = 8080`, "8080"},
		{"name", "myapp"},
		{":values", "myapp"},
		{`:set version "1.0.0"`, "1.0.0"},
		{"version", "1.0.0"},
		{`printf "%s:%d" .Values.name .Values.port`, "myapp:8080"},
	}

	for _, step := range steps {
		buf.Reset()
		repl.processLine(step.input)

		output := buf.String()
		if !strings.Contains(output, step.contains) {
			t.Errorf("step %q: output should contain %q, got: %s", step.input, step.contains, output)
		}
	}
}

func TestREPL_EmptyLine(t *testing.T) {
	var buf bytes.Buffer
	repl := NewREPL()
	repl.writer = &buf

	// Empty line should be handled in Run(), but processLine receives trimmed input
	// In the actual REPL, empty lines are skipped before processLine is called
	// Here we test that non-empty input produces output

	repl.processLine(`"test"`)
	if buf.Len() == 0 {
		t.Error("non-empty input should produce output")
	}
}

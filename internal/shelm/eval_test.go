package shelm

import (
	"strings"
	"testing"
)

func TestWrapExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "bare identifier",
			input:    "foo",
			expected: `{{ index .Values "foo" }}`,
		},
		{
			name:     "identifier with underscore",
			input:    "foo_bar",
			expected: `{{ index .Values "foo_bar" }}`,
		},
		{
			name:     "identifier with numbers",
			input:    "foo123",
			expected: `{{ index .Values "foo123" }}`,
		},
		{
			name:     "function call",
			input:    `upper "hello"`,
			expected: `{{ upper "hello" }}`,
		},
		{
			name:     "already wrapped",
			input:    `{{ upper "hello" }}`,
			expected: `{{ upper "hello" }}`,
		},
		{
			name:     "complex expression",
			input:    `.Values.foo.bar`,
			expected: `{{ .Values.foo.bar }}`,
		},
		{
			name:     "expression with spaces",
			input:    "  foo  ",
			expected: `{{ index .Values "foo" }}`,
		},
		{
			name:     "pipe expression",
			input:    `.Values.name | upper`,
			expected: `{{ .Values.name | upper }}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapExpression(tt.input)
			if result != tt.expected {
				t.Errorf("wrapExpression(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestWrapAssignment(t *testing.T) {
	tests := []struct {
		name     string
		varName  string
		expr     string
		contains string
	}{
		{
			name:     "simple expression",
			varName:  "foo",
			expr:     `upper "bar"`,
			contains: `shelmCapture "foo"`,
		},
		{
			name:     "wrapped expression",
			varName:  "result",
			expr:     `{{ add 1 2 }}`,
			contains: `shelmCapture "result"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapAssignment(tt.varName, tt.expr)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("wrapAssignment(%q, %q) = %q, should contain %q", tt.varName, tt.expr, result, tt.contains)
			}
		})
	}
}

func TestEvaluator_Eval(t *testing.T) {
	eval := NewEvaluator()

	tests := []struct {
		name     string
		expr     string
		setup    func(*Context)
		expected string
		wantErr  bool
	}{
		{
			name:     "simple string",
			expr:     `"hello"`,
			expected: "hello",
		},
		{
			name:     "upper function",
			expr:     `upper "hello"`,
			expected: "HELLO",
		},
		{
			name:     "add function",
			expr:     `add 2 3`,
			expected: "5",
		},
		{
			name:     "access Values",
			expr:     `.Values.name`,
			setup:    func(c *Context) { c.Set("name", "myapp") },
			expected: "myapp",
		},
		{
			name:     "nested Values",
			expr:     `.Values.config.port`,
			setup:    func(c *Context) { c.Set("config", map[string]any{"port": 8080}) },
			expected: "8080",
		},
		{
			name:     "bare identifier lookup",
			expr:     `foo`,
			setup:    func(c *Context) { c.Set("foo", "bar") },
			expected: "bar",
		},
		{
			name:     "default function",
			expr:     `default "fallback" .Values.missing`,
			expected: "fallback",
		},
		{
			name:     "list function",
			expr:     `list 1 2 3`,
			expected: "[1 2 3]",
		},
		{
			name:     "dict function",
			expr:     `dict "a" 1`,
			expected: "map[a:1]",
		},
		{
			name:     "printf function",
			expr:     `printf "%s:%d" "host" 8080`,
			expected: "host:8080",
		},
		{
			name:     "conditional true",
			expr:     `{{ if true }}yes{{ end }}`,
			expected: "yes",
		},
		{
			name:     "conditional false",
			expr:     `{{ if false }}yes{{ else }}no{{ end }}`,
			expected: "no",
		},
		{
			name:    "invalid template",
			expr:    `{{ invalid`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext()
			if tt.setup != nil {
				tt.setup(ctx)
			}

			result, err := eval.Eval(tt.expr, ctx)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			resultStr, ok := result.(string)
			if !ok {
				t.Errorf("result is not a string: %T", result)
				return
			}

			if resultStr != tt.expected {
				t.Errorf("Eval(%q) = %q, want %q", tt.expr, resultStr, tt.expected)
			}
		})
	}
}

func TestEvaluator_EvalAssignment(t *testing.T) {
	eval := NewEvaluator()

	tests := []struct {
		name        string
		varName     string
		expr        string
		setup       func(*Context)
		expectedVal any
		wantErr     bool
	}{
		{
			name:        "simple string",
			varName:     "result",
			expr:        `upper "hello"`,
			expectedVal: "HELLO",
		},
		{
			name:        "number",
			varName:     "sum",
			expr:        `add 10 20`,
			expectedVal: int64(30),
		},
		{
			name:        "fromYaml",
			varName:     "data",
			expr:        `fromYaml "foo: bar"`,
			expectedVal: map[string]any{"foo": "bar"},
		},
		{
			name:    "invalid expression",
			varName: "x",
			expr:    `{{ invalid`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext()
			if tt.setup != nil {
				tt.setup(ctx)
			}

			result, err := eval.EvalAssignment(tt.varName, tt.expr, ctx)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check that value was stored in context
			stored, ok := ctx.Get(tt.varName)
			if !ok {
				t.Errorf("value not stored in context")
				return
			}

			// For maps, compare string representation
			switch expected := tt.expectedVal.(type) {
			case string:
				if result != expected {
					t.Errorf("result = %v, want %v", result, expected)
				}
			case int64:
				if result != expected {
					t.Errorf("result = %v, want %v", result, expected)
				}
			case map[string]any:
				resultMap, ok := result.(map[string]any)
				if !ok {
					t.Errorf("result is not a map: %T", result)
					return
				}
				for k, v := range expected {
					if resultMap[k] != v {
						t.Errorf("result[%q] = %v, want %v", k, resultMap[k], v)
					}
				}
			}

			_ = stored // Value verified by checking context
		})
	}
}

func TestEvaluator_SprigFunctions(t *testing.T) {
	eval := NewEvaluator()
	ctx := NewContext()

	// Test various Sprig functions
	tests := []struct {
		expr     string
		expected string
	}{
		{`lower "HELLO"`, "hello"},
		{`upper "hello"`, "HELLO"},
		{`title "hello world"`, "Hello World"},
		{`trim "  hello  "`, "hello"},
		{`trimPrefix "hello" "helloworld"`, "world"},
		{`trimSuffix "world" "helloworld"`, "hello"},
		{`contains "ell" "hello"`, "true"},
		{`hasPrefix "hello" "helloworld"`, "true"},
		{`hasSuffix "world" "helloworld"`, "true"},
		{`replace "o" "0" "hello"`, "hell0"},
		{`repeat 3 "ab"`, "ababab"},
		{`substr 0 5 "hello world"`, "hello"},
		{`add 1 2`, "3"},
		{`sub 5 3`, "2"},
		{`mul 3 4`, "12"},
		{`div 10 2`, "5"},
		{`mod 10 3`, "1"},
		{`max 1 5 3`, "5"},
		{`min 1 5 3`, "1"},
		{`len "hello"`, "5"},
		{`default "default" ""`, "default"},
		{`default "default" "value"`, "value"},
		{`empty ""`, "true"},
		{`empty "value"`, "false"},
		{`coalesce "" "" "first"`, "first"},
		{`ternary "yes" "no" true`, "yes"},
		{`ternary "yes" "no" false`, "no"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := eval.Eval(tt.expr, ctx)
			if err != nil {
				t.Errorf("Eval(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("Eval(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestEvaluator_ShelmFunctions(t *testing.T) {
	eval := NewEvaluator()
	ctx := NewContext()

	// Test fromYaml
	result, err := eval.Eval(`fromYaml "foo: bar\nbaz: 123"`, ctx)
	if err != nil {
		t.Errorf("fromYaml error: %v", err)
	}
	if !strings.Contains(result.(string), "foo") {
		t.Errorf("fromYaml result should contain 'foo': %v", result)
	}

	// Test toYaml
	ctx.Set("data", map[string]any{"key": "value"})
	result, err = eval.Eval(`toYaml .Values.data`, ctx)
	if err != nil {
		t.Errorf("toYaml error: %v", err)
	}
	if !strings.Contains(result.(string), "key: value") {
		t.Errorf("toYaml result should contain 'key: value': %v", result)
	}

	// Test fromJson
	result, err = eval.Eval(`fromJson "{\"a\": 1}"`, ctx)
	if err != nil {
		t.Errorf("fromJson error: %v", err)
	}

	// Test toJson
	result, err = eval.Eval(`toJson .Values.data`, ctx)
	if err != nil {
		t.Errorf("toJson error: %v", err)
	}
	if !strings.Contains(result.(string), `"key":"value"`) {
		t.Errorf("toJson result should contain key:value: %v", result)
	}
}

func TestIdentifierPattern(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"foo", true},
		{"Foo", true},
		{"_foo", true},
		{"foo_bar", true},
		{"foo123", true},
		{"_", true},
		{"a", true},
		{"123", false},
		{"123foo", false},
		{"foo-bar", false},
		{"foo.bar", false},
		{"foo bar", false},
		{"", false},
		{" ", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := identifierPattern.MatchString(tt.input)
			if result != tt.expected {
				t.Errorf("identifierPattern.MatchString(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

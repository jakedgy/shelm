package shelm

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestCompleter_MetaCommands(t *testing.T) {
	ctx := NewContext()
	funcNames := []string{"upper", "lower", "add"}
	c := NewCompleter(ctx, funcNames)

	tests := []struct {
		name       string
		line       string
		wantPrefix string
		wantMatch  []string
	}{
		{
			name:       "complete from colon",
			line:       ":",
			wantPrefix: ":",
			wantMatch:  []string{":exit ", ":help ", ":load ", ":quit ", ":save ", ":set ", ":template ", ":unset ", ":values "},
		},
		{
			name:       "complete :h",
			line:       ":h",
			wantPrefix: ":h",
			wantMatch:  []string{":help "},
		},
		{
			name:       "complete :s",
			line:       ":s",
			wantPrefix: ":s",
			wantMatch:  []string{":save ", ":set "},
		},
		{
			name:       "complete :q",
			line:       ":q",
			wantPrefix: ":q",
			wantMatch:  []string{":quit "},
		},
		{
			name:       "no match",
			line:       ":xyz",
			wantPrefix: ":xyz",
			wantMatch:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix, completions := c.getCompletions(tt.line)
			if prefix != tt.wantPrefix {
				t.Errorf("prefix = %q, want %q", prefix, tt.wantPrefix)
			}
			if !stringSliceEqual(completions, tt.wantMatch) {
				t.Errorf("completions = %v, want %v", completions, tt.wantMatch)
			}
		})
	}
}

func TestCompleter_Functions(t *testing.T) {
	ctx := NewContext()
	funcNames := []string{"upper", "lower", "add", "sub", "list", "dict", "upper2"}
	c := NewCompleter(ctx, funcNames)

	tests := []struct {
		name       string
		line       string
		wantPrefix string
		wantMatch  []string
	}{
		{
			name:       "complete up",
			line:       "up",
			wantPrefix: "up",
			wantMatch:  []string{"upper ", "upper2 "},
		},
		{
			name:       "complete after {{",
			line:       "{{ up",
			wantPrefix: "up",
			wantMatch:  []string{"upper ", "upper2 "},
		},
		{
			name:       "complete after pipe",
			line:       `{{ "hello" | up`,
			wantPrefix: "up",
			wantMatch:  []string{"upper ", "upper2 "},
		},
		{
			name:       "complete l",
			line:       "l",
			wantPrefix: "l",
			wantMatch:  []string{"list ", "lower "},
		},
		{
			name:       "complete all from empty",
			line:       "",
			wantPrefix: "",
			wantMatch:  []string{"add ", "dict ", "list ", "lower ", "sub ", "upper ", "upper2 "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix, completions := c.getCompletions(tt.line)
			if prefix != tt.wantPrefix {
				t.Errorf("prefix = %q, want %q", prefix, tt.wantPrefix)
			}
			if !stringSliceEqual(completions, tt.wantMatch) {
				t.Errorf("completions = %v, want %v", completions, tt.wantMatch)
			}
		})
	}
}

func TestCompleter_Variables(t *testing.T) {
	ctx := NewContext()
	ctx.Set("name", "myapp")
	ctx.Set("namespace", "default")
	ctx.Set("port", 8080)
	ctx.Set("config", map[string]any{
		"debug":   true,
		"timeout": 30,
		"nested": map[string]any{
			"key1": "val1",
			"key2": "val2",
		},
	})

	funcNames := []string{"upper"}
	c := NewCompleter(ctx, funcNames)

	tests := []struct {
		name       string
		line       string
		wantPrefix string
		wantMatch  []string
	}{
		{
			name:       "complete .Values.na",
			line:       ".Values.na",
			wantPrefix: ".Values.na",
			wantMatch:  []string{".Values.name", ".Values.namespace"},
		},
		{
			name:       "complete .Values.config.",
			line:       ".Values.config.",
			wantPrefix: ".Values.config.",
			wantMatch:  []string{".Values.config.debug", ".Values.config.nested", ".Values.config.timeout"},
		},
		{
			name:       "complete nested .Values.config.nested.",
			line:       ".Values.config.nested.",
			wantPrefix: ".Values.config.nested.",
			wantMatch:  []string{".Values.config.nested.key1", ".Values.config.nested.key2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix, completions, ok := c.tryVariableCompletion(tt.line)
			if !ok {
				t.Error("expected variable completion")
				return
			}
			if prefix != tt.wantPrefix {
				t.Errorf("prefix = %q, want %q", prefix, tt.wantPrefix)
			}
			sort.Strings(completions)
			sort.Strings(tt.wantMatch)
			if !stringSliceEqual(completions, tt.wantMatch) {
				t.Errorf("completions = %v, want %v", completions, tt.wantMatch)
			}
		})
	}
}

func TestCompleter_VariableNames(t *testing.T) {
	ctx := NewContext()
	ctx.Set("name", "myapp")
	ctx.Set("namespace", "default")
	ctx.Set("port", 8080)

	c := NewCompleter(ctx, []string{})

	tests := []struct {
		name       string
		prefix     string
		wantPrefix string
		wantMatch  []string
	}{
		{
			name:       "complete n",
			prefix:     "n",
			wantPrefix: "n",
			wantMatch:  []string{"name", "namespace"},
		},
		{
			name:       "complete empty",
			prefix:     "",
			wantPrefix: "",
			wantMatch:  []string{"name", "namespace", "port"},
		},
		{
			name:       "complete p",
			prefix:     "p",
			wantPrefix: "p",
			wantMatch:  []string{"port"},
		},
		{
			name:       "no match",
			prefix:     "xyz",
			wantPrefix: "xyz",
			wantMatch:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix, completions := c.completeVariableNames(tt.prefix)
			if prefix != tt.wantPrefix {
				t.Errorf("prefix = %q, want %q", prefix, tt.wantPrefix)
			}
			sort.Strings(completions)
			sort.Strings(tt.wantMatch)
			if !stringSliceEqual(completions, tt.wantMatch) {
				t.Errorf("completions = %v, want %v", completions, tt.wantMatch)
			}
		})
	}
}

func TestCompleter_FilePaths(t *testing.T) {
	// Create temp directory with test files
	tmpDir := t.TempDir()

	// Create test files
	_ = os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte("test"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "config.json"), []byte("test"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "values.yaml"), []byte("test"), 0644)
	_ = os.MkdirAll(filepath.Join(tmpDir, "templates"), 0755)

	ctx := NewContext()
	c := NewCompleter(ctx, []string{})

	// Test file completion
	_, completions := c.completeFilePath(filepath.Join(tmpDir, "config"))

	if len(completions) != 2 {
		t.Errorf("expected 2 completions (config.yaml, config.json), got %d: %v", len(completions), completions)
	}

	// Verify both config files are in completions (with trailing space)
	hasYaml := false
	hasJson := false
	for _, comp := range completions {
		if strings.Contains(comp, "config.yaml ") {
			hasYaml = true
		}
		if strings.Contains(comp, "config.json ") {
			hasJson = true
		}
	}
	if !hasYaml || !hasJson {
		t.Errorf("expected config.yaml and config.json (with trailing space), got %v", completions)
	}

	// Test that directories have trailing separator
	_, completions = c.completeFilePath(filepath.Join(tmpDir, "temp"))
	for _, comp := range completions {
		if strings.Contains(comp, "templates") && !strings.HasSuffix(comp, string(os.PathSeparator)) {
			t.Errorf("directory completion should end with separator: %s", comp)
		}
	}
}

func TestCompleter_FilePathInQuotes(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "test.yaml"), []byte("test"), 0644)

	ctx := NewContext()
	c := NewCompleter(ctx, []string{})

	// Test completion inside quoted string
	line := `shelmReadFile "` + tmpDir + `/te`
	_, completions, ok := c.tryFilePathCompletion(line)

	if !ok {
		t.Error("expected file path completion inside quotes")
	}

	if len(completions) == 0 {
		t.Error("expected completions for file in quotes")
	}
}

func TestCompleter_LoadCommand(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "values.yaml"), []byte("test"), 0644)

	ctx := NewContext()
	c := NewCompleter(ctx, []string{})

	// Test :load with partial path
	line := ":load " + filepath.Join(tmpDir, "val")
	_, completions := c.completeMetaCommand(line)

	if len(completions) == 0 {
		t.Error("expected file completions for :load command")
	}
}

func TestCompleter_SetCommand_VariableCompletion(t *testing.T) {
	ctx := NewContext()
	ctx.Set("name", "myapp")
	ctx.Set("port", 8080)

	c := NewCompleter(ctx, []string{})

	// Test :unset with partial variable name
	_, completions := c.completeMetaCommand(":unset n")

	if len(completions) != 1 || completions[0] != "name" {
		t.Errorf("expected [name], got %v", completions)
	}
}

func TestCompleter_SaveCommand(t *testing.T) {
	ctx := NewContext()
	ctx.Set("myconfig", map[string]any{"key": "value"})
	ctx.Set("mydata", "test")

	c := NewCompleter(ctx, []string{})

	// Test :save with partial variable name (first arg)
	_, completions := c.completeMetaCommand(":save my")

	sort.Strings(completions)
	expected := []string{"myconfig", "mydata"}
	if !stringSliceEqual(completions, expected) {
		t.Errorf("expected %v, got %v", expected, completions)
	}
}

func TestCompleter_WordCompleter_Interface(t *testing.T) {
	ctx := NewContext()
	ctx.Set("foo", "bar")
	funcNames := []string{"upper", "lower"}
	c := NewCompleter(ctx, funcNames)

	// Test the Complete method (WordCompleter interface)
	head, completions, tail := c.Complete("up", 2)

	if head != "" {
		t.Errorf("head = %q, want empty", head)
	}
	if tail != "" {
		t.Errorf("tail = %q, want empty", tail)
	}
	if len(completions) != 1 || completions[0] != "upper " {
		t.Errorf("completions = %v, want [upper ]", completions)
	}

	// Test with cursor in middle
	head, _, tail = c.Complete("up world", 2)
	_ = head // verify head is assigned
	if tail != " world" {
		t.Errorf("tail = %q, want ' world'", tail)
	}
}

func TestCompleter_ValuesWithDot(t *testing.T) {
	ctx := NewContext()
	c := NewCompleter(ctx, []string{})

	// Test .Values completion
	prefix, completions, ok := c.tryVariableCompletion(".Values")
	if !ok {
		t.Error("expected variable completion for .Values")
	}
	if prefix != ".Values" {
		t.Errorf("prefix = %q, want .Values", prefix)
	}
	if len(completions) != 1 || completions[0] != ".Values." {
		t.Errorf("completions = %v, want [.Values.]", completions)
	}
}

func TestExtractFunctionPrefix(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{"upper", "upper"},
		{"{{ upper", "upper"},
		{"{{ .name | upper", "upper"},
		{"{{ .name | upp", "upp"},
		{"", ""},
		{"{{ ", ""},
		{"{{", ""},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got := extractFunctionPrefix(tt.line)
			if got != tt.want {
				t.Errorf("extractFunctionPrefix(%q) = %q, want %q", tt.line, got, tt.want)
			}
		})
	}
}

func TestGetFunctionNames(t *testing.T) {
	names := GetFunctionNames()

	if len(names) == 0 {
		t.Error("expected function names")
	}

	// Check some known functions exist
	expected := map[string]bool{
		"upper":         false,
		"lower":         false,
		"add":           false,
		"fromYaml":      false,
		"toYaml":        false,
		"shelmReadFile": false,
	}

	for _, name := range names {
		if _, ok := expected[name]; ok {
			expected[name] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("expected function %q not found", name)
		}
	}
}

func TestREPL_CompleterIntegration(t *testing.T) {
	repl := NewREPL()

	// Verify completer is initialized
	if repl.completer == nil {
		t.Error("completer should be initialized")
	}

	// Add some values and verify completion works
	repl.ctx.Set("testvar", "value")

	_, completions := repl.completer.completeVariableNames("test")
	if len(completions) != 1 || completions[0] != "testvar" {
		t.Errorf("expected [testvar], got %v", completions)
	}
}

// Helper function
func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

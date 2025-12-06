package shelm

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Completer provides tab completion for the shelm REPL.
type Completer struct {
	ctx          *Context
	funcNames    []string
	metaCommands []string
}

// NewCompleter creates a new Completer with the given context.
func NewCompleter(ctx *Context, funcNames []string) *Completer {
	metaCommands := []string{
		":exit", ":help", ":load", ":quit",
		":save", ":set", ":template", ":unset", ":values",
	}

	// Sort function names for consistent completion order
	sortedFuncs := make([]string, len(funcNames))
	copy(sortedFuncs, funcNames)
	sort.Strings(sortedFuncs)

	return &Completer{
		ctx:          ctx,
		funcNames:    sortedFuncs,
		metaCommands: metaCommands,
	}
}

// Complete implements liner.WordCompleter interface.
// It returns (head, completions, tail) where:
// - head: the part of line before the word being completed
// - completions: list of possible completions
// - tail: the part of line after the word being completed
func (c *Completer) Complete(line string, pos int) (string, []string, string) {
	// Get the portion of the line up to cursor
	if pos > len(line) {
		pos = len(line)
	}
	head := line[:pos]
	tail := line[pos:]

	// Determine completion context and get completions
	prefix, completions := c.getCompletions(head)

	// Calculate the head (everything before the word being completed)
	headEnd := len(head) - len(prefix)
	if headEnd < 0 {
		headEnd = 0
	}

	return head[:headEnd], completions, tail
}

// getCompletions determines the completion context and returns matching completions.
// Returns (prefix, completions) where prefix is the partial word being completed.
func (c *Completer) getCompletions(line string) (string, []string) {
	// 1. Meta-command completion (line starts with :)
	if strings.HasPrefix(line, ":") {
		return c.completeMetaCommand(line)
	}

	// 2. Check if we're completing a file path (in quotes)
	if prefix, completions, ok := c.tryFilePathCompletion(line); ok {
		return prefix, completions
	}

	// 3. Variable completion (check for .Values. or . prefix)
	if prefix, completions, ok := c.tryVariableCompletion(line); ok {
		return prefix, completions
	}

	// 4. Function completion (anywhere in template context)
	return c.completeFunctions(line)
}

// completeMetaCommand handles completion for meta-commands.
func (c *Completer) completeMetaCommand(line string) (string, []string) {
	// Check if we need file path completion for :load, :save, :template
	parts := strings.Fields(line)
	if len(parts) >= 1 {
		cmd := strings.ToLower(parts[0])

		// For :load, :save, :template - complete file paths
		if cmd == ":load" || cmd == ":save" || cmd == ":template" {
			if len(parts) >= 2 || strings.HasSuffix(line, " ") {
				// :save <var> <file> - second arg is variable, third is file
				if cmd == ":save" {
					if len(parts) == 2 && !strings.HasSuffix(line, " ") {
						// Still typing variable name
						return c.completeVariableNames(parts[1])
					}
					if len(parts) == 2 && strings.HasSuffix(line, " ") {
						// Space after :save, need variable name
						return c.completeVariableNames("")
					}
					if len(parts) >= 3 || (len(parts) == 2 && strings.HasSuffix(line, " ")) {
						// Need file path
						pathPrefix := ""
						if len(parts) >= 3 && !strings.HasSuffix(line, " ") {
							pathPrefix = parts[len(parts)-1]
						}
						return c.completeFilePath(pathPrefix)
					}
				}

				// :load <file> or :template <file>
				pathPrefix := ""
				if len(parts) >= 2 && !strings.HasSuffix(line, " ") {
					pathPrefix = parts[len(parts)-1]
				}
				return c.completeFilePath(pathPrefix)
			}
		}

		// For :set and :unset - complete variable names
		if (cmd == ":set" || cmd == ":unset") && (len(parts) >= 2 || strings.HasSuffix(line, " ")) {
			varPrefix := ""
			if len(parts) >= 2 && !strings.HasSuffix(line, " ") {
				varPrefix = parts[1]
			}
			return c.completeVariableNames(varPrefix)
		}
	}

	// Complete the command itself
	var completions []string
	for _, cmd := range c.metaCommands {
		if strings.HasPrefix(cmd, line) {
			completions = append(completions, cmd+" ") // Add trailing space
		}
	}
	return line, completions
}

// tryFilePathCompletion checks if the cursor is inside a quoted string
// and completes file paths if so.
func (c *Completer) tryFilePathCompletion(line string) (string, []string, bool) {
	// Find if we're inside a quoted string
	// Count quotes to determine if inside a string
	inDoubleQuote := false
	quoteStart := -1

	for i := 0; i < len(line); i++ {
		ch := line[i]
		if ch == '"' && (i == 0 || line[i-1] != '\\') {
			if !inDoubleQuote {
				inDoubleQuote = true
				quoteStart = i + 1
			} else {
				inDoubleQuote = false
			}
		}
	}

	if inDoubleQuote && quoteStart >= 0 {
		pathPrefix := line[quoteStart:]
		prefix, completions := c.completeFilePath(pathPrefix)
		return prefix, completions, true
	}

	return "", nil, false
}

// tryVariableCompletion checks if we're completing a variable access.
func (c *Completer) tryVariableCompletion(line string) (string, []string, bool) {
	// Check for .Values.xxx pattern
	if idx := strings.LastIndex(line, ".Values."); idx != -1 {
		after := line[idx+8:] // After ".Values."
		return c.completeNestedVariable(after, c.ctx.Values)
	}

	// Check for .Values (without trailing dot)
	if strings.HasSuffix(line, ".Values") {
		return ".Values", []string{".Values."}, true
	}

	return "", nil, false
}

// completeNestedVariable completes nested keys in a map.
func (c *Completer) completeNestedVariable(path string, data map[string]any) (string, []string, bool) {
	parts := strings.Split(path, ".")

	// Navigate to the right level
	current := data
	for i := 0; i < len(parts)-1; i++ {
		key := parts[i]
		if key == "" {
			continue
		}
		val, ok := current[key]
		if !ok {
			return ".Values." + path, nil, true
		}
		nested, ok := val.(map[string]any)
		if !ok {
			return ".Values." + path, nil, true
		}
		current = nested
	}

	// Complete the last part
	prefix := ""
	if len(parts) > 0 {
		prefix = parts[len(parts)-1]
	}

	var completions []string
	for key := range current {
		if strings.HasPrefix(key, prefix) {
			// Build full path
			fullPath := strings.Join(parts[:len(parts)-1], ".")
			if fullPath != "" {
				fullPath += "."
			}
			completions = append(completions, ".Values."+fullPath+key)
		}
	}
	sort.Strings(completions)

	return ".Values." + path, completions, true
}

// completeVariableNames completes top-level variable names.
func (c *Completer) completeVariableNames(prefix string) (string, []string) {
	var completions []string
	for key := range c.ctx.Values {
		if strings.HasPrefix(key, prefix) {
			completions = append(completions, key)
		}
	}
	sort.Strings(completions)
	return prefix, completions
}

// completeFunctions completes function names.
func (c *Completer) completeFunctions(line string) (string, []string) {
	// Find the word being typed (after {{ or |)
	prefix := extractFunctionPrefix(line)

	var completions []string
	for _, fn := range c.funcNames {
		if strings.HasPrefix(fn, prefix) {
			completions = append(completions, fn+" ") // Add trailing space
		}
	}

	return prefix, completions
}

// completeFilePath completes file system paths.
func (c *Completer) completeFilePath(prefix string) (string, []string) {
	if prefix == "" {
		prefix = "."
	}

	dir := filepath.Dir(prefix)
	base := filepath.Base(prefix)

	// Handle case where prefix ends with /
	if strings.HasSuffix(prefix, string(os.PathSeparator)) || prefix == "." {
		dir = prefix
		base = ""
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return prefix, nil
	}

	var completions []string
	for _, entry := range entries {
		name := entry.Name()
		// Skip hidden files unless explicitly requested
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(base, ".") {
			continue
		}
		if strings.HasPrefix(name, base) {
			fullPath := filepath.Join(dir, name)
			if entry.IsDir() {
				fullPath += string(os.PathSeparator) // Directories get trailing slash (no space)
			} else {
				fullPath += " " // Files get trailing space
			}
			completions = append(completions, fullPath)
		}
	}

	return prefix, completions
}

// extractFunctionPrefix extracts the function name prefix being typed.
// It looks for the word after {{ or | or at start of line.
func extractFunctionPrefix(line string) string {
	// Find the last position that could start a function name
	// This could be after {{, after |, after whitespace, or at start
	lastBrace := strings.LastIndex(line, "{{")
	lastPipe := strings.LastIndex(line, "|")

	start := 0
	if lastBrace > start {
		start = lastBrace + 2
	}
	if lastPipe > start {
		start = lastPipe + 1
	}

	// Skip whitespace
	for start < len(line) && line[start] == ' ' {
		start++
	}

	// If we're at end of line after {{ or |, return empty
	if start >= len(line) {
		return ""
	}

	// Extract the word
	end := len(line)
	word := line[start:end]

	// Find the start of the last word (in case of multiple words)
	lastSpace := strings.LastIndex(word, " ")
	if lastSpace >= 0 {
		word = word[lastSpace+1:]
	}

	// Don't return template delimiters as function prefix
	if word == "{{" || word == "}}" {
		return ""
	}

	return word
}

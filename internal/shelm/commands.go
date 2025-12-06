package shelm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const helpText = `shelm - Sprig + text/template REPL

Syntax:
  <expression>           Evaluate a template expression
  name = <expression>    Assign result to .Values

Meta-commands:
  :help                  Show this help
  :values                Display all values (.Values)
  :set <name> <expr>     Set a value
  :unset <name>          Delete a value
  :load <file> [name]    Load YAML/JSON file into .Values (or .Values.<name>)
  :save <name> <file>    Save .Values.<name> as YAML to file
  :template <file>       Execute template file
  :quit / :exit          Exit the REPL

Built-in functions:
  shelmReadFile(path)    Read file contents as string
  shelmWriteFile(p, d)   Write data to file
  fromYaml(str)          Parse YAML string to value
  toYaml(val)            Convert value to YAML string
  fromJson(str)          Parse JSON string to value
  toJson(val)            Convert value to JSON string
  toPrettyJson(val)      Convert value to pretty JSON string

All Sprig functions are available: https://masterminds.github.io/sprig/
`

// CommandResult represents the result of a meta-command.
type CommandResult struct {
	Output   string
	ShouldExit bool
	Error    error
}

// CommandHandler handles meta-commands.
type CommandHandler struct {
	ctx  *Context
	eval *Evaluator
}

// NewCommandHandler creates a new command handler.
func NewCommandHandler(ctx *Context, eval *Evaluator) *CommandHandler {
	return &CommandHandler{
		ctx:  ctx,
		eval: eval,
	}
}

// Handle processes a meta-command. Returns true if REPL should exit.
func (h *CommandHandler) Handle(line string) CommandResult {
	line = strings.TrimPrefix(line, ":")
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return CommandResult{Error: fmt.Errorf("empty command")}
	}

	cmd := strings.ToLower(parts[0])
	args := parts[1:]

	switch cmd {
	case "help":
		return CommandResult{Output: helpText}

	case "values":
		return CommandResult{Output: FormatValues(h.ctx.Values)}

	case "set":
		return h.handleSet(args)

	case "unset":
		return h.handleUnset(args)

	case "load":
		return h.handleLoad(args)

	case "save":
		return h.handleSave(args)

	case "template":
		return h.handleTemplate(args)

	case "quit", "exit":
		return CommandResult{ShouldExit: true}

	default:
		return CommandResult{Error: fmt.Errorf("unknown command: %s", cmd)}
	}
}

func (h *CommandHandler) handleSet(args []string) CommandResult {
	if len(args) < 2 {
		return CommandResult{Error: fmt.Errorf("usage: :set <name> <expression>")}
	}

	name := args[0]
	expr := strings.Join(args[1:], " ")

	result, err := h.eval.EvalAssignment(name, expr, h.ctx)
	if err != nil {
		return CommandResult{Error: err}
	}

	return CommandResult{Output: FormatOutput(result)}
}

func (h *CommandHandler) handleUnset(args []string) CommandResult {
	if len(args) < 1 {
		return CommandResult{Error: fmt.Errorf("usage: :unset <name>")}
	}

	name := args[0]
	if _, ok := h.ctx.Get(name); !ok {
		return CommandResult{Error: fmt.Errorf("variable not found: %s", name)}
	}

	h.ctx.Delete(name)
	return CommandResult{Output: fmt.Sprintf("unset %s", name)}
}

func (h *CommandHandler) handleLoad(args []string) CommandResult {
	if len(args) < 1 {
		return CommandResult{Error: fmt.Errorf("usage: :load <file> [var]")}
	}

	path := args[0]
	varName := ""
	if len(args) > 1 {
		varName = args[1]
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return CommandResult{Error: fmt.Errorf("failed to read file: %w", err)}
	}

	// Try to parse as YAML/JSON
	var parsed any
	ext := strings.ToLower(filepath.Ext(path))

	if ext == ".json" {
		if err := json.Unmarshal(data, &parsed); err != nil {
			return CommandResult{Error: fmt.Errorf("failed to parse JSON: %w", err)}
		}
	} else {
		// Default to YAML (which also handles JSON)
		if err := yaml.Unmarshal(data, &parsed); err != nil {
			return CommandResult{Error: fmt.Errorf("failed to parse YAML: %w", err)}
		}
	}

	if varName != "" {
		// Store in specific variable
		h.ctx.Set(varName, parsed)
		return CommandResult{Output: fmt.Sprintf("loaded %s into %s", path, varName)}
	} else {
		// Merge into .Values (shallow merge)
		if m, ok := parsed.(map[string]any); ok {
			h.ctx.Merge(m)
			return CommandResult{Output: fmt.Sprintf("merged %s into .Values", path)}
		}
		return CommandResult{Error: fmt.Errorf("cannot merge non-map data without a variable name")}
	}
}

func (h *CommandHandler) handleSave(args []string) CommandResult {
	if len(args) < 2 {
		return CommandResult{Error: fmt.Errorf("usage: :save <var> <file>")}
	}

	varName := args[0]
	path := args[1]

	val, ok := h.ctx.Get(varName)
	if !ok {
		return CommandResult{Error: fmt.Errorf("variable not found: %s", varName)}
	}

	data, err := yaml.Marshal(val)
	if err != nil {
		return CommandResult{Error: fmt.Errorf("failed to marshal to YAML: %w", err)}
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return CommandResult{Error: fmt.Errorf("failed to write file: %w", err)}
	}

	return CommandResult{Output: fmt.Sprintf("saved %s to %s", varName, path)}
}

func (h *CommandHandler) handleTemplate(args []string) CommandResult {
	if len(args) < 1 {
		return CommandResult{Error: fmt.Errorf("usage: :template <file>")}
	}

	path := args[0]

	// Read and parse the template file
	data, err := os.ReadFile(path)
	if err != nil {
		return CommandResult{Error: fmt.Errorf("failed to read file: %w", err)}
	}

	result, err := h.eval.Eval(string(data), h.ctx)
	if err != nil {
		return CommandResult{Error: err}
	}

	return CommandResult{Output: FormatOutput(result)}
}

package shelm

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/peterh/liner"
)

const (
	DefaultPrompt = "shelm> "
	HistoryFile   = ".shelm_history"
)

// assignmentPattern matches: identifier = expression
// identifier: [A-Za-z_][A-Za-z0-9_]*
var assignmentPattern = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.+)$`)

// REPL represents the shelm REPL.
type REPL struct {
	ctx        *Context
	eval       *Evaluator
	cmdHandler *CommandHandler
	prompt     string
	liner      *liner.State
	writer     io.Writer
	completer  *Completer
}

// NewREPL creates a new REPL instance.
func NewREPL() *REPL {
	ctx := NewContext()
	eval := NewEvaluator()
	cmdHandler := NewCommandHandler(ctx, eval)

	// Get all function names for completion
	funcNames := GetFunctionNames()
	completer := NewCompleter(ctx, funcNames)

	return &REPL{
		ctx:        ctx,
		eval:       eval,
		cmdHandler: cmdHandler,
		prompt:     DefaultPrompt,
		writer:     os.Stdout,
		completer:  completer,
	}
}

// Run starts the REPL loop.
func (r *REPL) Run() {
	// Initialize liner
	line := liner.NewLiner()
	defer func() { _ = line.Close() }()
	r.liner = line

	// Set up tab completion using WordCompleter
	line.SetWordCompleter(r.completer.Complete)

	// Set completion style to print list (like bash)
	line.SetTabCompletionStyle(liner.TabPrints)

	// Enable Ctrl+C to abort current line
	line.SetCtrlCAborts(true)

	// Load history
	r.loadHistory()
	defer r.saveHistory()

	// Set up signal handling for clean exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	go func() {
		<-sigChan
		r.saveHistory()
		_ = line.Close()
		os.Exit(0)
	}()

	for {
		input, err := line.Prompt(r.prompt)
		if err != nil {
			if err == liner.ErrPromptAborted {
				// Ctrl+C pressed, continue to next prompt
				_, _ = fmt.Fprintln(r.writer)
				continue
			}
			if err == io.EOF {
				_, _ = fmt.Fprintln(r.writer)
				break
			}
			_, _ = fmt.Fprintf(r.writer, "ERROR: %v\n", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Add to history
		line.AppendHistory(input)

		shouldExit := r.processLine(input)
		if shouldExit {
			break
		}
	}
}

// processLine handles a single line of input.
// Returns true if the REPL should exit.
func (r *REPL) processLine(line string) bool {
	// 1. Check for meta-commands (starting with :)
	if strings.HasPrefix(line, ":") {
		result := r.cmdHandler.Handle(line)
		if result.Error != nil {
			_, _ = fmt.Fprintf(r.writer, "ERROR: %v\n", result.Error)
		} else if result.Output != "" {
			_, _ = fmt.Fprintln(r.writer, result.Output)
		}
		return result.ShouldExit
	}

	// 2. Check for assignment syntax
	if match := assignmentPattern.FindStringSubmatch(line); match != nil {
		name := match[1]
		expr := match[2]
		r.handleAssignment(name, expr)
		return false
	}

	// 3. Default: template expression
	r.handleExpression(line)
	return false
}

// handleAssignment processes an assignment: name = expression
func (r *REPL) handleAssignment(name, expr string) {
	result, err := r.eval.EvalAssignment(name, expr, r.ctx)
	if err != nil {
		_, _ = fmt.Fprintf(r.writer, "ERROR: %v\n", err)
		return
	}

	output := FormatOutput(result)
	if output != "" {
		_, _ = fmt.Fprintln(r.writer, output)
	}
}

// handleExpression evaluates a template expression and prints the result.
func (r *REPL) handleExpression(expr string) {
	result, err := r.eval.Eval(expr, r.ctx)
	if err != nil {
		_, _ = fmt.Fprintf(r.writer, "ERROR: %v\n", err)
		return
	}

	output := FormatOutput(result)
	if output != "" {
		_, _ = fmt.Fprintln(r.writer, output)
	}
}

// Context returns the REPL's context for testing.
func (r *REPL) Context() *Context {
	return r.ctx
}

// loadHistory loads command history from file.
func (r *REPL) loadHistory() {
	if r.liner == nil {
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	histPath := filepath.Join(homeDir, HistoryFile)
	if f, err := os.Open(histPath); err == nil {
		_, _ = r.liner.ReadHistory(f)
		_ = f.Close()
	}
}

// saveHistory saves command history to file.
func (r *REPL) saveHistory() {
	if r.liner == nil {
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	histPath := filepath.Join(homeDir, HistoryFile)
	if f, err := os.Create(histPath); err == nil {
		_, _ = r.liner.WriteHistory(f)
		_ = f.Close()
	}
}

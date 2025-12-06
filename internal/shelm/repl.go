package shelm

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"regexp"
	"strings"
)

const (
	DefaultPrompt = "shelm> "
)

// assignmentPattern matches: identifier = expression
// identifier: [A-Za-z_][A-Za-z0-9_]*
var assignmentPattern = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.+)$`)

// REPL represents the shelm REPL.
type REPL struct {
	ctx     *Context
	eval    *Evaluator
	cmdHandler *CommandHandler
	prompt  string
	reader  *bufio.Reader
	writer  io.Writer
}

// NewREPL creates a new REPL instance.
func NewREPL() *REPL {
	ctx := NewContext()
	eval := NewEvaluator()
	cmdHandler := NewCommandHandler(ctx, eval)

	return &REPL{
		ctx:        ctx,
		eval:       eval,
		cmdHandler: cmdHandler,
		prompt:     DefaultPrompt,
		reader:     bufio.NewReader(os.Stdin),
		writer:     os.Stdout,
	}
}

// Run starts the REPL loop.
func (r *REPL) Run() {
	// Set up Ctrl+C handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	go func() {
		for range sigChan {
			// On Ctrl+C, print newline and prompt
			_, _ = fmt.Fprintln(r.writer)
			_, _ = fmt.Fprint(r.writer, r.prompt)
		}
	}()

	for {
		_, _ = fmt.Fprint(r.writer, r.prompt)

		line, err := r.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				_, _ = fmt.Fprintln(r.writer)
				break
			}
			_, _ = fmt.Fprintf(r.writer, "ERROR: %v\n", err)
			continue
		}

		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		shouldExit := r.processLine(line)
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

package shelm

import (
	"bytes"
	"regexp"
	"strings"
	"text/template"
)

// identifierPattern matches a bare identifier that should be looked up in .Vars
var identifierPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// Evaluator handles template evaluation.
type Evaluator struct {
	baseFuncMap template.FuncMap
}

// NewEvaluator creates a new template evaluator.
func NewEvaluator() *Evaluator {
	return &Evaluator{
		baseFuncMap: BuildFuncMap(nil),
	}
}

// wrapExpression wraps an expression in {{ }} if not already a template.
// If the expression is a bare identifier, it wraps it as a .Vars lookup.
func wrapExpression(expr string) string {
	expr = strings.TrimSpace(expr)
	if strings.Contains(expr, "{{") {
		return expr
	}
	// If it's a bare identifier, treat it as a variable lookup
	if identifierPattern.MatchString(expr) {
		return "{{ index .Vars \"" + expr + "\" }}"
	}
	return "{{ " + expr + " }}"
}

// wrapAssignment wraps an expression to capture its value.
func wrapAssignment(name, expr string) string {
	expr = strings.TrimSpace(expr)
	// If already has {{ }}, extract the inner expression
	if strings.HasPrefix(expr, "{{") && strings.HasSuffix(expr, "}}") {
		inner := strings.TrimSuffix(strings.TrimPrefix(expr, "{{"), "}}")
		return "{{ shelmCapture \"" + name + "\" (" + strings.TrimSpace(inner) + ") }}"
	}
	return "{{ shelmCapture \"" + name + "\" (" + expr + ") }}"
}

// Eval evaluates a template expression against the given context.
// It returns the raw result (before formatting) and any error.
func (e *Evaluator) Eval(expr string, ctx *Context) (any, error) {
	wrapped := wrapExpression(expr)
	return e.evalTemplate(wrapped, ctx)
}

// EvalAssignment evaluates an assignment expression and stores the result.
// It returns the captured value and any error.
func (e *Evaluator) EvalAssignment(name, expr string, ctx *Context) (any, error) {
	var captured any

	// Create a capture function that stores the value
	captureFunc := func(n string, v any) any {
		if n == name {
			captured = v
			ctx.Set(n, v)
		}
		return v
	}

	// Build function map with capture function
	funcMap := BuildFuncMap(captureFunc)

	wrapped := wrapAssignment(name, expr)

	tmpl, err := template.New("assign").
		Option("missingkey=error").
		Funcs(funcMap).
		Parse(wrapped)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return nil, err
	}

	return captured, nil
}

// evalTemplate executes a template string and returns the string output.
func (e *Evaluator) evalTemplate(tmplStr string, ctx *Context) (any, error) {
	tmpl, err := template.New("eval").
		Option("missingkey=error").
		Funcs(e.baseFuncMap).
		Parse(tmplStr)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return nil, err
	}

	return buf.String(), nil
}

// EvalFile evaluates a template file against the given context.
func (e *Evaluator) EvalFile(path string, ctx *Context) (string, error) {
	tmpl, err := template.New("file").
		Option("missingkey=error").
		Funcs(e.baseFuncMap).
		ParseFiles(path)
	if err != nil {
		return "", err
	}

	// ParseFiles creates a template with the base name of the file
	tmpl = tmpl.Lookup(path[strings.LastIndex(path, "/")+1:])
	if tmpl == nil {
		// Fallback to the first template
		tmpl = tmpl.Lookup("")
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", err
	}

	return buf.String(), nil
}

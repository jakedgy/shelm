# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

shelm is a REPL for evaluating Go `text/template` expressions with Sprig functions and a persistent variable environment. It's designed for testing/debugging Go templates, Sprig functions, and Helm chart templates.

## Build and Test Commands

```bash
# Build
go build -o shelm .

# Run tests with race detection and coverage
go test -v -race -coverprofile=coverage.out ./...

# Lint
golangci-lint run

# Install globally
go install github.com/jakedgy/shelm@latest
```

## Architecture

```
main.go                      # Entry point, creates REPL
internal/shelm/
├── repl.go                  # Main REPL loop, input parsing
├── eval.go                  # Template evaluation (Eval, EvalAssignment, EvalFile)
├── context.go               # Execution context with .Values map
├── commands.go              # Meta-commands (:load, :save, :set, :template, etc.)
├── funcs.go                 # Sprig + custom functions (fromYaml, toYaml, etc.)
└── output.go                # Result formatting (YAML for maps/slices, raw for strings)
```

**Input Types**: The REPL parses input into three categories:
1. Meta-commands (`:` prefix) - handled by CommandHandler
2. Variable assignments (`name = expression`) - uses `shelmCapture` to store results
3. Template expressions - evaluated via Go's text/template

**Context**: The `.Values` map persists across REPL iterations, matching Helm's template context structure.

## Key Dependencies

- `github.com/Masterminds/sprig/v3` - Template function library
- `gopkg.in/yaml.v3` - YAML parsing/marshaling

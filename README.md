# shelm

A REPL for evaluating Go `text/template` expressions with [Sprig](https://masterminds.github.io/sprig/) functions and a persistent variable environment.

## Installation

```bash
go install github.com/jakedgy/shelm@latest
```

Or build from source:

```bash
git clone https://github.com/jakedgy/shelm.git
cd shelm
go build -o shelm .
```

## Usage

Start the REPL:

```bash
shelm
```

### Expression Evaluation

Evaluate any template expression:

```
shelm> upper "hello"
HELLO

shelm> add 2 3
5

shelm> list "a" "b" "c"
[a b c]
```

### Variable Assignment

Assign values to variables using `name = expression`:

```
shelm> foo = upper "hello"
HELLO

shelm> foo
HELLO

shelm> count = add 10 20
30
```

Variables are stored in `.Values` and persist throughout the session.

### Meta-Commands

| Command | Description |
|---------|-------------|
| `:help` | Show help |
| `:values` | Display all values |
| `:set <name> <expr>` | Set a value |
| `:unset <name>` | Delete a value |
| `:load <file> [name]` | Load YAML/JSON file |
| `:save <name> <file>` | Save value as YAML |
| `:template <file>` | Execute template file |
| `:quit` / `:exit` | Exit the REPL |

### Working with Files

Load a YAML file into a variable:

```
shelm> :load config.yaml cfg
loaded config.yaml into cfg

shelm> cfg
port: 8080
name: myapp
```

Or merge directly into .Values:

```
shelm> :load config.yaml
merged config.yaml into .Values
```

Execute a template file:

```
shelm> :template deployment.yaml.tpl
apiVersion: apps/v1
kind: Deployment
...
```

### Built-in Functions

In addition to all Sprig functions, shelm provides:

| Function | Description |
|----------|-------------|
| `shelmReadFile(path)` | Read file contents as string |
| `shelmWriteFile(path, data)` | Write data to file |
| `fromYaml(str)` | Parse YAML string to value |
| `toYaml(val)` | Convert value to YAML string |
| `fromJson(str)` | Parse JSON string to value |
| `toJson(val)` | Convert value to JSON string |
| `toPrettyJson(val)` | Convert value to pretty JSON |

### Example Session

```
$ shelm
shelm> foo = upper "hello"
HELLO

shelm> :load values.yaml
merged values.yaml into .Values

shelm> :values
foo: HELLO
service:
    name: myapp
    port: 8080

shelm> {{ dig "service" "port" 0 .Values }}
8080

shelm> :template config.tmpl
PORT=8080
NAME=myapp

shelm> :quit
```

## Template Context

Templates receive a context with `.Values` - the same as Helm charts use.

Access values in templates:

```
{{ .Values.myvar }}
{{ index .Values "my-var" }}
{{ dig "nested" "key" "default" .Values.config }}
```

## Helm Template Debugging

shelm is useful for debugging Helm chart templates. Load your values and test templates interactively:

```
shelm> :load mychart/values.yaml
shelm> :template mychart/templates/deployment.yaml
```

See [docs/troubleshooting-helm-templates.md](docs/troubleshooting-helm-templates.md) for a complete guide.

## License

MIT

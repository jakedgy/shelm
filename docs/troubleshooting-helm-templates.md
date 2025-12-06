# Troubleshooting Helm Templates with shelm

This guide demonstrates how to use `shelm` to debug and troubleshoot Helm chart templates. We'll work through common issues using a sample chart.

## Setup

First, start shelm and load the chart's values:

```
$ shelm
shelm> :load docs/examples/broken-chart/values.yaml
merged docs/examples/broken-chart/values.yaml into .Values
```

Now all your Helm values are available in `.Values` - exactly like in Helm charts.

## Testing Actual Helm Templates

You can render complete Helm templates directly:

```
shelm> :template docs/examples/broken-chart/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
...
```

This helps you spot issues like:
- **Type errors**: `DB_PORT: 5432` should be `DB_PORT: "5432"` for env vars
- **Boolean coercion**: `CACHE_ENABLED: true` should be `CACHE_ENABLED: "true"`
- **Missing conditionals**: livenessProbe not rendered when `enableMetrics: false`

## Common Issues and How to Debug Them

### Issue 1: Accessing Nested Values

**Problem:** Your template fails with "nil pointer" or "can't evaluate field" errors.

Let's explore the values structure:

```
shelm> :values
```

To access nested values safely, use `dig`:

```
shelm> dig "config" "database" "host" "" .Values
postgres.default.svc

shelm> dig "config" "database" "missing_key" "default_value" .Values
default_value
```

Test accessing deeply nested values:

```
shelm> .Values.config.database.host
postgres.default.svc

shelm> .Values.config.features.enableCache
true
```

### Issue 2: Type Coercion Problems

**Problem:** Kubernetes expects string values for env vars, but your YAML has integers or booleans.

```
shelm> .Values.config.database.port
5432

shelm> printf "%v" .Values.config.database.port
5432
```

The port is an integer. In Kubernetes env vars, this needs to be a string:

```
shelm> printf "%d" .Values.config.database.port
5432

shelm> toString .Values.config.database.port
5432

shelm> quote (toString .Values.config.database.port)
"5432"
```

For booleans:

```
shelm> .Values.config.features.enableCache
true

shelm> quote (toString .Values.config.features.enableCache)
"true"
```

### Issue 3: Building Complex Strings

**Problem:** Need to construct URLs or connection strings from multiple values.

```
shelm> printf "redis://%s:%d" .Values.config.redis.host .Values.config.redis.port
redis://redis.default.svc:6379

shelm> printf "postgres://%s:%d/%s" .Values.config.database.host .Values.config.database.port .Values.config.database.name
postgres://postgres.default.svc:5432/myapp_db
```

### Issue 4: Working with Lists

**Problem:** Iterating over lists and accessing list elements.

Check the env list:

```
shelm> .Values.env
- name: LOG_LEVEL
  value: info
- name: APP_ENV
  value: production

shelm> index .Values.env 0
map[name:LOG_LEVEL value:info]

shelm> (index .Values.env 0).name
LOG_LEVEL
```

Get all env var names:

```
shelm> range .Values.env
```

This won't work directly - `range` needs to be in a template block:

```
shelm> {{ range .Values.env }}{{ .name }}: {{ .value }}{{ "\n" }}{{ end }}
LOG_LEVEL: info
APP_ENV: production
```

### Issue 5: Conditional Logic

**Problem:** Testing conditionals before using them in templates.

```
shelm> .Values.ingress.enabled
true

shelm> .Values.ingress.tls.enabled
true

shelm> and .Values.ingress.enabled .Values.ingress.tls.enabled
true

shelm> {{ if and .Values.ingress.enabled .Values.ingress.tls.enabled }}TLS is enabled{{ end }}
TLS is enabled
```

Test for missing or false values:

```
shelm> .Values.config.features.enableMetrics
false

shelm> {{ if .Values.config.features.enableMetrics }}Metrics enabled{{ else }}Metrics disabled{{ end }}
Metrics disabled
```

### Issue 6: Default Values

**Problem:** Handling missing values gracefully.

```
shelm> .Values.name
<no value>

shelm> default "myapp" .Values.name
myapp

shelm> .Values.image.tag | default "latest"
1.2.3
```

Coalesce multiple fallbacks:

```
shelm> coalesce .Values.name .Values.image.repository "unknown"
myapp
```

### Issue 7: YAML Formatting

**Problem:** Need to output nested structures as properly indented YAML.

```
shelm> .Values.resources
limits:
    cpu: 500m
    memory: 256Mi
requests:
    cpu: 100m
    memory: 128Mi

shelm> toYaml .Values.resources
limits:
    cpu: 500m
    memory: 256Mi
requests:
    cpu: 100m
    memory: 128Mi
```

Test nindent behavior (add newline + indent):

```
shelm> printf "resources:%s" (toYaml .Values.resources | nindent 2)
resources:
  limits:
    cpu: 500m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

### Issue 8: Label Generation

**Problem:** Generating labels from a map.

```
shelm> .Values.labels
app.kubernetes.io/name: myapp
app.kubernetes.io/version: "1.2.3"
```

Iterate over labels:

```
shelm> {{ range $k, $v := .Values.labels }}{{ $k }}: {{ $v }}{{ "\n" }}{{ end }}
app.kubernetes.io/name: myapp
app.kubernetes.io/version: 1.2.3
```

### Issue 9: Testing Template Snippets

Save a partial template to test:

```
shelm> :save config test-config.yaml
saved config to test-config.yaml
```

Load and transform external data:

```
shelm> data = fromYaml (shelmReadFile "docs/examples/broken-chart/values.yaml")
...

shelm> keys .Values.data
[config env image ingress labels replicaCount resources service]
```

### Issue 10: Debugging Image Tags

**Problem:** Constructing the full image reference.

```
shelm> printf "%s:%s" .Values.image.repository .Values.image.tag
myapp:1.2.3

shelm> .Values.image | toJson
{"pullPolicy":"IfNotPresent","repository":"myapp","tag":"1.2.3"}
```

## Testing a Full Template

Create a simple configmap template and test it:

```
shelm> {{ shelmWriteFile "test-cm.yaml" (printf "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: %s-config\ndata:\n  DB_HOST: %s\n  DB_PORT: \"%d\"\n" (default "myapp" .Values.name) .Values.config.database.host .Values.config.database.port) }}

shelm> shelmReadFile "test-cm.yaml"
apiVersion: v1
kind: ConfigMap
metadata:
  name: myapp-config
data:
  DB_HOST: postgres.default.svc
  DB_PORT: "5432"
```

## Quick Reference

| Task | Command |
|------|---------|
| Load values | `:load values.yaml` |
| View all values | `:values` |
| Access nested value | `.Values.path.to.value` |
| Safe nested access | `dig "key1" "key2" "default" .Values` |
| Default value | `default "fallback" .Values.key` |
| Convert to string | `toString .Values.intValue` |
| Quote string | `quote .Values.value` |
| Format string | `printf "fmt" args...` |
| Convert to YAML | `toYaml .Values.object` |
| Indent YAML | `toYaml .Values.obj \| nindent 2` |
| Check boolean | `{{ if .Values.enabled }}...{{ end }}` |
| Iterate list | `{{ range .Values.list }}...{{ end }}` |
| Iterate map | `{{ range $k, $v := .Values.map }}...{{ end }}` |
| Get map keys | `keys .Values.map` |

## Tips

1. **Start simple**: Load your values and explore with `:values` first
2. **Test incrementally**: Build complex expressions piece by piece
3. **Use `dig` for safety**: Avoid nil pointer errors with nested access
4. **Check types**: Use `printf "%T"` to see the Go type of a value
5. **Save your work**: Use `:save` to export tested configurations

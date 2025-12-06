package shelm

import (
	"encoding/json"
	"os"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"
)

// ShelmFuncs returns the shelm-specific template functions.
func ShelmFuncs() template.FuncMap {
	return template.FuncMap{
		"shelmReadFile":  shelmReadFile,
		"shelmWriteFile": shelmWriteFile,
		// YAML/JSON functions (commonly needed but not in TxtFuncMap)
		"fromYaml":     fromYaml,
		"toYaml":       toYaml,
		"fromJson":     fromJson,
		"toJson":       toJson,
		"toPrettyJson": toPrettyJson,
	}
}

// BuildFuncMap creates a combined function map with Sprig and shelm functions.
// The captureFunc parameter allows providing a custom capture function for storing values.
func BuildFuncMap(captureFunc func(string, any) any) template.FuncMap {
	// Start with Sprig functions
	funcs := sprig.TxtFuncMap()

	// Add shelm-specific functions
	for k, v := range ShelmFuncs() {
		funcs[k] = v
	}

	// Add capture function if provided
	if captureFunc != nil {
		funcs["shelmCapture"] = captureFunc
	}

	return funcs
}

// shelmReadFile reads a file from disk and returns its contents as a string.
func shelmReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// shelmWriteFile writes data to a file. Returns empty string on success.
func shelmWriteFile(path string, data string) (string, error) {
	err := os.WriteFile(path, []byte(data), 0644)
	if err != nil {
		return "", err
	}
	return "", nil
}

// fromYaml parses a YAML string into a map or slice.
func fromYaml(str string) (any, error) {
	var result any
	if err := yaml.Unmarshal([]byte(str), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// toYaml converts a value to YAML string.
func toYaml(v any) (string, error) {
	data, err := yaml.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// fromJson parses a JSON string into a map or slice.
func fromJson(str string) (any, error) {
	var result any
	if err := json.Unmarshal([]byte(str), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// toJson converts a value to JSON string.
func toJson(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// toPrettyJson converts a value to pretty-printed JSON string.
func toPrettyJson(v any) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetFunctionNames returns a list of all available function names.
// This includes all Sprig functions and shelm custom functions.
func GetFunctionNames() []string {
	funcMap := BuildFuncMap(nil)
	names := make([]string, 0, len(funcMap))
	for name := range funcMap {
		names = append(names, name)
	}
	return names
}

package shelm

import (
	"testing"
)

func TestNewContext(t *testing.T) {
	ctx := NewContext()
	if ctx == nil {
		t.Fatal("NewContext returned nil")
	}
	if ctx.Values == nil {
		t.Fatal("Values map is nil")
	}
	if len(ctx.Values) != 0 {
		t.Errorf("Values map should be empty, got %d items", len(ctx.Values))
	}
}

func TestContext_Set(t *testing.T) {
	ctx := NewContext()

	ctx.Set("foo", "bar")
	if v, ok := ctx.Values["foo"]; !ok || v != "bar" {
		t.Errorf("Set failed: expected 'bar', got %v", v)
	}

	// Test overwrite
	ctx.Set("foo", "baz")
	if v := ctx.Values["foo"]; v != "baz" {
		t.Errorf("Set overwrite failed: expected 'baz', got %v", v)
	}

	// Test different types
	ctx.Set("int", 42)
	ctx.Set("float", 3.14)
	ctx.Set("bool", true)
	ctx.Set("slice", []string{"a", "b"})
	ctx.Set("map", map[string]any{"key": "value"})

	if ctx.Values["int"] != 42 {
		t.Error("Failed to set int")
	}
	if ctx.Values["float"] != 3.14 {
		t.Error("Failed to set float")
	}
	if ctx.Values["bool"] != true {
		t.Error("Failed to set bool")
	}
}

func TestContext_Get(t *testing.T) {
	ctx := NewContext()
	ctx.Values["foo"] = "bar"

	v, ok := ctx.Get("foo")
	if !ok {
		t.Error("Get returned false for existing key")
	}
	if v != "bar" {
		t.Errorf("Get returned wrong value: expected 'bar', got %v", v)
	}

	_, ok = ctx.Get("nonexistent")
	if ok {
		t.Error("Get returned true for non-existent key")
	}
}

func TestContext_Delete(t *testing.T) {
	ctx := NewContext()
	ctx.Values["foo"] = "bar"
	ctx.Values["baz"] = "qux"

	ctx.Delete("foo")

	if _, ok := ctx.Values["foo"]; ok {
		t.Error("Delete failed to remove key")
	}
	if _, ok := ctx.Values["baz"]; !ok {
		t.Error("Delete removed wrong key")
	}

	// Deleting non-existent key should not panic
	ctx.Delete("nonexistent")
}

func TestContext_Merge(t *testing.T) {
	ctx := NewContext()
	ctx.Values["existing"] = "value"

	data := map[string]any{
		"new1": "value1",
		"new2": 42,
	}

	ctx.Merge(data)

	if ctx.Values["existing"] != "value" {
		t.Error("Merge removed existing key")
	}
	if ctx.Values["new1"] != "value1" {
		t.Error("Merge failed to add new1")
	}
	if ctx.Values["new2"] != 42 {
		t.Error("Merge failed to add new2")
	}

	// Test overwrite with merge
	overwrite := map[string]any{
		"existing": "overwritten",
	}
	ctx.Merge(overwrite)

	if ctx.Values["existing"] != "overwritten" {
		t.Error("Merge failed to overwrite existing key")
	}
}

func TestContext_MergeEmpty(t *testing.T) {
	ctx := NewContext()
	ctx.Values["foo"] = "bar"

	ctx.Merge(map[string]any{})

	if ctx.Values["foo"] != "bar" {
		t.Error("Merge with empty map modified existing values")
	}
}

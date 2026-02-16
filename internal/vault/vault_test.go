package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileVault(t *testing.T) {
	tmpDir := t.TempDir()
	v := &FileVault{
		path: filepath.Join(tmpDir, "vault.enc"),
		key:  deriveKey(),
	}

	// Test Set and Get
	if err := v.Set("TEST_KEY", "test-value-123"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, err := v.Get("TEST_KEY")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "test-value-123" {
		t.Errorf("expected 'test-value-123', got %q", val)
	}

	// Test Get nonexistent
	_, err = v.Get("NONEXISTENT")
	if err == nil {
		t.Error("expected error for nonexistent key")
	}

	// Test List
	v.Set("ANOTHER_KEY", "another-value")
	keys, err := v.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}

	// Test Delete
	if err := v.Delete("TEST_KEY"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	keys, _ = v.List()
	if len(keys) != 1 {
		t.Errorf("expected 1 key after delete, got %d", len(keys))
	}

	// Test Delete nonexistent
	if err := v.Delete("NONEXISTENT"); err == nil {
		t.Error("expected error when deleting nonexistent key")
	}
}

func TestFileVault_Persistence(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "vault.enc")
	key := deriveKey()

	// Write with one instance
	v1 := &FileVault{path: path, key: key}
	v1.Set("PERSIST_KEY", "persist-value")

	// Read with another instance
	v2 := &FileVault{path: path, key: key}
	val, err := v2.Get("PERSIST_KEY")
	if err != nil {
		t.Fatalf("Get from second instance failed: %v", err)
	}
	if val != "persist-value" {
		t.Errorf("expected 'persist-value', got %q", val)
	}
}

func TestFileVault_EmptyVault(t *testing.T) {
	tmpDir := t.TempDir()
	v := &FileVault{
		path: filepath.Join(tmpDir, "vault.enc"),
		key:  deriveKey(),
	}

	// List on empty vault should return empty
	keys, err := v.List()
	if err != nil {
		t.Fatalf("List on empty vault failed: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("expected 0 keys, got %d", len(keys))
	}
}

func TestFileVault_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()
	v := &FileVault{
		path: filepath.Join(tmpDir, "vault.enc"),
		key:  deriveKey(),
	}

	v.Set("KEY", "value1")
	v.Set("KEY", "value2")

	val, _ := v.Get("KEY")
	if val != "value2" {
		t.Errorf("expected 'value2' after overwrite, got %q", val)
	}

	keys, _ := v.List()
	if len(keys) != 1 {
		t.Errorf("expected 1 key (no duplicate), got %d", len(keys))
	}
}

func TestMask(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"short", "****"},
		{"12345678", "****"},
		{"sk-1234567890abcdef", "sk-1...cdef"},
		{"ANTHROPIC_API_KEY_VALUE_LONG", "ANTH...LONG"},
	}

	for _, tt := range tests {
		result := Mask(tt.input)
		if result != tt.expected {
			t.Errorf("Mask(%q): expected %q, got %q", tt.input, tt.expected, result)
		}
	}
}

func TestNewVault_ReturnsSomething(t *testing.T) {
	v := New()
	if v == nil {
		t.Fatal("New() returned nil")
	}
}

func TestDeriveKey(t *testing.T) {
	// Key should be deterministic for same machine
	key1 := deriveKey()
	key2 := deriveKey()
	if len(key1) != 32 {
		t.Errorf("expected 32-byte key, got %d", len(key1))
	}
	for i := range key1 {
		if key1[i] != key2[i] {
			t.Error("deriveKey should be deterministic")
			break
		}
	}
}

func TestNewFileVault_DefaultPath(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	v := NewFileVault()
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "palm", "vault.enc")
	if v.path != expected {
		t.Errorf("expected %q, got %q", expected, v.path)
	}
}

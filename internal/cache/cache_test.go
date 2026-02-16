package cache

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDir(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", "/tmp/test-cache")
	if Dir() != "/tmp/test-cache/palm" {
		t.Errorf("expected /tmp/test-cache/palm, got %q", Dir())
	}

	t.Setenv("XDG_CACHE_HOME", "")
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".cache", "palm")
	if Dir() != expected {
		t.Errorf("expected %q, got %q", expected, Dir())
	}
}

func TestIsCached_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmpDir)

	if IsCached("pip", "nonexistent") {
		t.Error("should not be cached")
	}
	if IsCached("npm", "nonexistent") {
		t.Error("should not be cached")
	}
	if IsCached("docker", "nonexistent") {
		t.Error("should not be cached")
	}
}

func TestSanitize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"org/image:tag", "org_image_tag"},
		{"@scope/pkg", "_scope_pkg"},
		{"a/b:c@d", "a_b_c_d"},
	}

	for _, tt := range tests {
		result := sanitize(tt.input)
		if result != tt.expected {
			t.Errorf("sanitize(%q): expected %q, got %q", tt.input, tt.expected, result)
		}
	}
}

func TestBundle_EmptyCache(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmpDir)

	err := Bundle(filepath.Join(tmpDir, "out.tar.gz"))
	if err == nil {
		t.Error("expected error for empty cache")
	}
}

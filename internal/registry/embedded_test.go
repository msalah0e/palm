package registry

import (
	"embed"
	"testing"
)

//go:embed testdata/*.toml
var testFS embed.FS

func TestLoadFromFS(t *testing.T) {
	reg, err := LoadFromFS(testFS, "testdata")
	if err != nil {
		t.Fatalf("LoadFromFS failed: %v", err)
	}

	tools := reg.All()
	if len(tools) != 3 {
		t.Errorf("expected 3 tools, got %d", len(tools))
	}

	// Verify specific tools loaded
	tool := reg.Get("test-tool-a")
	if tool == nil {
		t.Fatal("test-tool-a not found")
	}
	if tool.DisplayName != "Test Tool A" {
		t.Errorf("expected 'Test Tool A', got %q", tool.DisplayName)
	}
	if tool.Category != "coding" {
		t.Errorf("expected category 'coding', got %q", tool.Category)
	}
	if tool.Install.Pip != "test-tool-a" {
		t.Errorf("expected pip 'test-tool-a', got %q", tool.Install.Pip)
	}

	tool2 := reg.Get("test-tool-c")
	if tool2 == nil {
		t.Fatal("test-tool-c not found")
	}
	if tool2.Install.Docker != "testorg/tool-c:latest" {
		t.Errorf("expected docker image, got %q", tool2.Install.Docker)
	}
}

func TestLoadFromFS_InvalidDir(t *testing.T) {
	_, err := LoadFromFS(testFS, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

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

func TestLoadFromFS_SkipsPresets(t *testing.T) {
	reg, err := LoadFromFS(testFS, "testdata")
	if err != nil {
		t.Fatalf("LoadFromFS failed: %v", err)
	}

	// presets.toml should be skipped — only tool files should be loaded
	tools := reg.All()
	if len(tools) != 3 {
		t.Errorf("expected 3 tools (presets.toml skipped), got %d", len(tools))
	}
}

func TestLoadPresetsFromFS(t *testing.T) {
	presets, err := LoadPresetsFromFS(testFS, "testdata")
	if err != nil {
		t.Fatalf("LoadPresetsFromFS failed: %v", err)
	}

	if len(presets) != 2 {
		t.Errorf("expected 2 presets, got %d", len(presets))
	}

	if presets[0].Name != "test-preset" {
		t.Errorf("expected first preset name 'test-preset', got %q", presets[0].Name)
	}
	if presets[0].DisplayName != "Test Preset" {
		t.Errorf("expected display name 'Test Preset', got %q", presets[0].DisplayName)
	}
	if len(presets[0].Tools) != 2 {
		t.Errorf("expected 2 tools in first preset, got %d", len(presets[0].Tools))
	}
	if presets[0].Tools[0] != "test-tool-a" {
		t.Errorf("expected first tool 'test-tool-a', got %q", presets[0].Tools[0])
	}
}

func TestLoadPresetsFromFS_MissingFile(t *testing.T) {
	// Use a FS that doesn't have presets.toml
	var emptyFS embed.FS
	_, err := LoadPresetsFromFS(emptyFS, "nonexistent")
	if err == nil {
		t.Error("expected error for missing presets.toml")
	}
}

func TestLoadFromFS_InvalidDir(t *testing.T) {
	_, err := LoadFromFS(testFS, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

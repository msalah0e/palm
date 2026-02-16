package state

import (
	"os"
	"testing"
)

func TestState(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Empty state
	s := Load()
	if len(s.Installed) != 0 {
		t.Errorf("expected empty state, got %d tools", len(s.Installed))
	}

	// Record a tool
	if err := Record("aider", "0.72.1", "pip", "aider-chat", "/usr/local/bin/aider"); err != nil {
		t.Fatalf("Record failed: %v", err)
	}

	// Verify it persists
	s = Load()
	if len(s.Installed) != 1 {
		t.Errorf("expected 1 tool, got %d", len(s.Installed))
	}
	tool, ok := s.Installed["aider"]
	if !ok {
		t.Fatal("aider not found in state")
	}
	if tool.Version != "0.72.1" {
		t.Errorf("expected version '0.72.1', got %q", tool.Version)
	}
	if tool.Backend != "pip" {
		t.Errorf("expected backend 'pip', got %q", tool.Backend)
	}
	if tool.InstalledAt.IsZero() {
		t.Error("InstalledAt should be set")
	}

	// Update existing tool
	Record("aider", "0.73.0", "pip", "aider-chat", "/usr/local/bin/aider")
	s = Load()
	if s.Installed["aider"].Version != "0.73.0" {
		t.Errorf("expected updated version '0.73.0', got %q", s.Installed["aider"].Version)
	}

	// IsInstalled
	if !IsInstalled("aider") {
		t.Error("aider should be installed")
	}
	if IsInstalled("nonexistent") {
		t.Error("nonexistent should not be installed")
	}

	// ListInstalled
	installed := ListInstalled()
	if len(installed) != 1 {
		t.Errorf("expected 1 installed, got %d", len(installed))
	}

	// Remove
	if err := Remove("aider"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	if IsInstalled("aider") {
		t.Error("aider should be removed")
	}
}

func TestState_DefaultPath(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	path := statePath()
	home, _ := os.UserHomeDir()
	if path == "" {
		t.Fatal("statePath should not be empty")
	}
	if path != home+"/.config/palm/state.toml" {
		t.Errorf("unexpected path: %q", path)
	}
}

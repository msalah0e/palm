package stats

import (
	"os"
	"testing"
)

func TestStats(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Empty stats
	s, err := Summarize()
	if err != nil {
		t.Fatalf("Summarize on empty failed: %v", err)
	}
	if s.TotalCommands != 0 {
		t.Errorf("expected 0 commands, got %d", s.TotalCommands)
	}

	// Record some events
	Record("install", "aider", "aider-chat", true)
	Record("install", "ollama", "ollama", true)
	Record("install", "bad-tool", "", false)
	Record("search", "", "", true)

	// Verify summary
	s, err = Summarize()
	if err != nil {
		t.Fatalf("Summarize failed: %v", err)
	}
	if s.TotalCommands != 4 {
		t.Errorf("expected 4 commands, got %d", s.TotalCommands)
	}
	if s.ToolsInstalled != 2 {
		t.Errorf("expected 2 tools installed (successful only), got %d", s.ToolsInstalled)
	}
	if s.LastUsed.IsZero() {
		t.Error("LastUsed should be set")
	}
}

func TestHistoryPath(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	path := historyPath()
	home, _ := os.UserHomeDir()
	if path != home+"/.config/palm/history.jsonl" {
		t.Errorf("unexpected path: %q", path)
	}
}

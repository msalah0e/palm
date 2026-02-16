package session

import (
	"os"
	"testing"
	"time"
)

func TestStartAndEnd(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	s, err := Start("aider")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if s.Tool != "aider" {
		t.Errorf("expected tool 'aider', got %q", s.Tool)
	}
	if s.StartedAt.IsZero() {
		t.Error("StartedAt should not be zero")
	}

	time.Sleep(10 * time.Millisecond)
	if err := End(s, 0); err != nil {
		t.Fatalf("End failed: %v", err)
	}
	if s.Duration <= 0 {
		t.Error("duration should be > 0")
	}
	if s.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", s.ExitCode)
	}
}

func TestRecord(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	err := Record("claude-code", 5*time.Second, 0, 0.05, 1000, "anthropic")
	if err != nil {
		t.Fatalf("Record failed: %v", err)
	}

	err = Record("aider", 10*time.Second, 1, 0.02, 500, "openai")
	if err != nil {
		t.Fatalf("Record failed: %v", err)
	}

	sessions, err := List(0)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}

	// Most recent first
	if sessions[0].Tool != "aider" {
		t.Errorf("expected most recent to be aider, got %q", sessions[0].Tool)
	}
}

func TestList(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	// Empty
	sessions, err := List(10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(sessions))
	}

	// Add 5 sessions
	for i := 0; i < 5; i++ {
		_ = Record("tool", time.Second, 0, 0, 0, "")
	}

	// Get last 3
	sessions, err = List(3)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(sessions) != 3 {
		t.Errorf("expected 3 sessions, got %d", len(sessions))
	}
}

func TestSummarize(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	_ = Record("aider", 5*time.Second, 0, 0.10, 1000, "openai")
	_ = Record("aider", 3*time.Second, 0, 0.05, 500, "openai")
	_ = Record("claude-code", 10*time.Second, 0, 0.20, 2000, "anthropic")

	summary, err := Summarize()
	if err != nil {
		t.Fatalf("Summarize failed: %v", err)
	}
	if summary.TotalSessions != 3 {
		t.Errorf("expected 3 sessions, got %d", summary.TotalSessions)
	}
	if summary.TotalTokens != 3500 {
		t.Errorf("expected 3500 tokens, got %d", summary.TotalTokens)
	}

	aider := summary.ByTool["aider"]
	if aider.Sessions != 2 {
		t.Errorf("expected 2 aider sessions, got %d", aider.Sessions)
	}
}

func TestSummarizeEmpty(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	summary, err := Summarize()
	if err != nil {
		t.Fatalf("Summarize failed: %v", err)
	}
	if summary.TotalSessions != 0 {
		t.Errorf("expected 0 sessions, got %d", summary.TotalSessions)
	}
}

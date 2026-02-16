package budget

import (
	"os"
	"testing"
	"time"

	"github.com/msalah0e/palm/internal/session"
)

func TestLoadDefault(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	b := Load()
	if b.AlertAt != 0.8 {
		t.Errorf("expected default AlertAt 0.8, got %f", b.AlertAt)
	}
	if b.MonthlyLimit != 0 {
		t.Errorf("expected no monthly limit, got %f", b.MonthlyLimit)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	b := &Budget{
		MonthlyLimit: 50.0,
		DailyLimit:   10.0,
		AlertAt:      0.9,
		PerTool:      map[string]float64{"aider": 20.0},
	}
	if err := Save(b); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded := Load()
	if loaded.MonthlyLimit != 50.0 {
		t.Errorf("expected monthly 50, got %f", loaded.MonthlyLimit)
	}
	if loaded.DailyLimit != 10.0 {
		t.Errorf("expected daily 10, got %f", loaded.DailyLimit)
	}
	if loaded.PerTool["aider"] != 20.0 {
		t.Errorf("expected per-tool aider 20, got %f", loaded.PerTool["aider"])
	}
}

func TestGetStatus(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	// Set budget
	b := &Budget{
		MonthlyLimit: 100.0,
		DailyLimit:   20.0,
		AlertAt:      0.8,
		PerTool:      make(map[string]float64),
	}
	_ = Save(b)

	// Record sessions
	_ = session.Record("aider", 5*time.Second, 0, 10.0, 1000, "openai")
	_ = session.Record("claude-code", 5*time.Second, 0, 5.0, 500, "anthropic")

	status, err := GetStatus()
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if status.MonthlySpend != 15.0 {
		t.Errorf("expected monthly spend 15, got %f", status.MonthlySpend)
	}
	if status.PercentUsed != 15.0 {
		t.Errorf("expected 15%% used, got %f", status.PercentUsed)
	}
	if status.IsOverBudget {
		t.Error("should not be over budget")
	}
	if status.IsNearBudget {
		t.Error("should not be near budget at 15%%")
	}
}

func TestCheckBudget(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	// No budget set — should pass
	if err := CheckBudget("aider"); err != nil {
		t.Errorf("no budget should pass: %v", err)
	}

	// Set tight budget
	b := &Budget{
		MonthlyLimit: 1.0,
		AlertAt:      0.8,
		PerTool:      make(map[string]float64),
	}
	_ = Save(b)

	// Record a session that exceeds the limit
	_ = session.Record("aider", 5*time.Second, 0, 2.0, 1000, "openai")

	err := CheckBudget("aider")
	if err == nil {
		t.Error("expected budget exceeded error")
	}
}

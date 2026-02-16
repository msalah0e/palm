package cmd

import (
	"testing"
)

func TestPrintTruncatedOutput_Short(t *testing.T) {
	// Just verify it doesn't panic
	printTruncatedOutput("short text", 100)
}

func TestPrintTruncatedOutput_Long(t *testing.T) {
	// Create a long string
	long := ""
	for i := 0; i < 500; i++ {
		long += "a"
	}

	// Should not panic with small limit
	printTruncatedOutput(long, 100)
}

func TestPrintTruncatedOutput_Empty(t *testing.T) {
	printTruncatedOutput("", 100)
}

func TestSquadResult_Fields(t *testing.T) {
	r := SquadResult{
		Tool:     "test-tool",
		Output:   "output text",
		ExitCode: 0,
		Error:    "",
	}

	if r.Tool != "test-tool" {
		t.Errorf("expected tool 'test-tool', got %q", r.Tool)
	}
	if r.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", r.ExitCode)
	}
	if r.Error != "" {
		t.Errorf("expected empty error, got %q", r.Error)
	}
}

func TestSquadResult_Error(t *testing.T) {
	r := SquadResult{
		Tool:     "failing-tool",
		ExitCode: -1,
		Error:    "not installed",
	}

	if r.Error != "not installed" {
		t.Errorf("expected error 'not installed', got %q", r.Error)
	}
}

func TestHandleRaceMode_AllFailed(t *testing.T) {
	results := []SquadResult{
		{Tool: "tool1", Error: "not installed"},
		{Tool: "tool2", Error: "timeout"},
	}

	// Should not panic — prints "All tools failed"
	handleRaceMode(results)
}

func TestHandleAllMode_Empty(t *testing.T) {
	results := []SquadResult{
		{Tool: "tool1", Error: "not installed"},
	}

	// Should not panic
	handleAllMode(results, false)
}

func TestHandleVoteMode_InsufficientCandidates(t *testing.T) {
	results := []SquadResult{
		{Tool: "tool1", Output: "output", Error: ""},
		{Tool: "tool2", Error: "not installed"},
	}

	// Should warn about needing 2+ results, not panic
	handleVoteMode(results, "fake-judge", "task", nil, 1)
}

func TestHandleMergeMode_NoCandidates(t *testing.T) {
	results := []SquadResult{
		{Tool: "tool1", Error: "not installed"},
	}

	// Should warn about no results, not panic
	handleMergeMode(results, "fake-judge", "task", nil, 1)
}

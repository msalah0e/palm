package cmd

import (
	"testing"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0B"},
		{512, "512B"},
		{1023, "1023B"},
		{1024, "1.0KB"},
		{2048, "2.0KB"},
		{5120, "5.0KB"},
		{1536, "1.5KB"},
	}

	for _, tt := range tests {
		result := formatBytes(tt.input)
		if result != tt.expected {
			t.Errorf("formatBytes(%d) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatGrade(t *testing.T) {
	grades := []string{"A+", "A", "B", "C", "D", "F"}
	for _, g := range grades {
		result := formatGrade(g)
		if result == "" {
			t.Errorf("formatGrade(%q) returned empty string", g)
		}
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{1, 2, 2},
		{5, 3, 5},
		{0, 0, 0},
		{-1, -2, -1},
		{-1, 1, 1},
		{100, 100, 100},
	}

	for _, tt := range tests {
		result := max(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("max(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestSpeedResult_Fields(t *testing.T) {
	r := SpeedResult{
		Provider:  "Ollama",
		Model:     "llama3.3",
		OutputLen: 500,
		TokensEst: 125,
		TPS:       25.0,
		ExitCode:  0,
	}

	if r.Provider != "Ollama" {
		t.Errorf("expected provider 'Ollama', got %q", r.Provider)
	}
	if r.TokensEst != 125 {
		t.Errorf("expected tokens 125, got %d", r.TokensEst)
	}
	if r.TPS != 25.0 {
		t.Errorf("expected TPS 25.0, got %f", r.TPS)
	}
}

func TestSpeedResult_Error(t *testing.T) {
	r := SpeedResult{
		Provider: "Test",
		Model:    "test",
		ExitCode: -1,
		Error:    "timeout (90s)",
	}

	if r.Error != "timeout (90s)" {
		t.Errorf("expected timeout error, got %q", r.Error)
	}
}

func TestPrintSpeedtestHeader(t *testing.T) {
	// Should not panic
	printSpeedtestHeader()
}

func TestPrintSpeedtestResults_Empty(t *testing.T) {
	// Should not panic with empty results
	printSpeedtestResults([]SpeedResult{})
}

func TestPrintSpeedtestResults_WithErrors(t *testing.T) {
	results := []SpeedResult{
		{Provider: "Test1", Model: "m1", Error: "failed"},
		{Provider: "Test2", Model: "m2", TPS: 10.0, OutputLen: 100, TokensEst: 25},
	}

	// Should not panic
	printSpeedtestResults(results)
}

func TestPrintSpeedGrade_NoProviders(t *testing.T) {
	results := []SpeedResult{
		{Provider: "Test", Error: "failed"},
	}

	// Should print F grade, not panic
	printSpeedGrade(results)
}

func TestPrintSpeedGrade_FastProviders(t *testing.T) {
	results := []SpeedResult{
		{Provider: "Fast", Model: "m1", TPS: 120.0},
	}

	// Should print A+ grade
	printSpeedGrade(results)
}

func TestPrintSpeedGrade_SlowProviders(t *testing.T) {
	results := []SpeedResult{
		{Provider: "Slow", Model: "m1", TPS: 2.0},
	}

	// Should print F grade
	printSpeedGrade(results)
}

func TestPrintSpeedGrade_GradeThresholds(t *testing.T) {
	tests := []struct {
		tps float64
	}{
		{100}, // A+
		{50},  // A
		{25},  // B
		{10},  // C
		{5},   // D
		{1},   // F
	}

	for _, tt := range tests {
		results := []SpeedResult{
			{Provider: "Test", TPS: tt.tps},
		}
		// Should not panic for any threshold
		printSpeedGrade(results)
	}
}

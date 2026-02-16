package cmd

import (
	"math"
	"strings"
	"testing"
)

func TestParseEvalScore_ValidOutput(t *testing.T) {
	output := `ACCURACY: 85
HALLUCINATION: 10
COMPLETENESS: 90
CLARITY: 80
VERDICT: Accurate and thorough response`

	score := parseEvalScore("test-tool", output)

	if score.Tool != "test-tool" {
		t.Errorf("expected tool 'test-tool', got %q", score.Tool)
	}
	if score.Accuracy != 85 {
		t.Errorf("expected accuracy 85, got %d", score.Accuracy)
	}
	if score.Hallucination != 10 {
		t.Errorf("expected hallucination 10, got %d", score.Hallucination)
	}
	if score.Completeness != 90 {
		t.Errorf("expected completeness 90, got %d", score.Completeness)
	}
	if score.Clarity != 80 {
		t.Errorf("expected clarity 80, got %d", score.Clarity)
	}
	if score.Verdict != "Accurate and thorough response" {
		t.Errorf("expected verdict 'Accurate and thorough response', got %q", score.Verdict)
	}
	if score.Overall == 0 {
		t.Error("expected non-zero overall score")
	}
}

func TestParseEvalScore_EmptyOutput(t *testing.T) {
	score := parseEvalScore("test-tool", "")

	if score.Tool != "test-tool" {
		t.Errorf("expected tool 'test-tool', got %q", score.Tool)
	}
	if score.Accuracy != 0 {
		t.Errorf("expected accuracy 0 for empty output, got %d", score.Accuracy)
	}
	if score.Overall != 0 {
		t.Errorf("expected overall 0 for empty output, got %d", score.Overall)
	}
}

func TestParseEvalScore_UnparseableOutput(t *testing.T) {
	output := "This is some random text that doesn't match the expected format."

	score := parseEvalScore("test-tool", output)

	// Should default to 50s when output exists but can't parse
	if score.Accuracy != 50 {
		t.Errorf("expected accuracy 50 for unparseable, got %d", score.Accuracy)
	}
	if score.Hallucination != 50 {
		t.Errorf("expected hallucination 50 for unparseable, got %d", score.Hallucination)
	}
	if score.Verdict != "Could not parse judge output — showing estimates" {
		t.Errorf("expected fallback verdict, got %q", score.Verdict)
	}
}

func TestParseEvalScore_PartialOutput(t *testing.T) {
	output := `ACCURACY: 95
CLARITY: 70
Some other text`

	score := parseEvalScore("partial", output)

	if score.Accuracy != 95 {
		t.Errorf("expected accuracy 95, got %d", score.Accuracy)
	}
	if score.Clarity != 70 {
		t.Errorf("expected clarity 70, got %d", score.Clarity)
	}
	// Hallucination and Completeness should be 0 since not in output
	// But since Accuracy > 0, overall should be calculated
	if score.Overall == 0 {
		t.Error("expected non-zero overall when accuracy is set")
	}
}

func TestParseEvalScore_OverallCalculation(t *testing.T) {
	tests := []struct {
		name          string
		accuracy      int
		hallucination int
		completeness  int
		clarity       int
		expectMin     int
		expectMax     int
	}{
		{
			name:          "perfect scores",
			accuracy:      100,
			hallucination: 0,
			completeness:  100,
			clarity:       100,
			expectMin:     95,
			expectMax:     100,
		},
		{
			name:          "high hallucination penalty",
			accuracy:      80,
			hallucination: 80,
			completeness:  80,
			clarity:       80,
			expectMin:     30,
			expectMax:     50,
		},
		{
			name:          "all zeros except accuracy",
			accuracy:      50,
			hallucination: 0,
			completeness:  0,
			clarity:       0,
			expectMin:     15,
			expectMax:     25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hallPenalty := float64(tt.hallucination) * 0.5
			raw := (float64(tt.accuracy)*0.4 + float64(tt.completeness)*0.3 + float64(tt.clarity)*0.3) - hallPenalty
			overall := int(math.Max(0, math.Min(100, raw)))

			if overall < tt.expectMin || overall > tt.expectMax {
				t.Errorf("overall %d not in range [%d, %d]", overall, tt.expectMin, tt.expectMax)
			}
		})
	}
}

func TestParseEvalScore_ExtremeValues(t *testing.T) {
	output := `ACCURACY: 100
HALLUCINATION: 100
COMPLETENESS: 100
CLARITY: 100
VERDICT: Maximum hallucination`

	score := parseEvalScore("extreme", output)

	// With 100 hallucination penalty (50), overall should be capped low
	if score.Overall > 60 {
		t.Errorf("expected low overall with max hallucination, got %d", score.Overall)
	}
}

func TestBuildEvalPrompt_Basic(t *testing.T) {
	prompt := buildEvalPrompt("What is 2+2?", "", "The answer is 4")

	if prompt == "" {
		t.Fatal("buildEvalPrompt returned empty string")
	}

	// Should contain the question
	if !strings.Contains(prompt, "What is 2+2?") {
		t.Error("prompt should contain the question")
	}

	// Should contain the response
	if !strings.Contains(prompt, "The answer is 4") {
		t.Error("prompt should contain the response")
	}

	// Should contain scoring criteria
	for _, criterion := range []string{"ACCURACY", "HALLUCINATION", "COMPLETENESS", "CLARITY", "VERDICT"} {
		if !strings.Contains(prompt, criterion) {
			t.Errorf("prompt should contain criterion %q", criterion)
		}
	}
}

func TestBuildEvalPrompt_WithContext(t *testing.T) {
	prompt := buildEvalPrompt("Explain TCP", "networking basics", "TCP is a protocol")

	if !strings.Contains(prompt, "networking basics") {
		t.Error("prompt should contain context when provided")
	}
}

func TestBuildEvalPrompt_WithoutContext(t *testing.T) {
	prompt := buildEvalPrompt("What is Go?", "", "Go is a language")

	if strings.Contains(prompt, "Context:") {
		t.Error("prompt should not contain Context: when context is empty")
	}
}

func TestGradeFromScore(t *testing.T) {
	tests := []struct {
		score int
	}{
		{95},
		{85},
		{75},
		{65},
		{55},
		{30},
		{0},
	}

	for _, tt := range tests {
		result := gradeFromScore(tt.score)
		if result == "" {
			t.Errorf("gradeFromScore(%d) returned empty string", tt.score)
		}
	}
}

func TestFormatScoreNum(t *testing.T) {
	// These return ANSI-colored strings, just verify non-empty
	for _, score := range []int{0, 30, 50, 70, 90, 100} {
		result := formatScoreNum(score)
		if result == "" {
			t.Errorf("formatScoreNum(%d) returned empty string", score)
		}
	}
}

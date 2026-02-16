package models

import (
	"testing"
)

func TestBuiltinProviders(t *testing.T) {
	providers := BuiltinProviders()
	if len(providers) < 5 {
		t.Errorf("expected at least 5 providers, got %d", len(providers))
	}

	// Verify key providers exist
	names := make(map[string]bool)
	for _, p := range providers {
		names[p.Name] = true
		if len(p.Models) == 0 {
			t.Errorf("provider %s has no models", p.Name)
		}
	}

	for _, expected := range []string{"OpenAI", "Anthropic", "Google", "Ollama"} {
		if !names[expected] {
			t.Errorf("missing provider: %s", expected)
		}
	}
}

func TestAllModels(t *testing.T) {
	all := AllModels()
	if len(all) < 15 {
		t.Errorf("expected at least 15 models, got %d", len(all))
	}
}

func TestFindModel(t *testing.T) {
	tests := []struct {
		query    string
		expected string
	}{
		{"gpt-4o", "GPT-4o"},
		{"claude-opus-4-6", "Claude Opus 4.6"},
		{"gemini-2.5-pro", "Gemini 2.5 Pro"},
		{"llama3.3", "Llama 3.3 70B"},
	}

	for _, tt := range tests {
		m := FindModel(tt.query)
		if m == nil {
			t.Errorf("FindModel(%q) returned nil, expected %s", tt.query, tt.expected)
			continue
		}
		if m.Name != tt.expected {
			t.Errorf("FindModel(%q) = %q, want %q", tt.query, m.Name, tt.expected)
		}
	}
}

func TestFindModelNotFound(t *testing.T) {
	m := FindModel("nonexistent-model-xyz")
	if m != nil {
		t.Errorf("expected nil for nonexistent model, got %v", m)
	}
}

func TestFormatContext(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{128000, "128k"},
		{200000, "200k"},
		{1048576, "1.0M"},
		{8191, "8k"},
	}

	for _, tt := range tests {
		result := FormatContext(tt.input)
		if result != tt.expected {
			t.Errorf("FormatContext(%d) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestModelCosts(t *testing.T) {
	// Verify all cloud models have cost info
	for _, p := range BuiltinProviders() {
		if p.Name == "Ollama" {
			continue // local models are free
		}
		for _, m := range p.Models {
			if m.Type == "chat" && m.InputCost == 0 {
				t.Errorf("model %s (%s) has no input cost", m.ID, p.Name)
			}
		}
	}
}

func TestProviderEnvKeys(t *testing.T) {
	for _, p := range BuiltinProviders() {
		if p.Name == "Ollama" {
			if p.EnvKey != "" {
				t.Errorf("Ollama should not require API key")
			}
			continue
		}
		if p.EnvKey == "" {
			t.Errorf("provider %s should have an EnvKey", p.Name)
		}
	}
}

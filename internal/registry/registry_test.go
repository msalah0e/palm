package registry

import (
	"testing"
)

func sampleTools() []Tool {
	return []Tool{
		{
			Name:        "aider",
			DisplayName: "Aider",
			Description: "AI pair programming in your terminal",
			Category:    "coding",
			Tags:        []string{"ai", "coding", "pair-programming"},
			Homepage:    "https://aider.chat",
			Install:     Install{Pip: "aider-chat"},
			Keys:        Keys{Required: []string{"OPENAI_API_KEY"}, Optional: []string{"ANTHROPIC_API_KEY"}},
		},
		{
			Name:        "ollama",
			DisplayName: "Ollama",
			Description: "Run large language models locally",
			Category:    "llm",
			Tags:        []string{"ai", "llm", "local"},
			Homepage:    "https://ollama.com",
			Install:     Install{Brew: "ollama", Script: "https://ollama.com/install.sh"},
			Keys:        Keys{},
		},
		{
			Name:        "claude-code",
			DisplayName: "Claude Code",
			Description: "Anthropic's official CLI for Claude",
			Category:    "coding",
			Tags:        []string{"ai", "coding", "claude"},
			Install:     Install{Npm: "@anthropic-ai/claude-code"},
			Keys:        Keys{Required: []string{"ANTHROPIC_API_KEY"}},
		},
		{
			Name:        "vllm",
			DisplayName: "vLLM",
			Description: "High-throughput LLM serving engine",
			Category:    "infra",
			Tags:        []string{"ai", "llm", "serving"},
			Install:     Install{Pip: "vllm", Docker: "vllm/vllm-openai:latest"},
			Keys:        Keys{},
		},
	}
}

func TestNew(t *testing.T) {
	reg := New(sampleTools())
	if reg == nil {
		t.Fatal("New returned nil")
	}
	if len(reg.All()) != 4 {
		t.Errorf("expected 4 tools, got %d", len(reg.All()))
	}
}

func TestGet(t *testing.T) {
	reg := New(sampleTools())

	tool := reg.Get("aider")
	if tool == nil {
		t.Fatal("Get(aider) returned nil")
	}
	if tool.DisplayName != "Aider" {
		t.Errorf("expected DisplayName 'Aider', got %q", tool.DisplayName)
	}

	if reg.Get("nonexistent") != nil {
		t.Error("Get(nonexistent) should return nil")
	}
}

func TestSearch(t *testing.T) {
	reg := New(sampleTools())

	tests := []struct {
		query    string
		expected int
	}{
		{"coding", 2},       // category match
		{"aider", 1},        // name match
		{"llm", 2},          // tag + category match (ollama + vllm)
		{"local", 1},        // tag match
		{"terminal", 1},     // description match
		{"nonexistent", 0},  // no match
		{"claude", 1},       // name/tag match
		{"serving", 1},      // tag match
	}

	for _, tt := range tests {
		results := reg.Search(tt.query)
		if len(results) != tt.expected {
			t.Errorf("Search(%q): expected %d results, got %d", tt.query, tt.expected, len(results))
		}
	}
}

func TestByCategory(t *testing.T) {
	reg := New(sampleTools())

	coding := reg.ByCategory("coding")
	if len(coding) != 2 {
		t.Errorf("expected 2 coding tools, got %d", len(coding))
	}

	llm := reg.ByCategory("llm")
	if len(llm) != 1 {
		t.Errorf("expected 1 llm tool, got %d", len(llm))
	}

	empty := reg.ByCategory("nonexistent")
	if len(empty) != 0 {
		t.Errorf("expected 0 tools for nonexistent category, got %d", len(empty))
	}
}

func TestCategories(t *testing.T) {
	reg := New(sampleTools())
	cats := reg.Categories()

	if len(cats) != 3 { // coding, llm, infra
		t.Errorf("expected 3 categories, got %d: %v", len(cats), cats)
	}
}

func TestInstallMethod(t *testing.T) {
	tests := []struct {
		name            string
		tool            Tool
		expectedBackend string
		expectedPkg     string
	}{
		{
			name:            "brew priority",
			tool:            Tool{Install: Install{Brew: "ollama", Pip: "ollama-python", Docker: "ollama:latest"}},
			expectedBackend: "brew",
			expectedPkg:     "ollama",
		},
		{
			name:            "script priority over pip",
			tool:            Tool{Install: Install{Script: "https://example.com/install.sh", Pip: "something"}},
			expectedBackend: "script",
			expectedPkg:     "https://example.com/install.sh",
		},
		{
			name:            "pip when no brew/script",
			tool:            Tool{Install: Install{Pip: "aider-chat", Npm: "something"}},
			expectedBackend: "pip",
			expectedPkg:     "aider-chat",
		},
		{
			name:            "npm",
			tool:            Tool{Install: Install{Npm: "@anthropic-ai/claude-code"}},
			expectedBackend: "npm",
			expectedPkg:     "@anthropic-ai/claude-code",
		},
		{
			name:            "cargo",
			tool:            Tool{Install: Install{Cargo: "qdrant"}},
			expectedBackend: "cargo",
			expectedPkg:     "qdrant",
		},
		{
			name:            "go",
			tool:            Tool{Install: Install{Go: "github.com/example/tool@latest"}},
			expectedBackend: "go",
			expectedPkg:     "github.com/example/tool@latest",
		},
		{
			name:            "docker",
			tool:            Tool{Install: Install{Docker: "vllm/vllm:latest"}},
			expectedBackend: "docker",
			expectedPkg:     "vllm/vllm:latest",
		},
		{
			name:            "manual fallback",
			tool:            Tool{Homepage: "https://example.com"},
			expectedBackend: "manual",
			expectedPkg:     "https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend, pkg := tt.tool.InstallMethod()
			if backend != tt.expectedBackend {
				t.Errorf("expected backend %q, got %q", tt.expectedBackend, backend)
			}
			if pkg != tt.expectedPkg {
				t.Errorf("expected pkg %q, got %q", tt.expectedPkg, pkg)
			}
		})
	}
}

func TestNeedsAPIKey(t *testing.T) {
	reg := New(sampleTools())

	aider := reg.Get("aider")
	if !aider.NeedsAPIKey() {
		t.Error("aider should need API key")
	}

	ollama := reg.Get("ollama")
	if ollama.NeedsAPIKey() {
		t.Error("ollama should not need API key")
	}
}

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.2.3", "1.2.3"},
		{"v2.0.0", "v2.0.0"},
		{"ollama version 0.5.4", "0.5.4"},
		{"aider 0.72.1", "0.72.1"},
		{"2.1.42 (Claude Code)", "2.1.42"},
		{"Python 3.14.0a5", "3.14.0a5"},
		{"go version go1.24.0 darwin/arm64", "go1.24.0"},
		{"npm 11.1.0", "11.1.0"},
	}

	for _, tt := range tests {
		result := ExtractVersion(tt.input)
		if result != tt.expected {
			t.Errorf("ExtractVersion(%q): expected %q, got %q", tt.input, tt.expected, result)
		}
	}
}

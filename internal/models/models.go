package models

import "fmt"

// Provider represents an LLM provider.
type Provider struct {
	Name     string
	Endpoint string
	Models   []Model
	EnvKey   string
}

// Model represents a single LLM model.
type Model struct {
	ID          string  `toml:"id"`
	Name        string  `toml:"name"`
	Provider    string  `toml:"provider"`
	Context     int     `toml:"context"`
	InputCost   float64 `toml:"input_cost"`  // per 1M tokens
	OutputCost  float64 `toml:"output_cost"` // per 1M tokens
	Type        string  `toml:"type"` // chat, completion, embedding, image
	Released    string  `toml:"released"`
	Description string  `toml:"description"`
}

// BuiltinProviders returns the known LLM providers.
func BuiltinProviders() []Provider {
	return []Provider{
		{
			Name:     "OpenAI",
			Endpoint: "https://api.openai.com/v1",
			EnvKey:   "OPENAI_API_KEY",
			Models: []Model{
				{ID: "gpt-4o", Name: "GPT-4o", Provider: "openai", Context: 128000, InputCost: 2.50, OutputCost: 10.00, Type: "chat"},
				{ID: "gpt-4o-mini", Name: "GPT-4o Mini", Provider: "openai", Context: 128000, InputCost: 0.15, OutputCost: 0.60, Type: "chat"},
				{ID: "gpt-4.1", Name: "GPT-4.1", Provider: "openai", Context: 1047576, InputCost: 2.00, OutputCost: 8.00, Type: "chat"},
				{ID: "gpt-4.1-mini", Name: "GPT-4.1 Mini", Provider: "openai", Context: 1047576, InputCost: 0.40, OutputCost: 1.60, Type: "chat"},
				{ID: "gpt-4.1-nano", Name: "GPT-4.1 Nano", Provider: "openai", Context: 1047576, InputCost: 0.10, OutputCost: 0.40, Type: "chat"},
				{ID: "o3", Name: "o3", Provider: "openai", Context: 200000, InputCost: 2.00, OutputCost: 8.00, Type: "chat"},
				{ID: "o4-mini", Name: "o4-mini", Provider: "openai", Context: 200000, InputCost: 1.10, OutputCost: 4.40, Type: "chat"},
				{ID: "text-embedding-3-large", Name: "Embedding 3 Large", Provider: "openai", Context: 8191, InputCost: 0.13, OutputCost: 0, Type: "embedding"},
			},
		},
		{
			Name:     "Anthropic",
			Endpoint: "https://api.anthropic.com/v1",
			EnvKey:   "ANTHROPIC_API_KEY",
			Models: []Model{
				{ID: "claude-opus-4-6", Name: "Claude Opus 4.6", Provider: "anthropic", Context: 200000, InputCost: 15.00, OutputCost: 75.00, Type: "chat"},
				{ID: "claude-sonnet-4-5-20250929", Name: "Claude Sonnet 4.5", Provider: "anthropic", Context: 200000, InputCost: 3.00, OutputCost: 15.00, Type: "chat"},
				{ID: "claude-haiku-4-5-20251001", Name: "Claude Haiku 4.5", Provider: "anthropic", Context: 200000, InputCost: 0.80, OutputCost: 4.00, Type: "chat"},
			},
		},
		{
			Name:     "Google",
			Endpoint: "https://generativelanguage.googleapis.com/v1beta",
			EnvKey:   "GOOGLE_API_KEY",
			Models: []Model{
				{ID: "gemini-2.5-pro", Name: "Gemini 2.5 Pro", Provider: "google", Context: 1048576, InputCost: 1.25, OutputCost: 10.00, Type: "chat"},
				{ID: "gemini-2.5-flash", Name: "Gemini 2.5 Flash", Provider: "google", Context: 1048576, InputCost: 0.15, OutputCost: 0.60, Type: "chat"},
				{ID: "gemini-2.0-flash", Name: "Gemini 2.0 Flash", Provider: "google", Context: 1048576, InputCost: 0.10, OutputCost: 0.40, Type: "chat"},
			},
		},
		{
			Name:     "Ollama",
			Endpoint: "http://localhost:11434",
			EnvKey:   "",
			Models: []Model{
				{ID: "llama3.3", Name: "Llama 3.3 70B", Provider: "ollama", Context: 131072, Type: "chat"},
				{ID: "qwen3", Name: "Qwen 3", Provider: "ollama", Context: 40960, Type: "chat"},
				{ID: "deepseek-r1", Name: "DeepSeek R1", Provider: "ollama", Context: 131072, Type: "chat"},
				{ID: "mistral", Name: "Mistral 7B", Provider: "ollama", Context: 32768, Type: "chat"},
				{ID: "codellama", Name: "Code Llama", Provider: "ollama", Context: 16384, Type: "chat"},
				{ID: "nomic-embed-text", Name: "Nomic Embed", Provider: "ollama", Context: 8192, Type: "embedding"},
			},
		},
		{
			Name:     "Groq",
			Endpoint: "https://api.groq.com/openai/v1",
			EnvKey:   "GROQ_API_KEY",
			Models: []Model{
				{ID: "llama-3.3-70b-versatile", Name: "Llama 3.3 70B", Provider: "groq", Context: 128000, InputCost: 0.59, OutputCost: 0.79, Type: "chat"},
				{ID: "deepseek-r1-distill-llama-70b", Name: "DeepSeek R1 70B", Provider: "groq", Context: 128000, InputCost: 0.75, OutputCost: 0.99, Type: "chat"},
			},
		},
		{
			Name:     "Mistral",
			Endpoint: "https://api.mistral.ai/v1",
			EnvKey:   "MISTRAL_API_KEY",
			Models: []Model{
				{ID: "mistral-large-latest", Name: "Mistral Large", Provider: "mistral", Context: 128000, InputCost: 2.00, OutputCost: 6.00, Type: "chat"},
				{ID: "codestral-latest", Name: "Codestral", Provider: "mistral", Context: 256000, InputCost: 0.30, OutputCost: 0.90, Type: "chat"},
			},
		},
	}
}

// AllModels returns a flat list of all known models.
func AllModels() []Model {
	var all []Model
	for _, p := range BuiltinProviders() {
		all = append(all, p.Models...)
	}
	return all
}

// FindModel searches for a model by ID prefix.
func FindModel(query string) *Model {
	for _, m := range AllModels() {
		if m.ID == query {
			return &m
		}
	}
	// Fuzzy: prefix match
	for _, m := range AllModels() {
		if len(query) >= 3 && len(m.ID) >= len(query) && m.ID[:len(query)] == query {
			return &m
		}
	}
	return nil
}

// FormatContext returns a human-readable context window size.
func FormatContext(ctx int) string {
	if ctx >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(ctx)/1000000)
	}
	return fmt.Sprintf("%dk", ctx/1000)
}

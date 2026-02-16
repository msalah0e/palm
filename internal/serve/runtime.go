package serve

import (
	"fmt"
	"os/exec"
)

// Runtime represents a local LLM runtime.
type Runtime struct {
	Name    string // ollama, llama-cpp, vllm
	Path    string // binary path
	Version string
}

// DetectRuntime finds the best available LLM runtime.
func DetectRuntime() *Runtime {
	// Priority: ollama > llama-server > vllm
	if path, err := exec.LookPath("ollama"); err == nil {
		ver := ""
		if out, err := exec.Command(path, "--version").Output(); err == nil {
			ver = extractVersion(string(out))
		}
		return &Runtime{Name: "ollama", Path: path, Version: ver}
	}

	if path, err := exec.LookPath("llama-server"); err == nil {
		return &Runtime{Name: "llama-cpp", Path: path}
	}

	if path, err := exec.LookPath("llama-cpp"); err == nil {
		return &Runtime{Name: "llama-cpp", Path: path}
	}

	if path, err := exec.LookPath("vllm"); err == nil {
		return &Runtime{Name: "vllm", Path: path}
	}

	return nil
}

// Start launches the runtime with the given model.
func (r *Runtime) Start(model string, gpu bool) *exec.Cmd {
	switch r.Name {
	case "ollama":
		return exec.Command(r.Path, "run", model)
	case "llama-cpp":
		args := []string{"--model", model}
		if gpu {
			args = append(args, "--n-gpu-layers", "999")
		}
		return exec.Command(r.Path, args...)
	case "vllm":
		return exec.Command(r.Path, "serve", model)
	}
	return nil
}

// Pull downloads a model using the runtime.
func (r *Runtime) Pull(model string) *exec.Cmd {
	switch r.Name {
	case "ollama":
		return exec.Command(r.Path, "pull", model)
	default:
		return nil
	}
}

// ListModels returns available models from the runtime.
func (r *Runtime) ListModels() *exec.Cmd {
	switch r.Name {
	case "ollama":
		return exec.Command(r.Path, "list")
	default:
		return nil
	}
}

// IsRunning checks if the runtime is currently serving.
func (r *Runtime) IsRunning() bool {
	switch r.Name {
	case "ollama":
		cmd := exec.Command(r.Path, "list")
		return cmd.Run() == nil
	}
	return false
}

func extractVersion(s string) string {
	// Simple extraction: find version-like pattern
	for _, word := range splitWords(s) {
		if len(word) > 0 && (word[0] >= '0' && word[0] <= '9') {
			return word
		}
	}
	return ""
}

func splitWords(s string) []string {
	var words []string
	word := ""
	for _, c := range s {
		if c == ' ' || c == '\n' || c == '\t' || c == ':' {
			if word != "" {
				words = append(words, word)
				word = ""
			}
		} else {
			word += string(c)
		}
	}
	if word != "" {
		words = append(words, word)
	}
	return words
}

// String returns a display name for the runtime.
func (r *Runtime) String() string {
	if r.Version != "" {
		return fmt.Sprintf("%s %s", r.Name, r.Version)
	}
	return r.Name
}

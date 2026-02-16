package proxy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveProvider(t *testing.T) {
	srv := New(Config{Port: 4778})

	tests := []struct {
		path     string
		provider string
		target   string
	}{
		{"/openai/v1/chat/completions", "openai", "https://api.openai.com"},
		{"/anthropic/v1/messages", "anthropic", "https://api.anthropic.com"},
		{"/google/v1beta/models", "google", "https://generativelanguage.googleapis.com"},
		{"/ollama/api/generate", "ollama", "http://localhost:11434"},
		{"/unknown/path", "", ""},
	}

	for _, tt := range tests {
		provider, target, _ := srv.resolveProvider(tt.path)
		if provider != tt.provider {
			t.Errorf("resolveProvider(%q) provider = %q, want %q", tt.path, provider, tt.provider)
		}
		if target != tt.target {
			t.Errorf("resolveProvider(%q) target = %q, want %q", tt.path, target, tt.target)
		}
	}
}

func TestPidFile(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	pidFile := PidFile()
	if pidFile == "" {
		t.Error("PidFile should not be empty")
	}
	if filepath.Ext(pidFile) != ".pid" {
		t.Errorf("PidFile should end in .pid, got %s", pidFile)
	}
}

func TestIsRunningNoFile(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	running, pid := IsRunning()
	if running {
		t.Error("should not be running without PID file")
	}
	if pid != 0 {
		t.Errorf("expected pid 0, got %d", pid)
	}
}

func TestReadLogsEmpty(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	logs, err := ReadLogs(10)
	if err != nil {
		t.Fatalf("ReadLogs failed: %v", err)
	}
	if len(logs) != 0 {
		t.Errorf("expected 0 logs, got %d", len(logs))
	}
}

func TestReadLogs(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	// Write some log entries
	logPath := filepath.Join(dir, "palm", "proxy.jsonl")
	_ = os.MkdirAll(filepath.Dir(logPath), 0o755)

	f, _ := os.Create(logPath)
	for i := 0; i < 5; i++ {
		entry := RequestLog{
			Method:   "POST",
			Path:     "/v1/chat/completions",
			Provider: "openai",
			Status:   200,
			Duration: 100,
		}
		_ = json.NewEncoder(f).Encode(entry)
	}
	f.Close()

	logs, err := ReadLogs(3)
	if err != nil {
		t.Fatalf("ReadLogs failed: %v", err)
	}
	if len(logs) != 3 {
		t.Errorf("expected 3 logs, got %d", len(logs))
	}
}

func TestProviderRoutes(t *testing.T) {
	// Verify all expected providers are routed
	expected := []string{"openai", "anthropic", "google", "groq", "mistral", "ollama"}
	for _, name := range expected {
		prefix := "/" + name + "/"
		if _, ok := providerRoutes[prefix]; !ok {
			t.Errorf("missing route for provider: %s", name)
		}
	}
}

func TestProviderKeys(t *testing.T) {
	// Ollama should not have a key
	if _, ok := providerKeys["ollama"]; ok {
		t.Error("ollama should not have an API key requirement")
	}

	// Others should
	for _, name := range []string{"openai", "anthropic", "google"} {
		if _, ok := providerKeys[name]; !ok {
			t.Errorf("missing key for provider: %s", name)
		}
	}
}

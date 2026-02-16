package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if !cfg.UI.Emoji {
		t.Error("default emoji should be true")
	}
	if !cfg.UI.Color {
		t.Error("default color should be true")
	}
	if !cfg.Install.PreferUV {
		t.Error("default prefer_uv should be true")
	}
	if cfg.Stats.Enabled {
		t.Error("default stats should be disabled")
	}
	if cfg.Vault.Backend != "auto" {
		t.Errorf("expected vault backend 'auto', got %q", cfg.Vault.Backend)
	}
	if !cfg.Parallel.Enabled {
		t.Error("default parallel should be enabled")
	}
	if cfg.Parallel.Concurrency != 4 {
		t.Errorf("expected concurrency 4, got %d", cfg.Parallel.Concurrency)
	}
}

func TestConfigDir(t *testing.T) {
	// Test with XDG_CONFIG_HOME set
	t.Setenv("XDG_CONFIG_HOME", "/tmp/test-xdg")
	dir := ConfigDir()
	if dir != "/tmp/test-xdg/palm" {
		t.Errorf("expected /tmp/test-xdg/palm, got %q", dir)
	}

	// Test without XDG_CONFIG_HOME
	t.Setenv("XDG_CONFIG_HOME", "")
	dir = ConfigDir()
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "palm")
	if dir != expected {
		t.Errorf("expected %q, got %q", expected, dir)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg := Default()
	cfg.Parallel.Concurrency = 8
	cfg.Install.PreferUV = false

	if err := Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded := Load()
	if loaded.Parallel.Concurrency != 8 {
		t.Errorf("expected concurrency 8, got %d", loaded.Parallel.Concurrency)
	}
	if loaded.Install.PreferUV {
		t.Error("expected prefer_uv false after load")
	}
}

func TestEnsureExists(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	if err := EnsureExists(); err != nil {
		t.Fatalf("EnsureExists failed: %v", err)
	}

	path := filepath.Join(tmpDir, "palm", "config.toml")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("config file not created: %v", err)
	}

	// Second call should be no-op
	if err := EnsureExists(); err != nil {
		t.Fatalf("EnsureExists second call failed: %v", err)
	}
}

func TestSetupConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg := Default()
	if cfg.Setup.Complete {
		t.Error("default setup.complete should be false")
	}
	if cfg.Setup.Preset != "" {
		t.Error("default setup.preset should be empty")
	}

	cfg.Setup.Complete = true
	cfg.Setup.Preset = "essentials"

	if err := Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded := Load()
	if !loaded.Setup.Complete {
		t.Error("expected setup.complete true after load")
	}
	if loaded.Setup.Preset != "essentials" {
		t.Errorf("expected setup.preset 'essentials', got %q", loaded.Setup.Preset)
	}
}

func TestFindProjectConfig(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "a", "b", "c")
	os.MkdirAll(subDir, 0o755)

	// Write .palm.toml in the root tmpDir
	os.WriteFile(filepath.Join(tmpDir, ".palm.toml"), []byte("[install]\nprefer_uv = false\n"), 0o644)

	// Change to the deep subdirectory
	origDir, _ := os.Getwd()
	os.Chdir(subDir)
	defer os.Chdir(origDir)

	found := findProjectConfig()
	// Resolve symlinks (macOS /var -> /private/var)
	expectedResolved, _ := filepath.EvalSymlinks(filepath.Join(tmpDir, ".palm.toml"))
	foundResolved, _ := filepath.EvalSymlinks(found)
	if foundResolved != expectedResolved {
		t.Errorf("expected %q, got %q", expectedResolved, foundResolved)
	}
}

package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds tamr configuration.
type Config struct {
	UI       UIConfig       `toml:"ui"`
	Stats    StatsConfig    `toml:"stats"`
	Install  InstallConfig  `toml:"install"`
	Keys     KeysConfig     `toml:"keys"`
	Vault    VaultConfig    `toml:"vault"`
	Parallel ParallelConfig `toml:"parallel"`
	Hooks    HooksConfig    `toml:"hooks"`
}

// UIConfig controls display options.
type UIConfig struct {
	Emoji bool `toml:"emoji"`
	Color bool `toml:"color"`
}

// StatsConfig controls usage tracking.
type StatsConfig struct {
	Enabled bool `toml:"enabled"`
}

// InstallConfig controls installation behavior.
type InstallConfig struct {
	PreferUV     bool `toml:"prefer_uv"`
	CleanupAfter bool `toml:"cleanup_after"`
}

// KeysConfig controls API key behavior.
type KeysConfig struct {
	AutoExport bool `toml:"auto_export"`
}

// VaultConfig controls vault backend selection.
type VaultConfig struct {
	Backend string `toml:"backend"` // "auto", "keychain", "file"
}

// ParallelConfig controls concurrent execution.
type ParallelConfig struct {
	Enabled     bool `toml:"enabled"`
	Concurrency int  `toml:"concurrency"`
}

// HooksConfig defines lifecycle hook scripts.
type HooksConfig struct {
	PreInstall  string `toml:"pre_install"`
	PostInstall string `toml:"post_install"`
	PreRun      string `toml:"pre_run"`
	PostRun     string `toml:"post_run"`
	PreUpdate   string `toml:"pre_update"`
	PostUpdate  string `toml:"post_update"`
}

// Default returns the default configuration.
func Default() *Config {
	return &Config{
		UI:       UIConfig{Emoji: true, Color: true},
		Stats:    StatsConfig{Enabled: false},
		Install:  InstallConfig{PreferUV: true, CleanupAfter: false},
		Keys:     KeysConfig{AutoExport: false},
		Vault:    VaultConfig{Backend: "auto"},
		Parallel: ParallelConfig{Enabled: true, Concurrency: 4},
	}
}

// ConfigDir returns the tamr config directory path.
func ConfigDir() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "tamr")
}

func configPath() string {
	return filepath.Join(ConfigDir(), "config.toml")
}

// Load reads the config file, creating defaults if it doesn't exist.
func Load() *Config {
	cfg := Default()
	path := configPath()

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}

	_ = toml.Unmarshal(data, cfg)
	return cfg
}

// Save writes the config to disk.
func Save(cfg *Config) error {
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return toml.NewEncoder(f).Encode(cfg)
}

// EnsureExists creates the config file with defaults if it doesn't exist.
func EnsureExists() error {
	path := configPath()
	if _, err := os.Stat(path); err == nil {
		return nil // already exists
	}
	return Save(Default())
}

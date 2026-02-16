package state

import (
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

// InstalledTool tracks metadata about an installed tool.
type InstalledTool struct {
	Version     string    `toml:"version"`
	Backend     string    `toml:"backend"`
	Package     string    `toml:"package"`
	InstalledAt time.Time `toml:"installed_at"`
	UpdatedAt   time.Time `toml:"updated_at"`
	Path        string    `toml:"path"`
}

// State tracks all tamr-managed installations.
type State struct {
	Installed map[string]InstalledTool `toml:"installed"`
}

func statePath() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "tamr", "state.toml")
}

// Load reads the state file, returning empty state if it doesn't exist.
func Load() *State {
	s := &State{Installed: make(map[string]InstalledTool)}
	data, err := os.ReadFile(statePath())
	if err != nil {
		return s
	}
	_ = toml.Unmarshal(data, s)
	if s.Installed == nil {
		s.Installed = make(map[string]InstalledTool)
	}
	return s
}

// Save writes the state file to disk.
func Save(s *State) error {
	path := statePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(s)
}

// Record adds or updates an installed tool in the state.
func Record(name, version, backend, pkg, path string) error {
	s := Load()
	now := time.Now()
	existing, exists := s.Installed[name]
	if exists {
		existing.Version = version
		existing.Backend = backend
		existing.Package = pkg
		existing.Path = path
		existing.UpdatedAt = now
		s.Installed[name] = existing
	} else {
		s.Installed[name] = InstalledTool{
			Version:     version,
			Backend:     backend,
			Package:     pkg,
			InstalledAt: now,
			UpdatedAt:   now,
			Path:        path,
		}
	}
	return Save(s)
}

// Remove deletes a tool from the state.
func Remove(name string) error {
	s := Load()
	delete(s.Installed, name)
	return Save(s)
}

// IsInstalled checks if a tool is tracked in state.
func IsInstalled(name string) bool {
	s := Load()
	_, ok := s.Installed[name]
	return ok
}

// ListInstalled returns all tracked tool names.
func ListInstalled() map[string]InstalledTool {
	return Load().Installed
}

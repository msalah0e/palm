package stats

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Entry represents a single usage event.
type Entry struct {
	Timestamp time.Time `json:"ts"`
	Command   string    `json:"cmd"`
	Tool      string    `json:"tool,omitempty"`
	Package   string    `json:"pkg,omitempty"`
	OK        bool      `json:"ok"`
}

// Summary holds aggregated stats.
type Summary struct {
	TotalCommands  int
	AICommands     int
	BrewCommands   int
	ToolsInstalled int
	LastUsed       time.Time
}

func historyPath() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "tamr", "history.jsonl")
}

// Record appends an entry to the history file.
func Record(cmd, tool, pkg string, ok bool) error {
	path := historyPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	entry := Entry{
		Timestamp: time.Now(),
		Command:   cmd,
		Tool:      tool,
		Package:   pkg,
		OK:        ok,
	}
	return json.NewEncoder(f).Encode(entry)
}

// Summarize reads history and returns aggregated stats.
func Summarize() (*Summary, error) {
	path := historyPath()
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Summary{}, nil
		}
		return nil, err
	}
	defer f.Close()

	s := &Summary{}
	installed := make(map[string]bool)
	dec := json.NewDecoder(f)
	for dec.More() {
		var e Entry
		if err := dec.Decode(&e); err != nil {
			continue
		}
		s.TotalCommands++
		if e.Timestamp.After(s.LastUsed) {
			s.LastUsed = e.Timestamp
		}
		if len(e.Command) >= 2 && e.Command[:2] == "ai" {
			s.AICommands++
		} else {
			s.BrewCommands++
		}
		if e.Command == "ai install" && e.OK && e.Tool != "" {
			installed[e.Tool] = true
		}
	}
	s.ToolsInstalled = len(installed)
	return s, nil
}

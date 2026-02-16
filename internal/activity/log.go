package activity

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Entry represents a single activity log entry.
type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	Tool      string    `json:"tool,omitempty"`
	Details   string    `json:"details,omitempty"`
	Cost      float64   `json:"cost,omitempty"`
	Duration  float64   `json:"duration,omitempty"`
}

func logPath() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "palm", "activity.jsonl")
}

// Log appends an entry to the activity log.
func Log(action, tool, details string) error {
	entry := Entry{
		Timestamp: time.Now(),
		Action:    action,
		Tool:      tool,
		Details:   details,
	}
	return append_entry(entry)
}

// LogWithCost appends an entry with cost information.
func LogWithCost(action, tool, details string, cost, duration float64) error {
	entry := Entry{
		Timestamp: time.Now(),
		Action:    action,
		Tool:      tool,
		Details:   details,
		Cost:      cost,
		Duration:  duration,
	}
	return append_entry(entry)
}

func append_entry(entry Entry) error {
	path := logPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	data, _ := json.Marshal(entry)
	_, err = fmt.Fprintf(f, "%s\n", data)
	return err
}

// Read returns the last N entries from the log.
func Read(count int) ([]Entry, error) {
	data, err := os.ReadFile(logPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var entries []Entry
	for _, line := range splitLines(string(data)) {
		if line == "" {
			continue
		}
		var e Entry
		if json.Unmarshal([]byte(line), &e) == nil {
			entries = append(entries, e)
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})

	if count > 0 && len(entries) > count {
		entries = entries[:count]
	}
	return entries, nil
}

// Search finds entries matching a query.
func Search(query string, count int) ([]Entry, error) {
	all, err := Read(0)
	if err != nil {
		return nil, err
	}

	var results []Entry
	for _, e := range all {
		if contains(e.Action, query) || contains(e.Tool, query) || contains(e.Details, query) {
			results = append(results, e)
			if count > 0 && len(results) >= count {
				break
			}
		}
	}
	return results, nil
}

// Clear removes all log entries.
func Clear() error {
	return os.Remove(logPath())
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func contains(s, sub string) bool {
	if sub == "" {
		return true
	}
	sl := len(s)
	subl := len(sub)
	if sl < subl {
		return false
	}
	for i := 0; i <= sl-subl; i++ {
		match := true
		for j := 0; j < subl; j++ {
			sc := s[i+j]
			qc := sub[j]
			// Case-insensitive
			if sc >= 'A' && sc <= 'Z' {
				sc += 32
			}
			if qc >= 'A' && qc <= 'Z' {
				qc += 32
			}
			if sc != qc {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

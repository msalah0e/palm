package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Session tracks a single tool run session.
type Session struct {
	ID        string    `json:"id"`
	Tool      string    `json:"tool"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at,omitempty"`
	Duration  float64   `json:"duration_secs,omitempty"`
	ExitCode  int       `json:"exit_code,omitempty"`
	Cost      float64   `json:"cost,omitempty"`
	Tokens    int64     `json:"tokens,omitempty"`
	Provider  string    `json:"provider,omitempty"`
}

// Summary aggregates session data.
type Summary struct {
	TotalSessions int
	TotalDuration time.Duration
	TotalCost     float64
	TotalTokens   int64
	ByTool        map[string]ToolSummary
}

// ToolSummary tracks per-tool session metrics.
type ToolSummary struct {
	Sessions int
	Duration time.Duration
	Cost     float64
	Tokens   int64
}

func sessionsPath() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "palm", "sessions.jsonl")
}

// Start records a new session start and returns its ID.
func Start(tool string) (*Session, error) {
	s := &Session{
		ID:        time.Now().Format("20060102-150405"),
		Tool:      tool,
		StartedAt: time.Now(),
	}
	return s, nil
}

// End finalizes a session and writes it to disk.
func End(s *Session, exitCode int) error {
	s.EndedAt = time.Now()
	s.Duration = s.EndedAt.Sub(s.StartedAt).Seconds()
	s.ExitCode = exitCode
	return save(s)
}

// Record saves a complete session entry.
func Record(tool string, duration time.Duration, exitCode int, cost float64, tokens int64, provider string) error {
	s := &Session{
		ID:        time.Now().Format("20060102-150405"),
		Tool:      tool,
		StartedAt: time.Now().Add(-duration),
		EndedAt:   time.Now(),
		Duration:  duration.Seconds(),
		ExitCode:  exitCode,
		Cost:      cost,
		Tokens:    tokens,
		Provider:  provider,
	}
	return save(s)
}

func save(s *Session) error {
	path := sessionsPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(s)
}

// List returns the most recent n sessions.
func List(n int) ([]Session, error) {
	path := sessionsPath()
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var all []Session
	dec := json.NewDecoder(f)
	for dec.More() {
		var s Session
		if err := dec.Decode(&s); err != nil {
			continue
		}
		all = append(all, s)
	}

	// Return last n
	if n > 0 && len(all) > n {
		all = all[len(all)-n:]
	}

	// Reverse for most-recent-first
	for i, j := 0, len(all)-1; i < j; i, j = i+1, j-1 {
		all[i], all[j] = all[j], all[i]
	}

	return all, nil
}

// Summarize aggregates all session data.
func Summarize() (*Summary, error) {
	sessions, err := List(0)
	if err != nil {
		return nil, err
	}

	s := &Summary{
		ByTool: make(map[string]ToolSummary),
	}

	for _, sess := range sessions {
		s.TotalSessions++
		dur := time.Duration(sess.Duration * float64(time.Second))
		s.TotalDuration += dur
		s.TotalCost += sess.Cost
		s.TotalTokens += sess.Tokens

		ts := s.ByTool[sess.Tool]
		ts.Sessions++
		ts.Duration += dur
		ts.Cost += sess.Cost
		ts.Tokens += sess.Tokens
		s.ByTool[sess.Tool] = ts
	}

	return s, nil
}

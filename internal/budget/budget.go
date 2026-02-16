package budget

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/msalah0e/palm/internal/session"
)

// Budget defines spending limits.
type Budget struct {
	MonthlyLimit float64            `toml:"monthly_limit"`
	DailyLimit   float64            `toml:"daily_limit"`
	AlertAt      float64            `toml:"alert_at"` // percentage (0.8 = 80%)
	PerTool      map[string]float64 `toml:"per_tool"` // per-tool monthly limits
}

// Status represents current budget status.
type Status struct {
	MonthlyLimit float64
	MonthlySpend float64
	DailyLimit   float64
	DailySpend   float64
	PercentUsed  float64
	IsOverBudget bool
	IsNearBudget bool
	ByTool       map[string]float64
	ByProvider   map[string]float64
	TotalTokens  int64
	CurrentMonth string
}

func budgetPath() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "palm", "budget.toml")
}

// Load reads the budget configuration.
func Load() *Budget {
	b := &Budget{
		AlertAt: 0.8,
		PerTool: make(map[string]float64),
	}
	data, err := os.ReadFile(budgetPath())
	if err != nil {
		return b
	}
	_ = toml.Unmarshal(data, b)
	if b.PerTool == nil {
		b.PerTool = make(map[string]float64)
	}
	return b
}

// Save writes the budget configuration.
func Save(b *Budget) error {
	path := budgetPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(b)
}

// GetStatus computes current budget status from session history.
func GetStatus() (*Status, error) {
	b := Load()

	sessions, err := session.List(0)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	s := &Status{
		MonthlyLimit: b.MonthlyLimit,
		DailyLimit:   b.DailyLimit,
		ByTool:       make(map[string]float64),
		ByProvider:   make(map[string]float64),
		CurrentMonth: now.Format("January 2006"),
	}

	for _, sess := range sessions {
		if sess.StartedAt.After(monthStart) {
			s.MonthlySpend += sess.Cost
			s.TotalTokens += sess.Tokens
			s.ByTool[sess.Tool] += sess.Cost
			if sess.Provider != "" {
				s.ByProvider[sess.Provider] += sess.Cost
			}
		}
		if sess.StartedAt.After(dayStart) {
			s.DailySpend += sess.Cost
		}
	}

	if b.MonthlyLimit > 0 {
		s.PercentUsed = (s.MonthlySpend / b.MonthlyLimit) * 100
		s.IsOverBudget = s.MonthlySpend >= b.MonthlyLimit
		s.IsNearBudget = s.MonthlySpend >= b.MonthlyLimit*b.AlertAt
	}

	return s, nil
}

// CheckBudget returns an error if the budget would be exceeded.
func CheckBudget(tool string) error {
	b := Load()
	if b.MonthlyLimit == 0 && b.DailyLimit == 0 {
		return nil // no budget set
	}

	status, err := GetStatus()
	if err != nil {
		return nil // don't block on error
	}

	if status.IsOverBudget {
		return fmt.Errorf("monthly budget exceeded ($%.2f / $%.2f)", status.MonthlySpend, status.MonthlyLimit)
	}

	if b.DailyLimit > 0 && status.DailySpend >= b.DailyLimit {
		return fmt.Errorf("daily budget exceeded ($%.2f / $%.2f)", status.DailySpend, status.DailyLimit)
	}

	// Check per-tool limit
	if limit, ok := b.PerTool[tool]; ok {
		if spend, ok := status.ByTool[tool]; ok && spend >= limit {
			return fmt.Errorf("tool budget exceeded for %s ($%.2f / $%.2f)", tool, spend, limit)
		}
	}

	return nil
}

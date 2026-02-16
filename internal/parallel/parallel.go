package parallel

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/msalah0e/palm/internal/ui"
	"golang.org/x/sync/errgroup"
)

// Result holds the outcome of a parallel task.
type Result struct {
	Name    string
	OK      bool
	Err     error
	Output  string
	Elapsed time.Duration
}

// Task is a function that runs in parallel.
type Task struct {
	Name string
	Fn   func() (string, error)
}

// Run executes tasks in parallel with the given concurrency limit.
// Returns results in the order tasks were submitted.
func Run(tasks []Task, concurrency int) []Result {
	if concurrency < 1 {
		concurrency = 4
	}

	results := make([]Result, len(tasks))
	var mu sync.Mutex

	g, _ := errgroup.WithContext(context.Background())
	g.SetLimit(concurrency)

	for i, task := range tasks {
		i, task := i, task
		g.Go(func() error {
			start := time.Now()

			mu.Lock()
			fmt.Printf("  %s %s...\n", ui.Subtle.Sprint("⟳"), task.Name)
			mu.Unlock()

			output, err := task.Fn()
			elapsed := time.Since(start)

			mu.Lock()
			if err != nil {
				results[i] = Result{Name: task.Name, OK: false, Err: err, Output: output, Elapsed: elapsed}
				fmt.Printf("  %s %s %s\n", ui.StatusIcon(false), task.Name, ui.Bad.Sprintf("(%v)", err))
				// Show truncated output to help diagnose failures
				if output = strings.TrimSpace(output); output != "" {
					for _, line := range truncateLines(output, 5) {
						fmt.Printf("      %s\n", ui.Subtle.Sprint(line))
					}
				}
			} else {
				results[i] = Result{Name: task.Name, OK: true, Output: output, Elapsed: elapsed}
				fmt.Printf("  %s %s %s\n", ui.StatusIcon(true), task.Name, ui.Subtle.Sprintf("%.1fs", elapsed.Seconds()))
			}
			mu.Unlock()

			return nil // never fail the group — collect results instead
		})
	}

	_ = g.Wait()
	return results
}

// truncateLines splits text into lines and returns at most n lines.
func truncateLines(s string, n int) []string {
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return lines
	}
	out := lines[:n]
	out = append(out, fmt.Sprintf("... (%d more lines)", len(lines)-n))
	return out
}

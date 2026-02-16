package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/msalah0e/palm/internal/vault"
	"github.com/spf13/cobra"
)

// ComposeFile represents a .palm-compose.toml workflow definition.
type ComposeFile struct {
	Name        string        `toml:"name"`
	Description string        `toml:"description"`
	Steps       []ComposeStep `toml:"steps"`
}

// ComposeStep is a single step in a compose workflow.
type ComposeStep struct {
	Name      string   `toml:"name"`
	Run       string   `toml:"run"`
	Tool      string   `toml:"tool"`
	Args      []string `toml:"args"`
	Input     string   `toml:"input"`
	DependsOn []string `toml:"depends_on"`
	OnFail    string   `toml:"on_fail"` // continue, stop (default: stop)
	Timeout   int      `toml:"timeout"` // seconds, 0 = no timeout
}

// ComposeResult holds the result of running a step.
type ComposeResult struct {
	Step     string
	Output   string
	Duration time.Duration
	ExitCode int
	Error    string
}

func composeCmd() *cobra.Command {
	var (
		file    string
		dryRun  bool
		verbose bool
	)

	cmd := &cobra.Command{
		Use:   "compose [--file workflow.toml]",
		Short: "Run multi-tool AI workflows from a TOML file",
		Long: `Compose runs multi-step AI workflows defined in TOML files.
Each step can use a different AI tool, pass data between steps,
and declare dependencies for parallel execution.

Think of it as Docker Compose for AI tools.

Examples:
  palm compose                              # Run .palm-compose.toml
  palm compose --file review.toml           # Run a specific workflow
  palm compose init                         # Create a sample workflow
  palm compose --dry-run                    # Show what would run

Workflow file (.palm-compose.toml):
  name = "code-review"
  description = "Multi-tool code review pipeline"

  [[steps]]
  name = "analyze"
  run = "cat src/main.py"

  [[steps]]
  name = "review"
  tool = "ollama"
  args = ["run", "llama3.3"]
  input = "step:analyze"
  depends_on = ["analyze"]

  [[steps]]
  name = "test"
  run = "go test ./..."
  depends_on = ["review"]`,
		Aliases: []string{"workflow"},
		Run: func(cmd *cobra.Command, args []string) {
			// Handle "compose init" subcommand
			if len(args) > 0 && args[0] == "init" {
				composeInit()
				return
			}

			// Load workflow file
			workflow, err := loadComposeFile(file)
			if err != nil {
				ui.Bad.Printf("  Failed to load workflow: %v\n", err)
				os.Exit(1)
			}

			ui.Banner("compose")
			if workflow.Name != "" {
				fmt.Printf("  Workflow: %s\n", ui.Brand.Sprint(workflow.Name))
			}
			if workflow.Description != "" {
				fmt.Printf("  %s\n", ui.Subtle.Sprint(workflow.Description))
			}
			fmt.Printf("  Steps:    %d\n\n", len(workflow.Steps))

			if dryRun {
				composeDryRun(workflow)
				return
			}

			v := vault.New()
			env := buildVaultEnv(v)

			results := runCompose(workflow, env, verbose)

			// Print summary
			fmt.Println()
			fmt.Println("  " + strings.Repeat("═", 60))
			fmt.Printf("  %s Workflow complete\n\n", ui.Brand.Sprint("🌴"))

			headers := []string{"Step", "Time", "Status"}
			var rows [][]string
			allPassed := true

			for _, r := range results {
				status := ui.StatusIcon(true) + " ok"
				dur := fmt.Sprintf("%.2fs", r.Duration.Seconds())
				if r.Error != "" {
					status = ui.StatusIcon(false) + " " + r.Error
					allPassed = false
				}
				rows = append(rows, []string{r.Step, dur, status})
			}

			ui.Table(headers, rows)

			if !allPassed {
				fmt.Println()
				ui.Bad.Println("  Some steps failed")
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", ".palm-compose.toml", "Workflow file path")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would run without executing")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show step output")
	return cmd
}

func composeInit() {
	const sampleWorkflow = `# palm compose workflow
# Run with: palm compose

name = "code-review"
description = "Multi-tool code review pipeline"

# Step 1: Read the source file
[[steps]]
name = "read-code"
run = "cat main.go"

# Step 2: AI reviews the code (depends on step 1)
[[steps]]
name = "ai-review"
tool = "ollama"
args = ["run", "llama3.3", "Review this Go code for bugs and improvements:"]
input = "step:read-code"
depends_on = ["read-code"]

# Step 3: Run tests in parallel with review (no dependency on review)
[[steps]]
name = "run-tests"
run = "go test ./..."
timeout = 60

# Step 4: Generate summary after both review and tests
[[steps]]
name = "summary"
tool = "ollama"
args = ["run", "llama3.3", "Summarize the code review and test results:"]
input = "step:ai-review,step:run-tests"
depends_on = ["ai-review", "run-tests"]
`

	path := ".palm-compose.toml"
	if _, err := os.Stat(path); err == nil {
		ui.Warn.Printf("  %s already exists\n", path)
		return
	}

	if err := os.WriteFile(path, []byte(sampleWorkflow), 0644); err != nil {
		ui.Bad.Printf("  Failed to create %s: %v\n", path, err)
		os.Exit(1)
	}

	ui.Good.Printf("  Created %s\n", path)
	fmt.Println("  Edit it, then run: palm compose")
}

func loadComposeFile(file string) (*ComposeFile, error) {
	// Search up for the file
	path := file
	if !filepath.IsAbs(path) {
		dir, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		for {
			candidate := filepath.Join(dir, file)
			if _, err := os.Stat(candidate); err == nil {
				path = candidate
				break
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				return nil, fmt.Errorf("%s not found (run 'palm compose init' to create one)", file)
			}
			dir = parent
		}
	}

	var cf ComposeFile
	if _, err := toml.DecodeFile(path, &cf); err != nil {
		return nil, err
	}

	// Validate
	stepNames := make(map[string]bool)
	for _, s := range cf.Steps {
		if s.Name == "" {
			return nil, fmt.Errorf("step missing 'name'")
		}
		if s.Run == "" && s.Tool == "" {
			return nil, fmt.Errorf("step '%s': must have 'run' or 'tool'", s.Name)
		}
		if stepNames[s.Name] {
			return nil, fmt.Errorf("duplicate step name: '%s'", s.Name)
		}
		stepNames[s.Name] = true
	}

	// Validate dependencies exist
	for _, s := range cf.Steps {
		for _, dep := range s.DependsOn {
			if !stepNames[dep] {
				return nil, fmt.Errorf("step '%s' depends on unknown step '%s'", s.Name, dep)
			}
		}
	}

	return &cf, nil
}

func composeDryRun(wf *ComposeFile) {
	fmt.Printf("  %s Dry run — showing execution plan\n\n", ui.Info.Sprint("📋"))

	// Build dependency graph
	levels := resolveExecutionOrder(wf)

	for i, level := range levels {
		if len(level) > 1 {
			fmt.Printf("  %s Parallel group %d:\n", ui.Info.Sprint("⚡"), i+1)
		} else {
			fmt.Printf("  %s Step %d:\n", ui.Subtle.Sprint("→"), i+1)
		}

		for _, step := range level {
			cmd := step.Run
			if step.Tool != "" {
				cmd = step.Tool + " " + strings.Join(step.Args, " ")
			}
			fmt.Printf("    %s  %s\n", ui.Brand.Sprint(step.Name), ui.Subtle.Sprint(cmd))
			if step.Input != "" {
				fmt.Printf("           input: %s\n", ui.Info.Sprint(step.Input))
			}
			if len(step.DependsOn) > 0 {
				fmt.Printf("           after: %s\n", strings.Join(step.DependsOn, ", "))
			}
		}
		fmt.Println()
	}
}

// resolveExecutionOrder returns steps grouped into parallel execution levels.
// Steps in the same level have all their dependencies satisfied by prior levels.
func resolveExecutionOrder(wf *ComposeFile) [][]ComposeStep {
	stepMap := make(map[string]ComposeStep)
	for _, s := range wf.Steps {
		stepMap[s.Name] = s
	}

	var levels [][]ComposeStep
	resolved := make(map[string]bool)
	remaining := make(map[string]bool)
	for _, s := range wf.Steps {
		remaining[s.Name] = true
	}

	for len(remaining) > 0 {
		var level []ComposeStep

		for name := range remaining {
			step := stepMap[name]
			allDepsResolved := true
			for _, dep := range step.DependsOn {
				if !resolved[dep] {
					allDepsResolved = false
					break
				}
			}
			if allDepsResolved {
				level = append(level, step)
			}
		}

		if len(level) == 0 {
			// Circular dependency — add all remaining
			for name := range remaining {
				level = append(level, stepMap[name])
			}
			levels = append(levels, level)
			break
		}

		for _, s := range level {
			resolved[s.Name] = true
			delete(remaining, s.Name)
		}
		levels = append(levels, level)
	}

	return levels
}

func runCompose(wf *ComposeFile, env []string, verbose bool) []ComposeResult {
	levels := resolveExecutionOrder(wf)

	// Store outputs by step name for input references
	outputs := make(map[string]string)
	var mu sync.Mutex
	var allResults []ComposeResult

	for levelIdx, level := range levels {
		if len(level) > 1 {
			fmt.Printf("  %s Parallel group %d (%d steps)\n", ui.Info.Sprint("⚡"), levelIdx+1, len(level))
		}

		var wg sync.WaitGroup
		levelResults := make([]ComposeResult, len(level))

		for i, step := range level {
			wg.Add(1)
			go func(idx int, s ComposeStep) {
				defer wg.Done()

				displayName := s.Name
				fmt.Printf("  %s Running %s...\n", ui.Subtle.Sprint("→"), ui.Brand.Sprint(displayName))

				// Resolve input
				var stdinData string
				if s.Input != "" {
					stdinData = resolveInput(s.Input, outputs, &mu)
				}

				result := executeComposeStep(s, env, stdinData, verbose)

				mu.Lock()
				outputs[s.Name] = result.Output
				levelResults[idx] = result
				mu.Unlock()

				if result.Error != "" {
					ui.Bad.Printf("  %s %s failed: %s\n", ui.StatusIcon(false), displayName, result.Error)
				} else {
					fmt.Printf("  %s %s completed in %.2fs\n",
						ui.StatusIcon(true),
						ui.Brand.Sprint(displayName),
						result.Duration.Seconds())
				}

				if verbose && result.Output != "" {
					fmt.Println()
					printTruncatedOutput(result.Output, 500)
					fmt.Println()
				}
			}(i, step)
		}

		wg.Wait()

		// Check for failures
		for _, r := range levelResults {
			allResults = append(allResults, r)

			// Find original step to check on_fail
			for _, s := range level {
				if s.Name == r.Step && r.Error != "" && s.OnFail != "continue" {
					// Stop execution
					return allResults
				}
			}
		}
	}

	return allResults
}

func resolveInput(input string, outputs map[string]string, mu *sync.Mutex) string {
	parts := strings.Split(input, ",")
	var resolved []string

	for _, part := range parts {
		part = strings.TrimSpace(part)

		if strings.HasPrefix(part, "step:") {
			stepName := strings.TrimPrefix(part, "step:")
			mu.Lock()
			if out, ok := outputs[stepName]; ok {
				resolved = append(resolved, out)
			}
			mu.Unlock()
		} else if strings.HasPrefix(part, "file:") {
			filePath := strings.TrimPrefix(part, "file:")
			if data, err := os.ReadFile(filePath); err == nil {
				resolved = append(resolved, string(data))
			}
		} else if strings.HasPrefix(part, "git:") {
			gitCmd := strings.TrimPrefix(part, "git:")
			if gitCmd == "diff" {
				if out, err := exec.Command("git", "diff").Output(); err == nil {
					resolved = append(resolved, string(out))
				}
			} else if gitCmd == "log" {
				if out, err := exec.Command("git", "log", "--oneline", "-10").Output(); err == nil {
					resolved = append(resolved, string(out))
				}
			}
		} else {
			// Literal text
			resolved = append(resolved, part)
		}
	}

	return strings.Join(resolved, "\n\n")
}

func executeComposeStep(step ComposeStep, env []string, stdinData string, verbose bool) ComposeResult {
	var cmdArgs []string

	if step.Run != "" {
		// Shell command
		cmdArgs = []string{"sh", "-c", step.Run}
	} else if step.Tool != "" {
		// Tool with args
		cmdArgs = append([]string{step.Tool}, step.Args...)
	}

	var stdout, stderr bytes.Buffer
	c := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	c.Stdout = &stdout
	c.Stderr = &stderr
	c.Env = env

	if stdinData != "" {
		c.Stdin = strings.NewReader(stdinData)
	}

	start := time.Now()

	if step.Timeout > 0 {
		if err := c.Start(); err != nil {
			return ComposeResult{
				Step:     step.Name,
				Duration: time.Since(start),
				Output:   stderr.String(),
				ExitCode: 1,
				Error:    err.Error(),
			}
		}

		done := make(chan error, 1)
		go func() { done <- c.Wait() }()

		select {
		case err := <-done:
			elapsed := time.Since(start)
			if err != nil {
				return ComposeResult{
					Step:     step.Name,
					Duration: elapsed,
					Output:   stderr.String(),
					ExitCode: 1,
					Error:    err.Error(),
				}
			}
			return ComposeResult{
				Step:     step.Name,
				Duration: elapsed,
				Output:   stdout.String(),
				ExitCode: 0,
			}
		case <-time.After(time.Duration(step.Timeout) * time.Second):
			_ = c.Process.Kill()
			return ComposeResult{
				Step:     step.Name,
				Duration: time.Duration(step.Timeout) * time.Second,
				ExitCode: -1,
				Error:    "timeout",
			}
		}
	}

	err := c.Run()
	elapsed := time.Since(start)

	if err != nil {
		return ComposeResult{
			Step:     step.Name,
			Duration: elapsed,
			Output:   stderr.String(),
			ExitCode: 1,
			Error:    err.Error(),
		}
	}

	return ComposeResult{
		Step:     step.Name,
		Duration: elapsed,
		Output:   stdout.String(),
		ExitCode: 0,
	}
}

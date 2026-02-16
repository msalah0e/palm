package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/msalah0e/palm/internal/registry"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/msalah0e/palm/internal/vault"
	"github.com/spf13/cobra"
)

// SquadResult holds the result from one tool in the squad.
type SquadResult struct {
	Tool     string
	Output   string
	Duration time.Duration
	ExitCode int
	Error    string
}

func squadCmd() *cobra.Command {
	var (
		tools   string
		judge   string
		timeout int
		mode    string
		showAll bool
	)

	cmd := &cobra.Command{
		Use:   `squad "<task>" --tools tool1,tool2 [--judge ollama]`,
		Short: "Run multiple AI tools on the same task, pick the best result",
		Long: `Squad runs the same task through multiple AI tools in parallel,
then optionally uses a "judge" AI to evaluate and pick the best result.

Modes:
  race    First tool to finish wins (default)
  vote    All tools run, judge picks the best
  merge   All tools run, judge merges/synthesizes results
  all     Show all outputs side by side

Examples:
  palm squad "explain quicksort" --tools ollama,aider --mode race
  palm squad "fix the bug in main.py" --tools aider,codex --judge ollama --mode vote
  palm squad "write unit tests" --tools claude-code,aider,codex --mode merge --judge ollama
  palm squad "review this code" --tools ollama,aider --mode all`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			task := args[0]
			toolNames := strings.Split(tools, ",")

			if len(toolNames) < 2 {
				ui.Warn.Println("  Provide at least 2 tools: --tools tool1,tool2")
				os.Exit(1)
			}

			for i := range toolNames {
				toolNames[i] = strings.TrimSpace(toolNames[i])
			}

			// Validate mode
			switch mode {
			case "race", "vote", "merge", "all":
			default:
				ui.Bad.Printf("  Unknown mode: %s (use race, vote, merge, or all)\n", mode)
				os.Exit(1)
			}

			// Judge required for vote and merge modes
			if (mode == "vote" || mode == "merge") && judge == "" {
				ui.Warn.Println("  --judge is required for vote and merge modes")
				fmt.Println("  Example: --judge ollama")
				os.Exit(1)
			}

			ui.Banner("squad")
			fmt.Printf("  Task:  %s\n", ui.Brand.Sprint(task))
			fmt.Printf("  Tools: %s\n", strings.Join(toolNames, ", "))
			fmt.Printf("  Mode:  %s\n", ui.Info.Sprint(mode))
			if judge != "" {
				fmt.Printf("  Judge: %s\n", ui.Info.Sprint(judge))
			}
			fmt.Println()

			reg := loadRegistry()
			v := vault.New()

			// Build environment with all vault keys
			env := buildVaultEnv(v)

			// Run all tools in parallel
			results := runSquad(toolNames, task, reg, env, timeout)

			// Display results based on mode
			switch mode {
			case "race":
				handleRaceMode(results)
			case "all":
				handleAllMode(results, showAll)
			case "vote":
				handleVoteMode(results, judge, task, env, timeout)
			case "merge":
				handleMergeMode(results, judge, task, env, timeout)
			}
		},
	}

	cmd.Flags().StringVar(&tools, "tools", "", "Comma-separated list of tools (required)")
	cmd.Flags().StringVar(&judge, "judge", "", "Tool to judge/merge results (e.g., ollama)")
	cmd.Flags().IntVar(&timeout, "timeout", 60, "Timeout per tool in seconds")
	cmd.Flags().StringVar(&mode, "mode", "race", "Squad mode: race, vote, merge, all")
	cmd.Flags().BoolVar(&showAll, "verbose", false, "Show full output from each tool")
	_ = cmd.MarkFlagRequired("tools")
	return cmd
}

func buildVaultEnv(v vault.Vault) []string {
	env := os.Environ()
	keys, _ := v.List()
	for _, key := range keys {
		if val, err := v.Get(key); err == nil {
			if os.Getenv(key) == "" {
				env = append(env, fmt.Sprintf("%s=%s", key, val))
			}
		}
	}
	return env
}

func runSquad(toolNames []string, task string, reg *registry.Registry, env []string, timeout int) []SquadResult {
	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		results = make([]SquadResult, len(toolNames))
	)

	fmt.Printf("  %s Dispatching to %d tools...\n\n", ui.Info.Sprint("⚡"), len(toolNames))

	for i, name := range toolNames {
		wg.Add(1)
		go func(idx int, toolName string) {
			defer wg.Done()

			tool := reg.Get(toolName)
			bin := toolName
			if tool != nil && tool.Install.Verify.Command != "" {
				parts := strings.Fields(tool.Install.Verify.Command)
				if len(parts) > 0 {
					bin = parts[0]
				}
			}

			displayName := toolName
			if tool != nil {
				displayName = tool.DisplayName
			}

			// Check if tool is installed
			if _, err := exec.LookPath(bin); err != nil {
				mu.Lock()
				results[idx] = SquadResult{
					Tool:     displayName,
					ExitCode: -1,
					Error:    "not installed",
				}
				mu.Unlock()
				return
			}

			// Build command
			var cmdArgs []string
			switch toolName {
			case "ollama":
				cmdArgs = []string{bin, "run", "llama3.3", task}
			default:
				cmdArgs = []string{bin, task}
			}

			var stdout, stderr bytes.Buffer
			c := exec.Command(cmdArgs[0], cmdArgs[1:]...)
			c.Stdout = &stdout
			c.Stderr = &stderr
			c.Env = env
			c.Stdin = strings.NewReader(task)

			start := time.Now()
			if err := c.Start(); err != nil {
				mu.Lock()
				results[idx] = SquadResult{
					Tool:     displayName,
					ExitCode: 1,
					Error:    err.Error(),
				}
				mu.Unlock()
				return
			}

			done := make(chan error, 1)
			go func() { done <- c.Wait() }()

			var result SquadResult
			select {
			case err := <-done:
				elapsed := time.Since(start)
				result = SquadResult{
					Tool:     displayName,
					Duration: elapsed,
					ExitCode: 0,
				}
				if err != nil {
					result.ExitCode = 1
					result.Error = err.Error()
					result.Output = stderr.String()
				} else {
					result.Output = stdout.String()
				}
			case <-time.After(time.Duration(timeout) * time.Second):
				_ = c.Process.Kill()
				result = SquadResult{
					Tool:     displayName,
					Duration: time.Duration(timeout) * time.Second,
					ExitCode: -1,
					Error:    "timeout",
				}
			}

			mu.Lock()
			results[idx] = result
			mu.Unlock()
		}(i, name)
	}

	wg.Wait()
	return results
}

func handleRaceMode(results []SquadResult) {
	fmt.Printf("  %s %s mode — first successful result wins\n\n", ui.Info.Sprint("🏎️"), ui.Brand.Sprint("Race"))

	// Find first successful result (lowest duration)
	var winner *SquadResult
	for i := range results {
		r := &results[i]
		if r.Error == "" && (winner == nil || r.Duration < winner.Duration) {
			winner = r
		}
	}

	// Print summary table
	printSquadSummary(results)

	if winner != nil {
		fmt.Printf("\n  %s Winner: %s (%.2fs)\n", ui.Brand.Sprint("🏆"), ui.Brand.Sprint(winner.Tool), winner.Duration.Seconds())
		if winner.Output != "" {
			fmt.Println()
			fmt.Println("  " + strings.Repeat("─", 60))
			printTruncatedOutput(winner.Output, 2000)
		}
	} else {
		ui.Bad.Println("\n  All tools failed")
	}
}

func handleAllMode(results []SquadResult, verbose bool) {
	fmt.Printf("  %s %s mode — showing all results\n\n", ui.Info.Sprint("📋"), ui.Brand.Sprint("All"))

	printSquadSummary(results)

	fmt.Println()
	for _, r := range results {
		fmt.Printf("  %s %s", "━━━", ui.Brand.Sprint(r.Tool))
		if r.Error != "" {
			ui.Bad.Printf(" (%s)", r.Error)
		} else {
			fmt.Printf(" (%.2fs)", r.Duration.Seconds())
		}
		fmt.Println(" ━━━")

		if r.Output != "" {
			limit := 500
			if verbose {
				limit = 5000
			}
			printTruncatedOutput(r.Output, limit)
		}
		fmt.Println()
	}
}

func handleVoteMode(results []SquadResult, judge, task string, env []string, timeout int) {
	fmt.Printf("  %s %s mode — judge picks the best\n\n", ui.Info.Sprint("🗳️"), ui.Brand.Sprint("Vote"))

	printSquadSummary(results)

	// Collect successful outputs
	var candidates []string
	for i, r := range results {
		if r.Error == "" && r.Output != "" {
			// Truncate long outputs for the judge
			output := r.Output
			if len(output) > 1000 {
				output = output[:1000] + "..."
			}
			candidates = append(candidates, fmt.Sprintf("=== Candidate %d (%s, %.1fs) ===\n%s", i+1, r.Tool, r.Duration.Seconds(), output))
		}
	}

	if len(candidates) < 2 {
		ui.Warn.Println("\n  Need at least 2 successful results for voting")
		return
	}

	// Build judge prompt
	judgePrompt := fmt.Sprintf(`You are judging AI tool outputs. The task was: "%s"

Here are the candidates:

%s

Pick the BEST response. Reply with ONLY:
1. The candidate number (e.g., "Candidate 1")
2. A brief reason why (1 sentence)
3. Then paste the winning output`, task, strings.Join(candidates, "\n\n"))

	fmt.Printf("\n  %s Sending to judge (%s)...\n", ui.Info.Sprint("⚖️"), ui.Brand.Sprint(judge))

	judgeOutput := runJudge(judge, judgePrompt, env, timeout)
	if judgeOutput != "" {
		fmt.Println()
		fmt.Println("  " + strings.Repeat("─", 60))
		fmt.Printf("  %s Judge verdict:\n\n", ui.Brand.Sprint("⚖️"))
		printTruncatedOutput(judgeOutput, 3000)
	} else {
		ui.Bad.Println("  Judge failed to produce output")
	}
}

func handleMergeMode(results []SquadResult, judge, task string, env []string, timeout int) {
	fmt.Printf("  %s %s mode — judge synthesizes all results\n\n", ui.Info.Sprint("🔀"), ui.Brand.Sprint("Merge"))

	printSquadSummary(results)

	// Collect successful outputs
	var contributions []string
	for i, r := range results {
		if r.Error == "" && r.Output != "" {
			output := r.Output
			if len(output) > 1000 {
				output = output[:1000] + "..."
			}
			contributions = append(contributions, fmt.Sprintf("=== From %s (tool %d, %.1fs) ===\n%s", r.Tool, i+1, r.Duration.Seconds(), output))
		}
	}

	if len(contributions) == 0 {
		ui.Bad.Println("\n  No successful results to merge")
		return
	}

	// Build merge prompt
	mergePrompt := fmt.Sprintf(`You are synthesizing outputs from multiple AI tools. The original task was: "%s"

Here are the outputs from each tool:

%s

Create the BEST possible response by merging the strengths of each tool's output.
Take the best ideas, examples, and explanations from each, and produce a single high-quality result.
Do not mention the tools or that this is a merge — just produce the best answer.`, task, strings.Join(contributions, "\n\n"))

	fmt.Printf("\n  %s Synthesizing with %s...\n", ui.Info.Sprint("🔀"), ui.Brand.Sprint(judge))

	mergeOutput := runJudge(judge, mergePrompt, env, timeout)
	if mergeOutput != "" {
		fmt.Println()
		fmt.Println("  " + strings.Repeat("─", 60))
		fmt.Printf("  %s Merged result:\n\n", ui.Brand.Sprint("🔀"))
		printTruncatedOutput(mergeOutput, 5000)
	} else {
		ui.Bad.Println("  Merge failed to produce output")
	}
}

func runJudge(judge, prompt string, env []string, timeout int) string {
	var cmdArgs []string
	switch judge {
	case "ollama":
		cmdArgs = []string{"ollama", "run", "llama3.3", prompt}
	default:
		cmdArgs = []string{judge, prompt}
	}

	var stdout bytes.Buffer
	c := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	c.Stdout = &stdout
	c.Stderr = os.Stderr
	c.Env = env
	c.Stdin = strings.NewReader(prompt)

	if err := c.Start(); err != nil {
		return ""
	}

	done := make(chan error, 1)
	go func() { done <- c.Wait() }()

	select {
	case err := <-done:
		if err != nil {
			return ""
		}
		return stdout.String()
	case <-time.After(time.Duration(timeout) * time.Second):
		_ = c.Process.Kill()
		return ""
	}
}

func printSquadSummary(results []SquadResult) {
	headers := []string{"Tool", "Time", "Output", "Status"}
	var rows [][]string

	for _, r := range results {
		status := ui.StatusIcon(true) + " ok"
		dur := fmt.Sprintf("%.2fs", r.Duration.Seconds())
		outLen := fmt.Sprintf("%d chars", len(r.Output))

		if r.Error != "" {
			status = ui.StatusIcon(false) + " " + r.Error
			dur = "-"
			outLen = "-"
		}

		rows = append(rows, []string{r.Tool, dur, outLen, status})
	}

	ui.Table(headers, rows)
}

func printTruncatedOutput(output string, maxLen int) {
	output = strings.TrimSpace(output)
	if len(output) > maxLen {
		output = output[:maxLen] + "\n  ... (truncated)"
	}
	// Indent each line
	for _, line := range strings.Split(output, "\n") {
		fmt.Printf("  %s\n", line)
	}
}

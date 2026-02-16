package cmd

import (
	"bytes"
	"fmt"
	"math"
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

// SpeedResult holds a single speedtest result.
type SpeedResult struct {
	Provider  string
	Model     string
	Latency   time.Duration // Time to first byte (TTFB)
	TotalTime time.Duration
	OutputLen int
	TokensEst int // Estimated tokens (chars / 4)
	TPS       float64
	ExitCode  int
	Error     string
}

func speedtestCmd() *cobra.Command {
	var (
		prompt     string
		quick      bool
		tools      string
		timeout    int
		showOutput bool
	)

	cmd := &cobra.Command{
		Use:     "speedtest [prompt]",
		Short:   "AI speedtest — benchmark your LLM stack like an internet speed test",
		Aliases: []string{"speed", "benchmark", "bench"},
		Long: `Run a visual AI speedtest across all configured providers.
Tests latency, throughput, and quality — displayed with progress bars and a scorecard.

When --tools is provided, runs a direct comparison between specific tools (benchmark mode).

Examples:
  palm speedtest                                      # Test all configured providers
  palm speedtest --prompt "explain recursion"          # Custom prompt
  palm speedtest --quick                               # Faster test (shorter prompt)
  palm speedtest "explain quicksort" --tools ollama,mods  # Compare specific tools
  palm speedtest "fix the bug" --tools aider,codex --output  # Show tool output`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// If positional arg provided, use as prompt
			if len(args) > 0 {
				prompt = args[0]
			}

			// Benchmark mode: compare specific tools
			if tools != "" {
				runBenchmarkMode(prompt, tools, timeout, showOutput)
				return
			}

			// Speedtest mode: auto-detect and test all providers
			v := vault.New()
			env := buildVaultEnv(v)

			type testTarget struct {
				Provider string
				Model    string
				Cmd      []string
			}

			var targets []testTarget

			if _, err := exec.LookPath("ollama"); err == nil {
				targets = append(targets, testTarget{
					Provider: "Ollama",
					Model:    "llama3.3",
					Cmd:      []string{"ollama", "run", "llama3.3"},
				})
			}

			if _, err := exec.LookPath("aider"); err == nil {
				targets = append(targets, testTarget{
					Provider: "Aider",
					Model:    "default",
					Cmd:      []string{"aider", "--message"},
				})
			}

			if _, err := exec.LookPath("mods"); err == nil {
				targets = append(targets, testTarget{
					Provider: "Mods",
					Model:    "default",
					Cmd:      []string{"mods"},
				})
			}

			if _, err := exec.LookPath("llm"); err == nil {
				targets = append(targets, testTarget{
					Provider: "LLM",
					Model:    "default",
					Cmd:      []string{"llm"},
				})
			}

			if len(targets) == 0 {
				printSpeedtestHeader()
				fmt.Println()
				ui.Warn.Println("  No AI tools detected. Install some first:")
				fmt.Println("    palm install ollama mods llm")
				fmt.Println()
				return
			}

			if prompt == "" {
				if quick {
					prompt = "Say hello in 3 words"
				} else {
					prompt = "Explain the difference between a stack and a queue in 100 words"
				}
			}

			printSpeedtestHeader()
			fmt.Println()
			fmt.Printf("  Prompt:   %s\n", ui.Subtle.Sprint(prompt))
			fmt.Printf("  Targets:  %d providers\n", len(targets))
			fmt.Println()

			var mu sync.Mutex
			var wg sync.WaitGroup
			results := make([]SpeedResult, len(targets))

			for i, t := range targets {
				wg.Add(1)
				go func(idx int, target testTarget) {
					defer wg.Done()

					fmt.Printf("  %s Testing %s (%s)...\n",
						ui.Info.Sprint("⟳"),
						ui.Brand.Sprint(target.Provider),
						target.Model)

					result := runSpeedTest(target.Provider, target.Model, target.Cmd, prompt, env)

					mu.Lock()
					results[idx] = result
					mu.Unlock()

					if result.Error != "" {
						fmt.Printf("  %s %s: %s\n",
							ui.StatusIcon(false),
							target.Provider,
							ui.Bad.Sprint(result.Error))
					} else {
						fmt.Printf("  %s %s: %.2fs, ~%d tok/s\n",
							ui.StatusIcon(true),
							ui.Brand.Sprint(target.Provider),
							result.TotalTime.Seconds(),
							int(result.TPS))
					}
				}(i, t)
			}

			wg.Wait()
			fmt.Println()
			printSpeedtestResults(results)
		},
	}

	cmd.Flags().StringVar(&prompt, "prompt", "", "Custom test prompt")
	cmd.Flags().BoolVar(&quick, "quick", false, "Quick test with shorter prompt")
	cmd.Flags().StringVar(&tools, "tools", "", "Compare specific tools (e.g., ollama,mods)")
	cmd.Flags().IntVar(&timeout, "timeout", 30, "Timeout per tool in seconds (benchmark mode)")
	cmd.Flags().BoolVar(&showOutput, "output", false, "Show tool output (benchmark mode)")
	return cmd
}

// runBenchmarkMode compares specific tools on the same prompt.
func runBenchmarkMode(prompt, tools string, timeout int, showOutput bool) {
	reg := loadRegistry()
	v := vault.New()

	toolNames := strings.Split(tools, ",")
	if len(toolNames) < 2 {
		ui.Warn.Println("  Provide at least 2 tools to compare: --tools tool1,tool2")
		os.Exit(1)
	}

	if prompt == "" {
		prompt = "Explain the difference between a stack and a queue in 100 words"
	}

	ui.Banner("benchmark")
	fmt.Printf("  Prompt: %s\n", ui.Brand.Sprint(prompt))
	fmt.Printf("  Tools:  %s\n", strings.Join(toolNames, ", "))
	fmt.Printf("  Timeout: %ds\n\n", timeout)

	var results []BenchResult

	for _, name := range toolNames {
		name = strings.TrimSpace(name)
		tool := reg.Get(name)

		bin := name
		if tool != nil && tool.Install.Verify.Command != "" {
			parts := strings.Fields(tool.Install.Verify.Command)
			if len(parts) > 0 {
				bin = parts[0]
			}
		}

		if _, err := exec.LookPath(bin); err != nil {
			results = append(results, BenchResult{
				Tool:     name,
				ExitCode: -1,
				Error:    "not installed",
			})
			continue
		}

		fmt.Printf("  Running %s... ", ui.Brand.Sprint(name))

		result := runBenchmark(name, bin, prompt, tool, v, timeout)
		results = append(results, result)

		if result.Error != "" {
			ui.Bad.Printf("failed (%s)\n", result.Error)
		} else {
			ui.Good.Printf("%.2fs\n", result.Duration.Seconds())
		}
	}

	fmt.Println()
	headers := []string{"Tool", "Time", "Output Length", "Status"}
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

	if showOutput {
		fmt.Println()
		for _, r := range results {
			if r.Output != "" {
				fmt.Printf("  === %s ===\n", ui.Brand.Sprint(r.Tool))
				out := r.Output
				if len(out) > 500 {
					out = out[:500] + "\n  ... (truncated)"
				}
				fmt.Println(out)
				fmt.Println()
			}
		}
	}
}

// BenchResult holds the result of benchmarking a single tool.
type BenchResult struct {
	Tool     string
	Duration time.Duration
	Output   string
	ExitCode int
	Error    string
}

func runBenchmark(name, bin, prompt string, tool *registry.Tool, v vault.Vault, timeout int) BenchResult {
	env := os.Environ()
	if tool != nil {
		allKeys := append(tool.Keys.Required, tool.Keys.Optional...)
		for _, key := range allKeys {
			if os.Getenv(key) == "" {
				if val, err := v.Get(key); err == nil {
					env = append(env, fmt.Sprintf("%s=%s", key, val))
				}
			}
		}
	}

	var cmdArgs []string
	switch name {
	case "ollama":
		cmdArgs = []string{bin, "run", "llama3.3", prompt}
	default:
		cmdArgs = []string{bin, prompt}
	}

	var stdout, stderr bytes.Buffer
	c := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	c.Stdout = &stdout
	c.Stderr = &stderr
	c.Env = env
	c.Stdin = strings.NewReader(prompt)

	start := time.Now()
	if err := c.Start(); err != nil {
		return BenchResult{
			Tool:     name,
			Duration: time.Since(start),
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
			return BenchResult{
				Tool:     name,
				Duration: elapsed,
				Output:   stderr.String(),
				ExitCode: 1,
				Error:    err.Error(),
			}
		}
		return BenchResult{
			Tool:     name,
			Duration: elapsed,
			Output:   stdout.String(),
			ExitCode: 0,
		}
	case <-time.After(time.Duration(timeout) * time.Second):
		_ = c.Process.Kill()
		return BenchResult{
			Tool:     name,
			Duration: time.Duration(timeout) * time.Second,
			Error:    "timeout",
			ExitCode: -1,
		}
	}
}

func printSpeedtestHeader() {
	fmt.Println()
	fmt.Println(ui.Brand.Sprint("  ╔═══════════════════════════════════════════════╗"))
	fmt.Println(ui.Brand.Sprint("  ║") + "   🌴  " + ui.Brand.Sprint("palm speedtest") + "                          " + ui.Brand.Sprint("║"))
	fmt.Println(ui.Brand.Sprint("  ║") + "   " + ui.Subtle.Sprint("AI performance benchmark for your stack") + "   " + ui.Brand.Sprint("║"))
	fmt.Println(ui.Brand.Sprint("  ╚═══════════════════════════════════════════════╝"))
}

func runSpeedTest(provider, model string, cmdArgs []string, prompt string, env []string) SpeedResult {
	args := append(cmdArgs, prompt)

	var stdout bytes.Buffer
	c := exec.Command(args[0], args[1:]...)
	c.Stdout = &stdout
	c.Stderr = &bytes.Buffer{}
	c.Env = env
	c.Stdin = strings.NewReader(prompt)

	start := time.Now()
	if err := c.Start(); err != nil {
		return SpeedResult{
			Provider:  provider,
			Model:     model,
			TotalTime: time.Since(start),
			ExitCode:  1,
			Error:     err.Error(),
		}
	}

	done := make(chan error, 1)
	go func() { done <- c.Wait() }()

	select {
	case err := <-done:
		elapsed := time.Since(start)
		if err != nil {
			return SpeedResult{
				Provider:  provider,
				Model:     model,
				TotalTime: elapsed,
				ExitCode:  1,
				Error:     err.Error(),
			}
		}

		output := stdout.String()
		tokensEst := len(output) / 4
		var tps float64
		if elapsed.Seconds() > 0 {
			tps = float64(tokensEst) / elapsed.Seconds()
		}

		return SpeedResult{
			Provider:  provider,
			Model:     model,
			TotalTime: elapsed,
			OutputLen: len(output),
			TokensEst: tokensEst,
			TPS:       tps,
			ExitCode:  0,
		}

	case <-time.After(90 * time.Second):
		_ = c.Process.Kill()
		return SpeedResult{
			Provider:  provider,
			Model:     model,
			TotalTime: 90 * time.Second,
			ExitCode:  -1,
			Error:     "timeout (90s)",
		}
	}
}

func printSpeedtestResults(results []SpeedResult) {
	fmt.Println(ui.Brand.Sprint("  ┌─────────────────────────────────────────────────────────┐"))
	fmt.Println(ui.Brand.Sprint("  │") + "  " + ui.Brand.Sprint("RESULTS") + "                                                " + ui.Brand.Sprint("│"))
	fmt.Println(ui.Brand.Sprint("  ├─────────────────────────────────────────────────────────┤"))

	var maxTPS float64
	for _, r := range results {
		if r.TPS > maxTPS {
			maxTPS = r.TPS
		}
	}
	if maxTPS == 0 {
		maxTPS = 1
	}

	for _, r := range results {
		fmt.Println(ui.Brand.Sprint("  │") + "                                                         " + ui.Brand.Sprint("│"))

		if r.Error != "" {
			fmt.Printf(ui.Brand.Sprint("  │")+"  %-12s  %s%s\n",
				r.Provider,
				ui.Bad.Sprint("FAILED: "+r.Error),
				strings.Repeat(" ", max(0, 40-len("FAILED: "+r.Error)))+ui.Brand.Sprint("│"))
			continue
		}

		provLine := fmt.Sprintf("  %-12s  %s", r.Provider, ui.Subtle.Sprint(r.Model))
		pad1 := max(0, 55-len(r.Provider)-2-len(r.Model)-2)
		fmt.Println(ui.Brand.Sprint("  │") + provLine + strings.Repeat(" ", pad1) + ui.Brand.Sprint("│"))

		barWidth := 30
		filled := int(math.Round((r.TPS / maxTPS) * float64(barWidth)))
		if filled < 1 && r.TPS > 0 {
			filled = 1
		}

		bar := ui.Brand.Sprint(strings.Repeat("█", filled)) + ui.Subtle.Sprint(strings.Repeat("░", barWidth-filled))
		tpsStr := fmt.Sprintf("%.1f tok/s", r.TPS)
		pad2 := max(0, 19-len(tpsStr))
		fmt.Printf(ui.Brand.Sprint("  │")+"  %s  %s%s"+ui.Brand.Sprint("│")+"\n",
			bar, tpsStr, strings.Repeat(" ", pad2))

		timeStr := fmt.Sprintf("%.2fs", r.TotalTime.Seconds())
		outStr := formatBytes(r.OutputLen)
		tokStr := fmt.Sprintf("~%d tokens", r.TokensEst)
		statsLine := fmt.Sprintf("  %s  %s  %s",
			ui.Subtle.Sprint(timeStr),
			ui.Subtle.Sprint(outStr),
			ui.Subtle.Sprint(tokStr))
		pad3 := max(0, 55-len(timeStr)-len(outStr)-len(tokStr)-6)
		fmt.Println(ui.Brand.Sprint("  │") + statsLine + strings.Repeat(" ", pad3) + ui.Brand.Sprint("│"))
	}

	fmt.Println(ui.Brand.Sprint("  │") + "                                                         " + ui.Brand.Sprint("│"))
	fmt.Println(ui.Brand.Sprint("  └─────────────────────────────────────────────────────────┘"))

	var winner *SpeedResult
	for i := range results {
		if results[i].Error == "" && (winner == nil || results[i].TPS > winner.TPS) {
			winner = &results[i]
		}
	}

	if winner != nil {
		fmt.Println()
		fmt.Printf("  %s Fastest: %s at %.1f tok/s\n",
			ui.Brand.Sprint("🏆"),
			ui.Brand.Sprint(winner.Provider+" ("+winner.Model+")"),
			winner.TPS)
	}

	fmt.Println()
	printSpeedGrade(results)
	fmt.Println()
}

func printSpeedGrade(results []SpeedResult) {
	var totalTPS float64
	var count int
	for _, r := range results {
		if r.Error == "" {
			totalTPS += r.TPS
			count++
		}
	}

	if count == 0 {
		ui.Bad.Println("  Grade: F — no providers responded")
		return
	}

	avgTPS := totalTPS / float64(count)

	var grade string
	switch {
	case avgTPS >= 100:
		grade = "A+"
	case avgTPS >= 50:
		grade = "A"
	case avgTPS >= 25:
		grade = "B"
	case avgTPS >= 10:
		grade = "C"
	case avgTPS >= 5:
		grade = "D"
	default:
		grade = "F"
	}

	fmt.Printf("  Your AI Stack Grade: %s  (avg %.1f tok/s across %d providers)\n",
		formatGrade(grade), avgTPS, count)
}

func formatGrade(grade string) string {
	switch {
	case strings.HasPrefix(grade, "A"):
		return ui.Good.Sprint(grade)
	case grade == "B":
		return ui.Good.Sprint(grade)
	case grade == "C":
		return ui.Warn.Sprint(grade)
	default:
		return ui.Bad.Sprint(grade)
	}
}

func formatBytes(n int) string {
	if n < 1024 {
		return fmt.Sprintf("%dB", n)
	}
	return fmt.Sprintf("%.1fKB", float64(n)/1024)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/msalah0e/palm/internal/registry"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/msalah0e/palm/internal/vault"
	"github.com/spf13/cobra"
)

// BenchResult holds the result of benchmarking a single tool.
type BenchResult struct {
	Tool     string
	Duration time.Duration
	Output   string
	ExitCode int
	Error    string
}

func benchmarkCmd() *cobra.Command {
	var tools string
	var timeout int
	var showOutput bool

	cmd := &cobra.Command{
		Use:   "benchmark <prompt>",
		Short: "Compare AI tools by running the same prompt",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			prompt := args[0]

			reg := loadRegistry()
			v := vault.New()

			// Parse tool list
			toolNames := strings.Split(tools, ",")
			if len(toolNames) < 2 {
				ui.Warn.Println("  Provide at least 2 tools to compare: --tools tool1,tool2")
				os.Exit(1)
			}

			ui.Banner("benchmark")
			fmt.Printf("  Prompt: %s\n", ui.Brand.Sprint(prompt))
			fmt.Printf("  Tools:  %s\n", strings.Join(toolNames, ", "))
			fmt.Printf("  Timeout: %ds\n\n", timeout)

			var results []BenchResult

			for _, name := range toolNames {
				name = strings.TrimSpace(name)
				tool := reg.Get(name)

				// Determine the binary
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

			// Print results
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
		},
	}

	cmd.Flags().StringVar(&tools, "tools", "", "Comma-separated list of tools to benchmark (required)")
	cmd.Flags().IntVar(&timeout, "timeout", 30, "Timeout per tool in seconds")
	cmd.Flags().BoolVar(&showOutput, "output", false, "Show tool output")
	_ = cmd.MarkFlagRequired("tools")
	return cmd
}

func runBenchmark(name, bin, prompt string, tool *registry.Tool, v vault.Vault, timeout int) BenchResult {
	// Build environment with vault keys
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

	// Build command based on tool type
	var cmdArgs []string
	switch name {
	case "ollama":
		cmdArgs = []string{bin, "run", "llama3.3", prompt}
	default:
		// Generic: pipe prompt to stdin
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

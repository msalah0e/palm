package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

// pirateProvider defines a free AI tool with its quota detection.
type pirateProvider struct {
	Name      string
	Binary    string
	EnvKey    string
	RunCmd    []string
	QuotaErr  []string // strings in stderr that indicate quota exhaustion
	Priority  int      // lower = try first
}

var pirateProviders = []pirateProvider{
	{
		Name:     "ollama",
		Binary:   "ollama",
		RunCmd:   []string{"ollama", "run"},
		QuotaErr: []string{}, // local, no quota
		Priority: 1,
	},
	{
		Name:     "llama-cpp",
		Binary:   "llama-cli",
		RunCmd:   []string{"llama-cli", "-m"},
		QuotaErr: []string{},
		Priority: 2,
	},
	{
		Name:     "claude-code",
		Binary:   "claude",
		EnvKey:   "ANTHROPIC_API_KEY",
		RunCmd:   []string{"claude"},
		QuotaErr: []string{"rate_limit", "quota", "exceeded", "429", "overloaded"},
		Priority: 3,
	},
	{
		Name:     "aider",
		Binary:   "aider",
		EnvKey:   "OPENAI_API_KEY",
		RunCmd:   []string{"aider"},
		QuotaErr: []string{"rate limit", "quota exceeded", "429", "insufficient_quota"},
		Priority: 4,
	},
	{
		Name:     "codex",
		Binary:   "codex",
		EnvKey:   "OPENAI_API_KEY",
		RunCmd:   []string{"codex"},
		QuotaErr: []string{"rate limit", "quota", "429"},
		Priority: 5,
	},
	{
		Name:     "gemini",
		Binary:   "gemini",
		EnvKey:   "GOOGLE_API_KEY",
		RunCmd:   []string{"gemini"},
		QuotaErr: []string{"RESOURCE_EXHAUSTED", "quota", "429"},
		Priority: 6,
	},
}

func pirateCmd() *cobra.Command {
	var preferLocal bool
	var maxRetries int

	cmd := &cobra.Command{
		Use:     "pirate [prompt]",
		Aliases: []string{"arr", "free"},
		Short:   "Pirate mode — auto-switch between free AI tools when quota is reached",
		Long: "Pirate mode runs your prompt across available AI tools.\n" +
			"When one tool hits its rate limit or quota, palm automatically\n" +
			"switches to the next available tool — like a pirate hopping ships.\n\n" +
			"Priority: local models (ollama) → free tiers → paid APIs",
		Args: cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				showPirateStatus(preferLocal)
				return
			}

			prompt := strings.Join(args, " ")
			runPirate(prompt, preferLocal, maxRetries)
		},
	}

	cmd.Flags().BoolVar(&preferLocal, "local", false, "Prefer local models (ollama, llama-cpp)")
	cmd.Flags().IntVar(&maxRetries, "retries", 3, "Max tools to try before giving up")

	cmd.AddCommand(
		pirateStatusCmd(),
		pirateRunCmd(),
	)

	return cmd
}

func pirateStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show available AI providers and their status",
		Run: func(cmd *cobra.Command, args []string) {
			showPirateStatus(false)
		},
	}
}

func pirateRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run <tool> [args...]",
		Short: "Run a specific tool with pirate fallback on quota errors",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			toolName := args[0]
			var provider *pirateProvider
			for i := range pirateProviders {
				if pirateProviders[i].Name == toolName {
					provider = &pirateProviders[i]
					break
				}
			}
			if provider == nil {
				ui.Bad.Printf("  Unknown pirate provider: %s\n", toolName)
				os.Exit(1)
			}

			remaining := args[1:]
			runArgs := append(provider.RunCmd, remaining...)
			fmt.Printf("  %s Running: %s\n", ui.Brand.Sprint("🏴‍☠️"), strings.Join(runArgs, " "))

			c := exec.Command(runArgs[0], runArgs[1:]...)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			c.Run()
		},
	}
}

func showPirateStatus(preferLocal bool) {
	ui.Banner("pirate mode")
	fmt.Println("  Auto-switch between AI tools when quota is reached")
	fmt.Println()

	localCount := 0
	cloudCount := 0

	var rows [][]string
	for _, p := range pirateProviders {
		available := isProviderAvailable(p)
		status := ui.StatusIcon(available)
		kind := "cloud"
		if p.EnvKey == "" {
			kind = "local"
			if available {
				localCount++
			}
		} else if available {
			cloudCount++
		}

		keyStatus := "-"
		if p.EnvKey != "" {
			if os.Getenv(p.EnvKey) != "" {
				keyStatus = "set"
			} else {
				keyStatus = "missing"
			}
		}

		rows = append(rows, []string{status, p.Name, kind, keyStatus, fmt.Sprintf("%d", p.Priority)})
	}

	ui.Table([]string{"", "Provider", "Type", "Key", "Priority"}, rows)

	fmt.Println()
	fmt.Printf("  %d local + %d cloud providers available\n", localCount, cloudCount)
	fmt.Println()
	fmt.Println("  Usage: palm pirate \"your prompt here\"")
	fmt.Println("  The prompt will be sent to the highest-priority available tool.")
	fmt.Println("  If quota is hit, palm auto-switches to the next provider.")
}

func runPirate(prompt string, preferLocal bool, maxRetries int) {
	ui.Banner("pirate mode")
	fmt.Printf("  Prompt: %s\n\n", truncatePirate(prompt, 60))

	// Build ordered list of available providers
	available := []pirateProvider{}
	for _, p := range pirateProviders {
		if isProviderAvailable(p) {
			available = append(available, p)
		}
	}

	if len(available) == 0 {
		ui.Bad.Println("  No AI providers available!")
		fmt.Println("  Install ollama for free local AI: palm install ollama")
		os.Exit(1)
	}

	if preferLocal {
		// Move local providers to front
		var local, cloud []pirateProvider
		for _, p := range available {
			if p.EnvKey == "" {
				local = append(local, p)
			} else {
				cloud = append(cloud, p)
			}
		}
		available = append(local, cloud...)
	}

	tried := 0
	for _, p := range available {
		if tried >= maxRetries {
			break
		}
		tried++

		fmt.Printf("  Trying %s (attempt %d/%d)...\n", ui.Brand.Sprint(p.Name), tried, maxRetries)

		success, output := tryProvider(p, prompt)
		if success {
			fmt.Println()
			fmt.Println(output)
			return
		}

		// Check if it's a quota error
		isQuota := false
		for _, qe := range p.QuotaErr {
			if strings.Contains(strings.ToLower(output), strings.ToLower(qe)) {
				isQuota = true
				break
			}
		}

		if isQuota {
			ui.Warn.Printf("  %s %s hit quota limit — switching...\n", ui.WarnIcon(), p.Name)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// Non-quota error — still show output
		if output != "" {
			fmt.Println(output)
		}
		return
	}

	ui.Bad.Println("  All providers exhausted. Try again later or install local models.")
}

func tryProvider(p pirateProvider, prompt string) (bool, string) {
	var cmdArgs []string

	switch p.Name {
	case "ollama":
		cmdArgs = []string{"ollama", "run", "llama3.2", prompt}
	case "claude-code":
		cmdArgs = []string{"claude", "-p", prompt}
	case "aider":
		cmdArgs = []string{"aider", "--message", prompt}
	case "codex":
		cmdArgs = []string{"codex", prompt}
	case "gemini":
		cmdArgs = []string{"gemini", prompt}
	default:
		cmdArgs = append(p.RunCmd, prompt)
	}

	c := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	out, err := c.CombinedOutput()
	output := strings.TrimSpace(string(out))

	if err != nil {
		return false, output
	}
	return true, output
}

func isProviderAvailable(p pirateProvider) bool {
	// Check binary exists
	if _, err := exec.LookPath(p.Binary); err != nil {
		return false
	}
	// Check API key if needed
	if p.EnvKey != "" && os.Getenv(p.EnvKey) == "" {
		return false
	}
	return true
}

func truncatePirate(s string, max int) string {
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}

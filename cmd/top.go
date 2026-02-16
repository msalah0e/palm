package cmd

import (
	"strings"
	"time"

	"github.com/msalah0e/palm/internal/registry"
	"github.com/msalah0e/palm/internal/top"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

func topCmd() *cobra.Command {
	var interval int

	cmd := &cobra.Command{
		Use:     "top",
		Aliases: []string{"monitor", "htop"},
		Short:   "Live monitor for running AI tool processes",
		Long:    ui.Brand.Sprint(ui.Palm+" palm top") + " \u2014 htop-like dashboard for AI tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			reg := loadRegistry()
			known := buildKnownBinaries(reg)

			cfg := top.Config{
				RefreshInterval: time.Duration(interval) * time.Second,
				KnownBinaries:  known,
			}

			return top.Run(cfg)
		},
	}

	cmd.Flags().IntVar(&interval, "interval", 1, "Refresh interval in seconds")

	return cmd
}

// buildKnownBinaries extracts binary names from registry verify commands
// and adds hardcoded extras for common AI tools.
func buildKnownBinaries(reg *registry.Registry) map[string]string {
	known := make(map[string]string)

	// Extract from registry
	for _, t := range reg.All() {
		if t.Install.Verify.Command == "" {
			continue
		}
		parts := strings.Fields(t.Install.Verify.Command)
		if len(parts) == 0 {
			continue
		}
		bin := parts[0]
		// Skip interpreters — we'll match on subsequent args
		if bin == "python3" || bin == "python" || bin == "node" || bin == "sh" || bin == "bash" {
			if len(parts) > 1 {
				// Use the module/script name
				bin = parts[1]
				if bin == "-m" && len(parts) > 2 {
					bin = parts[2]
				}
			}
		}
		displayName := t.DisplayName
		if displayName == "" {
			displayName = t.Name
		}
		known[bin] = displayName
	}

	// Hardcoded extras for common AI tool binaries
	extras := map[string]string{
		"claude":      "Claude Code",
		"aider":       "Aider",
		"ollama":      "Ollama",
		"codex":       "Codex CLI",
		"copilot":     "GitHub Copilot",
		"cursor":      "Cursor",
		"cody":        "Sourcegraph Cody",
		"continue":    "Continue",
		"tabby":       "TabbyML",
		"llama-server": "Llama.cpp",
		"llamafile":   "Llamafile",
		"vllm":        "vLLM",
		"tgi":         "Text Gen Inference",
		"sgpt":        "Shell GPT",
		"fabric":      "Fabric",
		"goose":       "Goose",
		"mentat":      "Mentat",
		"sweep":       "Sweep",
		"gpt-engineer": "GPT Engineer",
		"open-interpreter": "Open Interpreter",
		"interpreter": "Open Interpreter",
	}

	for bin, name := range extras {
		if _, exists := known[bin]; !exists {
			known[bin] = name
		}
	}

	return known
}

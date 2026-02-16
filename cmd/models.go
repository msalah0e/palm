package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/msalah0e/palm/internal/models"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/msalah0e/palm/internal/vault"
	"github.com/spf13/cobra"
)

func modelsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "models",
		Short: "Manage LLM models across providers",
	}

	cmd.AddCommand(
		modelsListCmd(),
		modelsPullCmd(),
		modelsInfoCmd(),
		modelsProvidersCmd(),
	)

	return cmd
}

func modelsListCmd() *cobra.Command {
	var provider string
	var modelType string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available models across all providers",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("available models")

			providers := models.BuiltinProviders()
			v := vault.New()

			for _, p := range providers {
				if provider != "" && !strings.EqualFold(p.Name, provider) {
					continue
				}

				// Check if API key is available
				keyStatus := ui.Subtle.Sprint("(no key)")
				if p.EnvKey == "" {
					keyStatus = ui.Good.Sprint("(local)")
				} else if os.Getenv(p.EnvKey) != "" {
					keyStatus = ui.Good.Sprint("(env)")
				} else if _, err := v.Get(p.EnvKey); err == nil {
					keyStatus = ui.Good.Sprint("(vault)")
				}

				fmt.Printf("  %s %s\n", ui.Brand.Sprint(p.Name), keyStatus)

				for _, m := range p.Models {
					if modelType != "" && m.Type != modelType {
						continue
					}

					ctx := models.FormatContext(m.Context)
					cost := ""
					if m.InputCost > 0 {
						cost = fmt.Sprintf("$%.2f/$%.2f", m.InputCost, m.OutputCost)
					}

					fmt.Printf("    %-35s %-6s %-8s %s\n",
						m.ID, ctx, m.Type, ui.Subtle.Sprint(cost))
				}
				fmt.Println()
			}
		},
	}

	cmd.Flags().StringVarP(&provider, "provider", "p", "", "Filter by provider (openai, anthropic, google, ollama)")
	cmd.Flags().StringVarP(&modelType, "type", "t", "", "Filter by type (chat, embedding, image)")
	return cmd
}

func modelsPullCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pull <model>",
		Short: "Pull a model (for Ollama and local providers)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			modelID := args[0]

			ui.Banner("pulling model")

			m := models.FindModel(modelID)
			if m != nil && m.Provider != "ollama" {
				fmt.Printf("  %s is a cloud model (%s) — no download needed\n", m.Name, m.Provider)
				fmt.Println("  Just set the API key: palm keys add " + getProviderKey(m.Provider))
				return
			}

			// Check if ollama is installed
			if _, err := exec.LookPath("ollama"); err != nil {
				ui.Bad.Println("  Ollama not found. Install it first:")
				fmt.Println("  palm install ollama")
				os.Exit(1)
			}

			fmt.Printf("  Pulling %s via Ollama...\n\n", modelID)

			c := exec.Command("ollama", "pull", modelID)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				ui.Bad.Printf("\n  Pull failed: %v\n", err)
				os.Exit(1)
			}

			fmt.Println()
			ui.Good.Printf("  %s %s pulled successfully\n", ui.StatusIcon(true), modelID)
		},
	}
}

func modelsInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <model>",
		Short: "Show detailed info about a model",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			m := models.FindModel(args[0])
			if m == nil {
				ui.Warn.Printf("  Model %q not found in registry\n", args[0])
				os.Exit(1)
			}

			ui.Banner("model info")

			fmt.Printf("  %s\n", ui.Brand.Sprint(m.Name))
			fmt.Printf("  ID:        %s\n", m.ID)
			fmt.Printf("  Provider:  %s\n", m.Provider)
			fmt.Printf("  Type:      %s\n", m.Type)
			fmt.Printf("  Context:   %s tokens\n", models.FormatContext(m.Context))
			if m.InputCost > 0 {
				fmt.Printf("  Input:     $%.2f / 1M tokens\n", m.InputCost)
				fmt.Printf("  Output:    $%.2f / 1M tokens\n", m.OutputCost)
			} else {
				fmt.Println("  Cost:      free (local)")
			}
		},
	}
}

func modelsProvidersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "providers",
		Short: "List supported LLM providers",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("LLM providers")

			v := vault.New()

			for _, p := range models.BuiltinProviders() {
				keyStatus := ui.StatusIcon(false) + " no key"
				if p.EnvKey == "" {
					keyStatus = ui.StatusIcon(true) + " local"
				} else if os.Getenv(p.EnvKey) != "" {
					keyStatus = ui.StatusIcon(true) + " env"
				} else if _, err := v.Get(p.EnvKey); err == nil {
					keyStatus = ui.StatusIcon(true) + " vault"
				}

				fmt.Printf("  %-12s %d models  %s", ui.Brand.Sprint(p.Name), len(p.Models), keyStatus)
				if p.EnvKey != "" {
					fmt.Printf("  (%s)", p.EnvKey)
				}
				fmt.Println()
			}
		},
	}
}

func getProviderKey(provider string) string {
	for _, p := range models.BuiltinProviders() {
		if strings.EqualFold(p.Name, provider) {
			return p.EnvKey
		}
	}
	return "API_KEY"
}

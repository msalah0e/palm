package cmd

import (
	"fmt"

	"github.com/msalah0e/tamr/internal/ui"
	"github.com/spf13/cobra"
)

func discoverCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "discover",
		Short: "Browse curated AI tools by category",
		Run: func(cmd *cobra.Command, args []string) {
			reg := loadRegistry()

			ui.Banner("discover AI tools")

			categoryLabels := map[string]string{
				"coding": "\U0001F4BB Coding",
				"llm":    "\U0001F9E0 LLM & Inference",
				"agents": "\U0001F916 Agents & Automation",
				"media":  "\U0001F3A8 Creative & Media",
				"infra":  "\u2699\ufe0f  Infrastructure",
				"data":   "\U0001F4CA Data & Vector DBs",
			}

			categoryOrder := []string{"coding", "llm", "agents", "media", "infra", "data"}

			for _, cat := range categoryOrder {
				tools := reg.ByCategory(cat)
				if len(tools) == 0 {
					continue
				}

				label := categoryLabels[cat]
				if label == "" {
					label = cat
				}

				fmt.Printf("  %s\n", ui.Brand.Sprint(label))
				for _, t := range tools {
					desc := t.Description
					if len(desc) > 50 {
						desc = desc[:47] + "..."
					}
					fmt.Printf("  %s %-20s %s\n", ui.Subtle.Sprint("\u2500\u2500"), t.Name, desc)
				}
				fmt.Println()
			}

			fmt.Println("  `tamr info <tool>` for details · `tamr install <tool>` to install")
		},
	}
}

package cmd

import (
	"fmt"

	"github.com/msalah0e/palm/internal/ui"
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
				"coding":        "\U0001F4BB Coding",
				"llm":           "\U0001F9E0 LLM & Inference",
				"agents":        "\U0001F916 Agents & Automation",
				"chat":          "\U0001F4AC Chat & Assistants",
				"devtools":      "\U0001F6E0\ufe0f  Dev Tools",
				"media":         "\U0001F3A8 Creative & Media",
				"infra":         "\u2699\ufe0f  Infrastructure",
				"data":          "\U0001F4CA Data & Vector DBs",
				"testing":       "\U0001F9EA Testing & Evaluation",
				"security":      "\U0001F6E1\ufe0f  Security & Safety",
				"observability": "\U0001F4E1 Observability",
				"search":       "\U0001F50D Search & RAG",
				"writing":       "\u270D\ufe0f  Writing & Documents",
			}

			categoryOrder := []string{"coding", "llm", "agents", "chat", "devtools", "media", "infra", "data", "testing", "security", "observability", "search", "writing"}

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

			fmt.Println("  `palm info <tool>` for details · `palm install <tool>` to install")
		},
	}
}

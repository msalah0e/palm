package cmd

import (
	"fmt"
	"os"

	"github.com/msalah0e/tamr/internal/ui"
	"github.com/spf13/cobra"
)

func aiSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search the AI tool registry",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			reg := loadRegistry()
			query := args[0]
			results := reg.Search(query)

			ui.Banner(fmt.Sprintf("search results for %q", query))

			if len(results) == 0 {
				fmt.Println("  No tools found matching your query.")
				os.Exit(0)
			}

			headers := []string{"Tool", "Category", "Install via", "Description"}
			var rows [][]string
			for _, t := range results {
				backend, _ := t.InstallMethod()
				desc := t.Description
				if len(desc) > 45 {
					desc = desc[:42] + "..."
				}
				rows = append(rows, []string{t.Name, t.Category, backend, desc})
			}

			ui.Table(headers, rows)

			fmt.Printf("\n  %d results \u00b7 `tamr ai install <tool>` to install\n", len(results))
		},
	}
}

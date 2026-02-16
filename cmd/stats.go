package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/msalah0e/palm/internal/stats"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

func statsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Show local usage statistics",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("usage statistics")

			summary, err := stats.Summarize()
			if err != nil {
				ui.Bad.Printf("  Failed to read stats: %v\n", err)
				os.Exit(1)
			}

			if summary.TotalCommands == 0 {
				fmt.Println("  No usage data recorded yet.")
				fmt.Println("  Enable stats in ~/.config/palm/config.toml")
				return
			}

			fmt.Printf("  Total commands:     %d\n", summary.TotalCommands)
			fmt.Printf("  Tools installed:    %d\n", summary.ToolsInstalled)

			if !summary.LastUsed.IsZero() {
				ago := time.Since(summary.LastUsed).Round(time.Second)
				fmt.Printf("  Last used:          %s ago\n", ago)
			}
		},
	}
}

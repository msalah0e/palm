package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/msalah0e/palm/internal/activity"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

func actlogCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "log",
		Aliases: []string{"activity", "logs"},
		Short:   "Unified activity log — track AI tool actions across sessions",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("activity log")

			entries, err := activity.Read(20)
			if err != nil || len(entries) == 0 {
				fmt.Println("  No activity recorded yet.")
				fmt.Println("  Activity is logged when using `palm run <tool>`")
				return
			}

			var rows [][]string
			for _, e := range entries {
				timeStr := e.Timestamp.Format("Jan 02 15:04")
				costStr := "-"
				if e.Cost > 0 {
					costStr = fmt.Sprintf("$%.4f", e.Cost)
				}
				durStr := "-"
				if e.Duration > 0 {
					durStr = formatDuration(time.Duration(e.Duration * float64(time.Second)))
				}
				rows = append(rows, []string{timeStr, e.Action, e.Tool, durStr, costStr, truncateLog(e.Details, 30)})
			}
			ui.Table([]string{"Time", "Action", "Tool", "Duration", "Cost", "Details"}, rows)
			fmt.Printf("\n  Showing %d most recent entries\n", len(entries))
		},
	}

	cmd.AddCommand(
		actlogSearchCmd(),
		actlogClearCmd(),
		actlogExportCmd(),
		actlogStatsCmd(),
	)

	return cmd
}

func actlogSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search activity log entries",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			results, err := activity.Search(args[0], 50)
			if err != nil || len(results) == 0 {
				fmt.Printf("  No entries matching %q\n", args[0])
				return
			}

			ui.Banner("search results")
			var rows [][]string
			for _, e := range results {
				rows = append(rows, []string{
					e.Timestamp.Format("Jan 02 15:04"),
					e.Action,
					e.Tool,
					truncateLog(e.Details, 40),
				})
			}
			ui.Table([]string{"Time", "Action", "Tool", "Details"}, rows)
			fmt.Printf("\n  %d results\n", len(results))
		},
	}
}

func actlogClearCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Clear the activity log",
		Run: func(cmd *cobra.Command, args []string) {
			if err := activity.Clear(); err != nil {
				ui.Bad.Printf("  Failed to clear: %v\n", err)
				os.Exit(1)
			}
			ui.Good.Printf("  %s Activity log cleared\n", ui.StatusIcon(true))
		},
	}
}

func actlogExportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "export",
		Short: "Export activity log as JSON",
		Run: func(cmd *cobra.Command, args []string) {
			entries, err := activity.Read(0)
			if err != nil {
				ui.Bad.Printf("  %v\n", err)
				os.Exit(1)
			}
			data, _ := json.MarshalIndent(entries, "", "  ")
			fmt.Println(string(data))
		},
	}
}

func actlogStatsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Show activity statistics",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("activity stats")

			entries, err := activity.Read(0)
			if err != nil || len(entries) == 0 {
				fmt.Println("  No activity data")
				return
			}

			toolCounts := make(map[string]int)
			actionCounts := make(map[string]int)
			var totalCost float64

			for _, e := range entries {
				toolCounts[e.Tool]++
				actionCounts[e.Action]++
				totalCost += e.Cost
			}

			fmt.Printf("  Total entries: %d\n\n", len(entries))

			fmt.Println("  By tool:")
			for tool, count := range toolCounts {
				fmt.Printf("    %-20s %d\n", tool, count)
			}

			fmt.Println("\n  By action:")
			for action, count := range actionCounts {
				fmt.Printf("    %-20s %d\n", action, count)
			}

			if totalCost > 0 {
				fmt.Printf("\n  Total cost: $%.4f\n", totalCost)
			}
		},
	}
}

func truncateLog(s string, max int) string {
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}

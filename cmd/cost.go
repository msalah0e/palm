package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/msalah0e/palm/internal/session"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

func costCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cost",
		Short:   "Track AI tool spending across providers",
		Aliases: []string{"costs", "spend"},
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("cost tracking")

			summary, err := session.Summarize()
			if err != nil || summary.TotalSessions == 0 {
				fmt.Println("  No cost data recorded yet.")
				fmt.Println("  Costs are tracked when using `palm run <tool>`")
				fmt.Println()
				fmt.Println("  Tip: Use `palm budget set --monthly 50` to set limits")
				return
			}

			var rows [][]string
			for tool, ts := range summary.ByTool {
				costStr := "-"
				if ts.Cost > 0 {
					costStr = fmt.Sprintf("$%.4f", ts.Cost)
				}
				rows = append(rows, []string{
					tool,
					fmt.Sprintf("%d", ts.Sessions),
					formatDuration(ts.Duration),
					costStr,
				})
			}

			ui.Table([]string{"Tool", "Sessions", "Time", "Cost"}, rows)

			fmt.Println()
			fmt.Printf("  Total: %d sessions", summary.TotalSessions)
			if summary.TotalCost > 0 {
				fmt.Printf(" · $%.4f", summary.TotalCost)
			}
			fmt.Println()
		},
	}

	cmd.AddCommand(
		costTodayCmd(),
		costWeekCmd(),
		costExportCmd(),
	)

	return cmd
}

func costTodayCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "today",
		Short: "Show today's spending",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("today's costs")

			sessions, err := session.List(100)
			if err != nil {
				ui.Bad.Printf("  %v\n", err)
				os.Exit(1)
			}

			today := time.Now().Truncate(24 * time.Hour)
			var totalCost float64
			count := 0

			for _, s := range sessions {
				if s.StartedAt.After(today) {
					totalCost += s.Cost
					count++
				}
			}

			if count == 0 {
				fmt.Println("  No sessions today")
				return
			}

			fmt.Printf("  Sessions: %d\n", count)
			fmt.Printf("  Cost:     $%.4f\n", totalCost)
		},
	}
}

func costWeekCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "week",
		Short: "Show this week's spending",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("weekly costs")

			sessions, err := session.List(500)
			if err != nil {
				ui.Bad.Printf("  %v\n", err)
				os.Exit(1)
			}

			weekAgo := time.Now().Add(-7 * 24 * time.Hour)
			var totalCost float64
			count := 0

			for _, s := range sessions {
				if s.StartedAt.After(weekAgo) {
					totalCost += s.Cost
					count++
				}
			}

			if count == 0 {
				fmt.Println("  No sessions this week")
				return
			}

			fmt.Printf("  Sessions: %d\n", count)
			fmt.Printf("  Cost:     $%.4f\n", totalCost)
			fmt.Printf("  Avg/day:  $%.4f\n", totalCost/7)
		},
	}
}

func costExportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "export",
		Short: "Export cost data as JSON",
		Run: func(cmd *cobra.Command, args []string) {
			summary, err := session.Summarize()
			if err != nil {
				ui.Bad.Printf("  %v\n", err)
				os.Exit(1)
			}
			data, _ := json.MarshalIndent(summary, "", "  ")
			fmt.Println(string(data))
		},
	}
}

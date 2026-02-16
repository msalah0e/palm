package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/msalah0e/palm/internal/session"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

func sessionsCmd() *cobra.Command {
	var count int
	var cost bool

	cmd := &cobra.Command{
		Use:     "sessions",
		Aliases: []string{"history"},
		Short:   "Show recent tool run sessions",
		Run: func(cmd *cobra.Command, args []string) {
			if cost {
				showSessionCosts()
				return
			}

			ui.Banner("recent sessions")

			sessions, err := session.List(count)
			if err != nil {
				ui.Bad.Printf("  Failed to read sessions: %v\n", err)
				os.Exit(1)
			}

			if len(sessions) == 0 {
				fmt.Println("  No sessions recorded yet.")
				fmt.Println("  Sessions are tracked when using `palm run <tool>`")
				return
			}

			headers := []string{"Time", "Tool", "Duration", "Exit", "Cost"}
			var rows [][]string

			for _, s := range sessions {
				dur := formatDuration(time.Duration(s.Duration * float64(time.Second)))
				exit := ui.StatusIcon(s.ExitCode == 0)
				costStr := "-"
				if s.Cost > 0 {
					costStr = fmt.Sprintf("$%.4f", s.Cost)
				}
				timeStr := s.StartedAt.Format("Jan 02 15:04")

				rows = append(rows, []string{timeStr, s.Tool, dur, exit, costStr})
			}

			ui.Table(headers, rows)
			fmt.Printf("\n  Showing %d most recent sessions\n", len(sessions))
		},
	}

	cmd.Flags().IntVarP(&count, "count", "n", 20, "Number of sessions to show")
	cmd.Flags().BoolVar(&cost, "cost", false, "Show cost breakdown by tool")
	return cmd
}

func showSessionCosts() {
	ui.Banner("session costs")

	summary, err := session.Summarize()
	if err != nil {
		ui.Bad.Printf("  Failed to read sessions: %v\n", err)
		os.Exit(1)
	}

	if summary.TotalSessions == 0 {
		fmt.Println("  No sessions recorded yet.")
		return
	}

	headers := []string{"Tool", "Sessions", "Total Time", "Cost", "Tokens"}
	var rows [][]string

	for tool, ts := range summary.ByTool {
		costStr := "-"
		if ts.Cost > 0 {
			costStr = fmt.Sprintf("$%.4f", ts.Cost)
		}
		tokStr := "-"
		if ts.Tokens > 0 {
			tokStr = fmt.Sprintf("%d", ts.Tokens)
		}
		rows = append(rows, []string{
			tool,
			fmt.Sprintf("%d", ts.Sessions),
			formatDuration(ts.Duration),
			costStr,
			tokStr,
		})
	}

	ui.Table(headers, rows)

	fmt.Println()
	fmt.Printf("  Total: %d sessions · %s",
		summary.TotalSessions,
		formatDuration(summary.TotalDuration))
	if summary.TotalCost > 0 {
		fmt.Printf(" · $%.4f", summary.TotalCost)
	}
	fmt.Println()
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}

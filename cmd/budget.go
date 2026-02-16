package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/msalah0e/palm/internal/budget"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

func budgetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "budget",
		Short: "Manage AI spending controls",
	}

	cmd.AddCommand(
		budgetStatusCmd(),
		budgetSetCmd(),
		budgetResetCmd(),
	)

	return cmd
}

func budgetStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current budget status",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("budget status")

			status, err := budget.GetStatus()
			if err != nil {
				ui.Bad.Printf("  Failed to check budget: %v\n", err)
				os.Exit(1)
			}

			if status.MonthlyLimit == 0 && status.DailyLimit == 0 {
				fmt.Println("  No budget limits configured.")
				fmt.Println("  Set one: palm budget set --monthly 50")
				return
			}

			fmt.Printf("  %s\n\n", ui.Brand.Sprint(status.CurrentMonth))

			if status.MonthlyLimit > 0 {
				bar := progressBar(status.PercentUsed, 30)
				icon := ui.StatusIcon(true)
				if status.IsOverBudget {
					icon = ui.StatusIcon(false)
				} else if status.IsNearBudget {
					icon = ui.WarnIcon()
				}
				fmt.Printf("  Monthly:  %s $%.2f / $%.2f (%.0f%%)\n", icon, status.MonthlySpend, status.MonthlyLimit, status.PercentUsed)
				fmt.Printf("            %s\n", bar)
			}

			if status.DailyLimit > 0 {
				pct := 0.0
				if status.DailyLimit > 0 {
					pct = (status.DailySpend / status.DailyLimit) * 100
				}
				fmt.Printf("  Daily:    $%.2f / $%.2f (%.0f%%)\n", status.DailySpend, status.DailyLimit, pct)
			}

			if len(status.ByTool) > 0 {
				fmt.Println()
				fmt.Println("  By tool:")
				for tool, cost := range status.ByTool {
					fmt.Printf("    %-20s $%.4f\n", tool, cost)
				}
			}

			if len(status.ByProvider) > 0 {
				fmt.Println()
				fmt.Println("  By provider:")
				for provider, cost := range status.ByProvider {
					fmt.Printf("    %-20s $%.4f\n", provider, cost)
				}
			}

			if status.TotalTokens > 0 {
				fmt.Printf("\n  Total tokens: %d\n", status.TotalTokens)
			}
		},
	}
}

func budgetSetCmd() *cobra.Command {
	var monthly, daily float64
	var tool string

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set budget limits",
		Run: func(cmd *cobra.Command, args []string) {
			b := budget.Load()

			if monthly > 0 {
				b.MonthlyLimit = monthly
				ui.Good.Printf("  %s Monthly limit set to $%.2f\n", ui.StatusIcon(true), monthly)
			}
			if daily > 0 {
				b.DailyLimit = daily
				ui.Good.Printf("  %s Daily limit set to $%.2f\n", ui.StatusIcon(true), daily)
			}
			if tool != "" && len(args) > 0 {
				limit, err := strconv.ParseFloat(args[0], 64)
				if err != nil {
					ui.Bad.Printf("  Invalid amount: %s\n", args[0])
					os.Exit(1)
				}
				b.PerTool[tool] = limit
				ui.Good.Printf("  %s Limit for %s set to $%.2f/month\n", ui.StatusIcon(true), tool, limit)
			}

			if monthly == 0 && daily == 0 && tool == "" {
				fmt.Println("  Usage:")
				fmt.Println("    palm budget set --monthly 50")
				fmt.Println("    palm budget set --daily 10")
				fmt.Println("    palm budget set --tool aider 20")
				return
			}

			if err := budget.Save(b); err != nil {
				ui.Bad.Printf("  Failed to save budget: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().Float64Var(&monthly, "monthly", 0, "Monthly spending limit in USD")
	cmd.Flags().Float64Var(&daily, "daily", 0, "Daily spending limit in USD")
	cmd.Flags().StringVar(&tool, "tool", "", "Set per-tool monthly limit")
	return cmd
}

func budgetResetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset all budget limits",
		Run: func(cmd *cobra.Command, args []string) {
			b := &budget.Budget{
				AlertAt: 0.8,
				PerTool: make(map[string]float64),
			}
			if err := budget.Save(b); err != nil {
				ui.Bad.Printf("  Failed to reset budget: %v\n", err)
				os.Exit(1)
			}
			ui.Good.Printf("  %s Budget limits cleared\n", ui.StatusIcon(true))
		},
	}
}

func progressBar(percent float64, width int) string {
	if percent > 100 {
		percent = 100
	}
	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "\u2588" // █
		} else {
			bar += "\u2591" // ░
		}
	}
	return bar
}

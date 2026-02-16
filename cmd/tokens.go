package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/msalah0e/palm/internal/tokens"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

func tokensCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tokens",
		Aliases: []string{"tok"},
		Short:   "Count tokens and estimate context budget",
	}

	cmd.AddCommand(
		tokensCountCmd(),
		tokensBudgetCmd(),
		tokensTopCmd(),
	)

	return cmd
}

func tokensCountCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "count <file|dir>",
		Short: "Count tokens in files or directories",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			target := args[0]
			info, err := os.Stat(target)
			if err != nil {
				ui.Bad.Printf("  %v\n", err)
				os.Exit(1)
			}

			if !info.IsDir() {
				fr, err := tokens.CountFile(target)
				if err != nil {
					ui.Bad.Printf("  %v\n", err)
					os.Exit(1)
				}
				if jsonOutput {
					data, _ := json.MarshalIndent(fr, "", "  ")
					fmt.Println(string(data))
					return
				}
				fmt.Printf("  %s  %s tokens (%d lines, %d bytes)\n",
					ui.Brand.Sprint(target), tokens.FormatTokens(fr.Tokens), fr.Lines, fr.Bytes)
				return
			}

			result, err := tokens.ScanDir(target)
			if err != nil {
				ui.Bad.Printf("  Scan failed: %v\n", err)
				os.Exit(1)
			}

			if jsonOutput {
				data, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(data))
				return
			}

			ui.Banner("token count")
			fmt.Printf("  %s  %s\n", ui.Brand.Sprintf("%-12s", "Directory"), target)
			fmt.Printf("  %s  %d\n", ui.Brand.Sprintf("%-12s", "Files"), len(result.Files))
			fmt.Printf("  %s  %s\n", ui.Brand.Sprintf("%-12s", "Tokens"), tokens.FormatTokens(result.Total))
			fmt.Printf("  %s  %d\n", ui.Brand.Sprintf("%-12s", "Lines"), result.TotalLines)
			fmt.Printf("  %s  %.1f KB\n", ui.Brand.Sprintf("%-12s", "Size"), float64(result.TotalBytes)/1024)

			// Top 10 files
			fmt.Println()
			fmt.Println("  Top files by token count:")
			limit := 10
			if len(result.Files) < limit {
				limit = len(result.Files)
			}
			for i := 0; i < limit; i++ {
				f := result.Files[i]
				pct := float64(f.Tokens) / float64(result.Total) * 100
				fmt.Printf("  %s  %-6s  %5.1f%%  %s\n",
					ui.Subtle.Sprintf("%2d.", i+1),
					tokens.FormatTokens(f.Tokens),
					pct,
					f.Path)
			}
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

func tokensBudgetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "budget [dir]",
		Short: "Show how your project fits in model context windows",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}

			absDir, _ := filepath.Abs(dir)
			result, err := tokens.ScanDir(dir)
			if err != nil {
				ui.Bad.Printf("  Scan failed: %v\n", err)
				os.Exit(1)
			}

			ui.Banner("context budget")
			fmt.Printf("  Project: %s (%s tokens, %d files)\n\n",
				filepath.Base(absDir), tokens.FormatTokens(result.Total), len(result.Files))

			budgets := tokens.Budget(result.Total)
			var rows [][]string
			for _, b := range budgets {
				status := ui.StatusIcon(b.Fits)
				pctStr := fmt.Sprintf("%.1f%%", b.Percent)
				bar := renderBar(b.Percent, 20)
				rows = append(rows, []string{status, b.Model, tokens.FormatTokens(b.Window), pctStr, bar})
			}
			ui.Table([]string{"", "Model", "Context", "Used", "Budget"}, rows)
		},
	}
}

func tokensTopCmd() *cobra.Command {
	var count int

	cmd := &cobra.Command{
		Use:   "top [dir]",
		Short: "Show files with highest token counts",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}

			result, err := tokens.ScanDir(dir)
			if err != nil {
				ui.Bad.Printf("  Scan failed: %v\n", err)
				os.Exit(1)
			}

			ui.Banner("top files by tokens")
			limit := count
			if len(result.Files) < limit {
				limit = len(result.Files)
			}

			var rows [][]string
			for i := 0; i < limit; i++ {
				f := result.Files[i]
				pct := float64(f.Tokens) / float64(result.Total) * 100
				rows = append(rows, []string{
					tokens.FormatTokens(f.Tokens),
					fmt.Sprintf("%.1f%%", pct),
					fmt.Sprintf("%d", f.Lines),
					f.Path,
				})
			}
			ui.Table([]string{"Tokens", "%", "Lines", "File"}, rows)
			fmt.Printf("\n  Total: %s tokens across %d files\n", tokens.FormatTokens(result.Total), len(result.Files))
		},
	}

	cmd.Flags().IntVarP(&count, "count", "n", 20, "Number of files to show")
	return cmd
}

func renderBar(pct float64, width int) string {
	if pct > 100 {
		pct = 100
	}
	filled := int(pct / 100 * float64(width))
	if filled > width {
		filled = width
	}
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "\u2588"
		} else {
			bar += "\u2591"
		}
	}
	return bar
}

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

func benchmarkCmd() *cobra.Command {
	var iterations int

	cmd := &cobra.Command{
		Use:     "benchmark <command>",
		Aliases: []string{"bench"},
		Short:   "Benchmark AI tool commands — measure speed, reliability, and output quality",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			command := strings.Join(args, " ")

			ui.Banner("benchmark")
			fmt.Printf("  Command:    %s\n", ui.Brand.Sprint(command))
			fmt.Printf("  Iterations: %d\n\n", iterations)

			var durations []time.Duration
			successes := 0

			for i := 0; i < iterations; i++ {
				fmt.Printf("  Run %d/%d... ", i+1, iterations)

				start := time.Now()
				c := exec.Command("sh", "-c", command)
				c.Stdout = nil
				c.Stderr = nil
				err := c.Run()
				elapsed := time.Since(start)

				durations = append(durations, elapsed)
				if err == nil {
					successes++
					ui.Good.Printf("%s\n", elapsed.Round(time.Millisecond))
				} else {
					ui.Bad.Printf("%s (failed)\n", elapsed.Round(time.Millisecond))
				}
			}

			if len(durations) == 0 {
				return
			}

			var total time.Duration
			fastest := durations[0]
			slowest := durations[0]
			for _, d := range durations {
				total += d
				if d < fastest {
					fastest = d
				}
				if d > slowest {
					slowest = d
				}
			}
			avg := total / time.Duration(len(durations))

			fmt.Println()
			fmt.Printf("  %s  %s\n", ui.Brand.Sprintf("%-10s", "Fastest"), fastest.Round(time.Millisecond))
			fmt.Printf("  %s  %s\n", ui.Brand.Sprintf("%-10s", "Slowest"), slowest.Round(time.Millisecond))
			fmt.Printf("  %s  %s\n", ui.Brand.Sprintf("%-10s", "Average"), avg.Round(time.Millisecond))
			fmt.Printf("  %s  %d/%d (%.0f%%)\n", ui.Brand.Sprintf("%-10s", "Success"), successes, iterations,
				float64(successes)/float64(iterations)*100)
		},
	}

	cmd.Flags().IntVarP(&iterations, "iterations", "n", 3, "Number of iterations")
	return cmd
}

func benchmarkCompareCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "compare <cmd1> -- <cmd2>",
		Short: "Compare performance of two commands",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Split on "--"
			var cmd1Args, cmd2Args []string
			inSecond := false
			for _, a := range args {
				if a == "--" {
					inSecond = true
					continue
				}
				if inSecond {
					cmd2Args = append(cmd2Args, a)
				} else {
					cmd1Args = append(cmd1Args, a)
				}
			}

			if len(cmd1Args) == 0 || len(cmd2Args) == 0 {
				ui.Bad.Println("  Usage: palm benchmark compare <cmd1> -- <cmd2>")
				os.Exit(1)
			}

			c1 := strings.Join(cmd1Args, " ")
			c2 := strings.Join(cmd2Args, " ")

			ui.Banner("benchmark compare")

			d1 := benchRun(c1, 3)
			d2 := benchRun(c2, 3)

			fmt.Println()
			var rows [][]string
			rows = append(rows, []string{c1, d1.Round(time.Millisecond).String()})
			rows = append(rows, []string{c2, d2.Round(time.Millisecond).String()})
			ui.Table([]string{"Command", "Avg Time"}, rows)

			if d1 < d2 {
				fmt.Printf("\n  %s is %.1fx faster\n", ui.Brand.Sprint(c1), float64(d2)/float64(d1))
			} else if d2 < d1 {
				fmt.Printf("\n  %s is %.1fx faster\n", ui.Brand.Sprint(c2), float64(d1)/float64(d2))
			} else {
				fmt.Println("\n  Both commands are equally fast")
			}
		},
	}
}

func benchRun(command string, n int) time.Duration {
	var total time.Duration
	for i := 0; i < n; i++ {
		start := time.Now()
		c := exec.Command("sh", "-c", command)
		c.Stdout = nil
		c.Stderr = nil
		c.Run()
		total += time.Since(start)
	}
	return total / time.Duration(n)
}

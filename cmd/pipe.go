package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/msalah0e/palm/internal/ui"
	"github.com/msalah0e/palm/internal/vault"
	"github.com/spf13/cobra"
)

func pipeCmd() *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:   "pipe <tool1> | <tool2> [| tool3...]",
		Short: "Chain AI tools — pipe output from one tool to another",
		Long: `Chain multiple AI tools together, piping the output of each as input to the next.

  Examples:
    palm pipe "ollama run llama3.3 'summarize this'" "|" "aider --message"
    palm pipe "echo 'explain quicksort'" "|" "ollama run llama3.3"
    palm pipe "cat README.md" "|" "ollama run llama3.3 'review this code'"

  Each segment between | runs as a separate command with vault keys injected.
  The stdout of each command becomes the stdin of the next.`,
		Args:               cobra.MinimumNArgs(1),
		DisableFlagParsing: false,
		Run: func(cmd *cobra.Command, args []string) {
			// Parse pipeline segments by splitting on "|"
			segments := parsePipeSegments(args)
			if len(segments) < 2 {
				ui.Warn.Println("  Provide at least 2 commands separated by |")
				fmt.Println("  Example: palm pipe \"echo hello\" \"|\" \"ollama run llama3.3\"")
				os.Exit(1)
			}

			ui.Banner("pipe")
			v := vault.New()
			reg := loadRegistry()

			// Build environment with all vault keys
			env := os.Environ()
			keys, _ := v.List()
			for _, key := range keys {
				if val, err := v.Get(key); err == nil {
					if os.Getenv(key) == "" {
						env = append(env, fmt.Sprintf("%s=%s", key, val))
					}
				}
			}

			var lastOutput bytes.Buffer
			totalStart := time.Now()

			for i, segment := range segments {
				if len(segment) == 0 {
					continue
				}

				// Determine tool name for display
				toolName := segment[0]
				if t := reg.Get(toolName); t != nil {
					toolName = t.DisplayName
				}

				step := fmt.Sprintf("[%d/%d]", i+1, len(segments))
				if verbose {
					fmt.Printf("  %s %s: %s\n", ui.Subtle.Sprint(step), ui.Brand.Sprint(toolName), strings.Join(segment, " "))
				}

				var stdout bytes.Buffer
				c := exec.Command(segment[0], segment[1:]...)
				c.Env = env
				c.Stdout = &stdout
				c.Stderr = os.Stderr

				// First command gets no stdin, subsequent get previous output
				if i > 0 {
					c.Stdin = bytes.NewReader(lastOutput.Bytes())
				}

				start := time.Now()
				if err := c.Run(); err != nil {
					ui.Bad.Printf("  %s %s failed: %v\n", step, toolName, err)
					os.Exit(1)
				}
				elapsed := time.Since(start)

				if verbose {
					fmt.Printf("  %s %s completed in %s (%d bytes)\n",
						ui.StatusIcon(true), toolName,
						ui.Subtle.Sprintf("%.1fs", elapsed.Seconds()),
						stdout.Len())
				}

				lastOutput = stdout
			}

			totalElapsed := time.Since(totalStart)

			// Print final output
			if lastOutput.Len() > 0 {
				if verbose {
					fmt.Println()
					fmt.Printf("  %s Pipeline complete in %.1fs\n\n", ui.StatusIcon(true), totalElapsed.Seconds())
					fmt.Println("  " + strings.Repeat("─", 50))
				}
				fmt.Print(lastOutput.String())
			}
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show each pipeline step")
	return cmd
}

// parsePipeSegments splits args by "|" into command segments.
func parsePipeSegments(args []string) [][]string {
	var segments [][]string
	var current []string

	for _, arg := range args {
		if arg == "|" {
			if len(current) > 0 {
				segments = append(segments, current)
				current = nil
			}
		} else {
			// Split single arg that contains the full command
			parts := strings.Fields(arg)
			current = append(current, parts...)
		}
	}
	if len(current) > 0 {
		segments = append(segments, current)
	}

	return segments
}

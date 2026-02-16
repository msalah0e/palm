package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/msalah0e/tamr/internal/ui"
	"github.com/msalah0e/tamr/internal/vault"
	"github.com/spf13/cobra"
)

func runCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "run <tool> [args...]",
		Short:              "Run an AI tool with vault keys auto-injected",
		Args:               cobra.MinimumNArgs(1),
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			reg := loadRegistry()
			toolName := args[0]
			toolArgs := args[1:]

			tool := reg.Get(toolName)

			// Find the binary to execute
			bin := toolName
			if tool != nil {
				// Use the first word of the verify command as the binary name
				if tool.Install.Verify.Command != "" {
					parts := strings.Fields(tool.Install.Verify.Command)
					if len(parts) > 0 {
						bin = parts[0]
					}
				}
			}

			binPath, err := exec.LookPath(bin)
			if err != nil {
				ui.Bad.Printf("tamr: %s not found in PATH\n", bin)
				if tool == nil {
					fmt.Println("  This tool is not in the registry either.")
					fmt.Println("  Run `tamr search` to find available tools")
				} else {
					fmt.Printf("  Run `tamr install %s` first\n", toolName)
				}
				os.Exit(1)
			}

			// Build environment with vault keys injected
			env := os.Environ()
			v := vault.New()

			if tool != nil {
				allKeys := append(tool.Keys.Required, tool.Keys.Optional...)
				injected := 0
				for _, key := range allKeys {
					// Don't override keys already set in environment
					if os.Getenv(key) != "" {
						continue
					}
					val, err := v.Get(key)
					if err == nil {
						env = append(env, fmt.Sprintf("%s=%s", key, val))
						injected++
					}
				}
				if injected > 0 {
					ui.Subtle.Fprintf(os.Stderr, "tamr: injected %d key(s) from vault\n", injected)
				}
			}

			// Replace this process with the tool
			if err := syscall.Exec(binPath, append([]string{bin}, toolArgs...), env); err != nil {
				ui.Bad.Printf("tamr: failed to exec %s: %v\n", bin, err)
				os.Exit(1)
			}
		},
	}
}

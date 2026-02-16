package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/msalah0e/palm/internal/registry"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

func healthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "health",
		Aliases: []string{"status"},
		Short:   "System health overview — tools, runtimes, configs, and disk usage",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("system health")

			// OS info
			fmt.Printf("  %s  %s/%s\n", ui.Brand.Sprintf("%-12s", "Platform"), runtime.GOOS, runtime.GOARCH)
			fmt.Printf("  %s  %s\n", ui.Brand.Sprintf("%-12s", "Go"), runtime.Version())

			// palm config dir
			configDir := palmConfigDir()
			configSize := dirSize(configDir)
			fmt.Printf("  %s  %s (%.1f KB)\n", ui.Brand.Sprintf("%-12s", "Config"), configDir, float64(configSize)/1024)

			// Installed tools
			reg := loadRegistry()
			detected := registry.DetectInstalled(reg)
			fmt.Printf("  %s  %d installed, %d in registry\n", ui.Brand.Sprintf("%-12s", "Tools"), len(detected), len(reg.All()))

			// Data files
			fmt.Println()
			fmt.Println("  Data files:")
			dataFiles := map[string]string{
				"vault.enc":       "API key vault",
				"graph.enc":       "Knowledge graph",
				"sessions.jsonl":  "Session history",
				"activity.jsonl":  "Activity log",
				"budget.json":     "Budget config",
				"state.json":      "State tracking",
			}
			for f, desc := range dataFiles {
				path := filepath.Join(configDir, f)
				if info, err := os.Stat(path); err == nil {
					fmt.Printf("    %s %-20s %s (%.1f KB)\n", ui.StatusIcon(true), f, ui.Subtle.Sprint(desc), float64(info.Size())/1024)
				}
			}

			// Prompts
			promptDir := filepath.Join(configDir, "prompts")
			if entries, err := os.ReadDir(promptDir); err == nil && len(entries) > 0 {
				fmt.Printf("    %s %-20s %d prompts\n", ui.StatusIcon(true), "prompts/", len(entries))
			}
		},
	}

	cmd.AddCommand(
		healthCheckCmd(),
	)

	return cmd
}

func healthCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Run comprehensive health checks",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("health check")

			checks := []struct {
				name  string
				check func() (bool, string)
			}{
				{"Config directory", func() (bool, string) {
					dir := palmConfigDir()
					_, err := os.Stat(dir)
					return err == nil, dir
				}},
				{"Vault encryption", func() (bool, string) {
					path := filepath.Join(palmConfigDir(), "vault.enc")
					info, err := os.Stat(path)
					if err != nil {
						return false, "no vault file"
					}
					return info.Size() > 0, fmt.Sprintf("%.1f KB", float64(info.Size())/1024)
				}},
				{"Graph encryption", func() (bool, string) {
					path := filepath.Join(palmConfigDir(), "graph.enc")
					_, err := os.Stat(path)
					return err == nil, "graph.enc exists"
				}},
				{"Git available", func() (bool, string) {
					out, err := exec.Command("git", "--version").Output()
					if err != nil {
						return false, "not found"
					}
					return true, strings.TrimSpace(string(out))
				}},
				{"Shell completion", func() (bool, string) {
					shell := os.Getenv("SHELL")
					if shell == "" {
						return false, "SHELL not set"
					}
					return true, filepath.Base(shell)
				}},
				{"Disk space", func() (bool, string) {
					size := dirSize(palmConfigDir())
					return size < 100*1024*1024, fmt.Sprintf("%.1f MB used", float64(size)/(1024*1024))
				}},
			}

			passed := 0
			for _, c := range checks {
				ok, detail := c.check()
				if ok {
					passed++
				}
				fmt.Printf("  %s %-25s %s\n", ui.StatusIcon(ok), c.name, ui.Subtle.Sprint(detail))
			}

			fmt.Printf("\n  %d/%d checks passed\n", passed, len(checks))
		},
	}
}

func dirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		size += info.Size()
		return nil
	})
	return size
}

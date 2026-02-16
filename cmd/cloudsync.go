package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

func cloudsyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sync",
		Aliases: []string{"cloud"},
		Short:   "Cross-machine sync — backup and restore palm state",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("sync status")

			configDir := palmConfigDir()
			files := []string{"vault.enc", "graph.enc", "sessions.jsonl", "activity.jsonl", "budget.json", "state.json"}

			fmt.Printf("  Config dir: %s\n\n", configDir)

			existing := 0
			for _, f := range files {
				path := filepath.Join(configDir, f)
				info, err := os.Stat(path)
				if err == nil {
					size := float64(info.Size()) / 1024
					fmt.Printf("  %s %-20s  %.1f KB\n", ui.StatusIcon(true), f, size)
					existing++
				} else {
					ui.Subtle.Printf("  %s %-20s  not found\n", "-", f)
				}
			}

			// Check for prompts
			promptDir := filepath.Join(configDir, "prompts")
			if entries, err := os.ReadDir(promptDir); err == nil && len(entries) > 0 {
				fmt.Printf("  %s %-20s  %d files\n", ui.StatusIcon(true), "prompts/", len(entries))
				existing++
			}

			fmt.Printf("\n  %d data files found\n", existing)
			fmt.Println()
			fmt.Println("  Run `palm sync export <path>` to backup")
			fmt.Println("  Run `palm sync import <path>` to restore")
		},
	}

	cmd.AddCommand(
		syncExportCmd(),
		syncImportCmd(),
	)

	return cmd
}

func syncExportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "export <path>",
		Short: "Export palm state to a backup directory",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dest := args[0]
			ui.Banner("sync export")

			if err := os.MkdirAll(dest, 0o755); err != nil {
				ui.Bad.Printf("  Failed to create %s: %v\n", dest, err)
				os.Exit(1)
			}

			configDir := palmConfigDir()
			files := []string{"vault.enc", "graph.enc", "sessions.jsonl", "activity.jsonl", "budget.json", "state.json"}

			copied := 0
			for _, f := range files {
				src := filepath.Join(configDir, f)
				if _, err := os.Stat(src); os.IsNotExist(err) {
					continue
				}
				data, err := os.ReadFile(src)
				if err != nil {
					ui.Bad.Printf("  Failed to read %s: %v\n", f, err)
					continue
				}
				dstPath := filepath.Join(dest, f)
				if err := os.WriteFile(dstPath, data, 0o600); err != nil {
					ui.Bad.Printf("  Failed to write %s: %v\n", dstPath, err)
					continue
				}
				ui.Good.Printf("  %s %s\n", ui.StatusIcon(true), f)
				copied++
			}

			// Copy prompts directory
			promptDir := filepath.Join(configDir, "prompts")
			if entries, err := os.ReadDir(promptDir); err == nil && len(entries) > 0 {
				destPrompts := filepath.Join(dest, "prompts")
				os.MkdirAll(destPrompts, 0o755)
				for _, e := range entries {
					data, _ := os.ReadFile(filepath.Join(promptDir, e.Name()))
					os.WriteFile(filepath.Join(destPrompts, e.Name()), data, 0o644)
				}
				ui.Good.Printf("  %s prompts/ (%d files)\n", ui.StatusIcon(true), len(entries))
				copied++
			}

			fmt.Printf("\n  Exported %d items to %s\n", copied, dest)
		},
	}
}

func syncImportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "import <path>",
		Short: "Import palm state from a backup directory",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			src := args[0]
			ui.Banner("sync import")

			if _, err := os.Stat(src); os.IsNotExist(err) {
				ui.Bad.Printf("  Backup not found: %s\n", src)
				os.Exit(1)
			}

			configDir := palmConfigDir()
			os.MkdirAll(configDir, 0o755)

			files := []string{"vault.enc", "graph.enc", "sessions.jsonl", "activity.jsonl", "budget.json", "state.json"}

			restored := 0
			for _, f := range files {
				srcPath := filepath.Join(src, f)
				if _, err := os.Stat(srcPath); os.IsNotExist(err) {
					continue
				}
				data, err := os.ReadFile(srcPath)
				if err != nil {
					ui.Bad.Printf("  Failed to read %s: %v\n", f, err)
					continue
				}
				dstPath := filepath.Join(configDir, f)
				if err := os.WriteFile(dstPath, data, 0o600); err != nil {
					ui.Bad.Printf("  Failed to write %s: %v\n", f, err)
					continue
				}
				ui.Good.Printf("  %s %s\n", ui.StatusIcon(true), f)
				restored++
			}

			// Restore prompts
			srcPrompts := filepath.Join(src, "prompts")
			if entries, err := os.ReadDir(srcPrompts); err == nil && len(entries) > 0 {
				destPrompts := filepath.Join(configDir, "prompts")
				os.MkdirAll(destPrompts, 0o755)
				for _, e := range entries {
					data, _ := os.ReadFile(filepath.Join(srcPrompts, e.Name()))
					os.WriteFile(filepath.Join(destPrompts, e.Name()), data, 0o644)
				}
				ui.Good.Printf("  %s prompts/ (%d files)\n", ui.StatusIcon(true), len(entries))
				restored++
			}

			fmt.Printf("\n  Restored %d items from %s\n", restored, src)
		},
	}
}

func palmConfigDir() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "palm")
}

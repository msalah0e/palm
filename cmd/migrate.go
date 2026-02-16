package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

// migrationPath defines a tool-to-tool migration with config mappings.
type migrationPath struct {
	From       string
	To         string
	ConfigMap  map[string]string // source file → destination file
	KeyMap     map[string]string // source env var → destination env var
}

var migrationPaths = []migrationPath{
	{
		From: "cursor",
		To:   "claude-code",
		ConfigMap: map[string]string{
			".cursorrules":      "CLAUDE.md",
			".cursor/rules/*.mdc": "CLAUDE.md",
		},
		KeyMap: map[string]string{
			"OPENAI_API_KEY": "ANTHROPIC_API_KEY",
		},
	},
	{
		From: "copilot",
		To:   "claude-code",
		ConfigMap: map[string]string{
			".github/copilot-instructions.md": "CLAUDE.md",
		},
	},
	{
		From: "cursor",
		To:   "windsurf",
		ConfigMap: map[string]string{
			".cursorrules": ".windsurfrules",
		},
	},
	{
		From: "windsurf",
		To:   "cursor",
		ConfigMap: map[string]string{
			".windsurfrules": ".cursorrules",
		},
	},
	{
		From: "aider",
		To:   "claude-code",
		ConfigMap: map[string]string{
			".aider.conf.yml": "CLAUDE.md",
		},
	},
	{
		From: "codex",
		To:   "claude-code",
		ConfigMap: map[string]string{
			"AGENTS.md": "CLAUDE.md",
		},
	},
}

func migrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "migrate <from> <to>",
		Aliases: []string{"mig"},
		Short:   "Migrate configs and keys between AI tools",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			from, to := args[0], args[1]

			ui.Banner("tool migration")
			fmt.Printf("  %s → %s\n\n", ui.Brand.Sprint(from), ui.Brand.Sprint(to))

			var path *migrationPath
			for i := range migrationPaths {
				if migrationPaths[i].From == from && migrationPaths[i].To == to {
					path = &migrationPaths[i]
					break
				}
			}

			if path == nil {
				ui.Warn.Printf("  No migration path from %s to %s\n", from, to)
				fmt.Println()
				fmt.Println("  Available migrations:")
				for _, mp := range migrationPaths {
					fmt.Printf("    %s → %s\n", mp.From, mp.To)
				}
				return
			}

			migrated := 0
			for src, dst := range path.ConfigMap {
				if _, err := os.Stat(src); os.IsNotExist(err) {
					ui.Subtle.Printf("  %s: not found, skipping\n", src)
					continue
				}

				data, err := os.ReadFile(src)
				if err != nil {
					ui.Bad.Printf("  Failed to read %s: %v\n", src, err)
					continue
				}

				content := wrapMigrated(to, from, string(data))

				if _, err := os.Stat(dst); err == nil {
					ui.Warn.Printf("  %s %s already exists — appending migrated content\n", ui.WarnIcon(), dst)
					existing, _ := os.ReadFile(dst)
					content = string(existing) + "\n\n# Migrated from " + from + " by palm\n" + content
				}

				if err := os.WriteFile(dst, []byte(content), 0o644); err != nil {
					ui.Bad.Printf("  Failed to write %s: %v\n", dst, err)
					continue
				}
				ui.Good.Printf("  %s %s → %s\n", ui.StatusIcon(true), src, dst)
				migrated++
			}

			if len(path.KeyMap) > 0 {
				fmt.Println()
				fmt.Println("  API key mappings:")
				for src, dst := range path.KeyMap {
					if val := os.Getenv(src); val != "" {
						ui.Good.Printf("    %s %s is set → use as %s\n", ui.StatusIcon(true), src, dst)
					} else {
						ui.Subtle.Printf("    - %s not set (maps to %s)\n", src, dst)
					}
				}
			}

			fmt.Println()
			if migrated > 0 {
				fmt.Printf("  %d config files migrated\n", migrated)
			} else {
				fmt.Println("  No config files found to migrate")
			}
		},
	}

	cmd.AddCommand(migrateListCmd())
	return cmd
}

func migrateListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Show available migration paths",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("migration paths")

			var rows [][]string
			for _, mp := range migrationPaths {
				files := []string{}
				for src, dst := range mp.ConfigMap {
					files = append(files, src+" → "+dst)
				}
				rows = append(rows, []string{mp.From, mp.To, strings.Join(files, ", ")})
			}
			ui.Table([]string{"From", "To", "Config Mapping"}, rows)
			fmt.Printf("\n  %d migration paths\n", len(migrationPaths))
		},
	}
}

func wrapMigrated(tool, from, content string) string {
	header := fmt.Sprintf("# Migrated from %s by palm\n# Review and adjust for %s conventions\n\n", from, tool)
	return header + content
}

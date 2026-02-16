package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

// ruleFiles maps AI tool → its rules file (superset of contextFiles).
var ruleFiles = map[string]string{
	"claude-code": "CLAUDE.md",
	"cursor":      ".cursor/rules/palm.mdc",
	"copilot":     ".github/copilot-instructions.md",
	"codex":       "AGENTS.md",
	"windsurf":    ".windsurfrules",
	"aider":       ".aider.conf.yml",
	"gemini":      "GEMINI.md",
	"trae":        ".trae/rules/palm.md",
}

func rulesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rules",
		Short:   "Universal AI rules — sync instructions across all tools",
		Aliases: []string{"rule"},
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("AI rules")

			// Check for .palm-rules.md or .palm-context.md
			source := findRulesSource()
			if source == "" {
				fmt.Println("  No rules source found.")
				fmt.Println("  Run `palm rules init` to create .palm-rules.md")
				return
			}

			fmt.Printf("  Source: %s\n\n", ui.Brand.Sprint(source))

			synced := 0
			for tool, file := range ruleFiles {
				if _, err := os.Stat(file); err == nil {
					fmt.Printf("  %s %-14s → %s\n", ui.StatusIcon(true), tool, file)
					synced++
				}
			}

			if synced == 0 {
				fmt.Println("  No tool-specific files found")
				fmt.Println("  Run `palm rules sync` to generate them")
			} else {
				fmt.Printf("\n  %d tool configs synced\n", synced)
			}
		},
	}

	cmd.AddCommand(
		rulesInitCmd(),
		rulesSyncCmd(),
		rulesAddCmd(),
		rulesCheckCmd(),
	)

	return cmd
}

func rulesInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create .palm-rules.md as your single source of truth",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("rules init")

			path := ".palm-rules.md"

			// If .palm-context.md exists, migrate it
			if data, err := os.ReadFile(".palm-context.md"); err == nil {
				if _, err := os.Stat(path); os.IsNotExist(err) {
					content := "# Project Rules\n\n" +
						"<!-- palm rules — edit this file and run `palm rules sync` -->\n" +
						"<!-- This is your single source of truth for all AI tool instructions -->\n\n" +
						string(data)
					os.WriteFile(path, []byte(content), 0o644)
					ui.Good.Printf("  %s Created %s (migrated from .palm-context.md)\n", ui.StatusIcon(true), path)
					return
				}
			}

			if _, err := os.Stat(path); err == nil {
				fmt.Printf("  %s already exists\n", path)
				return
			}

			lang, framework := detectProject()

			var b strings.Builder
			b.WriteString("# Project Rules\n\n")
			b.WriteString("<!-- palm rules — edit this file and run `palm rules sync` -->\n")
			b.WriteString("<!-- This is your single source of truth for all AI tool instructions -->\n\n")
			b.WriteString(fmt.Sprintf("Language: %s\n", lang))
			if framework != "" {
				b.WriteString(fmt.Sprintf("Framework: %s\n", framework))
			}
			b.WriteString("\n## Guidelines\n\n")
			b.WriteString("- Follow existing code patterns and conventions\n")
			b.WriteString("- Write tests for new functionality\n")
			b.WriteString("- Keep changes focused and minimal\n")
			b.WriteString("- Do not add unnecessary comments or documentation\n")
			b.WriteString("\n## Project Structure\n\n")
			b.WriteString("<!-- Describe your project structure here -->\n")
			b.WriteString("\n## Key Files\n\n")
			b.WriteString("<!-- List important files and their purpose -->\n")
			b.WriteString("\n## Do NOT\n\n")
			b.WriteString("- Do not modify files outside the scope of the task\n")
			b.WriteString("- Do not introduce new dependencies without asking\n")
			b.WriteString("- Do not change formatting/style of code you didn't write\n")

			os.WriteFile(path, []byte(b.String()), 0o644)
			ui.Good.Printf("  %s Created %s\n", ui.StatusIcon(true), path)
			fmt.Println()
			fmt.Println("  Edit it with your project rules, then run `palm rules sync`")
		},
	}
}

func rulesSyncCmd() *cobra.Command {
	var tools []string

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync rules to all AI tool config files",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("rules sync")

			source := findRulesSource()
			if source == "" {
				ui.Warn.Println("  No .palm-rules.md or .palm-context.md found")
				fmt.Println("  Run `palm rules init` first")
				os.Exit(1)
			}

			baseContent, err := os.ReadFile(source)
			if err != nil {
				ui.Bad.Printf("  Failed to read %s: %v\n", source, err)
				os.Exit(1)
			}

			targetTools := ruleFiles
			if len(tools) > 0 {
				targetTools = make(map[string]string)
				for _, t := range tools {
					if f, ok := ruleFiles[t]; ok {
						targetTools[t] = f
					}
				}
			}

			synced := 0
			for tool, file := range targetTools {
				dir := filepath.Dir(file)
				if dir != "." {
					os.MkdirAll(dir, 0o755)
				}

				content := wrapRulesForTool(tool, string(baseContent))
				if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
					ui.Bad.Printf("  %s %s: %v\n", ui.StatusIcon(false), file, err)
					continue
				}
				ui.Good.Printf("  %s synced → %s (%s)\n", ui.StatusIcon(true), file, tool)
				synced++
			}

			fmt.Printf("\n  %d files synced from %s\n", synced, source)
		},
	}

	cmd.Flags().StringSliceVar(&tools, "tools", nil, "Specific tools to sync (default: all)")
	return cmd
}

func rulesAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <rule>",
		Short: "Append a rule to .palm-rules.md",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			source := findRulesSource()
			if source == "" {
				source = ".palm-rules.md"
			}

			f, err := os.OpenFile(source, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if err != nil {
				ui.Bad.Printf("  Failed to open %s: %v\n", source, err)
				os.Exit(1)
			}
			defer f.Close()

			fmt.Fprintf(f, "\n- %s\n", args[0])
			ui.Good.Printf("  %s Added rule to %s\n", ui.StatusIcon(true), source)
			fmt.Println("  Run `palm rules sync` to apply")
		},
	}
}

func rulesCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Check which tool configs are in sync with rules",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("rules check")

			source := findRulesSource()
			if source == "" {
				fmt.Println("  No rules source found")
				return
			}

			sourceInfo, _ := os.Stat(source)
			fmt.Printf("  Source: %s (modified %s)\n\n", source, sourceInfo.ModTime().Format("Jan 02 15:04"))

			for tool, file := range ruleFiles {
				info, err := os.Stat(file)
				if err != nil {
					ui.Subtle.Printf("  %s %-14s  not created\n", "-", tool)
					continue
				}
				if info.ModTime().Before(sourceInfo.ModTime()) {
					ui.Warn.Printf("  %s %-14s  STALE (older than source)\n", ui.WarnIcon(), tool)
				} else {
					fmt.Printf("  %s %-14s  in sync\n", ui.StatusIcon(true), tool)
				}
			}
		},
	}
}

func findRulesSource() string {
	for _, name := range []string{".palm-rules.md", ".palm-context.md"} {
		if _, err := os.Stat(name); err == nil {
			return name
		}
	}
	return ""
}

func wrapRulesForTool(tool, content string) string {
	header := fmt.Sprintf("# %s Rules\n# Generated by palm rules — edit .palm-rules.md and run `palm rules sync`\n# Do not edit this file directly.\n\n", titleCase(tool))
	return header + content
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

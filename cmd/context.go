package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

// contextFile maps AI tool → its context file.
var contextFiles = map[string]string{
	"claude-code": "CLAUDE.md",
	"cursor":      ".cursorrules",
	"aider":       ".aider.conf.yml",
	"continue":    ".continuerc.json",
	"copilot":     ".github/copilot-instructions.md",
	"windsurf":    ".windsurfrules",
	"codex":       "AGENTS.md",
}

func contextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Manage unified AI tool context/instructions",
	}

	cmd.AddCommand(
		contextInitCmd(),
		contextShowCmd(),
		contextSyncCmd(),
	)

	return cmd
}

func contextInitCmd() *cobra.Command {
	var tools []string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Generate context files for AI tools in this project",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("context init")

			// Detect project type
			lang, framework := detectProject()
			fmt.Printf("  Detected: %s", lang)
			if framework != "" {
				fmt.Printf(" + %s", framework)
			}
			fmt.Println()

			// Create .palm-context.md as the single source of truth
			contextPath := ".palm-context.md"
			if _, err := os.Stat(contextPath); err == nil {
				fmt.Printf("  %s already exists\n", contextPath)
			} else {
				content := generateContext(lang, framework)
				if err := os.WriteFile(contextPath, []byte(content), 0o644); err != nil {
					ui.Bad.Printf("  Failed to create %s: %v\n", contextPath, err)
					os.Exit(1)
				}
				ui.Good.Printf("  %s Created %s\n", ui.StatusIcon(true), contextPath)
			}

			// Generate tool-specific files
			if len(tools) == 0 {
				// Auto-detect which tools are installed
				reg := loadRegistry()
				for toolName := range contextFiles {
					if t := reg.Get(toolName); t != nil {
						tools = append(tools, toolName)
					}
				}
				// Always generate for common ones
				if len(tools) == 0 {
					tools = []string{"claude-code", "cursor", "copilot"}
				}
			}

			for _, tool := range tools {
				file, ok := contextFiles[tool]
				if !ok {
					ui.Warn.Printf("  Unknown context format for %s\n", tool)
					continue
				}

				// Create directory if needed
				dir := filepath.Dir(file)
				if dir != "." {
					_ = os.MkdirAll(dir, 0o755)
				}

				if _, err := os.Stat(file); err == nil {
					ui.Subtle.Printf("  %s already exists, skipping\n", file)
					continue
				}

				content := generateToolContext(tool, lang, framework)
				if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
					ui.Bad.Printf("  Failed to create %s: %v\n", file, err)
					continue
				}
				ui.Good.Printf("  %s Created %s (%s)\n", ui.StatusIcon(true), file, tool)
			}

			fmt.Println()
			fmt.Println("  Edit .palm-context.md with your project-specific instructions")
			fmt.Println("  Run `palm context sync` to update tool-specific files")
		},
	}

	cmd.Flags().StringSliceVar(&tools, "tools", nil, "Specific tools to generate context for")
	return cmd
}

func contextShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Display current project context",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("project context")

			contextPath := ".palm-context.md"
			if data, err := os.ReadFile(contextPath); err == nil {
				fmt.Printf("  Source: %s\n\n", contextPath)
				fmt.Println(string(data))
			} else {
				fmt.Println("  No .palm-context.md found.")
				fmt.Println("  Run `palm context init` to create one")
				return
			}

			fmt.Println()
			fmt.Println("  Tool-specific files:")
			for tool, file := range contextFiles {
				if _, err := os.Stat(file); err == nil {
					fmt.Printf("  %s %-15s → %s\n", ui.StatusIcon(true), tool, file)
				}
			}
		},
	}
}

func contextSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync .palm-context.md to tool-specific files",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("context sync")

			contextPath := ".palm-context.md"
			baseContent, err := os.ReadFile(contextPath)
			if err != nil {
				ui.Warn.Println("  No .palm-context.md found. Run `palm context init` first")
				os.Exit(1)
			}

			synced := 0
			for tool, file := range contextFiles {
				if _, err := os.Stat(file); err != nil {
					continue // only sync existing files
				}

				content := wrapForTool(tool, string(baseContent))

				dir := filepath.Dir(file)
				if dir != "." {
					_ = os.MkdirAll(dir, 0o755)
				}

				if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
					ui.Bad.Printf("  %s %s: %v\n", ui.StatusIcon(false), file, err)
					continue
				}
				ui.Good.Printf("  %s synced → %s\n", ui.StatusIcon(true), file)
				synced++
			}

			if synced == 0 {
				fmt.Println("  No tool context files found to sync.")
				fmt.Println("  Run `palm context init` first")
			} else {
				fmt.Printf("\n  %d files synced from .palm-context.md\n", synced)
			}
		},
	}
}

func detectProject() (lang, framework string) {
	checks := []struct {
		file      string
		lang      string
		framework string
	}{
		{"go.mod", "Go", ""},
		{"Cargo.toml", "Rust", ""},
		{"package.json", "JavaScript/TypeScript", ""},
		{"pyproject.toml", "Python", ""},
		{"requirements.txt", "Python", ""},
		{"Gemfile", "Ruby", ""},
		{"pom.xml", "Java", "Maven"},
		{"build.gradle", "Java", "Gradle"},
		{"mix.exs", "Elixir", ""},
	}

	for _, c := range checks {
		if _, err := os.Stat(c.file); err == nil {
			lang = c.lang
			framework = c.framework
			break
		}
	}

	if lang == "" {
		lang = "Unknown"
	}

	// Detect framework from package.json
	if lang == "JavaScript/TypeScript" {
		if data, err := os.ReadFile("package.json"); err == nil {
			content := string(data)
			switch {
			case strings.Contains(content, "\"react\""):
				framework = "React"
			case strings.Contains(content, "\"next\""):
				framework = "Next.js"
			case strings.Contains(content, "\"vue\""):
				framework = "Vue"
			case strings.Contains(content, "\"svelte\""):
				framework = "Svelte"
			case strings.Contains(content, "\"express\""):
				framework = "Express"
			}
		}
	}

	// Detect Python framework
	if lang == "Python" {
		if data, err := os.ReadFile("pyproject.toml"); err == nil {
			content := string(data)
			switch {
			case strings.Contains(content, "django"):
				framework = "Django"
			case strings.Contains(content, "fastapi"):
				framework = "FastAPI"
			case strings.Contains(content, "flask"):
				framework = "Flask"
			}
		}
	}

	return
}

func generateContext(lang, framework string) string {
	var b strings.Builder
	b.WriteString("# Project Context\n\n")
	b.WriteString(fmt.Sprintf("Language: %s\n", lang))
	if framework != "" {
		b.WriteString(fmt.Sprintf("Framework: %s\n", framework))
	}
	b.WriteString("\n## Guidelines\n\n")
	b.WriteString("- Follow existing code patterns and conventions\n")
	b.WriteString("- Write tests for new functionality\n")
	b.WriteString("- Keep changes focused and minimal\n")

	switch lang {
	case "Go":
		b.WriteString("- Use `gofmt` for formatting\n")
		b.WriteString("- Handle errors explicitly, don't ignore them\n")
		b.WriteString("- Follow Go naming conventions (camelCase for unexported)\n")
	case "Python":
		b.WriteString("- Use type hints\n")
		b.WriteString("- Follow PEP 8 style guide\n")
	case "JavaScript/TypeScript":
		b.WriteString("- Use TypeScript when possible\n")
		b.WriteString("- Prefer functional patterns\n")
	case "Rust":
		b.WriteString("- Use `cargo fmt` for formatting\n")
		b.WriteString("- Handle errors with Result, avoid unwrap in production code\n")
	}

	b.WriteString("\n## Project Structure\n\n")
	b.WriteString("<!-- Describe your project structure here -->\n\n")
	b.WriteString("## Key Files\n\n")
	b.WriteString("<!-- List important files and their purpose -->\n")

	return b.String()
}

func generateToolContext(tool, lang, framework string) string {
	base := generateContext(lang, framework)

	switch tool {
	case "aider":
		return "# aider configuration\n# See: https://aider.chat/docs/config/aider_conf.html\n\n" +
			"auto-commits: false\nmap-tokens: 2048\n"
	case "continue":
		return `{
  "models": [],
  "customCommands": [],
  "contextProviders": [],
  "slashCommands": []
}
`
	default:
		return wrapForTool(tool, base)
	}
}

func wrapForTool(tool, content string) string {
	switch tool {
	case "cursor":
		return "# Cursor Rules\n# Generated by palm context — edit .palm-context.md and run `palm context sync`\n\n" + content
	case "copilot":
		return "# GitHub Copilot Instructions\n# Generated by palm context — edit .palm-context.md and run `palm context sync`\n\n" + content
	case "claude-code":
		return "# Claude Code Instructions\n# Generated by palm context — edit .palm-context.md and run `palm context sync`\n\n" + content
	case "codex":
		return "# OpenAI Codex Agent Instructions\n# Generated by palm context — edit .palm-context.md and run `palm context sync`\n\n" + content
	case "windsurf":
		return "# Windsurf Rules\n# Generated by palm context — edit .palm-context.md and run `palm context sync`\n\n" + content
	default:
		return content
	}
}

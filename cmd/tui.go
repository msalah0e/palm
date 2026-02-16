package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/msalah0e/palm/internal/registry"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/msalah0e/palm/internal/vault"
	"github.com/spf13/cobra"
)

// project represents a discovered project directory.
type project struct {
	Path        string
	Name        string
	Marker      string // go.mod, package.json, etc.
	Tools       []string
	HasPalmTOML bool
}

func tuiCmd() *cobra.Command {
	var scanDir string

	cmd := &cobra.Command{
		Use:     "ui",
		Aliases: []string{"tui"},
		Short:   "Interactive project navigator and tool browser",
		Long: `Browse projects and AI tools in a visual terminal interface.

Scans for project directories and shows which AI tools are configured.
In v1.1.0, this runs in static/list mode. A full interactive TUI
with bubbletea will be added in a future release.

  palm ui                  # Scan current directory
  palm ui --dir ~/Projects # Scan specific directory`,
		Run: func(cmd *cobra.Command, args []string) {
			if scanDir == "" {
				scanDir, _ = os.Getwd()
			}

			reg := loadRegistry()
			v := vault.New()

			printUIHeader()

			// Discover projects
			projects := discoverProjects(scanDir)

			if len(projects) == 0 {
				fmt.Println("  No projects found in", scanDir)
				fmt.Println("  Try: palm ui --dir ~/Projects")
				return
			}

			// Count vault keys
			keys, _ := v.List()
			keyCount := len(keys)

			fmt.Printf("  Scanning: %s\n", ui.Subtle.Sprint(scanDir))
			fmt.Printf("  Found:    %d projects · %d vault keys\n", len(projects), keyCount)
			fmt.Println()

			for _, p := range projects {
				// Check installed tools
				var installedTools []string
				var missingTools []string

				for _, toolName := range p.Tools {
					tool := reg.Get(toolName)
					if tool == nil {
						continue
					}
					dt := registry.DetectOne(*tool)
					if dt.Installed {
						installedTools = append(installedTools, toolName)
					} else {
						missingTools = append(missingTools, toolName)
					}
				}

				// Project header
				icon := "📁"
				if p.HasPalmTOML {
					icon = "🌴"
				}
				fmt.Printf("  %s %s\n", icon, ui.Brand.Sprint(p.Name))
				fmt.Printf("     %s %s\n", ui.Subtle.Sprint("→"), p.Path)

				if p.Marker != "" {
					fmt.Printf("     Type: %s\n", p.Marker)
				}

				if len(installedTools) > 0 {
					fmt.Printf("     Tools: %s %s\n", ui.StatusIcon(true), strings.Join(installedTools, ", "))
				}
				if len(missingTools) > 0 {
					fmt.Printf("     Missing: %s %s\n", ui.WarnIcon(), strings.Join(missingTools, ", "))
				}

				if len(p.Tools) == 0 {
					fmt.Printf("     Tools: %s\n", ui.Subtle.Sprint("none configured"))
				}

				fmt.Println()
			}

			fmt.Println("  " + strings.Repeat("─", 50))
			fmt.Println("  Tip: Add a .palm.toml to projects for tool configuration")
			fmt.Println("       palm workspace init")
		},
	}

	cmd.Flags().StringVar(&scanDir, "dir", "", "Directory to scan for projects")
	return cmd
}

func printUIHeader() {
	fmt.Println()
	fmt.Println(ui.Brand.Sprint("  ┌─────────────────────────────────────────────────┐"))
	fmt.Println(ui.Brand.Sprint("  │") + "  🌴 " + ui.Brand.Sprint("palm ui") + " — project navigator              " + ui.Brand.Sprint("│"))
	fmt.Println(ui.Brand.Sprint("  └─────────────────────────────────────────────────┘"))
	fmt.Println()
}

// discoverProjects scans a directory for project directories.
func discoverProjects(root string) []project {
	var projects []project

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}

	markers := []struct {
		file string
		desc string
	}{
		{".palm.toml", "palm workspace"},
		{"go.mod", "Go"},
		{"package.json", "Node.js"},
		{"pyproject.toml", "Python"},
		{"Cargo.toml", "Rust"},
		{"pom.xml", "Java (Maven)"},
		{"build.gradle", "Java (Gradle)"},
		{"mix.exs", "Elixir"},
		{"Gemfile", "Ruby"},
		{"requirements.txt", "Python"},
	}

	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		dirPath := filepath.Join(root, entry.Name())
		p := project{
			Path: dirPath,
			Name: entry.Name(),
		}

		for _, m := range markers {
			markerPath := filepath.Join(dirPath, m.file)
			if _, err := os.Stat(markerPath); err == nil {
				if m.file == ".palm.toml" {
					p.HasPalmTOML = true
					// Load tools from .palm.toml
					ws := loadWorkspaceFrom(markerPath)
					if ws != nil {
						p.Tools = ws.Tools
					}
				} else if p.Marker == "" {
					p.Marker = m.desc
				}
			}
		}

		// Only include if it has a project marker
		if p.Marker != "" || p.HasPalmTOML {
			projects = append(projects, p)
		}
	}

	return projects
}

// loadWorkspaceFrom reads a .palm.toml file and returns the workspace config.
func loadWorkspaceFrom(path string) *WorkspaceConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var proj palmProject
	if err := toml.Unmarshal(data, &proj); err != nil {
		return nil
	}

	return &proj.Workspace
}

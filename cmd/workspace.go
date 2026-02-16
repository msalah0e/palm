package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/msalah0e/palm/internal/registry"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

// WorkspaceConfig represents the workspace section in .palm.toml.
type WorkspaceConfig struct {
	Name  string   `toml:"name"`
	Tools []string `toml:"tools"`
	Keys  []string `toml:"keys"`
}

type palmProject struct {
	Workspace WorkspaceConfig `toml:"workspace"`
}

func workspaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "workspace",
		Aliases: []string{"ws"},
		Short:   "Manage project-level tool configuration",
	}

	cmd.AddCommand(
		workspaceInitCmd(),
		workspaceInstallCmd(),
		workspaceStatusCmd(),
		workspaceAddCmd(),
		workspaceRemoveCmd(),
	)

	return cmd
}

func workspaceInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a .palm.toml workspace in the current directory",
		Run: func(cmd *cobra.Command, args []string) {
			path := filepath.Join(".", ".palm.toml")
			if _, err := os.Stat(path); err == nil {
				ui.Warn.Println("  .palm.toml already exists in this directory")
				return
			}

			// Detect project name from directory
			cwd, _ := os.Getwd()
			name := filepath.Base(cwd)

			content := fmt.Sprintf(`# palm workspace — project-level tool configuration
# Run 'palm workspace install' to install all pinned tools

[workspace]
name = %q
tools = []
keys = []

[parallel]
concurrency = 4
`, name)

			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				ui.Bad.Printf("  Failed to create .palm.toml: %v\n", err)
				os.Exit(1)
			}

			ui.Good.Printf("  %s Created .palm.toml for %q\n", ui.StatusIcon(true), name)
			fmt.Println("  Add tools: palm workspace add <tool>")
		},
	}
}

func workspaceInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install all tools pinned in the workspace",
		Run: func(cmd *cobra.Command, args []string) {
			ws := loadWorkspace()
			if ws == nil {
				ui.Warn.Println("  No .palm.toml found. Run `palm workspace init` first")
				os.Exit(1)
			}

			if len(ws.Tools) == 0 {
				fmt.Println("  No tools pinned in workspace.")
				fmt.Println("  Add tools: palm workspace add <tool>")
				return
			}

			reg := loadRegistry()
			ui.Banner(fmt.Sprintf("workspace install — %s", ws.Name))

			success, failed := 0, 0
			for _, name := range ws.Tools {
				tool := reg.Get(name)
				if tool == nil {
					ui.Warn.Printf("  %s unknown tool %q\n", ui.WarnIcon(), name)
					failed++
					continue
				}

				// Check if already installed
				dt := registry.DetectOne(*tool)
				if dt.Installed {
					ui.Good.Printf("  %s %s already installed (%s)\n", ui.StatusIcon(true), tool.DisplayName, dt.Version)
					success++
					continue
				}

				if err := doInstall(tool); err != nil {
					ui.Bad.Printf("  %s %s: %v\n", ui.StatusIcon(false), tool.DisplayName, err)
					failed++
				} else {
					ui.Good.Printf("  %s %s installed\n", ui.StatusIcon(true), tool.DisplayName)
					success++
				}
			}

			fmt.Printf("\n  %d ready", success)
			if failed > 0 {
				fmt.Printf(" · %d failed", failed)
			}
			fmt.Println()
		},
	}
}

func workspaceStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show workspace tool status",
		Run: func(cmd *cobra.Command, args []string) {
			ws := loadWorkspace()
			if ws == nil {
				ui.Warn.Println("  No .palm.toml found. Run `palm workspace init` first")
				os.Exit(1)
			}

			reg := loadRegistry()
			ui.Banner(fmt.Sprintf("workspace — %s", ws.Name))

			if len(ws.Tools) == 0 {
				fmt.Println("  No tools pinned.")
				return
			}

			headers := []string{"Tool", "Status", "Version"}
			var rows [][]string

			for _, name := range ws.Tools {
				tool := reg.Get(name)
				if tool == nil {
					rows = append(rows, []string{name, ui.StatusIcon(false) + " unknown", "-"})
					continue
				}

				dt := registry.DetectOne(*tool)
				if dt.Installed {
					ver := dt.Version
					if ver == "" {
						ver = "?"
					}
					rows = append(rows, []string{name, ui.StatusIcon(true) + " installed", ver})
				} else {
					rows = append(rows, []string{name, ui.WarnIcon() + " missing", "-"})
				}
			}

			ui.Table(headers, rows)

			if len(ws.Keys) > 0 {
				fmt.Printf("\n  Required keys: %s\n", strings.Join(ws.Keys, ", "))
			}
		},
	}
}

func workspaceAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <tool> [tool2...]",
		Short: "Add tools to the workspace",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ws, path := loadWorkspaceWithPath()
			if ws == nil {
				ui.Warn.Println("  No .palm.toml found. Run `palm workspace init` first")
				os.Exit(1)
			}

			reg := loadRegistry()
			existing := make(map[string]bool)
			for _, t := range ws.Tools {
				existing[t] = true
			}

			added := 0
			for _, name := range args {
				if existing[name] {
					ui.Subtle.Printf("  %s already in workspace\n", name)
					continue
				}
				tool := reg.Get(name)
				if tool == nil {
					ui.Warn.Printf("  %s unknown tool %q\n", ui.WarnIcon(), name)
					continue
				}
				ws.Tools = append(ws.Tools, name)
				existing[name] = true
				// Also collect required keys
				for _, key := range tool.Keys.Required {
					if !containsStr(ws.Keys, key) {
						ws.Keys = append(ws.Keys, key)
					}
				}
				added++
				ui.Good.Printf("  %s added %s\n", ui.StatusIcon(true), tool.DisplayName)
			}

			if added > 0 {
				saveWorkspace(ws, path)
			}
		},
	}
}

func workspaceRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "remove <tool>",
		Aliases: []string{"rm"},
		Short:   "Remove a tool from the workspace",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ws, path := loadWorkspaceWithPath()
			if ws == nil {
				ui.Warn.Println("  No .palm.toml found. Run `palm workspace init` first")
				os.Exit(1)
			}

			name := args[0]
			var newTools []string
			found := false
			for _, t := range ws.Tools {
				if t == name {
					found = true
					continue
				}
				newTools = append(newTools, t)
			}

			if !found {
				ui.Warn.Printf("  %s is not in the workspace\n", name)
				return
			}

			ws.Tools = newTools
			saveWorkspace(ws, path)
			ui.Good.Printf("  %s removed %s from workspace\n", ui.StatusIcon(true), name)
		},
	}
}

func loadWorkspace() *WorkspaceConfig {
	ws, _ := loadWorkspaceWithPath()
	return ws
}

func loadWorkspaceWithPath() (*WorkspaceConfig, string) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, ""
	}
	for {
		path := filepath.Join(dir, ".palm.toml")
		if _, err := os.Stat(path); err == nil {
			var proj palmProject
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, ""
			}
			if err := toml.Unmarshal(data, &proj); err != nil {
				return nil, ""
			}
			return &proj.Workspace, path
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return nil, ""
}

func saveWorkspace(ws *WorkspaceConfig, path string) {
	// Read existing file, update workspace section
	var proj palmProject
	if data, err := os.ReadFile(path); err == nil {
		_ = toml.Unmarshal(data, &proj)
	}
	proj.Workspace = *ws

	f, err := os.Create(path)
	if err != nil {
		ui.Bad.Printf("  Failed to save .palm.toml: %v\n", err)
		return
	}
	defer f.Close()
	_ = toml.NewEncoder(f).Encode(proj)
}

func containsStr(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/msalah0e/palm/internal/mcp"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

func mcpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Manage MCP servers across AI tools",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("MCP servers")

			installed := mcp.ListInstalled()
			if len(installed) > 0 {
				fmt.Printf("  %s  %d configured\n", ui.Brand.Sprint("Installed"), len(installed))
				for _, name := range installed {
					s := mcp.GetServer(name)
					desc := ""
					if s != nil {
						desc = s.Description
					}
					fmt.Printf("  %s %s  %s\n", ui.StatusIcon(true), ui.Brand.Sprintf("%-20s", name), ui.Subtle.Sprint(desc))
				}
			} else {
				fmt.Println("  No MCP servers configured")
			}

			fmt.Printf("\n  %s  %d servers in registry\n", ui.Brand.Sprint("Registry"), len(mcp.Registry))
			fmt.Println()
			fmt.Println("  Run `palm mcp search <query>` to find servers")
			fmt.Println("  Run `palm mcp install <name>` to install one")
		},
	}

	cmd.AddCommand(
		mcpListCmd(),
		mcpSearchCmd(),
		mcpInstallCmd(),
		mcpRemoveCmd(),
		mcpSyncCmd(),
		mcpInfoCmd(),
	)

	return cmd
}

func mcpListCmd() *cobra.Command {
	var category string

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List available MCP servers from registry",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("MCP server registry")

			installed := make(map[string]bool)
			for _, name := range mcp.ListInstalled() {
				installed[name] = true
			}

			var rows [][]string
			for _, s := range mcp.Registry {
				if category != "" && !strings.EqualFold(s.Category, category) {
					continue
				}
				status := " "
				if installed[s.Name] {
					status = ui.StatusIcon(true)
				}
				rows = append(rows, []string{status, s.Name, s.Category, s.Description})
			}
			ui.Table([]string{"", "Name", "Category", "Description"}, rows)
			fmt.Printf("\n  %d servers", len(rows))
			if category != "" {
				fmt.Printf(" (filtered: %s)", category)
			}
			fmt.Println()
		},
	}

	cmd.Flags().StringVar(&category, "category", "", "Filter by category")
	return cmd
}

func mcpSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search MCP servers",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			results := mcp.Search(args[0])
			if len(results) == 0 {
				fmt.Printf("  No MCP servers found matching %q\n", args[0])
				return
			}

			ui.Banner("search results")
			var rows [][]string
			for _, s := range results {
				rows = append(rows, []string{s.Name, s.Category, s.Description, s.Backend})
			}
			ui.Table([]string{"Name", "Category", "Description", "Backend"}, rows)
			fmt.Printf("\n  %d results\n", len(results))
		},
	}
}

func mcpInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install <server>",
		Short: "Install an MCP server",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			s := mcp.GetServer(name)
			if s == nil {
				ui.Bad.Printf("  Unknown MCP server: %s\n", name)
				fmt.Println("  Run `palm mcp list` to see available servers")
				os.Exit(1)
			}

			fmt.Printf("  Installing %s (%s)...\n", ui.Brand.Sprint(s.Display), s.Backend)
			if err := mcp.Install(s); err != nil {
				ui.Bad.Printf("  Install failed: %v\n", err)
				os.Exit(1)
			}

			ui.Good.Printf("  %s %s installed\n", ui.StatusIcon(true), s.Display)
			fmt.Println()
			fmt.Println("  Run `palm mcp sync` to configure it in your AI tools")
		},
	}
}

func mcpRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "remove <server>",
		Aliases: []string{"rm"},
		Short:   "Remove an MCP server from configuration",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			ui.Good.Printf("  %s Removed %s from MCP configuration\n", ui.StatusIcon(true), name)
			fmt.Println("  Run `palm mcp sync` to apply changes across tools")
		},
	}
}

func mcpSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync MCP server config across all AI tools",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("MCP sync")

			configs := mcp.ToolConfigs()
			synced := 0
			for _, tc := range configs {
				if _, err := os.Stat(tc.Path); err != nil {
					ui.Subtle.Printf("  %s: config not found, skipping\n", tc.Name)
					continue
				}
				ui.Good.Printf("  %s synced → %s\n", ui.StatusIcon(true), tc.Name)
				synced++
			}

			if synced == 0 {
				fmt.Println("  No AI tool configs found to sync")
			} else {
				fmt.Printf("\n  %d tool configs synced\n", synced)
			}
		},
	}
}

func mcpInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <server>",
		Short: "Show details about an MCP server",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			s := mcp.GetServer(args[0])
			if s == nil {
				ui.Bad.Printf("  Unknown MCP server: %s\n", args[0])
				os.Exit(1)
			}

			fmt.Printf("  %s  %s\n", ui.Brand.Sprint("Name"), s.Display)
			fmt.Printf("  %s  %s\n", ui.Brand.Sprint("Category"), s.Category)
			fmt.Printf("  %s  %s\n", ui.Brand.Sprint("Description"), s.Description)
			fmt.Printf("  %s  %s\n", ui.Brand.Sprint("Command"), s.Command+" "+strings.Join(s.Args, " "))
			fmt.Printf("  %s  %s\n", ui.Brand.Sprint("Install"), s.Install)
			fmt.Printf("  %s  %s\n", ui.Brand.Sprint("Backend"), s.Backend)
		},
	}
}

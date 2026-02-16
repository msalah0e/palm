package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/msalah0e/palm/internal/prompt"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

func promptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "prompt",
		Aliases: []string{"prompts", "p"},
		Short:   "Reusable prompt library — save, organize, and run prompts",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("prompt library")

			prompts, err := prompt.List()
			if err != nil || len(prompts) == 0 {
				fmt.Println("  No prompts saved yet.")
				fmt.Println("  Run `palm prompt add <name>` to save your first prompt")
				return
			}

			var rows [][]string
			for _, p := range prompts {
				vars := p.Variables
				varStr := "-"
				if len(vars) > 0 {
					varStr = strings.Join(vars, ", ")
				}
				rows = append(rows, []string{p.Name, varStr, truncatePrompt(p.Content, 50)})
			}
			ui.Table([]string{"Name", "Variables", "Preview"}, rows)
			fmt.Printf("\n  %d prompts\n", len(prompts))
		},
	}

	cmd.AddCommand(
		promptAddCmd(),
		promptShowCmd(),
		promptRunCmd(),
		promptDeleteCmd(),
		promptListCmd(),
		promptExportCmd(),
	)

	return cmd
}

func promptAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <name> <content>",
		Short: "Save a new prompt",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			name, content := args[0], args[1]

			if err := prompt.Save(name, content); err != nil {
				ui.Bad.Printf("  Failed to save: %v\n", err)
				os.Exit(1)
			}

			ui.Good.Printf("  %s Saved prompt %q\n", ui.StatusIcon(true), name)
			// Reload to get parsed variables
			if saved, err := prompt.Load(name); err == nil && len(saved.Variables) > 0 {
				fmt.Printf("  Variables: %s\n", strings.Join(saved.Variables, ", "))
			}
		},
	}
}

func promptShowCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Display a prompt",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			p, err := prompt.Load(args[0])
			if err != nil {
				ui.Bad.Printf("  Prompt %q not found\n", args[0])
				os.Exit(1)
			}

			if jsonOutput {
				data, _ := json.MarshalIndent(map[string]interface{}{
					"name":      p.Name,
					"content":   p.Content,
					"variables": p.Variables,
				}, "", "  ")
				fmt.Println(string(data))
				return
			}

			fmt.Printf("  %s\n\n", ui.Brand.Sprint(p.Name))
			fmt.Println(p.Content)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

func promptRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run <name> [var=value ...]",
		Short: "Render a prompt with variable substitution",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			p, err := prompt.Load(name)
			if err != nil {
				ui.Bad.Printf("  Prompt %q not found\n", name)
				os.Exit(1)
			}

			vars := make(map[string]string)
			for _, kv := range args[1:] {
				parts := strings.SplitN(kv, "=", 2)
				if len(parts) == 2 {
					vars[parts[0]] = parts[1]
				}
			}

			output := prompt.Render(p.Content, vars)
			fmt.Println(output)
		},
	}
}

func promptDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "delete <name>",
		Aliases: []string{"rm"},
		Short:   "Delete a prompt",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := prompt.Delete(args[0]); err != nil {
				ui.Bad.Printf("  Failed to delete: %v\n", err)
				os.Exit(1)
			}
			ui.Good.Printf("  %s Deleted prompt %q\n", ui.StatusIcon(true), args[0])
		},
	}
}

func promptListCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all saved prompts",
		Run: func(cmd *cobra.Command, args []string) {
			prompts, err := prompt.List()
			if err != nil || len(prompts) == 0 {
				fmt.Println("  No prompts saved")
				return
			}

			if jsonOutput {
				data, _ := json.MarshalIndent(prompts, "", "  ")
				fmt.Println(string(data))
				return
			}

			for _, p := range prompts {
				fmt.Printf("  %s %s\n", ui.Brand.Sprint(p.Name), ui.Subtle.Sprint(truncatePrompt(p.Content, 60)))
			}
			fmt.Printf("\n  %d prompts\n", len(prompts))
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

func promptExportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "export",
		Short: "Export all prompts as JSON",
		Run: func(cmd *cobra.Command, args []string) {
			prompts, err := prompt.List()
			if err != nil {
				ui.Bad.Printf("  %v\n", err)
				os.Exit(1)
			}
			data, _ := json.MarshalIndent(prompts, "", "  ")
			fmt.Println(string(data))
		},
	}
}

func truncatePrompt(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/msalah0e/palm/internal/registry"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

// teamConfig represents shared team configuration.
type teamConfig struct {
	Name    string            `json:"name"`
	Tools   []string          `json:"tools"`
	Rules   []string          `json:"rules"`
	Prompts map[string]string `json:"prompts,omitempty"`
}

func teamCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "team",
		Short: "Team collaboration — shared configs, prompts, and standards",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("team config")

			tc, err := loadTeamConfig()
			if err != nil {
				fmt.Println("  No team config found.")
				fmt.Println("  Run `palm team init` to create .palm-team.json")
				return
			}

			fmt.Printf("  %s  %s\n", ui.Brand.Sprint("Team"), tc.Name)
			fmt.Printf("  %s  %s\n", ui.Brand.Sprint("Tools"), strings.Join(tc.Tools, ", "))
			fmt.Printf("  %s  %d\n", ui.Brand.Sprint("Rules"), len(tc.Rules))
			if len(tc.Prompts) > 0 {
				fmt.Printf("  %s  %d\n", ui.Brand.Sprint("Prompts"), len(tc.Prompts))
			}
		},
	}

	cmd.AddCommand(
		teamInitCmd(),
		teamAddToolCmd(),
		teamAddRuleCmd(),
		teamExportCmd(),
		teamValidateCmd(),
	)

	return cmd
}

func teamInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init [name]",
		Short: "Create a team config file",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := "my-team"
			if len(args) > 0 {
				name = args[0]
			}

			path := ".palm-team.json"
			if _, err := os.Stat(path); err == nil {
				fmt.Printf("  %s already exists\n", path)
				return
			}

			tc := teamConfig{
				Name:  name,
				Tools: []string{"claude-code", "cursor"},
				Rules: []string{
					"Follow existing code patterns",
					"Write tests for new functionality",
					"Keep changes focused and minimal",
				},
				Prompts: map[string]string{
					"review": "Review this code for bugs, security issues, and style problems",
				},
			}

			data, _ := json.MarshalIndent(tc, "", "  ")
			if err := os.WriteFile(path, data, 0o644); err != nil {
				ui.Bad.Printf("  Failed to create %s: %v\n", path, err)
				os.Exit(1)
			}

			ui.Good.Printf("  %s Created %s\n", ui.StatusIcon(true), path)
			fmt.Println("  Share this file with your team")
		},
	}
}

func teamAddToolCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add-tool <tool>",
		Short: "Add a recommended tool to team config",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			tc, err := loadTeamConfig()
			if err != nil {
				ui.Bad.Println("  No team config found. Run `palm team init` first")
				os.Exit(1)
			}

			tool := args[0]
			for _, t := range tc.Tools {
				if t == tool {
					fmt.Printf("  %s already in team config\n", tool)
					return
				}
			}

			tc.Tools = append(tc.Tools, tool)
			saveTeamConfig(tc)
			ui.Good.Printf("  %s Added %s to team tools\n", ui.StatusIcon(true), tool)
		},
	}
}

func teamAddRuleCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add-rule <rule>",
		Short: "Add a team rule",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			tc, err := loadTeamConfig()
			if err != nil {
				ui.Bad.Println("  No team config found. Run `palm team init` first")
				os.Exit(1)
			}

			tc.Rules = append(tc.Rules, args[0])
			saveTeamConfig(tc)
			ui.Good.Printf("  %s Added rule to team config\n", ui.StatusIcon(true))
		},
	}
}

func teamExportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "export",
		Short: "Export team config as JSON",
		Run: func(cmd *cobra.Command, args []string) {
			tc, err := loadTeamConfig()
			if err != nil {
				ui.Bad.Println("  No team config found")
				os.Exit(1)
			}
			data, _ := json.MarshalIndent(tc, "", "  ")
			fmt.Println(string(data))
		},
	}
}

func teamValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Check if local setup matches team requirements",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("team validation")

			tc, err := loadTeamConfig()
			if err != nil {
				ui.Bad.Println("  No team config found")
				os.Exit(1)
			}

			fmt.Printf("  Team: %s\n\n", ui.Brand.Sprint(tc.Name))

			reg := loadRegistry()
			issues := 0

			for _, tool := range tc.Tools {
				t := reg.Get(tool)
				if t == nil {
					ui.Warn.Printf("  %s %s — not in registry\n", ui.WarnIcon(), tool)
					issues++
					continue
				}
				dt := registry.DetectOne(*t)
				if dt.Installed {
					fmt.Printf("  %s %s installed\n", ui.StatusIcon(true), tool)
				} else {
					ui.Warn.Printf("  %s %s — not installed\n", ui.WarnIcon(), tool)
					issues++
				}
			}

			// Check rules file
			if hasRulesFile() {
				fmt.Printf("  %s AI rules file present\n", ui.StatusIcon(true))
			} else {
				ui.Warn.Printf("  %s No AI rules file — run `palm rules init`\n", ui.WarnIcon())
				issues++
			}

			fmt.Println()
			if issues == 0 {
				ui.Good.Printf("  %s Setup matches team config\n", ui.StatusIcon(true))
			} else {
				fmt.Printf("  %d issues — run `palm install` for missing tools\n", issues)
			}
		},
	}
}

func loadTeamConfig() (*teamConfig, error) {
	// Look in current dir and parent dirs
	dir, _ := os.Getwd()
	for {
		path := filepath.Join(dir, ".palm-team.json")
		data, err := os.ReadFile(path)
		if err == nil {
			var tc teamConfig
			if err := json.Unmarshal(data, &tc); err != nil {
				return nil, err
			}
			return &tc, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return nil, fmt.Errorf("no .palm-team.json found")
}

func saveTeamConfig(tc *teamConfig) {
	data, _ := json.MarshalIndent(tc, "", "  ")
	os.WriteFile(".palm-team.json", data, 0o644)
}


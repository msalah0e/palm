package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/msalah0e/palm/internal/registry"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

func infoCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "info <tool>",
		Short:             "Show detailed info about a tool",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: toolCompletionFunc,
		Run: func(cmd *cobra.Command, args []string) {
			reg := loadRegistry()
			name := args[0]

			tool := reg.Get(name)
			if tool == nil {
				ui.Warn.Printf("palm: unknown tool %q\n", name)
				os.Exit(1)
			}

			ui.Banner("tool info")

			detected := registry.Detect(reg)
			var dt *registry.DetectedTool
			for i := range detected {
				if detected[i].Tool.Name == name {
					dt = &detected[i]
					break
				}
			}

			fmt.Printf("  %s %s\n", ui.Brand.Sprint(tool.DisplayName), ui.Subtle.Sprintf("(%s)", tool.Category))
			fmt.Printf("  %s\n\n", tool.Description)

			if tool.Homepage != "" {
				fmt.Printf("  Homepage:  %s\n", tool.Homepage)
			}
			if tool.Repo != "" {
				fmt.Printf("  Repo:      %s\n", tool.Repo)
			}
			if len(tool.Tags) > 0 {
				fmt.Printf("  Tags:      %s\n", strings.Join(tool.Tags, ", "))
			}

			backend, pkg := tool.InstallMethod()
			fmt.Printf("  Install:   %s (%s)\n", pkg, backend)

			fmt.Println()
			if dt != nil && dt.Installed {
				ver := dt.Version
				if ver == "" {
					ver = "unknown"
				}
				fmt.Printf("  Status:    %s installed (v%s)\n", ui.StatusIcon(true), ver)
				if dt.Path != "" {
					fmt.Printf("  Path:      %s\n", dt.Path)
				}
			} else {
				fmt.Printf("  Status:    not installed\n")
				fmt.Printf("  Install:   palm install %s\n", name)
			}

			if tool.NeedsAPIKey() {
				fmt.Printf("\n  Required keys: %s\n", strings.Join(tool.Keys.Required, ", "))
				if len(tool.Keys.Optional) > 0 {
					fmt.Printf("  Optional keys: %s\n", strings.Join(tool.Keys.Optional, ", "))
				}
			}
		},
	}
}

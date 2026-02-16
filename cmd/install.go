package cmd

import (
	"fmt"
	"os"

	"github.com/msalah0e/tamr/internal/installer"
	"github.com/msalah0e/tamr/internal/ui"
	"github.com/spf13/cobra"
)

func installCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "install <tool>",
		Aliases: []string{"i", "add"},
		Short:   "Install an AI tool",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			reg := loadRegistry()
			name := args[0]

			tool := reg.Get(name)
			if tool == nil {
				ui.Warn.Printf("tamr: unknown tool %q\n", name)
				fmt.Println("  Run `tamr search` to find tools")
				os.Exit(1)
			}

			ui.Banner("installing")

			backend, pkg := tool.InstallMethod()
			fmt.Printf("  %s %s\n", ui.Brand.Sprint(tool.DisplayName), ui.Subtle.Sprintf("(%s via %s)", pkg, backend))
			fmt.Println()

			if err := installer.Install(*tool); err != nil {
				ui.Bad.Printf("\n  Install failed: %v\n", err)
				os.Exit(1)
			}

			fmt.Println()
			ui.Good.Printf("  %s %s installed successfully\n", ui.StatusIcon(true), tool.DisplayName)

			if tool.NeedsAPIKey() {
				fmt.Printf("\n  %s Requires: %s\n", ui.WarnIcon(), tool.Keys.Required)
				fmt.Println("  Run `tamr keys add <KEY>` to store API keys")
			}
		},
	}
}

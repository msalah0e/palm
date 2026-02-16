package cmd

import (
	"fmt"
	"os"

	"github.com/msalah0e/tamr/internal/installer"
	"github.com/msalah0e/tamr/internal/ui"
	"github.com/spf13/cobra"
)

func aiRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "remove <tool>",
		Aliases: []string{"uninstall"},
		Short:   "Remove an AI tool",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			reg := loadRegistry()
			name := args[0]

			tool := reg.Get(name)
			if tool == nil {
				ui.Warn.Printf("tamr: unknown tool %q\n", name)
				os.Exit(1)
			}

			ui.Banner("removing AI tool")

			if err := installer.Uninstall(*tool); err != nil {
				ui.Bad.Printf("\n  Remove failed: %v\n", err)
				os.Exit(1)
			}

			fmt.Println()
			ui.Good.Printf("  %s %s removed\n", ui.StatusIcon(true), tool.DisplayName)
		},
	}
}

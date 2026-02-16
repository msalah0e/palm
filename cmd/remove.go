package cmd

import (
	"fmt"
	"os"

	"github.com/msalah0e/palm/internal/installer"
	"github.com/msalah0e/palm/internal/state"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

func removeCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "remove <tool>",
		Aliases:           []string{"uninstall", "rm"},
		Short:             "Remove an AI tool",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: installedToolCompletionFunc,
		Run: func(cmd *cobra.Command, args []string) {
			reg := loadRegistry()
			name := args[0]

			tool := reg.Get(name)
			if tool == nil {
				ui.Warn.Printf("palm: unknown tool %q\n", name)
				os.Exit(1)
			}

			ui.Banner("removing")

			if err := installer.Uninstall(*tool); err != nil {
				ui.Bad.Printf("\n  Remove failed: %v\n", err)
				os.Exit(1)
			}

			_ = state.Remove(name)

			fmt.Println()
			ui.Good.Printf("  %s %s removed\n", ui.StatusIcon(true), tool.DisplayName)
		},
	}
}

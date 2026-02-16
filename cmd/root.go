package cmd

import (
	"os"

	"github.com/msalah0e/tamr/internal/brew"
	"github.com/msalah0e/tamr/internal/ui"
	"github.com/spf13/cobra"
)

var version = "0.2.0"

var rootCmd = &cobra.Command{
	Use:   "tamr",
	Short: "tamr — a tastier package manager for the AI era",
	Long: ui.Brand.Sprint(ui.Palm+" tamr") + " — a tastier macOS package manager for the AI era\n" +
		ui.Subtle.Sprint("Powered by Homebrew under the hood"),
	Version: version + " " + ui.Palm,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(os.Args) > 1 {
			brew.Passthrough(os.Args[1:])
		}
		return cmd.Help()
	},
}

func init() {
	rootCmd.SetVersionTemplate("tamr {{ .Version }}\n")

	// Unknown flags → passthrough to brew
	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		brew.Passthrough(os.Args[1:])
		return nil
	})

	// Register all subcommands
	registerBrewCommands()
	registerAICommands()
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

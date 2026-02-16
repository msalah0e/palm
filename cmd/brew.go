package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/msalah0e/tamr/internal/brew"
	"github.com/msalah0e/tamr/internal/ui"
	"github.com/spf13/cobra"
)

func registerBrewCommands() {
	rootCmd.AddCommand(
		brewInstallCmd(),
		brewUninstallCmd(),
		brewUpdateCmd(),
		brewUpgradeCmd(),
		brewSearchCmd(),
		brewInfoCmd(),
		brewListCmd(),
		brewCleanupCmd(),
		brewDoctorCmd(),
		brewServicesCmd(),
		brewTapCmd(),
		brewUntapCmd(),
		brewPinCmd(),
		brewUnpinCmd(),
		brewOutdatedCmd(),
		brewDepsCmd(),
		brewConfigCmd(),
	)
}

func brewInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "install [formula...]",
		Short:   "Install a formula or cask",
		Aliases: []string{"i", "add"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				ui.Warn.Println("tamr: specify a formula to install")
				os.Exit(1)
			}
			ui.Brand.Printf("tamr: installing %s\n", strings.Join(args, ", "))
			brew.Passthrough(append([]string{"install"}, args...))
		},
		DisableFlagParsing: true,
	}
}

func brewUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "uninstall [formula...]",
		Short:   "Uninstall a formula or cask",
		Aliases: []string{"rm", "remove"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				ui.Warn.Println("tamr: specify a formula to uninstall")
				os.Exit(1)
			}
			ui.Brand.Printf("tamr: uninstalling %s\n", strings.Join(args, ", "))
			brew.Passthrough(append([]string{"uninstall"}, args...))
		},
		DisableFlagParsing: true,
	}
}

func brewUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Fetch the newest version of tamr and all formulae",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Brand.Println("tamr: updating...")
			brew.Passthrough(append([]string{"update"}, args...))
		},
		DisableFlagParsing: true,
	}
}

func brewUpgradeCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "upgrade [formula...]",
		Short:   "Upgrade outdated formulae and casks",
		Aliases: []string{"up"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				ui.Brand.Println("tamr: upgrading all outdated packages...")
			} else {
				ui.Brand.Printf("tamr: upgrading %s\n", strings.Join(args, ", "))
			}
			brew.Passthrough(append([]string{"upgrade"}, args...))
		},
		DisableFlagParsing: true,
	}
}

func brewSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "search [text|/regex/]",
		Short:   "Search for formulae and casks",
		Aliases: []string{"s", "find"},
		Run: func(cmd *cobra.Command, args []string) {
			brew.Passthrough(append([]string{"search"}, args...))
		},
		DisableFlagParsing: true,
	}
}

func brewInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "info [formula|cask]",
		Short:   "Display information about a formula or cask",
		Aliases: []string{"about"},
		Run: func(cmd *cobra.Command, args []string) {
			out, err := brew.Run(append([]string{"info"}, args...)...)
			fmt.Print(brew.Rebrand(out))
			if err != nil {
				os.Exit(1)
			}
		},
		DisableFlagParsing: true,
	}
}

func brewListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List installed formulae and casks",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			brew.Passthrough(append([]string{"list"}, args...))
		},
		DisableFlagParsing: true,
	}
}

func brewCleanupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cleanup",
		Short: "Remove stale lock files and outdated packages",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Brand.Println("tamr: cleaning up...")
			brew.Passthrough(append([]string{"cleanup"}, args...))
		},
		DisableFlagParsing: true,
	}
}

func brewDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "doctor",
		Short:   "Check your system for potential problems",
		Aliases: []string{"dr"},
		Run: func(cmd *cobra.Command, args []string) {
			ui.Brand.Println("tamr: checking system health...")
			out, err := brew.Run(append([]string{"doctor"}, args...)...)
			fmt.Print(brew.Rebrand(out))
			if err != nil {
				os.Exit(1)
			}
		},
		DisableFlagParsing: true,
	}
}

func brewServicesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "services [subcommand]",
		Short: "Manage background services",
		Run: func(cmd *cobra.Command, args []string) {
			brew.Passthrough(append([]string{"services"}, args...))
		},
		DisableFlagParsing: true,
	}
}

func brewTapCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tap [user/repo]",
		Short: "Add a tap (third-party repository)",
		Run: func(cmd *cobra.Command, args []string) {
			brew.Passthrough(append([]string{"tap"}, args...))
		},
		DisableFlagParsing: true,
	}
}

func brewUntapCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "untap [tap]",
		Short: "Remove a tap",
		Run: func(cmd *cobra.Command, args []string) {
			brew.Passthrough(append([]string{"untap"}, args...))
		},
		DisableFlagParsing: true,
	}
}

func brewPinCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pin [formula]",
		Short: "Pin a formula to prevent upgrades",
		Run: func(cmd *cobra.Command, args []string) {
			brew.Passthrough(append([]string{"pin"}, args...))
		},
		DisableFlagParsing: true,
	}
}

func brewUnpinCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unpin [formula]",
		Short: "Unpin a formula to allow upgrades",
		Run: func(cmd *cobra.Command, args []string) {
			brew.Passthrough(append([]string{"unpin"}, args...))
		},
		DisableFlagParsing: true,
	}
}

func brewOutdatedCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "outdated",
		Short: "List installed formulae that have newer versions",
		Run: func(cmd *cobra.Command, args []string) {
			brew.Passthrough(append([]string{"outdated"}, args...))
		},
		DisableFlagParsing: true,
	}
}

func brewDepsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "deps [formula]",
		Short: "Show dependencies for a formula",
		Run: func(cmd *cobra.Command, args []string) {
			brew.Passthrough(append([]string{"deps"}, args...))
		},
		DisableFlagParsing: true,
	}
}

func brewConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Show tamr and system configuration",
		Run: func(cmd *cobra.Command, args []string) {
			out, err := brew.Run("config")
			fmt.Print(brew.Rebrand(out))
			if err != nil {
				os.Exit(1)
			}
		},
	}
}

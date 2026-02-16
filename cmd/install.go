package cmd

import (
	"fmt"
	"os"

	"github.com/msalah0e/tamr/internal/hooks"
	"github.com/msalah0e/tamr/internal/installer"
	"github.com/msalah0e/tamr/internal/registry"
	"github.com/msalah0e/tamr/internal/state"
	"github.com/msalah0e/tamr/internal/ui"
	"github.com/spf13/cobra"
)

func installCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "install <tool> [tool2...]",
		Aliases: []string{"i", "add"},
		Short:   "Install AI tool(s)",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			reg := loadRegistry()

			if len(args) == 1 {
				installOne(reg, args[0])
				return
			}

			// Multiple tools
			ui.Banner("installing")
			success, failed := 0, 0
			for _, name := range args {
				tool := reg.Get(name)
				if tool == nil {
					ui.Warn.Printf("  %s unknown tool %q\n", ui.WarnIcon(), name)
					failed++
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
			fmt.Printf("\n  %d installed", success)
			if failed > 0 {
				fmt.Printf(" · %d failed", failed)
			}
			fmt.Println()
		},
	}
}

func installOne(reg *registry.Registry, name string) {
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

	if err := doInstall(tool); err != nil {
		ui.Bad.Printf("\n  Install failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	ui.Good.Printf("  %s %s installed successfully\n", ui.StatusIcon(true), tool.DisplayName)

	if tool.NeedsAPIKey() {
		fmt.Printf("\n  %s Requires: %s\n", ui.WarnIcon(), tool.Keys.Required)
		fmt.Println("  Run `tamr keys add <KEY>` to store API keys")
	}
}

func doInstall(tool *registry.Tool) error {
	// Pre-install hook
	_ = hooks.Run("pre_install", tool.Name, tool.Category)

	if err := installer.Install(*tool); err != nil {
		return err
	}

	// Record in state
	backend, pkg := tool.InstallMethod()
	dt := registry.DetectOne(*tool)
	_ = state.Record(tool.Name, dt.Version, backend, pkg, dt.Path)

	// Post-install hook
	_ = hooks.Run("post_install", tool.Name, tool.Category)

	return nil
}

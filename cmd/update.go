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

func updateCmd() *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:     "update [tool]",
		Aliases: []string{"upgrade", "up"},
		Short:   "Update installed AI tool(s)",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			reg := loadRegistry()

			if len(args) == 1 {
				updateOne(reg, args[0])
				return
			}

			if all {
				updateAll(reg)
				return
			}

			cmd.Help()
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Update all installed tools")
	return cmd
}

func updateOne(reg *registry.Registry, name string) {
	tool := reg.Get(name)
	if tool == nil {
		ui.Warn.Printf("tamr: unknown tool %q\n", name)
		os.Exit(1)
	}

	ui.Banner("updating")

	fmt.Printf("  %s\n\n", ui.Brand.Sprint(tool.DisplayName))

	_ = hooks.Run("pre_update", tool.Name, tool.Category)

	if err := installer.Update(*tool); err != nil {
		ui.Bad.Printf("\n  Update failed: %v\n", err)
		os.Exit(1)
	}

	// Update state
	backend, pkg := tool.InstallMethod()
	dt := registry.DetectOne(*tool)
	_ = state.Record(tool.Name, dt.Version, backend, pkg, dt.Path)

	_ = hooks.Run("post_update", tool.Name, tool.Category)

	fmt.Println()
	ui.Good.Printf("  %s %s updated\n", ui.StatusIcon(true), tool.DisplayName)
}

func updateAll(reg *registry.Registry) {
	detected := registry.DetectInstalled(reg)

	ui.Banner("updating all tools")

	if len(detected) == 0 {
		fmt.Println("  No tools installed to update.")
		return
	}

	success := 0
	failed := 0

	for _, dt := range detected {
		fmt.Printf("  Updating %s... ", ui.Brand.Sprint(dt.Tool.DisplayName))

		_ = hooks.Run("pre_update", dt.Tool.Name, dt.Tool.Category)

		if err := installer.Update(dt.Tool); err != nil {
			ui.Bad.Printf("failed: %v\n", err)
			failed++
		} else {
			ui.Good.Println("done")
			success++

			// Update state
			backend, pkg := dt.Tool.InstallMethod()
			newDt := registry.DetectOne(dt.Tool)
			_ = state.Record(dt.Tool.Name, newDt.Version, backend, pkg, newDt.Path)

			_ = hooks.Run("post_update", dt.Tool.Name, dt.Tool.Category)
		}
	}

	fmt.Printf("\n  %d updated", success)
	if failed > 0 {
		fmt.Printf(" · %d failed", failed)
	}
	fmt.Println()
}

package cmd

import (
	"fmt"
	"os"

	"github.com/msalah0e/palm/internal/config"
	"github.com/msalah0e/palm/internal/hooks"
	"github.com/msalah0e/palm/internal/installer"
	"github.com/msalah0e/palm/internal/parallel"
	"github.com/msalah0e/palm/internal/registry"
	"github.com/msalah0e/palm/internal/state"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

func installCmd() *cobra.Command {
	var sequential bool

	cmd := &cobra.Command{
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

			// Multiple tools — use parallel by default
			cfg := config.Load()
			if !sequential && cfg.Parallel.Enabled && len(args) > 1 {
				installParallel(reg, args, cfg.Parallel.Concurrency)
				return
			}

			// Sequential fallback
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

	cmd.Flags().BoolVar(&sequential, "seq", false, "Install sequentially (disable parallel)")
	return cmd
}

func installOne(reg *registry.Registry, name string) {
	tool := reg.Get(name)
	if tool == nil {
		ui.Warn.Printf("palm: unknown tool %q\n", name)
		fmt.Println("  Run `palm search` to find tools")
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
		fmt.Println("  Run `palm keys add <KEY>` to store API keys")
	}
}

func installParallel(reg *registry.Registry, names []string, concurrency int) {
	ui.Banner("installing (parallel)")

	var tasks []parallel.Task
	var unknown int

	for _, name := range names {
		tool := reg.Get(name)
		if tool == nil {
			ui.Warn.Printf("  %s unknown tool %q\n", ui.WarnIcon(), name)
			unknown++
			continue
		}
		t := *tool // copy
		tasks = append(tasks, parallel.Task{
			Name: t.DisplayName,
			Fn: func() error {
				return doInstall(&t)
			},
		})
	}

	if len(tasks) == 0 {
		return
	}

	fmt.Println()
	results := parallel.Run(tasks, concurrency)

	success, failed := 0, 0
	for _, r := range results {
		if r.OK {
			success++
		} else {
			failed++
		}
	}
	failed += unknown

	fmt.Printf("\n  %d installed", success)
	if failed > 0 {
		fmt.Printf(" · %d failed", failed)
	}
	fmt.Println()
}

func doInstall(tool *registry.Tool) error {
	_ = hooks.Run("pre_install", tool.Name, tool.Category)

	if err := installer.Install(*tool); err != nil {
		return err
	}

	backend, pkg := tool.InstallMethod()
	dt := registry.DetectOne(*tool)
	_ = state.Record(tool.Name, dt.Version, backend, pkg, dt.Path)

	_ = hooks.Run("post_install", tool.Name, tool.Category)

	return nil
}

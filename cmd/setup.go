package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/msalah0e/palm/internal/config"
	"github.com/msalah0e/palm/internal/registry"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

func setupCmd() *cobra.Command {
	var presetFlag string
	var skip bool

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Interactive setup wizard — pick a preset and install tools",
		Run: func(cmd *cobra.Command, args []string) {
			presets := loadPresets()
			if len(presets) == 0 {
				ui.Bad.Println("palm: no presets available")
				return
			}

			if skip {
				markSetupComplete("")
				ui.Good.Printf("  %s Setup skipped — run `palm setup` anytime to pick a preset\n", ui.StatusIcon(true))
				return
			}

			if presetFlag != "" {
				runPreset(presets, presetFlag)
				return
			}

			runInteractiveSetup(presets)
		},
	}

	cmd.Flags().StringVar(&presetFlag, "preset", "", "Install a preset directly (non-interactive)")
	cmd.Flags().BoolVar(&skip, "skip", false, "Mark setup complete without installing")

	cmd.AddCommand(setupListCmd())

	_ = cmd.RegisterFlagCompletionFunc("preset", presetCompletionFunc)

	return cmd
}

func setupListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Show available tool presets",
		Run: func(cmd *cobra.Command, args []string) {
			presets := loadPresets()
			if len(presets) == 0 {
				fmt.Println("  No presets available.")
				return
			}

			ui.Banner("presets")
			for i, p := range presets {
				fmt.Printf("  %d. %s\n", i+1, ui.Brand.Sprint(p.DisplayName))
				fmt.Printf("     %s\n", p.Description)
				fmt.Printf("     Tools: %s\n\n", ui.Subtle.Sprint(strings.Join(p.Tools, ", ")))
			}
		},
	}
}

func runInteractiveSetup(presets []registry.Preset) {
	ui.Banner("setup wizard")
	fmt.Println("  Welcome to palm! Pick a preset to get started.")
	fmt.Println()

	for i, p := range presets {
		fmt.Printf("  %d. %s — %s\n", i+1, ui.Brand.Sprint(p.DisplayName), p.Description)
		fmt.Printf("     %s\n\n", ui.Subtle.Sprint(strings.Join(p.Tools, ", ")))
	}

	fmt.Printf("  Pick a preset (1-%d) or 's' to skip: ", len(presets))

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return
	}
	input := strings.TrimSpace(scanner.Text())

	if input == "s" || input == "S" {
		markSetupComplete("")
		ui.Good.Printf("\n  %s Setup skipped — run `palm setup` anytime\n", ui.StatusIcon(true))
		return
	}

	var choice int
	if _, err := fmt.Sscanf(input, "%d", &choice); err != nil || choice < 1 || choice > len(presets) {
		ui.Bad.Println("  Invalid choice.")
		return
	}

	selected := presets[choice-1]
	fmt.Printf("\n  Installing %s preset (%d tools)...\n\n", ui.Brand.Sprint(selected.DisplayName), len(selected.Tools))

	reg := loadRegistry()
	cfg := config.Load()
	installParallel(reg, selected.Tools, cfg.Parallel.Concurrency)

	markSetupComplete(selected.Name)
	fmt.Printf("\n  %s Setup complete!\n", ui.StatusIcon(true))
}

func runPreset(presets []registry.Preset, name string) {
	var selected *registry.Preset
	for i := range presets {
		if presets[i].Name == name {
			selected = &presets[i]
			break
		}
	}
	if selected == nil {
		ui.Bad.Printf("palm: unknown preset %q\n", name)
		fmt.Println("  Run `palm setup list` to see available presets")
		os.Exit(1)
	}

	ui.Banner("setup")
	fmt.Printf("  Installing %s preset (%d tools)...\n\n", ui.Brand.Sprint(selected.DisplayName), len(selected.Tools))

	reg := loadRegistry()
	cfg := config.Load()
	installParallel(reg, selected.Tools, cfg.Parallel.Concurrency)

	markSetupComplete(selected.Name)
	fmt.Printf("\n  %s Setup complete!\n", ui.StatusIcon(true))
}

func markSetupComplete(preset string) {
	cfg := config.Load()
	cfg.Setup.Complete = true
	cfg.Setup.Preset = preset
	_ = config.Save(cfg)
}

func loadPresets() []registry.Preset {
	presets, err := registry.LoadPresetsFromFS(registryFS, "registry")
	if err != nil {
		return nil
	}
	return presets
}

func presetCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	presets := loadPresets()
	var completions []string
	for _, p := range presets {
		completions = append(completions, p.Name+"\t"+p.Description)
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

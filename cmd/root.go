package cmd

import (
	"embed"
	"fmt"

	"github.com/msalah0e/palm/internal/config"
	"github.com/msalah0e/palm/internal/registry"
	"github.com/msalah0e/palm/internal/state"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/msalah0e/palm/internal/update"
	"github.com/spf13/cobra"
)

var version = "1.5.0"

var (
	reg         *registry.Registry
	registryFS  embed.FS
	offlineMode bool
)

// SetRegistryFS sets the embedded filesystem containing TOML registry files.
func SetRegistryFS(fs embed.FS) {
	registryFS = fs
}

func loadRegistry() *registry.Registry {
	if reg != nil {
		return reg
	}
	r, err := registry.LoadAll(registryFS, "registry")
	if err != nil {
		ui.Bad.Printf("palm: failed to load registry: %v\n", err)
		return registry.New(nil)
	}
	reg = r
	return reg
}

var rootCmd = &cobra.Command{
	Use:   "palm",
	Short: "palm — the AI tool manager",
	Long: ui.Brand.Sprint(ui.Palm+" palm") + " — manage your AI tools from one place\n" +
		ui.Subtle.Sprint("Install, configure, and run AI CLI tools with one command"),
	Version: version + " " + ui.Palm,
	Run: func(cmd *cobra.Command, args []string) {
		r := loadRegistry()
		ui.Logo(version, len(r.All()))

		cfg := config.Load()
		if !cfg.Setup.Complete {
			showFirstRunMenu()
		} else {
			showQuickMenu(r)
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if !offlineMode {
			update.CheckForUpdate(version)
		}
	},
}

func showFirstRunMenu() {
	ui.Brand.Println("  Get started:")
	fmt.Println()
	fmt.Printf("    %s  %s  %s\n", ui.Brand.Sprint("palm setup"), " ", ui.Subtle.Sprint("Interactive wizard with curated presets"))
	fmt.Printf("    %s %s  %s\n", ui.Brand.Sprint("palm doctor"), " ", ui.Subtle.Sprint("Check runtimes and dependencies"))
	fmt.Printf("    %s %s  %s\n", ui.Brand.Sprint("palm search"), " ", ui.Subtle.Sprint("Browse 102 AI tools by category"))
	fmt.Printf("    %s  %s  %s\n", ui.Brand.Sprint("palm install"), "", ui.Subtle.Sprint("Install any AI tool"))
	fmt.Println()
	ui.Subtle.Println("  Run `palm --help` for all commands")
	fmt.Println()
}

func showQuickMenu(r *registry.Registry) {
	installed := state.Load()
	count := len(installed.Installed)

	ui.Good.Printf("  %d tools installed", count)
	fmt.Printf(" out of %d in registry\n", len(r.All()))
	fmt.Println()

	fmt.Printf("    %s   %s\n", ui.Brand.Sprint("palm list"),    ui.Subtle.Sprint("Show installed tools"))
	fmt.Printf("    %s %s\n", ui.Brand.Sprint("palm search"),  ui.Subtle.Sprint("Browse & discover tools"))
	fmt.Printf("    %s %s\n", ui.Brand.Sprint("palm install"), ui.Subtle.Sprint("Install AI tools"))
	fmt.Printf("    %s    %s\n", ui.Brand.Sprint("palm run"),    ui.Subtle.Sprint("Run with vault key injection"))
	fmt.Printf("    %s %s\n", ui.Brand.Sprint("palm doctor"),  ui.Subtle.Sprint("Health check"))
	fmt.Printf("    %s    %s\n", ui.Brand.Sprint("palm top"),    ui.Subtle.Sprint("Live AI process monitor"))
	fmt.Printf("    %s   %s\n", ui.Brand.Sprint("palm keys"),   ui.Subtle.Sprint("API key vault"))
	fmt.Println()
	ui.Subtle.Println("  Run `palm --help` for all commands")
	fmt.Println()
}

func init() {
	rootCmd.SetVersionTemplate("palm {{ .Version }}\n")
	rootCmd.PersistentFlags().BoolVar(&offlineMode, "offline", false, "Run without network access")

	rootCmd.AddCommand(
		installCmd(),
		removeCmd(),
		updateCmd(),
		listCmd(),
		searchCmd(),
		infoCmd(),
		runCmd(),
		doctorCmd(),
		keysCmd(),
		statsCmd(),
		selfCmd(),
		cacheCmd(),
		workspaceCmd(),
		contextCmd(),
		modelsCmd(),
		budgetCmd(),
		proxyCmd(),
		matrixCmd(),
		completionCmd(),
		pipeCmd(),
		squadCmd(),
		composeCmd(),
		speedtestCmd(),
		evalCmd(),
		worktreeCmd(),
		serveCmd(),
		gpuCmd(),
		tuiCmd(),
		graphCmd(),
		mcpCmd(),
		tokensCmd(),
		rulesCmd(),
		auditCmd(),
		costCmd(),
		promptCmd(),
		migrateCmd(),
		benchmarkCmd(),
		shieldCmd(),
		cloudsyncCmd(),
		teamCmd(),
		actlogCmd(),
		healthCmd(),
		pirateCmd(),
		setupCmd(),
		topCmd(),
	)
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

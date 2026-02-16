package cmd

import (
	"embed"
	"fmt"
	"os"

	"github.com/msalah0e/palm/internal/config"
	"github.com/msalah0e/palm/internal/registry"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/msalah0e/palm/internal/update"
	"github.com/spf13/cobra"
)

var version = "1.4.0"

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
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat(config.ConfigDir()); os.IsNotExist(err) {
			cfg := config.Load()
			if !cfg.Setup.Complete {
				fmt.Println(ui.Subtle.Sprint("  Tip: Run `palm setup` to get started with curated tool presets"))
				fmt.Println()
			}
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if !offlineMode {
			update.CheckForUpdate(version)
		}
	},
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
	)
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

package cmd

import (
	"embed"

	"github.com/msalah0e/tamr/internal/registry"
	"github.com/msalah0e/tamr/internal/ui"
	"github.com/spf13/cobra"
)

var version = "0.3.0"

var (
	reg        *registry.Registry
	registryFS embed.FS
)

// SetRegistryFS sets the embedded filesystem containing TOML registry files.
func SetRegistryFS(fs embed.FS) {
	registryFS = fs
}

func loadRegistry() *registry.Registry {
	if reg != nil {
		return reg
	}
	r, err := registry.LoadFromFS(registryFS, "registry")
	if err != nil {
		ui.Bad.Printf("tamr: failed to load registry: %v\n", err)
		return registry.New(nil)
	}
	reg = r
	return reg
}

var rootCmd = &cobra.Command{
	Use:   "tamr",
	Short: "tamr — the AI tool manager",
	Long: ui.Brand.Sprint(ui.Palm+" tamr") + " — manage your AI tools from one place\n" +
		ui.Subtle.Sprint("Install, configure, and run AI CLI tools with one command"),
	Version: version + " " + ui.Palm,
}

func init() {
	rootCmd.SetVersionTemplate("tamr {{ .Version }}\n")

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
		discoverCmd(),
		statsCmd(),
	)
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

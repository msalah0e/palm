package cmd

import (
	"fmt"
	"os"

	"github.com/msalah0e/palm/internal/cache"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

func cacheCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage offline cache — download tools and create bundles",
	}

	cmd.AddCommand(
		cacheFetchCmd(),
		cacheBundleCmd(),
	)

	return cmd
}

func cacheFetchCmd() *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "fetch [tool...]",
		Short: "Pre-download tools for offline install",
		Long:  "Download tool packages to local cache for later offline installation",
		Run: func(cmd *cobra.Command, args []string) {
			reg := loadRegistry()

			if all {
				ui.Banner("fetching all tools")
				tools := reg.All()
				success, failed := 0, 0
				for _, tool := range tools {
					backend, pkg := tool.InstallMethod()
					if backend == "manual" {
						continue
					}
					fmt.Printf("  Caching %s (%s)... ", ui.Brand.Sprint(tool.DisplayName), backend)
					if err := cache.Fetch(backend, pkg); err != nil {
						ui.Bad.Printf("failed: %v\n", err)
						failed++
					} else {
						ui.Good.Println("done")
						success++
					}
				}
				fmt.Printf("\n  %d cached", success)
				if failed > 0 {
					fmt.Printf(" · %d failed", failed)
				}
				fmt.Println()
				return
			}

			if len(args) == 0 {
				cmd.Help()
				return
			}

			ui.Banner("fetching")
			for _, name := range args {
				tool := reg.Get(name)
				if tool == nil {
					ui.Warn.Printf("  %s unknown tool %q\n", ui.WarnIcon(), name)
					continue
				}
				backend, pkg := tool.InstallMethod()
				fmt.Printf("  Caching %s (%s)... ", ui.Brand.Sprint(tool.DisplayName), backend)
				if err := cache.Fetch(backend, pkg); err != nil {
					ui.Bad.Printf("failed: %v\n", err)
				} else {
					ui.Good.Println("done")
				}
			}

			fmt.Printf("\n  Cache: %s\n", cache.Dir())
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Fetch all tools in registry")
	return cmd
}

func cacheBundleCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bundle <output.tar.gz>",
		Short: "Create portable bundle of cached tools",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			output := args[0]
			ui.Banner("bundling")

			if err := cache.Bundle(output); err != nil {
				ui.Bad.Printf("  Bundle failed: %v\n", err)
				os.Exit(1)
			}

			ui.Good.Printf("  %s Bundle created: %s\n", ui.StatusIcon(true), output)
		},
	}
}

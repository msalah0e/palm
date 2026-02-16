package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/msalah0e/tamr/internal/ui"
	"github.com/msalah0e/tamr/internal/update"
	"github.com/spf13/cobra"
)

func selfUpdateCmd() *cobra.Command {
	var check bool

	cmd := &cobra.Command{
		Use:     "self-update",
		Aliases: []string{"selfupdate"},
		Short:   "Update tamr itself to the latest version",
		Run: func(cmd *cobra.Command, args []string) {
			if check {
				ui.Banner("version check")
				update.CheckNow(version)
				return
			}

			ui.Banner("self-update")

			if hasGo() {
				fmt.Println("  Updating via go install...")
				c := exec.Command("go", "install", "github.com/msalah0e/tamr@latest")
				c.Stdout = os.Stdout
				c.Stderr = os.Stderr
				if err := c.Run(); err != nil {
					ui.Bad.Printf("\n  Update failed: %v\n", err)
					fmt.Println("  Try: curl -fsSL https://msalah0e.github.io/tamr/install.sh | sh")
					os.Exit(1)
				}
				fmt.Println()
				ui.Good.Printf("  %s tamr updated successfully\n", ui.StatusIcon(true))
				return
			}

			// Fallback to install script
			fmt.Println("  Go not found, using install script...")
			c := exec.Command("sh", "-c", "curl -fsSL https://msalah0e.github.io/tamr/install.sh | sh")
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				ui.Bad.Printf("\n  Update failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Println()
			ui.Good.Printf("  %s tamr updated successfully\n", ui.StatusIcon(true))
		},
	}

	cmd.Flags().BoolVar(&check, "check", false, "Only check for updates, don't install")
	return cmd
}

func hasGo() bool {
	_, err := exec.LookPath("go")
	return err == nil
}

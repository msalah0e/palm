package cmd

import (
	"fmt"

	"github.com/msalah0e/tamr/internal/registry"
	"github.com/msalah0e/tamr/internal/ui"
	"github.com/spf13/cobra"
)

func aiListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed AI tools with status",
		Run: func(cmd *cobra.Command, args []string) {
			reg := loadRegistry()
			detected := registry.DetectInstalled(reg)

			ui.Banner("installed AI tools")

			if len(detected) == 0 {
				fmt.Println("  No AI tools detected.")
				fmt.Println("  Run `tamr ai discover` to explore available tools")
				return
			}

			headers := []string{"Tool", "Version", "Category", "Keys", "Status"}
			var rows [][]string
			needsAttention := 0

			for _, dt := range detected {
				keyStatus := "(local)"
				status := ui.StatusIcon(true) + " ready"

				if dt.Tool.NeedsAPIKey() {
					if len(dt.KeysMissing) == 0 {
						keyStatus = ui.StatusIcon(true) + " API key"
					} else {
						keyStatus = ui.StatusIcon(false) + " missing"
						status = ui.WarnIcon() + " needs key"
						needsAttention++
					}
				}

				ver := dt.Version
				if ver == "" {
					ver = "?"
				}

				rows = append(rows, []string{
					dt.Tool.Name,
					ver,
					dt.Tool.Category,
					keyStatus,
					status,
				})
			}

			ui.Table(headers, rows)

			fmt.Println()
			msg := fmt.Sprintf("  %d tools installed", len(detected))
			if needsAttention > 0 {
				msg += fmt.Sprintf(" \u00b7 %d needs attention", needsAttention)
			}
			fmt.Println(msg)
		},
	}
}

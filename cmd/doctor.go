package cmd

import (
	"fmt"
	"os/exec"

	"github.com/msalah0e/palm/internal/registry"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

func doctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "doctor",
		Aliases: []string{"dr"},
		Short:   "Health check — verify tools, keys, and runtimes",
		Run: func(cmd *cobra.Command, args []string) {
			reg := loadRegistry()
			detected := registry.DetectInstalled(reg)

			ui.Banner("health check")

			healthy := 0
			warnings := 0

			for _, dt := range detected {
				ver := dt.Version
				if ver == "" {
					ver = "?"
				}

				if len(dt.KeysMissing) > 0 {
					fmt.Printf("  %s %s %s — missing %s\n",
						ui.WarnIcon(), dt.Tool.Name, ver, dt.KeysMissing)
					warnings++
				} else {
					extra := ""
					if dt.Tool.NeedsAPIKey() && len(dt.KeysSet) > 0 {
						extra = fmt.Sprintf(" — %s set", dt.KeysSet[0])
					}
					fmt.Printf("  %s %s %s%s\n",
						ui.StatusIcon(true), dt.Tool.Name, ver, extra)
					healthy++
				}
			}

			if len(detected) == 0 {
				fmt.Println("  No AI tools installed.")
			}

			fmt.Println()
			checkRuntime("Python", "python3", "--version")
			checkRuntime("uv", "uv", "--version")
			checkRuntime("Node", "node", "--version")
			checkRuntime("npm", "npm", "--version")
			checkRuntime("Go", "go", "version")
			checkRuntime("Cargo", "cargo", "--version")
			checkRuntime("Docker", "docker", "--version")

			if len(detected) > 0 {
				fmt.Printf("\n  %d/%d tools healthy", healthy, len(detected))
				if warnings > 0 {
					fmt.Printf(" · %d warning(s)", warnings)
				}
				fmt.Println()
			}
		},
	}
}

func checkRuntime(name, bin string, args ...string) {
	if path, err := exec.LookPath(bin); err == nil {
		cmd := exec.Command(path, args...)
		out, _ := cmd.Output()
		ver := registry.ExtractVersion(string(out))
		fmt.Printf("  %s %s: %s\n", ui.StatusIcon(true), name, ver)
	} else {
		fmt.Printf("  %s %s: not found\n", ui.Subtle.Sprint("-"), name)
	}
}

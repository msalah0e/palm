package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/msalah0e/palm/internal/registry"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

func doctorCmd() *cobra.Command {
	var deep bool

	cmd := &cobra.Command{
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

			if runtime.GOOS == "linux" {
				for _, pm := range []struct{ name, bin string }{
					{"apt-get", "apt-get"},
					{"dnf", "dnf"},
					{"pacman", "pacman"},
				} {
					checkRuntime(pm.name, pm.bin, "--version")
				}
			}

			if len(detected) > 0 {
				fmt.Printf("\n  %d/%d tools healthy", healthy, len(detected))
				if warnings > 0 {
					fmt.Printf(" · %d warning(s)", warnings)
				}
				fmt.Println()
			}

			if deep {
				fmt.Println()
				runDeepChecks()
			}
		},
	}

	cmd.Flags().BoolVar(&deep, "deep", false, "Run extended health checks (configs, disk, network)")
	return cmd
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

func runDeepChecks() {
	ui.Banner("deep checks")

	// Config directory
	configDir := palmConfigDir()
	if _, err := os.Stat(configDir); err == nil {
		size := dirSizeDoctor(configDir)
		fmt.Printf("  %s Config dir: %s (%.1f KB)\n", ui.StatusIcon(true), configDir, float64(size)/1024)
	} else {
		fmt.Printf("  %s Config dir: not found\n", ui.StatusIcon(false))
	}

	// Vault check
	vaultPath := filepath.Join(configDir, "vault.enc")
	if info, err := os.Stat(vaultPath); err == nil {
		fmt.Printf("  %s Vault: %s (%.1f KB)\n", ui.StatusIcon(true), vaultPath, float64(info.Size())/1024)
	} else {
		ui.Subtle.Printf("  - Vault: not created\n")
	}

	// Graph check
	graphPath := filepath.Join(configDir, "graph.enc")
	if info, err := os.Stat(graphPath); err == nil {
		fmt.Printf("  %s Graph: %s (%.1f KB)\n", ui.StatusIcon(true), graphPath, float64(info.Size())/1024)
	} else {
		ui.Subtle.Printf("  - Graph: not created\n")
	}

	// AI rules files check
	fmt.Println()
	fmt.Println("  AI config files:")
	ruleFileChecks := map[string]string{
		"CLAUDE.md":                       "Claude Code",
		".cursorrules":                    "Cursor (legacy)",
		".cursor/rules/palm.mdc":          "Cursor (rules)",
		".github/copilot-instructions.md": "GitHub Copilot",
		"AGENTS.md":                       "OpenAI Codex",
		".windsurfrules":                  "Windsurf",
		".aider.conf.yml":                 "Aider",
		"GEMINI.md":                       "Gemini",
		".palm-rules.md":                  "palm rules",
		".palm-context.md":                "palm context",
		".palm-team.json":                 "palm team",
	}
	found := 0
	for file, tool := range ruleFileChecks {
		if _, err := os.Stat(file); err == nil {
			fmt.Printf("    %s %-38s %s\n", ui.StatusIcon(true), file, ui.Subtle.Sprint(tool))
			found++
		}
	}
	if found == 0 {
		ui.Subtle.Println("    No AI config files found in current directory")
	}

	// Network check
	fmt.Println()
	fmt.Print("  Network: ")
	if out, err := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "--connect-timeout", "3", "https://api.github.com").Output(); err == nil {
		code := strings.TrimSpace(string(out))
		if code == "200" || code == "403" {
			ui.Good.Printf("reachable (github.com → %s)\n", code)
		} else {
			ui.Warn.Printf("unexpected response: %s\n", code)
		}
	} else {
		ui.Bad.Println("unreachable")
	}

	// Git check
	if out, err := exec.Command("git", "config", "--global", "user.name").Output(); err == nil {
		name := strings.TrimSpace(string(out))
		if name != "" {
			fmt.Printf("  %s Git user: %s\n", ui.StatusIcon(true), name)
		}
	}
}

func dirSizeDoctor(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		size += info.Size()
		return nil
	})
	return size
}

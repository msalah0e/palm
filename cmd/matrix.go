package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/msalah0e/palm/internal/budget"
	"github.com/msalah0e/palm/internal/models"
	"github.com/msalah0e/palm/internal/proxy"
	"github.com/msalah0e/palm/internal/registry"
	"github.com/msalah0e/palm/internal/session"
	"github.com/msalah0e/palm/internal/ui"
	"github.com/msalah0e/palm/internal/vault"
	"github.com/spf13/cobra"
)

func matrixCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "matrix",
		Aliases: []string{"dashboard", "dash"},
		Short:   "Terminal dashboard — tools, keys, sessions, budget at a glance",
		Run: func(cmd *cobra.Command, args []string) {
			reg := loadRegistry()
			v := vault.New()

			// Header
			fmt.Printf("\n  %s %s v%s — Control Plane\n", ui.Palm, ui.Brand.Sprint("palm"), version)
			fmt.Println("  " + strings.Repeat("═", 60))

			// Section 1: Installed Tools
			fmt.Printf("\n  %s\n\n", ui.Brand.Sprint("Installed Tools"))
			detected := registry.DetectInstalled(reg)
			if len(detected) == 0 {
				fmt.Println("    No tools installed")
			} else {
				for _, dt := range detected {
					ver := dt.Version
					if ver == "" {
						ver = "?"
					}
					status := ui.StatusIcon(true)
					extra := ""
					if len(dt.KeysMissing) > 0 {
						status = ui.WarnIcon()
						extra = " — missing: " + strings.Join(dt.KeysMissing, ", ")
					}
					fmt.Printf("    %s %-20s %s%s\n", status, dt.Tool.Name, ui.Subtle.Sprint(ver), extra)
				}
				fmt.Printf("\n    %d tools installed", len(detected))
			}

			// Section 2: Runtimes
			fmt.Printf("\n\n  %s\n\n", ui.Brand.Sprint("Runtimes"))
			runtimes := []struct {
				name string
				bin  string
				args []string
			}{
				{"Python", "python3", []string{"--version"}},
				{"Node", "node", []string{"--version"}},
				{"Go", "go", []string{"version"}},
				{"Docker", "docker", []string{"--version"}},
			}
			for _, rt := range runtimes {
				if path, err := exec.LookPath(rt.bin); err == nil {
					c := exec.Command(path, rt.args...)
					out, _ := c.Output()
					ver := registry.ExtractVersion(string(out))
					fmt.Printf("    %s %-12s %s\n", ui.StatusIcon(true), rt.name, ver)
				} else {
					fmt.Printf("    %s %-12s not found\n", ui.Subtle.Sprint("-"), rt.name)
				}
			}

			// Section 3: API Keys
			fmt.Printf("\n  %s\n\n", ui.Brand.Sprint("Vault Keys"))
			keys, err := v.List()
			if err == nil && len(keys) > 0 {
				for _, key := range keys {
					val, err := v.Get(key)
					masked := "****"
					if err == nil {
						masked = vault.Mask(val)
					}
					fmt.Printf("    %s %-30s %s\n", ui.StatusIcon(true), key, ui.Subtle.Sprint(masked))
				}
				fmt.Printf("\n    %d keys stored", len(keys))
			} else {
				fmt.Println("    No API keys stored")
			}

			// Section 4: Providers
			fmt.Printf("\n\n  %s\n\n", ui.Brand.Sprint("LLM Providers"))
			for _, p := range models.BuiltinProviders() {
				available := false
				if p.EnvKey == "" {
					available = true
				} else if os.Getenv(p.EnvKey) != "" {
					available = true
				} else if _, err := v.Get(p.EnvKey); err == nil {
					available = true
				}
				icon := ui.StatusIcon(available)
				fmt.Printf("    %s %-12s %d models\n", icon, p.Name, len(p.Models))
			}

			// Section 5: Budget
			fmt.Printf("\n  %s\n\n", ui.Brand.Sprint("Budget"))
			budgetStatus, err := budget.GetStatus()
			if err == nil && (budgetStatus.MonthlyLimit > 0 || budgetStatus.DailyLimit > 0) {
				if budgetStatus.MonthlyLimit > 0 {
					icon := ui.StatusIcon(true)
					if budgetStatus.IsOverBudget {
						icon = ui.StatusIcon(false)
					} else if budgetStatus.IsNearBudget {
						icon = ui.WarnIcon()
					}
					bar := progressBar(budgetStatus.PercentUsed, 20)
					fmt.Printf("    %s Monthly: $%.2f / $%.2f  %s\n", icon, budgetStatus.MonthlySpend, budgetStatus.MonthlyLimit, bar)
				}
				if budgetStatus.DailyLimit > 0 {
					fmt.Printf("    Daily: $%.2f / $%.2f\n", budgetStatus.DailySpend, budgetStatus.DailyLimit)
				}
			} else {
				fmt.Println("    No budget configured")
			}

			// Section 6: Recent Sessions
			fmt.Printf("\n  %s\n\n", ui.Brand.Sprint("Recent Sessions"))
			sessions, err := session.List(5)
			if err == nil && len(sessions) > 0 {
				for _, s := range sessions {
					dur := formatDuration(time.Duration(s.Duration * float64(time.Second)))
					icon := ui.StatusIcon(s.ExitCode == 0)
					ago := time.Since(s.StartedAt).Round(time.Second)
					fmt.Printf("    %s %-15s %s  %s ago\n", icon, s.Tool, dur, ui.Subtle.Sprint(ago))
				}
			} else {
				fmt.Println("    No sessions recorded")
			}

			// Section 7: Proxy
			fmt.Printf("\n  %s\n\n", ui.Brand.Sprint("Proxy"))
			if running, pid := proxy.IsRunning(); running {
				fmt.Printf("    %s Running (PID %d)\n", ui.StatusIcon(true), pid)
			} else {
				fmt.Printf("    %s Not running\n", ui.Subtle.Sprint("-"))
			}

			// Registry stats
			fmt.Printf("\n  %s\n\n", ui.Brand.Sprint("Registry"))
			allTools := reg.All()
			cats := reg.Categories()
			fmt.Printf("    %d tools across %d categories\n", len(allTools), len(cats))

			fmt.Println("\n  " + strings.Repeat("═", 60))
			fmt.Println()
		},
	}
}

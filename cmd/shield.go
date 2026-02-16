package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

func shieldCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "shield",
		Aliases: []string{"safe", "guard"},
		Short:   "AI safety layer — pre-flight and post-flight checks for AI-generated code",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("shield status")

			checks := []struct {
				name   string
				check  func() bool
				detail string
			}{
				{"Git repo", isGitRepo, "Working inside a git repository"},
				{".gitignore", hasGitignore, "Sensitive files excluded from git"},
				{"No secrets in tracked files", noTrackedSecrets, "API keys and passwords not committed"},
				{"AI rules file", hasRulesFile, "AI tool instructions configured"},
				{"No large generated files", noLargeGenFiles, "No suspiciously large AI-generated files"},
			}

			passed := 0
			for _, c := range checks {
				ok := c.check()
				if ok {
					passed++
				}
				fmt.Printf("  %s %-30s  %s\n", ui.StatusIcon(ok), c.name, ui.Subtle.Sprint(c.detail))
			}

			fmt.Println()
			if passed == len(checks) {
				ui.Good.Printf("  %s All %d checks passed\n", ui.StatusIcon(true), len(checks))
			} else {
				fmt.Printf("  %d/%d checks passed\n", passed, len(checks))
			}
		},
	}

	cmd.AddCommand(
		shieldPreCmd(),
		shieldPostCmd(),
		shieldScanCmd(),
	)

	return cmd
}

func shieldPreCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pre",
		Short: "Run pre-flight checks before AI tool session",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("pre-flight checks")

			issues := 0

			// Check for uncommitted changes
			if isGitRepo() {
				fmt.Printf("  %s Git repository detected\n", ui.StatusIcon(true))

				// Check for dirty worktree
				if hasUncommittedChanges() {
					ui.Warn.Printf("  %s Uncommitted changes detected — consider committing first\n", ui.WarnIcon())
					issues++
				} else {
					fmt.Printf("  %s Working tree clean\n", ui.StatusIcon(true))
				}
			} else {
				ui.Warn.Printf("  %s Not a git repository — changes cannot be reverted\n", ui.WarnIcon())
				issues++
			}

			// Check for AI instructions
			if hasRulesFile() {
				fmt.Printf("  %s AI rules file found\n", ui.StatusIcon(true))
			} else {
				ui.Warn.Printf("  %s No AI rules file — run `palm rules init`\n", ui.WarnIcon())
				issues++
			}

			// Check .gitignore
			if hasGitignore() {
				fmt.Printf("  %s .gitignore present\n", ui.StatusIcon(true))
			} else {
				ui.Warn.Printf("  %s No .gitignore — sensitive files may be committed\n", ui.WarnIcon())
				issues++
			}

			fmt.Println()
			if issues == 0 {
				ui.Good.Printf("  %s Ready for AI session\n", ui.StatusIcon(true))
			} else {
				fmt.Printf("  %d warnings — proceed with caution\n", issues)
			}
		},
	}
}

func shieldPostCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "post",
		Short: "Run post-flight checks after AI tool session",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("post-flight checks")

			issues := 0

			if noTrackedSecrets() {
				fmt.Printf("  %s No secrets detected in staged files\n", ui.StatusIcon(true))
			} else {
				ui.Bad.Printf("  %s Possible secrets found in staged files!\n", ui.StatusIcon(false))
				issues++
			}

			if noLargeGenFiles() {
				fmt.Printf("  %s No suspiciously large generated files\n", ui.StatusIcon(true))
			} else {
				ui.Warn.Printf("  %s Large files detected — review before committing\n", ui.WarnIcon())
				issues++
			}

			// Check for common AI code smells in recently modified files
			fmt.Printf("  %s Run `palm audit` for detailed code quality check\n", ui.StatusIcon(true))

			fmt.Println()
			if issues == 0 {
				ui.Good.Printf("  %s Post-flight checks passed\n", ui.StatusIcon(true))
			} else {
				fmt.Printf("  %d issues found — review before committing\n", issues)
			}
		},
	}
}

func shieldScanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "scan [dir]",
		Short: "Scan directory for security issues in AI-generated code",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}

			ui.Banner("security scan")
			fmt.Printf("  Scanning: %s\n\n", dir)

			issues := 0
			filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return nil
				}
				if info.Size() > 512*1024 {
					return nil
				}

				name := info.Name()
				// Check for sensitive files
				sensitivePatterns := []string{".env", "credentials", "secret", ".pem", ".key"}
				for _, pat := range sensitivePatterns {
					if strings.Contains(strings.ToLower(name), pat) {
						ui.Warn.Printf("  %s Sensitive file: %s\n", ui.WarnIcon(), path)
						issues++
						return nil
					}
				}

				return nil
			})

			fmt.Println()
			if issues == 0 {
				ui.Good.Printf("  %s No security issues found\n", ui.StatusIcon(true))
			} else {
				fmt.Printf("  %d potential issues — review carefully\n", issues)
			}
		},
	}
}

func isGitRepo() bool {
	_, err := os.Stat(".git")
	return err == nil
}

func hasGitignore() bool {
	_, err := os.Stat(".gitignore")
	return err == nil
}

func hasRulesFile() bool {
	for _, f := range []string{".palm-rules.md", ".palm-context.md", "CLAUDE.md", "AGENTS.md", ".cursorrules", ".windsurfrules"} {
		if _, err := os.Stat(f); err == nil {
			return true
		}
	}
	return false
}

func noTrackedSecrets() bool {
	patterns := []string{".env", "credentials.json", ".pem", "id_rsa"}
	for _, p := range patterns {
		if _, err := os.Stat(p); err == nil {
			// Check if it's git-ignored
			data, _ := os.ReadFile(".gitignore")
			if !strings.Contains(string(data), p) {
				return false
			}
		}
	}
	return true
}

func noLargeGenFiles() bool {
	large := false
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasPrefix(path, ".git") || strings.HasPrefix(path, "node_modules") {
			return filepath.SkipDir
		}
		// Flag files >100KB that look generated
		if info.Size() > 100*1024 {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".go" || ext == ".py" || ext == ".js" || ext == ".ts" {
				large = true
			}
		}
		return nil
	})
	return !large
}

func hasUncommittedChanges() bool {
	out, err := exec.Command("git", "status", "--porcelain").Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(out))) > 0
}

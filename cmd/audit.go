package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/msalah0e/palm/internal/ui"
	"github.com/spf13/cobra"
)

type auditIssue struct {
	File     string
	Line     int
	Severity string // "error", "warning", "info"
	Message  string
}

func auditCmd() *cobra.Command {
	var fix bool

	cmd := &cobra.Command{
		Use:   "audit [file|dir]",
		Short: "AI code quality gate — detect common AI-generated code issues",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			target := "."
			if len(args) > 0 {
				target = args[0]
			}

			ui.Banner("code audit")

			info, err := os.Stat(target)
			if err != nil {
				ui.Bad.Printf("  %v\n", err)
				os.Exit(1)
			}

			var issues []auditIssue
			if info.IsDir() {
				filepath.Walk(target, func(path string, fi os.FileInfo, err error) error {
					if err != nil || fi.IsDir() {
						return nil
					}
					if fi.Size() > 512*1024 { // Skip >512KB
						return nil
					}
					ext := strings.ToLower(filepath.Ext(path))
					if ext == ".go" || ext == ".py" || ext == ".js" || ext == ".ts" || ext == ".tsx" {
						fileIssues := auditFile(path)
						issues = append(issues, fileIssues...)
					}
					return nil
				})
			} else {
				issues = auditFile(target)
			}

			if len(issues) == 0 {
				ui.Good.Printf("  %s No issues found\n", ui.StatusIcon(true))
				return
			}

			errors, warnings, infos := 0, 0, 0
			for _, issue := range issues {
				icon := ui.StatusIcon(false)
				switch issue.Severity {
				case "warning":
					icon = ui.WarnIcon()
					warnings++
				case "info":
					icon = ui.Info.Sprint("i")
					infos++
				default:
					errors++
				}
				fmt.Printf("  %s %s:%d — %s\n", icon, issue.File, issue.Line, issue.Message)
			}

			fmt.Println()
			fmt.Printf("  %d errors, %d warnings, %d info\n", errors, warnings, infos)

			if fix {
				fmt.Println()
				ui.Info.Println("  Auto-fix is not yet available. Review issues manually.")
			}
		},
	}

	cmd.Flags().BoolVar(&fix, "fix", false, "Attempt to auto-fix issues (coming soon)")
	return cmd
}

func auditFile(path string) []auditIssue {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	var issues []auditIssue

	ext := strings.ToLower(filepath.Ext(path))
	relPath, _ := filepath.Rel(".", path)

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Common AI code smells
		if strings.Contains(trimmed, "TODO: implement") || strings.Contains(trimmed, "TODO: add") {
			issues = append(issues, auditIssue{relPath, lineNum, "warning", "Placeholder TODO — likely unimplemented AI suggestion"})
		}
		if strings.Contains(trimmed, "// This function") && strings.Contains(trimmed, "...") {
			issues = append(issues, auditIssue{relPath, lineNum, "warning", "Truncated AI comment"})
		}
		if strings.Contains(trimmed, "pass  #") || trimmed == "pass" {
			if ext == ".py" {
				issues = append(issues, auditIssue{relPath, lineNum, "info", "Empty pass statement — may be AI placeholder"})
			}
		}
		if strings.Contains(trimmed, "console.log(") && (ext == ".ts" || ext == ".tsx" || ext == ".js") {
			issues = append(issues, auditIssue{relPath, lineNum, "info", "Debug console.log left in code"})
		}
		if strings.Contains(trimmed, "fmt.Println(\"debug") || strings.Contains(trimmed, "fmt.Println(\"DEBUG") {
			issues = append(issues, auditIssue{relPath, lineNum, "info", "Debug print statement"})
		}

		// Security checks
		if strings.Contains(line, "password") && strings.Contains(line, "=") && strings.Contains(line, "\"") {
			if !strings.Contains(trimmed, "//") && !strings.Contains(trimmed, "#") && !strings.HasPrefix(trimmed, "*") {
				issues = append(issues, auditIssue{relPath, lineNum, "error", "Possible hardcoded password"})
			}
		}
		if strings.Contains(line, "api_key") && strings.Contains(line, "\"sk-") {
			issues = append(issues, auditIssue{relPath, lineNum, "error", "Possible hardcoded API key"})
		}
		if strings.Contains(line, "secret") && strings.Contains(line, "=") && len(line) > 50 {
			if !strings.Contains(trimmed, "//") && !strings.Contains(trimmed, "#") && !strings.HasPrefix(trimmed, "*") && !strings.HasPrefix(trimmed, "os.") && !strings.HasPrefix(trimmed, "env") {
				issues = append(issues, auditIssue{relPath, lineNum, "warning", "Possible hardcoded secret"})
			}
		}

		// Unused imports / dead code patterns
		if ext == ".go" && strings.HasPrefix(trimmed, "_ = ") {
			issues = append(issues, auditIssue{relPath, lineNum, "info", "Blank identifier assignment — possibly suppressing unused error"})
		}

		// Overly long lines (common in AI output)
		if len(line) > 200 {
			issues = append(issues, auditIssue{relPath, lineNum, "info", fmt.Sprintf("Very long line (%d chars) — consider breaking up", len(line))})
		}
	}

	return issues
}

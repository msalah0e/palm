package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/msalah0e/palm/internal/ui"
	"github.com/msalah0e/palm/internal/vault"
	"github.com/spf13/cobra"
)

func worktreeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "worktree",
		Short: "Manage git worktrees for parallel branch work",
		Long: `Work on multiple branches simultaneously with git worktrees.
Each worktree is a separate checkout, so you can run AI tools on one
branch without disturbing your current work.

Examples:
  palm worktree add feature-auth             # Create worktree for branch
  palm worktree list                         # List active worktrees
  palm worktree run feature-auth aider       # Run tool in worktree
  palm worktree remove feature-auth          # Clean up worktree`,
	}

	cmd.AddCommand(
		worktreeAddCmd(),
		worktreeListCmd(),
		worktreeRemoveCmd(),
		worktreeRunCmd(),
	)

	return cmd
}

func worktreeAddCmd() *cobra.Command {
	var path string

	cmd := &cobra.Command{
		Use:   "add <branch>",
		Short: "Create a worktree for the given branch",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			branch := args[0]

			if path == "" {
				// Default: ../<repo>-<branch>
				cwd, _ := os.Getwd()
				repoName := filepath.Base(cwd)
				path = filepath.Join(filepath.Dir(cwd), repoName+"-"+branch)
			}

			ui.Banner("worktree add")

			// Check if branch exists
			var stderr bytes.Buffer
			check := exec.Command("git", "rev-parse", "--verify", branch)
			check.Stderr = &stderr
			branchExists := check.Run() == nil

			var gitArgs []string
			if branchExists {
				gitArgs = []string{"worktree", "add", path, branch}
			} else {
				gitArgs = []string{"worktree", "add", "-b", branch, path}
			}

			c := exec.Command("git", gitArgs...)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr

			if err := c.Run(); err != nil {
				ui.Bad.Printf("  Failed to create worktree: %v\n", err)
				os.Exit(1)
			}

			ui.Good.Printf("  %s Worktree created at %s\n", ui.StatusIcon(true), path)
			if !branchExists {
				fmt.Printf("  New branch %s created\n", ui.Brand.Sprint(branch))
			}
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "Custom path for the worktree")
	return cmd
}

func worktreeListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List active worktrees",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Banner("worktrees")

			out, err := exec.Command("git", "worktree", "list", "--porcelain").Output()
			if err != nil {
				ui.Bad.Printf("  Failed to list worktrees: %v\n", err)
				os.Exit(1)
			}

			// Parse porcelain output
			type worktree struct {
				path   string
				branch string
				bare   bool
			}

			var trees []worktree
			var current worktree

			for _, line := range strings.Split(string(out), "\n") {
				if strings.HasPrefix(line, "worktree ") {
					if current.path != "" {
						trees = append(trees, current)
					}
					current = worktree{path: strings.TrimPrefix(line, "worktree ")}
				} else if strings.HasPrefix(line, "branch ") {
					ref := strings.TrimPrefix(line, "branch ")
					current.branch = strings.TrimPrefix(ref, "refs/heads/")
				} else if line == "bare" {
					current.bare = true
				}
			}
			if current.path != "" {
				trees = append(trees, current)
			}

			if len(trees) == 0 {
				fmt.Println("  No worktrees found")
				return
			}

			headers := []string{"Branch", "Path"}
			var rows [][]string

			for _, t := range trees {
				branch := t.branch
				if branch == "" && t.bare {
					branch = "(bare)"
				}
				rows = append(rows, []string{branch, t.path})
			}

			ui.Table(headers, rows)
			fmt.Printf("\n  %d worktrees\n", len(trees))
		},
	}
}

func worktreeRemoveCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:     "remove <branch>",
		Aliases: []string{"rm"},
		Short:   "Remove a worktree",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			branch := args[0]

			// Find the worktree path for this branch
			out, err := exec.Command("git", "worktree", "list", "--porcelain").Output()
			if err != nil {
				ui.Bad.Printf("  Failed to list worktrees: %v\n", err)
				os.Exit(1)
			}

			var targetPath string
			var currentPath string

			for _, line := range strings.Split(string(out), "\n") {
				if strings.HasPrefix(line, "worktree ") {
					currentPath = strings.TrimPrefix(line, "worktree ")
				} else if strings.HasPrefix(line, "branch refs/heads/"+branch) {
					targetPath = currentPath
				}
			}

			if targetPath == "" {
				ui.Bad.Printf("  No worktree found for branch %q\n", branch)
				os.Exit(1)
			}

			gitArgs := []string{"worktree", "remove", targetPath}
			if force {
				gitArgs = []string{"worktree", "remove", "--force", targetPath}
			}

			c := exec.Command("git", gitArgs...)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr

			if err := c.Run(); err != nil {
				ui.Bad.Printf("  Failed to remove worktree: %v\n", err)
				fmt.Println("  Use --force to remove with uncommitted changes")
				os.Exit(1)
			}

			ui.Good.Printf("  %s Worktree for %s removed\n", ui.StatusIcon(true), ui.Brand.Sprint(branch))
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force removal even with changes")
	return cmd
}

func worktreeRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run <branch> <tool> [args...]",
		Short: "Run an AI tool inside a worktree",
		Long: `Run an AI tool in the context of a specific worktree.
Vault keys are automatically injected.

  palm worktree run feature-auth aider "add login form"
  palm worktree run fix-bug claude-code`,
		Args:               cobra.MinimumNArgs(2),
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			branch := args[0]
			toolName := args[1]
			toolArgs := args[2:]

			// Find the worktree path
			out, err := exec.Command("git", "worktree", "list", "--porcelain").Output()
			if err != nil {
				ui.Bad.Printf("  Failed to list worktrees: %v\n", err)
				os.Exit(1)
			}

			var targetPath string
			var currentPath string

			for _, line := range strings.Split(string(out), "\n") {
				if strings.HasPrefix(line, "worktree ") {
					currentPath = strings.TrimPrefix(line, "worktree ")
				} else if strings.HasPrefix(line, "branch refs/heads/"+branch) {
					targetPath = currentPath
				}
			}

			if targetPath == "" {
				ui.Bad.Printf("  No worktree found for branch %q\n", branch)
				fmt.Printf("  Create one first: palm worktree add %s\n", branch)
				os.Exit(1)
			}

			// Find the tool binary
			reg := loadRegistry()
			tool := reg.Get(toolName)

			bin := toolName
			if tool != nil && tool.Install.Verify.Command != "" {
				parts := strings.Fields(tool.Install.Verify.Command)
				if len(parts) > 0 {
					bin = parts[0]
				}
			}

			binPath, err := exec.LookPath(bin)
			if err != nil {
				ui.Bad.Printf("  %s not found in PATH\n", bin)
				os.Exit(1)
			}

			// Build env with vault keys
			env := os.Environ()
			v := vault.New()
			if tool != nil {
				allKeys := append(tool.Keys.Required, tool.Keys.Optional...)
				for _, key := range allKeys {
					if os.Getenv(key) == "" {
						if val, err := v.Get(key); err == nil {
							env = append(env, fmt.Sprintf("%s=%s", key, val))
						}
					}
				}
			}

			fmt.Printf("  Running %s in worktree %s (%s)\n\n",
				ui.Brand.Sprint(toolName),
				ui.Brand.Sprint(branch),
				targetPath)

			c := exec.Command(binPath, toolArgs...)
			c.Dir = targetPath
			c.Env = env
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr

			if err := c.Run(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					os.Exit(exitErr.ExitCode())
				}
				ui.Bad.Printf("  Failed to run %s: %v\n", toolName, err)
				os.Exit(1)
			}
		},
	}
}

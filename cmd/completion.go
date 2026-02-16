package cmd

import (
	"github.com/msalah0e/palm/internal/models"
	"github.com/msalah0e/palm/internal/vault"
	"github.com/spf13/cobra"
)

// completionCmd generates shell completion scripts.
func completionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate completion scripts for your shell.

  # Bash (add to ~/.bashrc)
  eval "$(palm completion bash)"

  # Zsh (add to ~/.zshrc)
  eval "$(palm completion zsh)"

  # Fish
  palm completion fish | source

  # PowerShell
  palm completion powershell | Out-String | Invoke-Expression`,
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				_ = rootCmd.GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				_ = rootCmd.GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				_ = rootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				_ = rootCmd.GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
			}
		},
	}

	return cmd
}

// toolCompletionFunc provides dynamic completion for tool names.
func toolCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	reg := loadRegistry()
	var completions []string
	for _, t := range reg.All() {
		completions = append(completions, t.Name+"\t"+t.Description)
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

// installedToolCompletionFunc provides completion for installed tools.
func installedToolCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	reg := loadRegistry()
	var completions []string
	for _, t := range reg.All() {
		completions = append(completions, t.Name+"\t"+t.Description)
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

// modelCompletionFunc provides dynamic completion for model names.
func modelCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var completions []string
	for _, m := range models.AllModels() {
		completions = append(completions, m.ID+"\t"+m.Name+" ("+m.Provider+")")
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

// keyCompletionFunc provides dynamic completion for vault key names.
func keyCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	v := vault.New()
	keys, err := v.List()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return keys, cobra.ShellCompDirectiveNoFileComp
}

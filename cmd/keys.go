package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/msalah0e/palm/internal/ui"
	"github.com/msalah0e/palm/internal/vault"
	"github.com/spf13/cobra"
)

func keysCmd() *cobra.Command {
	keysCmd := &cobra.Command{
		Use:   "keys",
		Short: "Manage API keys in the vault",
	}

	keysCmd.AddCommand(
		keysAddCmd(),
		keysRmCmd(),
		keysListCmd(),
		keysExportCmd(),
	)

	return keysCmd
}

func keysAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <KEY_NAME>",
		Short: "Store an API key in the vault",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			keyName := args[0]
			v := vault.New()

			fmt.Printf("  Enter value for %s: ", ui.Brand.Sprint(keyName))
			reader := bufio.NewReader(os.Stdin)
			value, _ := reader.ReadString('\n')
			value = strings.TrimSpace(value)

			if value == "" {
				ui.Warn.Println("  Empty value — key not stored")
				return
			}

			if err := v.Set(keyName, value); err != nil {
				ui.Bad.Printf("  Failed to store key: %v\n", err)
				os.Exit(1)
			}

			ui.Good.Printf("  %s %s stored in vault\n", ui.StatusIcon(true), keyName)
		},
	}
}

func keysRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "rm <KEY_NAME>",
		Aliases: []string{"remove", "delete"},
		Short:   "Remove an API key from the vault",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			keyName := args[0]
			v := vault.New()

			if err := v.Delete(keyName); err != nil {
				ui.Bad.Printf("  Failed to remove key: %v\n", err)
				os.Exit(1)
			}

			ui.Good.Printf("  %s %s removed from vault\n", ui.StatusIcon(true), keyName)
		},
	}
}

func keysListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List stored API keys (masked values)",
		Run: func(cmd *cobra.Command, args []string) {
			v := vault.New()

			ui.Banner("stored API keys")

			keys, err := v.List()
			if err != nil {
				ui.Bad.Printf("  Failed to list keys: %v\n", err)
				os.Exit(1)
			}

			if len(keys) == 0 {
				fmt.Println("  No API keys stored.")
				fmt.Println("  Run `palm keys add <KEY>` to add one")
				return
			}

			for _, key := range keys {
				val, err := v.Get(key)
				masked := "****"
				if err == nil {
					masked = vault.Mask(val)
				}
				fmt.Printf("  %s  %s\n", ui.Brand.Sprintf("%-30s", key), ui.Subtle.Sprint(masked))
			}

			fmt.Printf("\n  %d keys stored\n", len(keys))
		},
	}
}

func keysExportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "export",
		Short: "Print export statements for shell (eval-able)",
		Run: func(cmd *cobra.Command, args []string) {
			v := vault.New()

			keys, err := v.List()
			if err != nil {
				ui.Bad.Printf("  Failed to list keys: %v\n", err)
				os.Exit(1)
			}

			if len(keys) == 0 {
				fmt.Println("# No API keys stored in palm vault")
				return
			}

			fmt.Println("# palm vault — eval $(palm keys export)")
			for _, key := range keys {
				val, err := v.Get(key)
				if err == nil {
					fmt.Printf("export %s=%q\n", key, val)
				}
			}
		},
	}
}

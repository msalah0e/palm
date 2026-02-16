package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/msalah0e/tamr/internal/ui"
	"github.com/msalah0e/tamr/internal/vault"
	"github.com/spf13/cobra"
)

func aiKeysCmd() *cobra.Command {
	keysCmd := &cobra.Command{
		Use:   "keys",
		Short: "Manage API keys in macOS Keychain",
	}

	keysCmd.AddCommand(
		aiKeysAddCmd(),
		aiKeysRmCmd(),
		aiKeysListCmd(),
		aiKeysExportCmd(),
	)

	return keysCmd
}

func aiKeysAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <KEY_NAME>",
		Short: "Store an API key in macOS Keychain",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			keyName := args[0]
			v := vault.NewKeychain()

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

			ui.Good.Printf("  %s %s stored in Keychain\n", ui.StatusIcon(true), keyName)
		},
	}
}

func aiKeysRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "rm <KEY_NAME>",
		Aliases: []string{"remove", "delete"},
		Short:   "Remove an API key from Keychain",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			keyName := args[0]
			v := vault.NewKeychain()

			if err := v.Delete(keyName); err != nil {
				ui.Bad.Printf("  Failed to remove key: %v\n", err)
				os.Exit(1)
			}

			ui.Good.Printf("  %s %s removed from Keychain\n", ui.StatusIcon(true), keyName)
		},
	}
}

func aiKeysListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List stored API keys (masked values)",
		Run: func(cmd *cobra.Command, args []string) {
			v := vault.NewKeychain()

			ui.Banner("stored API keys")

			keys, err := v.List()
			if err != nil {
				ui.Bad.Printf("  Failed to list keys: %v\n", err)
				os.Exit(1)
			}

			if len(keys) == 0 {
				fmt.Println("  No API keys stored.")
				fmt.Println("  Run `tamr ai keys add <KEY>` to add one")
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

func aiKeysExportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "export",
		Short: "Print export statements for shell (eval-able)",
		Run: func(cmd *cobra.Command, args []string) {
			v := vault.NewKeychain()

			keys, err := v.List()
			if err != nil {
				ui.Bad.Printf("  Failed to list keys: %v\n", err)
				os.Exit(1)
			}

			if len(keys) == 0 {
				fmt.Println("# No API keys stored in tamr vault")
				return
			}

			fmt.Println("# tamr vault — eval $(tamr ai keys export)")
			for _, key := range keys {
				val, err := v.Get(key)
				if err == nil {
					fmt.Printf("export %s=%q\n", key, val)
				}
			}
		},
	}
}

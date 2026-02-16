package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/msalah0e/palm/internal/registry"
	"github.com/msalah0e/palm/internal/vault"
	"github.com/spf13/cobra"
)

func envCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "env",
		Short: "Print shell exports for vault keys and tool paths",
		Long:  "Print export statements for eval — usage: eval $(palm env)",
		Run: func(cmd *cobra.Command, args []string) {
			v := vault.New()
			reg := loadRegistry()

			fmt.Println("# palm env — eval $(palm env)")

			// Export all vault keys
			keys, err := v.List()
			if err == nil {
				for _, key := range keys {
					val, err := v.Get(key)
					if err == nil {
						fmt.Printf("export %s=%q\n", key, val)
					}
				}
			}

			// Collect unique directories from installed tool paths
			detected := registry.DetectInstalled(reg)
			seen := make(map[string]bool)
			var paths []string
			for _, dt := range detected {
				if dt.Path != "" {
					dir := filepath.Dir(dt.Path)
					if dir != "" && !seen[dir] {
						paths = append(paths, dir)
						seen[dir] = true
					}
				}
			}

			// Add common tool directories if they exist
			home, _ := os.UserHomeDir()
			if home != "" {
				for _, rel := range []string{".local/bin", "go/bin", ".cargo/bin"} {
					dir := filepath.Join(home, rel)
					if info, err := os.Stat(dir); err == nil && info.IsDir() && !seen[dir] {
						paths = append(paths, dir)
						seen[dir] = true
					}
				}
			}

			if len(paths) > 0 {
				fmt.Printf("export PATH=\"%s:$PATH\"\n", strings.Join(paths, ":"))
			}

			fmt.Println("# end palm env")
		},
	}
}

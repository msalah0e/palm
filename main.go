package main

import (
	"embed"
	"os"

	"github.com/msalah0e/palm/cmd"
)

//go:embed registry/*.toml
var registryFS embed.FS

func main() {
	cmd.SetRegistryFS(registryFS)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

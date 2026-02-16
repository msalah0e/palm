package registry

import (
	"embed"
	"fmt"

	"github.com/BurntSushi/toml"
)

type toolFile struct {
	Tools []Tool `toml:"tools"`
}

// LoadFromFS loads all tools from an embed.FS containing TOML files in a "registry" directory.
func LoadFromFS(fs embed.FS, dir string) (*Registry, error) {
	entries, err := fs.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading embedded registry: %w", err)
	}

	var allTools []Tool
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := fs.ReadFile(dir + "/" + entry.Name())
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", entry.Name(), err)
		}

		var tf toolFile
		if err := toml.Unmarshal(data, &tf); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", entry.Name(), err)
		}
		allTools = append(allTools, tf.Tools...)
	}

	return New(allTools), nil
}

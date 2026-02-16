package registry

import (
	"embed"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/msalah0e/tamr/internal/config"
)

// LoadAll merges the embedded registry with external plugin files from ~/.config/tamr/plugins/.
func LoadAll(fs embed.FS, dir string) (*Registry, error) {
	// Load embedded (built-in) tools
	reg, err := LoadFromFS(fs, dir)
	if err != nil {
		return nil, err
	}
	tools := reg.All()

	// Load external plugin files
	pluginDir := filepath.Join(config.ConfigDir(), "plugins")
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		// No plugins directory is fine
		return New(tools), nil
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(pluginDir, entry.Name()))
		if err != nil {
			continue
		}
		var tf toolFile
		if err := toml.Unmarshal(data, &tf); err != nil {
			continue
		}
		tools = append(tools, tf.Tools...)
	}

	return New(dedup(tools)), nil
}

// dedup removes duplicate tools by name, keeping the last occurrence (external overrides embedded).
func dedup(tools []Tool) []Tool {
	seen := make(map[string]int, len(tools))
	for i, t := range tools {
		seen[t.Name] = i
	}
	result := make([]Tool, 0, len(seen))
	added := make(map[string]bool, len(seen))
	for _, t := range tools {
		idx := seen[t.Name]
		if !added[t.Name] && idx == indexOf(tools, t.Name) {
			// Keep the last occurrence
			result = append(result, tools[seen[t.Name]])
			added[t.Name] = true
		}
	}
	return result
}

func indexOf(tools []Tool, name string) int {
	// Return the last index of this name
	last := -1
	for i, t := range tools {
		if t.Name == name {
			last = i
		}
	}
	return last
}

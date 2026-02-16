package registry

import (
	"embed"
	"fmt"

	"github.com/BurntSushi/toml"
)

type toolFile struct {
	Tools []Tool `toml:"tools"`
}

type presetFile struct {
	Presets []Preset `toml:"presets"`
}

// LoadFromFS loads all tools from an embed.FS containing TOML files in a "registry" directory.
func LoadFromFS(fs embed.FS, dir string) (*Registry, error) {
	entries, err := fs.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading embedded registry: %w", err)
	}

	var allTools []Tool
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "presets.toml" {
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

// LoadPresetsFromFS loads preset definitions from presets.toml in the embedded FS.
func LoadPresetsFromFS(fs embed.FS, dir string) ([]Preset, error) {
	data, err := fs.ReadFile(dir + "/presets.toml")
	if err != nil {
		return nil, fmt.Errorf("reading presets.toml: %w", err)
	}

	var pf presetFile
	if err := toml.Unmarshal(data, &pf); err != nil {
		return nil, fmt.Errorf("parsing presets.toml: %w", err)
	}
	return pf.Presets, nil
}

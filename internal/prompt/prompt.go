package prompt

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Prompt represents a stored prompt template.
type Prompt struct {
	Name      string
	Content   string
	Variables []string
	CreatedAt time.Time
}

// promptDir returns the prompts directory path.
func promptDir() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "palm", "prompts")
}

// Save stores a prompt to disk.
func Save(name, content string) error {
	dir := promptDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, name+".md")
	return os.WriteFile(path, []byte(content), 0o644)
}

// Load reads a prompt from disk.
func Load(name string) (*Prompt, error) {
	path := filepath.Join(promptDir(), name+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("prompt not found: %s", name)
	}
	info, _ := os.Stat(path)
	content := string(data)
	return &Prompt{
		Name:      name,
		Content:   content,
		Variables: extractVariables(content),
		CreatedAt: info.ModTime(),
	}, nil
}

// Delete removes a prompt.
func Delete(name string) error {
	path := filepath.Join(promptDir(), name+".md")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("prompt not found: %s", name)
	}
	return os.Remove(path)
}

// List returns all stored prompts.
func List() ([]Prompt, error) {
	dir := promptDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var prompts []Prompt
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".md")
		p, err := Load(name)
		if err != nil {
			continue
		}
		prompts = append(prompts, *p)
	}
	sort.Slice(prompts, func(i, j int) bool {
		return prompts[i].Name < prompts[j].Name
	})
	return prompts, nil
}

// Render substitutes variables in a prompt.
func Render(content string, vars map[string]string) string {
	result := content
	for k, v := range vars {
		result = strings.ReplaceAll(result, "{{"+k+"}}", v)
	}
	return result
}

// extractVariables finds all {{var}} patterns in content.
func extractVariables(content string) []string {
	seen := make(map[string]bool)
	var vars []string
	for {
		start := strings.Index(content, "{{")
		if start == -1 {
			break
		}
		end := strings.Index(content[start:], "}}")
		if end == -1 {
			break
		}
		name := content[start+2 : start+end]
		name = strings.TrimSpace(name)
		if name != "" && !seen[name] {
			vars = append(vars, name)
			seen[name] = true
		}
		content = content[start+end+2:]
	}
	return vars
}

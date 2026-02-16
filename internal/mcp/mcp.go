package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// Server represents an MCP server in the registry.
type Server struct {
	Name        string `json:"name"`
	Display     string `json:"display"`
	Description string `json:"description"`
	Command     string `json:"command"`
	Args        []string `json:"args"`
	Install     string `json:"install"`
	Backend     string `json:"backend"`
	URL         string `json:"url"`
	Category    string `json:"category"`
}

// ToolConfig represents how a specific AI tool stores MCP configuration.
type ToolConfig struct {
	Name     string
	Path     string
	Format   string // "json-servers", "json-mcp"
}

// Registry is the built-in list of popular MCP servers.
var Registry = []Server{
	{Name: "filesystem", Display: "Filesystem", Description: "Read/write local files", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-filesystem"}, Install: "npm install -g @modelcontextprotocol/server-filesystem", Backend: "npm", Category: "Core"},
	{Name: "postgres", Display: "PostgreSQL", Description: "Query PostgreSQL databases", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-postgres"}, Install: "npm install -g @modelcontextprotocol/server-postgres", Backend: "npm", Category: "Database"},
	{Name: "sqlite", Display: "SQLite", Description: "Query SQLite databases", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-sqlite"}, Install: "npm install -g @modelcontextprotocol/server-sqlite", Backend: "npm", Category: "Database"},
	{Name: "github", Display: "GitHub", Description: "GitHub repos, issues, PRs", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-github"}, Install: "npm install -g @modelcontextprotocol/server-github", Backend: "npm", Category: "Dev"},
	{Name: "gitlab", Display: "GitLab", Description: "GitLab repos and pipelines", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-gitlab"}, Install: "npm install -g @modelcontextprotocol/server-gitlab", Backend: "npm", Category: "Dev"},
	{Name: "slack", Display: "Slack", Description: "Read/send Slack messages", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-slack"}, Install: "npm install -g @modelcontextprotocol/server-slack", Backend: "npm", Category: "Communication"},
	{Name: "memory", Display: "Memory", Description: "Persistent knowledge graph", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-memory"}, Install: "npm install -g @modelcontextprotocol/server-memory", Backend: "npm", Category: "Core"},
	{Name: "brave-search", Display: "Brave Search", Description: "Web search via Brave", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-brave-search"}, Install: "npm install -g @modelcontextprotocol/server-brave-search", Backend: "npm", Category: "Search"},
	{Name: "puppeteer", Display: "Puppeteer", Description: "Browser automation", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-puppeteer"}, Install: "npm install -g @modelcontextprotocol/server-puppeteer", Backend: "npm", Category: "Browser"},
	{Name: "playwright", Display: "Playwright", Description: "Browser automation", Command: "npx", Args: []string{"-y", "@anthropic/mcp-server-playwright"}, Install: "npm install -g @anthropic/mcp-server-playwright", Backend: "npm", Category: "Browser"},
	{Name: "fetch", Display: "Fetch", Description: "HTTP fetching", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-fetch"}, Install: "npm install -g @modelcontextprotocol/server-fetch", Backend: "npm", Category: "Core"},
	{Name: "sentry", Display: "Sentry", Description: "Error tracking and monitoring", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-sentry"}, Install: "npm install -g @modelcontextprotocol/server-sentry", Backend: "npm", Category: "Monitoring"},
	{Name: "sequential-thinking", Display: "Sequential Thinking", Description: "Step-by-step reasoning", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-sequential-thinking"}, Install: "npm install -g @modelcontextprotocol/server-sequential-thinking", Backend: "npm", Category: "Reasoning"},
	{Name: "context7", Display: "Context7", Description: "Up-to-date library documentation", Command: "npx", Args: []string{"-y", "@upstash/context7-mcp"}, Install: "npm install -g @upstash/context7-mcp", Backend: "npm", Category: "Docs"},
	{Name: "redis", Display: "Redis", Description: "Redis database operations", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-redis"}, Install: "npm install -g @modelcontextprotocol/server-redis", Backend: "npm", Category: "Database"},
	{Name: "docker", Display: "Docker", Description: "Docker container management", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-docker"}, Install: "npm install -g @modelcontextprotocol/server-docker", Backend: "npm", Category: "Infra"},
	{Name: "kubernetes", Display: "Kubernetes", Description: "K8s cluster management", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-kubernetes"}, Install: "npm install -g @modelcontextprotocol/server-kubernetes", Backend: "npm", Category: "Infra"},
	{Name: "google-maps", Display: "Google Maps", Description: "Maps and geocoding", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-google-maps"}, Install: "npm install -g @modelcontextprotocol/server-google-maps", Backend: "npm", Category: "API"},
	{Name: "stripe", Display: "Stripe", Description: "Stripe payments API", Command: "npx", Args: []string{"-y", "@stripe/mcp"}, Install: "npm install -g @stripe/mcp", Backend: "npm", Category: "API"},
	{Name: "firebase", Display: "Firebase", Description: "Firebase services", Command: "npx", Args: []string{"-y", "firebase-mcp"}, Install: "npm install -g firebase-mcp", Backend: "npm", Category: "Cloud"},
}

// GetServer returns a server by name.
func GetServer(name string) *Server {
	for i := range Registry {
		if Registry[i].Name == name {
			return &Registry[i]
		}
	}
	return nil
}

// Search finds servers matching a query.
func Search(query string) []Server {
	q := strings.ToLower(query)
	var results []Server
	for _, s := range Registry {
		if strings.Contains(strings.ToLower(s.Name), q) ||
			strings.Contains(strings.ToLower(s.Description), q) ||
			strings.Contains(strings.ToLower(s.Category), q) {
			results = append(results, s)
		}
	}
	return results
}

// Categories returns sorted unique categories.
func Categories() []string {
	seen := make(map[string]bool)
	for _, s := range Registry {
		seen[s.Category] = true
	}
	cats := make([]string, 0, len(seen))
	for c := range seen {
		cats = append(cats, c)
	}
	sort.Strings(cats)
	return cats
}

// ToolConfigs returns the config paths for each AI tool that supports MCP.
func ToolConfigs() []ToolConfig {
	home, _ := os.UserHomeDir()
	configs := []ToolConfig{
		{Name: "claude-code", Path: filepath.Join(home, ".claude", "settings.json"), Format: "json-servers"},
		{Name: "cursor", Path: filepath.Join(home, ".cursor", "mcp.json"), Format: "json-servers"},
	}

	// VS Code settings
	var vscPath string
	switch runtime.GOOS {
	case "darwin":
		vscPath = filepath.Join(home, "Library", "Application Support", "Code", "User", "settings.json")
	case "linux":
		vscPath = filepath.Join(home, ".config", "Code", "User", "settings.json")
	default:
		vscPath = filepath.Join(home, "AppData", "Roaming", "Code", "User", "settings.json")
	}
	configs = append(configs, ToolConfig{Name: "vscode", Path: vscPath, Format: "json-mcp"})

	return configs
}

// ReadClaudeConfig reads existing MCP servers from Claude Code settings.
func ReadClaudeConfig() (map[string]interface{}, error) {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".claude", "settings.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, err
	}
	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return config, nil
}

// Install installs an MCP server package.
func Install(s *Server) error {
	parts := strings.Fields(s.Install)
	if len(parts) == 0 {
		return fmt.Errorf("no install command for %s", s.Name)
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ListInstalled checks which registry servers appear in Claude Code config.
func ListInstalled() []string {
	config, err := ReadClaudeConfig()
	if err != nil {
		return nil
	}
	servers, ok := config["mcpServers"]
	if !ok {
		return nil
	}
	serverMap, ok := servers.(map[string]interface{})
	if !ok {
		return nil
	}
	names := make([]string, 0, len(serverMap))
	for name := range serverMap {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

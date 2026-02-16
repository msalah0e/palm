package registry

import (
	"strings"
)

// Registry holds all known AI tools.
type Registry struct {
	tools  []Tool
	byName map[string]*Tool
}

// New creates a registry from a list of tools.
func New(tools []Tool) *Registry {
	r := &Registry{
		tools:  tools,
		byName: make(map[string]*Tool, len(tools)),
	}
	for i := range r.tools {
		r.byName[r.tools[i].Name] = &r.tools[i]
	}
	return r
}

// All returns all tools in the registry.
func (r *Registry) All() []Tool {
	return r.tools
}

// Get returns a tool by name, or nil if not found.
func (r *Registry) Get(name string) *Tool {
	return r.byName[name]
}

// Search finds tools matching a query against name, description, category, and tags.
func (r *Registry) Search(query string) []Tool {
	q := strings.ToLower(query)
	var results []Tool
	for _, t := range r.tools {
		if matches(t, q) {
			results = append(results, t)
		}
	}
	return results
}

// ByCategory returns tools filtered by category.
func (r *Registry) ByCategory(category string) []Tool {
	var results []Tool
	for _, t := range r.tools {
		if t.Category == category {
			results = append(results, t)
		}
	}
	return results
}

// Categories returns all unique categories.
func (r *Registry) Categories() []string {
	seen := make(map[string]bool)
	var cats []string
	for _, t := range r.tools {
		if !seen[t.Category] {
			seen[t.Category] = true
			cats = append(cats, t.Category)
		}
	}
	return cats
}

func matches(t Tool, query string) bool {
	if strings.Contains(strings.ToLower(t.Name), query) {
		return true
	}
	if strings.Contains(strings.ToLower(t.DisplayName), query) {
		return true
	}
	if strings.Contains(strings.ToLower(t.Description), query) {
		return true
	}
	if strings.ToLower(t.Category) == query {
		return true
	}
	for _, tag := range t.Tags {
		if strings.ToLower(tag) == query {
			return true
		}
	}
	return false
}

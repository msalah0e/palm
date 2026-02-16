package tokens

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Model context windows (approximate token limits).
var ModelContextWindows = map[string]int{
	"gpt-4o":              128000,
	"gpt-4o-mini":         128000,
	"gpt-4-turbo":         128000,
	"gpt-4":               8192,
	"gpt-3.5-turbo":       16385,
	"o1":                  200000,
	"o3":                  200000,
	"claude-opus-4":       200000,
	"claude-sonnet-4":     200000,
	"claude-sonnet-4-5":   200000,
	"claude-haiku-3-5":    200000,
	"claude-3-opus":       200000,
	"gemini-2.5-pro":      1000000,
	"gemini-2.5-flash":    1000000,
	"gemini-2.0-flash":    1000000,
	"llama-3.3":           128000,
	"mistral-large":       128000,
	"deepseek-v3":         128000,
	"codestral":           256000,
}

// FileResult holds token count info for a single file.
type FileResult struct {
	Path   string
	Tokens int
	Lines  int
	Bytes  int
}

// ScanResult holds results for a directory scan.
type ScanResult struct {
	Files      []FileResult
	Total      int
	TotalBytes int
	TotalLines int
}

// EstimateTokens estimates token count from byte length.
// Average ~4 characters per token for English/code (tiktoken approximation).
func EstimateTokens(content []byte) int {
	return (len(content) + 3) / 4
}

// CountFile counts tokens for a single file.
func CountFile(path string) (FileResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FileResult{}, err
	}
	lines := 1
	for _, b := range data {
		if b == '\n' {
			lines++
		}
	}
	return FileResult{
		Path:   path,
		Tokens: EstimateTokens(data),
		Lines:  lines,
		Bytes:  len(data),
	}, nil
}

// defaultIgnore lists directories to skip.
var defaultIgnore = map[string]bool{
	".git": true, "node_modules": true, "__pycache__": true,
	".venv": true, "venv": true, ".tox": true, "dist": true,
	"build": true, ".next": true, ".nuxt": true, "target": true,
	"vendor": true, ".idea": true, ".vscode": true,
}

// codeExtensions lists file extensions to count.
var codeExtensions = map[string]bool{
	".go": true, ".py": true, ".js": true, ".ts": true, ".tsx": true,
	".jsx": true, ".rs": true, ".java": true, ".c": true, ".cpp": true,
	".h": true, ".hpp": true, ".cs": true, ".rb": true, ".php": true,
	".swift": true, ".kt": true, ".scala": true, ".ex": true, ".exs": true,
	".md": true, ".txt": true, ".yaml": true, ".yml": true, ".toml": true,
	".json": true, ".xml": true, ".html": true, ".css": true, ".scss": true,
	".sql": true, ".sh": true, ".bash": true, ".zsh": true, ".fish": true,
	".dockerfile": true, ".tf": true, ".hcl": true, ".proto": true,
	".vue": true, ".svelte": true, ".astro": true,
}

// ScanDir counts tokens for all code files in a directory.
func ScanDir(root string) (*ScanResult, error) {
	result := &ScanResult{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if info.IsDir() {
			if defaultIgnore[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		// Also match Dockerfile, Makefile etc
		base := strings.ToLower(info.Name())
		if !codeExtensions[ext] && base != "dockerfile" && base != "makefile" && base != "cmakelists.txt" {
			return nil
		}
		// Skip large files (>1MB)
		if info.Size() > 1024*1024 {
			return nil
		}

		fr, err := CountFile(path)
		if err != nil {
			return nil
		}
		fr.Path, _ = filepath.Rel(root, path)
		result.Files = append(result.Files, fr)
		result.Total += fr.Tokens
		result.TotalBytes += fr.Bytes
		result.TotalLines += fr.Lines
		return nil
	})

	sort.Slice(result.Files, func(i, j int) bool {
		return result.Files[i].Tokens > result.Files[j].Tokens
	})

	return result, err
}

// ContextBudget shows how a token count fits within model context windows.
type ContextBudget struct {
	Model    string
	Window   int
	Used     int
	Percent  float64
	Fits     bool
}

// Budget calculates context budget for all known models.
func Budget(totalTokens int) []ContextBudget {
	var budgets []ContextBudget
	for model, window := range ModelContextWindows {
		pct := float64(totalTokens) / float64(window) * 100
		budgets = append(budgets, ContextBudget{
			Model:   model,
			Window:  window,
			Used:    totalTokens,
			Percent: pct,
			Fits:    totalTokens <= window,
		})
	}
	sort.Slice(budgets, func(i, j int) bool {
		return budgets[i].Window < budgets[j].Window
	})
	return budgets
}

// FormatTokens returns a human-readable token count.
func FormatTokens(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

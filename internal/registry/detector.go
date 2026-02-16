package registry

import (
	"os"
	"os/exec"
	"strings"
)

// DetectedTool holds detection results for a single tool.
type DetectedTool struct {
	Tool        Tool
	Installed   bool
	Version     string
	Path        string
	KeysSet     []string
	KeysMissing []string
}

// Detect scans the system for installed AI tools from the registry.
func Detect(reg *Registry) []DetectedTool {
	var results []DetectedTool
	for _, tool := range reg.All() {
		dt := detectOne(tool)
		results = append(results, dt)
	}
	return results
}

// DetectInstalled returns only tools that are installed.
func DetectInstalled(reg *Registry) []DetectedTool {
	var results []DetectedTool
	for _, tool := range reg.All() {
		dt := detectOne(tool)
		if dt.Installed {
			results = append(results, dt)
		}
	}
	return results
}

func detectOne(tool Tool) DetectedTool {
	dt := DetectedTool{Tool: tool}

	if tool.Install.Verify.Command == "" {
		return dt
	}

	// Run the full verify command via shell to handle pipes, subshells, etc.
	cmd := exec.Command("sh", "-c", tool.Install.Verify.Command)
	out, err := cmd.Output()
	if err != nil {
		// Command failed → tool not installed
		return dt
	}

	dt.Installed = true
	dt.Version = ExtractVersion(strings.TrimSpace(string(out)))

	// Try to find the binary path (use the first word of the verify command)
	parts := strings.Fields(tool.Install.Verify.Command)
	if len(parts) > 0 {
		if path, err := exec.LookPath(parts[0]); err == nil {
			dt.Path = path
		}
	}

	// Check API keys
	for _, key := range tool.Keys.Required {
		if os.Getenv(key) != "" {
			dt.KeysSet = append(dt.KeysSet, key)
		} else {
			dt.KeysMissing = append(dt.KeysMissing, key)
		}
	}

	return dt
}

// ExtractVersion tries to pull a version number from command output.
func ExtractVersion(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Look for a field that looks like a version number (starts with digit)
		fields := strings.Fields(line)
		for _, f := range fields {
			if len(f) > 0 && f[0] >= '0' && f[0] <= '9' {
				// Strip trailing parenthetical like "(Claude Code)"
				return f
			}
		}
		// Fallback: return last field
		if len(fields) == 1 {
			return fields[0]
		}
		return fields[len(fields)-1]
	}
	return output
}

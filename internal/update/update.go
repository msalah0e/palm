package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/msalah0e/palm/internal/ui"
)

const (
	repo        = "msalah0e/palm"
	checkEvery  = 24 * time.Hour
)

type releaseInfo struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

type checkCache struct {
	LastCheck time.Time `json:"last_check"`
	Latest    string    `json:"latest"`
}

func cachePath() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "palm", "update-check.json")
}

// CheckForUpdate checks GitHub for a newer version and prints a message if available.
// Only checks once per day to avoid slowing down every invocation.
func CheckForUpdate(currentVersion string) {
	path := cachePath()

	// Check cache first
	var cache checkCache
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &cache)
		if time.Since(cache.LastCheck) < checkEvery {
			if cache.Latest != "" && cache.Latest != currentVersion && cache.Latest != "v"+currentVersion {
				printUpdateMessage(currentVersion, cache.Latest)
			}
			return
		}
	}

	// Fetch latest release in background-safe way
	go func() {
		latest, err := fetchLatest()
		if err != nil {
			return
		}
		cache := checkCache{LastCheck: time.Now(), Latest: latest}
		data, _ := json.Marshal(cache)
		_ = os.MkdirAll(filepath.Dir(path), 0o755)
		_ = os.WriteFile(path, data, 0o644)
	}()
}

// CheckNow forces an immediate version check and prints the result.
func CheckNow(currentVersion string) {
	latest, err := fetchLatest()
	if err != nil {
		ui.Subtle.Printf("  Could not check for updates: %v\n", err)
		return
	}

	// Update cache
	cache := checkCache{LastCheck: time.Now(), Latest: latest}
	data, _ := json.Marshal(cache)
	_ = os.MkdirAll(filepath.Dir(cachePath()), 0o755)
	_ = os.WriteFile(cachePath(), data, 0o644)

	clean := latest
	if len(clean) > 0 && clean[0] == 'v' {
		clean = clean[1:]
	}
	if clean == currentVersion {
		ui.Good.Printf("  %s palm is up to date (%s)\n", ui.StatusIcon(true), currentVersion)
	} else {
		printUpdateMessage(currentVersion, latest)
	}
}

func fetchLatest() (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release releaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	return release.TagName, nil
}

func printUpdateMessage(current, latest string) {
	fmt.Println()
	ui.Warn.Printf("  Update available: %s → %s\n", current, latest)
	fmt.Printf("  Run: go install github.com/%s@latest\n", repo)
	fmt.Printf("  Or:  curl -fsSL https://%s.github.io/palm/install.sh | sh\n", "msalah0e")
}

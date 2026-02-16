package cache

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Dir returns the cache directory path.
func Dir() string {
	dir := os.Getenv("XDG_CACHE_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".cache")
	}
	return filepath.Join(dir, "palm")
}

// Fetch pre-downloads a package to the local cache.
func Fetch(backend, pkg string) error {
	dir := filepath.Join(Dir(), backend)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	switch backend {
	case "pip":
		return fetchPip(pkg, dir)
	case "npm":
		return fetchNpm(pkg, dir)
	case "docker":
		return fetchDocker(pkg, dir)
	case "brew":
		return fetchBrew(pkg, dir)
	default:
		// For go, cargo, script, binary — just record as fetchable
		marker := filepath.Join(dir, sanitize(pkg)+".fetch")
		return os.WriteFile(marker, []byte(pkg), 0o644)
	}
}

// IsCached returns true if a package exists in cache.
func IsCached(backend, pkg string) bool {
	dir := filepath.Join(Dir(), backend)
	switch backend {
	case "pip":
		matches, _ := filepath.Glob(filepath.Join(dir, sanitize(pkg)+"*"))
		return len(matches) > 0
	case "npm":
		return fileExists(filepath.Join(dir, sanitize(pkg)+".tgz"))
	case "docker":
		return fileExists(filepath.Join(dir, sanitize(pkg)+".tar"))
	default:
		return fileExists(filepath.Join(dir, sanitize(pkg)+".fetch"))
	}
}

// Bundle creates a tar.gz archive of the entire cache directory.
func Bundle(output string) error {
	cacheDir := Dir()
	if _, err := os.Stat(cacheDir); err != nil {
		return fmt.Errorf("cache is empty — run `palm fetch` first")
	}

	f, err := os.Create(output)
	if err != nil {
		return err
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	return filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(cacheDir, path)
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = rel
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(tw, file)
		return err
	})
}

func fetchPip(pkg, dir string) error {
	return runCmd("pip3", "download", "-d", dir, pkg)
}

func fetchNpm(pkg, dir string) error {
	return runCmd("npm", "pack", pkg, "--pack-destination", dir)
}

func fetchDocker(image, dir string) error {
	out := filepath.Join(dir, sanitize(image)+".tar")
	// Pull first, then save
	if err := runCmd("docker", "pull", image); err != nil {
		return err
	}
	return runCmd("docker", "save", "-o", out, image)
}

func fetchBrew(pkg, dir string) error {
	return runCmd("brew", "fetch", pkg, "--retry")
}

func sanitize(s string) string {
	r := strings.NewReplacer("/", "_", ":", "_", "@", "_")
	return r.Replace(s)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

package installer

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/msalah0e/tamr/internal/registry"
	"github.com/msalah0e/tamr/internal/ui"
)

// Install installs a tool using the best available backend.
func Install(tool registry.Tool) error {
	backend, pkg := tool.InstallMethod()
	if backend == "manual" {
		return fmt.Errorf("no automated install method — visit %s", pkg)
	}

	fmt.Printf("  Installing %s via %s (%s)...\n", ui.Brand.Sprint(tool.DisplayName), backend, pkg)

	switch backend {
	case "brew":
		return brewInstall(pkg)
	case "pip":
		return pipInstall(pkg)
	case "npm":
		return npmInstall(pkg)
	case "cargo":
		return cargoInstall(pkg)
	case "go":
		return goInstall(pkg)
	case "script":
		return scriptInstall(pkg)
	case "binary":
		return fmt.Errorf("binary install not yet supported — download from %s", pkg)
	default:
		return fmt.Errorf("unknown backend: %s", backend)
	}
}

// Update updates a tool by re-running its install with upgrade flags.
func Update(tool registry.Tool) error {
	backend, pkg := tool.InstallMethod()

	switch backend {
	case "brew":
		return runCmd("brew", "upgrade", pkg)
	case "pip":
		return pipUpdate(pkg)
	case "npm":
		return runCmd("npm", "update", "-g", pkg)
	case "cargo":
		return cargoInstall(pkg) // cargo install re-compiles latest
	case "go":
		return goInstall(pkg) // go install @latest gets latest
	case "script":
		return scriptInstall(pkg) // re-run the script
	default:
		return fmt.Errorf("cannot auto-update %s tools", backend)
	}
}

// Uninstall removes a tool using its install backend.
func Uninstall(tool registry.Tool) error {
	backend, pkg := tool.InstallMethod()

	fmt.Printf("  Removing %s via %s...\n", ui.Brand.Sprint(tool.DisplayName), backend)

	switch backend {
	case "brew":
		return runCmd("brew", "uninstall", pkg)
	case "pip":
		return pipUninstall(pkg)
	case "npm":
		return runCmd("npm", "uninstall", "-g", pkg)
	default:
		return fmt.Errorf("cannot auto-uninstall %s tools", backend)
	}
}

func brewInstall(pkg string) error {
	return runCmd("brew", "install", pkg)
}

func pipInstall(pkg string) error {
	if hasCommand("uv") {
		return runCmd("uv", "pip", "install", pkg)
	}
	if hasCommand("pip3") {
		return runCmd("pip3", "install", pkg)
	}
	if hasCommand("pip") {
		return runCmd("pip", "install", pkg)
	}
	return fmt.Errorf("no pip/uv found — install Python first")
}

func pipUpdate(pkg string) error {
	if hasCommand("uv") {
		return runCmd("uv", "pip", "install", "--upgrade", pkg)
	}
	if hasCommand("pip3") {
		return runCmd("pip3", "install", "--upgrade", pkg)
	}
	return fmt.Errorf("no pip/uv found — install Python first")
}

func pipUninstall(pkg string) error {
	if hasCommand("uv") {
		return runCmd("uv", "pip", "uninstall", pkg)
	}
	return runCmd("pip3", "uninstall", "-y", pkg)
}

func npmInstall(pkg string) error {
	if !hasCommand("npm") {
		return fmt.Errorf("npm not found — install Node.js first")
	}
	return runCmd("npm", "install", "-g", pkg)
}

func cargoInstall(pkg string) error {
	if !hasCommand("cargo") {
		return fmt.Errorf("cargo not found — install Rust first")
	}
	return runCmd("cargo", "install", pkg)
}

func goInstall(pkg string) error {
	if !hasCommand("go") {
		return fmt.Errorf("go not found — install Go first")
	}
	return runCmd("go", "install", pkg)
}

func scriptInstall(url string) error {
	if !hasCommand("curl") {
		return fmt.Errorf("curl not found")
	}
	return runCmd("sh", "-c", fmt.Sprintf("curl -fsSL %s | sh", url))
}

func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

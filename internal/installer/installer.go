package installer

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/msalah0e/palm/internal/registry"
	"github.com/msalah0e/palm/internal/ui"
)

// Install installs a tool using the best available backend.
func Install(tool registry.Tool) error {
	backend, pkg := tool.InstallMethod()
	if backend == "manual" {
		return fmt.Errorf("no automated install method — visit %s", pkg)
	}

	fmt.Printf("  Installing %s via %s (%s)...\n", ui.Brand.Sprint(tool.DisplayName), backend, pkg)

	switch backend {
	case "linux":
		return linuxInstall(pkg)
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
	case "docker":
		return dockerPull(pkg)
	case "script":
		return scriptInstall(pkg)
	case "binary":
		return fmt.Errorf("binary install not yet supported — download from %s", pkg)
	default:
		return fmt.Errorf("unknown backend: %s", backend)
	}
}

// InstallQuiet installs a tool, capturing all output instead of printing it.
// Returns the captured output (useful for showing on failure) and any error.
// Used during parallel installs to prevent interleaved terminal output.
func InstallQuiet(tool registry.Tool) (string, error) {
	backend, pkg := tool.InstallMethod()
	if backend == "manual" {
		return "", fmt.Errorf("no automated install method — visit %s", pkg)
	}

	switch backend {
	case "linux":
		return linuxInstallQuiet(pkg)
	case "brew":
		return runCmdQuiet("brew", "install", pkg)
	case "pip":
		return pipInstallQuiet(pkg)
	case "npm":
		return npmInstallQuiet(pkg)
	case "cargo":
		return cargoInstallQuiet(pkg)
	case "go":
		return goInstallQuiet(pkg)
	case "docker":
		return runCmdQuiet("docker", "pull", pkg)
	case "script":
		return scriptInstallQuiet(pkg)
	case "binary":
		return "", fmt.Errorf("binary install not yet supported — download from %s", pkg)
	default:
		return "", fmt.Errorf("unknown backend: %s", backend)
	}
}

// Update updates a tool by re-running its install with upgrade flags.
func Update(tool registry.Tool) error {
	backend, pkg := tool.InstallMethod()

	switch backend {
	case "linux":
		return linuxUpdate(pkg)
	case "brew":
		return runCmd("brew", "upgrade", pkg)
	case "pip":
		return pipUpdate(pkg)
	case "npm":
		return runCmd("npm", "update", "-g", pkg)
	case "cargo":
		return cargoInstall(pkg)
	case "go":
		return goInstall(pkg)
	case "docker":
		return dockerPull(pkg)
	case "script":
		return scriptInstall(pkg)
	default:
		return fmt.Errorf("cannot auto-update %s tools", backend)
	}
}

// Uninstall removes a tool using its install backend.
func Uninstall(tool registry.Tool) error {
	backend, pkg := tool.InstallMethod()

	fmt.Printf("  Removing %s via %s...\n", ui.Brand.Sprint(tool.DisplayName), backend)

	switch backend {
	case "linux":
		return linuxUninstall(pkg)
	case "brew":
		return runCmd("brew", "uninstall", pkg)
	case "pip":
		return pipUninstall(pkg)
	case "npm":
		return runCmd("npm", "uninstall", "-g", pkg)
	case "docker":
		return runCmd("docker", "rmi", pkg)
	default:
		return fmt.Errorf("cannot auto-uninstall %s tools", backend)
	}
}

func brewInstall(pkg string) error {
	return runCmd("brew", "install", pkg)
}

func pipInstall(pkg string) error {
	if hasCommand("uv") {
		return runCmd("uv", "tool", "install", pkg)
	}
	if hasCommand("pipx") {
		return runCmd("pipx", "install", pkg)
	}
	if hasCommand("pip3") {
		return runCmd("pip3", "install", pkg)
	}
	if hasCommand("pip") {
		return runCmd("pip", "install", pkg)
	}
	return fmt.Errorf("no pip/uv/pipx found — install Python first")
}

func pipUpdate(pkg string) error {
	if hasCommand("uv") {
		return runCmd("uv", "tool", "upgrade", pkg)
	}
	if hasCommand("pipx") {
		return runCmd("pipx", "upgrade", pkg)
	}
	if hasCommand("pip3") {
		return runCmd("pip3", "install", "--upgrade", pkg)
	}
	return fmt.Errorf("no pip/uv/pipx found — install Python first")
}

func pipUninstall(pkg string) error {
	if hasCommand("uv") {
		return runCmd("uv", "tool", "uninstall", pkg)
	}
	if hasCommand("pipx") {
		return runCmd("pipx", "uninstall", pkg)
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

func dockerPull(image string) error {
	if !hasCommand("docker") {
		return fmt.Errorf("docker not found — install Docker first")
	}
	return runCmd("docker", "pull", image)
}

func scriptInstall(script string) error {
	if strings.HasPrefix(script, "http://") || strings.HasPrefix(script, "https://") {
		if !hasCommand("curl") {
			return fmt.Errorf("curl not found")
		}
		return runCmd("sh", "-c", fmt.Sprintf("curl -fsSL %s | sh", script))
	}
	return runCmd("sh", "-c", script)
}

func detectLinuxPM() (string, error) {
	if hasCommand("apt-get") {
		return "apt-get", nil
	}
	if hasCommand("dnf") {
		return "dnf", nil
	}
	if hasCommand("pacman") {
		return "pacman", nil
	}
	return "", fmt.Errorf("no supported package manager found (need apt-get, dnf, or pacman)")
}

func linuxInstall(pkg string) error {
	pm, err := detectLinuxPM()
	if err != nil {
		return err
	}
	switch pm {
	case "apt-get":
		return runCmd("sudo", "apt-get", "install", "-y", pkg)
	case "dnf":
		return runCmd("sudo", "dnf", "install", "-y", pkg)
	case "pacman":
		return runCmd("sudo", "pacman", "-S", "--noconfirm", pkg)
	}
	return fmt.Errorf("unsupported package manager: %s", pm)
}

func linuxUpdate(pkg string) error {
	pm, err := detectLinuxPM()
	if err != nil {
		return err
	}
	switch pm {
	case "apt-get":
		return runCmd("sudo", "apt-get", "upgrade", "-y", pkg)
	case "dnf":
		return runCmd("sudo", "dnf", "upgrade", "-y", pkg)
	case "pacman":
		return runCmd("sudo", "pacman", "-Syu", "--noconfirm")
	}
	return fmt.Errorf("unsupported package manager: %s", pm)
}

func linuxUninstall(pkg string) error {
	pm, err := detectLinuxPM()
	if err != nil {
		return err
	}
	switch pm {
	case "apt-get":
		return runCmd("sudo", "apt-get", "remove", "-y", pkg)
	case "dnf":
		return runCmd("sudo", "dnf", "remove", "-y", pkg)
	case "pacman":
		return runCmd("sudo", "pacman", "-R", "--noconfirm", pkg)
	}
	return fmt.Errorf("unsupported package manager: %s", pm)
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

// runCmdQuiet runs a command and captures all output instead of printing it.
func runCmdQuiet(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func pipInstallQuiet(pkg string) (string, error) {
	if hasCommand("uv") {
		return runCmdQuiet("uv", "tool", "install", pkg)
	}
	if hasCommand("pipx") {
		return runCmdQuiet("pipx", "install", pkg)
	}
	if hasCommand("pip3") {
		return runCmdQuiet("pip3", "install", pkg)
	}
	if hasCommand("pip") {
		return runCmdQuiet("pip", "install", pkg)
	}
	return "", fmt.Errorf("no pip/uv/pipx found — install Python first")
}

func npmInstallQuiet(pkg string) (string, error) {
	if !hasCommand("npm") {
		return "", fmt.Errorf("npm not found — install Node.js first")
	}
	return runCmdQuiet("npm", "install", "-g", pkg)
}

func cargoInstallQuiet(pkg string) (string, error) {
	if !hasCommand("cargo") {
		return "", fmt.Errorf("cargo not found — install Rust first")
	}
	return runCmdQuiet("cargo", "install", pkg)
}

func goInstallQuiet(pkg string) (string, error) {
	if !hasCommand("go") {
		return "", fmt.Errorf("go not found — install Go first")
	}
	return runCmdQuiet("go", "install", pkg)
}

func scriptInstallQuiet(script string) (string, error) {
	if strings.HasPrefix(script, "http://") || strings.HasPrefix(script, "https://") {
		if !hasCommand("curl") {
			return "", fmt.Errorf("curl not found")
		}
		return runCmdQuiet("sh", "-c", fmt.Sprintf("curl -fsSL %s | sh", script))
	}
	return runCmdQuiet("sh", "-c", script)
}

func linuxInstallQuiet(pkg string) (string, error) {
	pm, err := detectLinuxPM()
	if err != nil {
		return "", err
	}
	switch pm {
	case "apt-get":
		return runCmdQuiet("sudo", "apt-get", "install", "-y", pkg)
	case "dnf":
		return runCmdQuiet("sudo", "dnf", "install", "-y", pkg)
	case "pacman":
		return runCmdQuiet("sudo", "pacman", "-S", "--noconfirm", pkg)
	}
	return "", fmt.Errorf("unsupported package manager: %s", pm)
}

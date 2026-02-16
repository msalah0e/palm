package brew

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// Path returns the path to the brew executable, or exits if not found.
func Path() string {
	path, err := exec.LookPath("brew")
	if err != nil {
		fmt.Fprintln(os.Stderr, "tamr: homebrew not found. Install it first: https://brew.sh")
		os.Exit(1)
	}
	return path
}

// Passthrough replaces the current process with brew, forwarding all args.
func Passthrough(args []string) {
	brew := Path()
	env := os.Environ()
	err := syscall.Exec(brew, append([]string{"brew"}, args...), env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tamr: failed to exec brew: %v\n", err)
		os.Exit(1)
	}
}

// Run executes brew with the given args and returns captured output.
func Run(args ...string) (string, error) {
	brew := Path()
	cmd := exec.Command(brew, args...)
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// Rebrand replaces "brew"/"Homebrew" with "tamr"/"Tamr" in output.
func Rebrand(s string) string {
	s = strings.ReplaceAll(s, "Homebrew", "Tamr")
	s = strings.ReplaceAll(s, "homebrew", "tamr")
	s = strings.ReplaceAll(s, "brew ", "tamr ")
	s = strings.ReplaceAll(s, "brew\n", "tamr\n")
	return s
}

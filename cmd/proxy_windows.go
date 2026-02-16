//go:build windows

package cmd

import (
	"os"
	"os/exec"
)

func setDetached(cmd *exec.Cmd) {
	// Windows: no Setsid equivalent needed for background processes
}

func stopProcess(proc *os.Process) error {
	return proc.Kill()
}

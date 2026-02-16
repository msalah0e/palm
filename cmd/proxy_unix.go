//go:build !windows

package cmd

import (
	"os"
	"os/exec"
	"syscall"
)

func setDetached(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}

func stopProcess(proc *os.Process) error {
	return proc.Signal(syscall.SIGTERM)
}

package hooks

import (
	"os"
	"os/exec"

	"github.com/msalah0e/palm/internal/config"
)

// Run executes the hook script for the given phase, if configured.
func Run(phase, toolName, category string) error {
	cfg := config.Load()
	script := getHook(cfg.Hooks, phase)
	if script == "" {
		return nil
	}

	cmd := exec.Command("sh", "-c", script)
	cmd.Env = append(os.Environ(),
		"PALM_TOOL="+toolName,
		"PALM_PHASE="+phase,
		"PALM_CATEGORY="+category,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func getHook(h config.HooksConfig, phase string) string {
	switch phase {
	case "pre_install":
		return h.PreInstall
	case "post_install":
		return h.PostInstall
	case "pre_run":
		return h.PreRun
	case "post_run":
		return h.PostRun
	case "pre_update":
		return h.PreUpdate
	case "post_update":
		return h.PostUpdate
	default:
		return ""
	}
}

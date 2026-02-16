//go:build e2e

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var palmBin string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "palm-e2e-*")
	if err != nil {
		panic("failed to create temp dir: " + err.Error())
	}
	defer os.RemoveAll(tmp)

	palmBin = filepath.Join(tmp, "palm")
	build := exec.Command("go", "build", "-ldflags", "-X github.com/msalah0e/palm/cmd.version=1.5.0-test", "-o", palmBin, ".")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		panic("failed to build palm: " + err.Error())
	}

	os.Exit(m.Run())
}

// runPalm executes the palm binary with an isolated HOME directory.
func runPalm(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(palmBin, args...)
	home := t.TempDir()
	cmd.Env = append(os.Environ(),
		"HOME="+home,
		"XDG_CONFIG_HOME="+filepath.Join(home, ".config"),
		"NO_COLOR=1",
	)

	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run palm %v: %v", args, err)
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}

// --- Core CLI ---

func TestE2E_Version(t *testing.T) {
	out, _, code := runPalm(t, "--version")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out, "1.5.0") {
		t.Errorf("expected version output to contain '1.5.0', got %q", out)
	}
}

func TestE2E_Help(t *testing.T) {
	out, _, code := runPalm(t, "--help")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out, "Available Commands") {
		t.Errorf("expected help to contain 'Available Commands', got %q", out)
	}
}

func TestE2E_BareCommand(t *testing.T) {
	out, _, code := runPalm(t)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out, "PALM") && !strings.Contains(out, "palm") {
		t.Errorf("expected logo or name in output, got %q", out)
	}
}

// --- Tool management (no actual installs) ---

func TestE2E_List(t *testing.T) {
	_, _, code := runPalm(t, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

func TestE2E_Search(t *testing.T) {
	out, _, code := runPalm(t, "search", "coding")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if len(out) == 0 {
		t.Error("expected search output, got empty")
	}
}

func TestE2E_SearchBrowse(t *testing.T) {
	_, _, code := runPalm(t, "search")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

func TestE2E_Info(t *testing.T) {
	out, _, code := runPalm(t, "info", "claude-code")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out, "Claude Code") {
		t.Errorf("expected output to contain 'Claude Code', got %q", out)
	}
}

func TestE2E_InfoUnknown(t *testing.T) {
	_, _, code := runPalm(t, "info", "nonexistent-tool-xyz")
	if code == 0 {
		t.Fatal("expected non-zero exit for unknown tool")
	}
}

// --- Vault ---

func TestE2E_KeysListEmpty(t *testing.T) {
	_, _, code := runPalm(t, "keys", "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

func TestE2E_KeysExport(t *testing.T) {
	_, _, code := runPalm(t, "keys", "export")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

// --- Health ---

func TestE2E_Doctor(t *testing.T) {
	_, _, code := runPalm(t, "doctor")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

func TestE2E_Health(t *testing.T) {
	_, _, code := runPalm(t, "health")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

// --- Models ---

func TestE2E_ModelsList(t *testing.T) {
	_, _, code := runPalm(t, "models", "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

func TestE2E_ModelsProviders(t *testing.T) {
	_, _, code := runPalm(t, "models", "providers")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

// --- Budget ---

func TestE2E_BudgetStatus(t *testing.T) {
	_, _, code := runPalm(t, "budget", "status")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

// --- Stats ---

func TestE2E_Stats(t *testing.T) {
	_, _, code := runPalm(t, "stats")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

// --- GPU ---

func TestE2E_GPU(t *testing.T) {
	_, _, code := runPalm(t, "gpu")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

// --- Top ---

func TestE2E_TopHelp(t *testing.T) {
	out, _, code := runPalm(t, "top", "--help")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out, "interval") {
		t.Errorf("expected help to mention 'interval', got %q", out)
	}
}

// --- Workspace ---

func TestE2E_WorkspaceInit(t *testing.T) {
	_, _, code := runPalm(t, "workspace", "init")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

// --- Graph ---

func TestE2E_GraphAdd(t *testing.T) {
	_, _, code := runPalm(t, "graph", "add", "test-entity", "--type", "person")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

// --- Tokens ---

func TestE2E_TokensCount(t *testing.T) {
	// Count tokens in go.mod (a small file that always exists)
	_, _, code := runPalm(t, "tokens", "count", "go.mod")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

// --- Rules ---

func TestE2E_RulesInit(t *testing.T) {
	_, _, code := runPalm(t, "rules", "init")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

// --- MCP ---

func TestE2E_MCPList(t *testing.T) {
	_, _, code := runPalm(t, "mcp", "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

// --- Cost ---

func TestE2E_Cost(t *testing.T) {
	_, _, code := runPalm(t, "cost")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

// --- Prompt ---

func TestE2E_PromptList(t *testing.T) {
	_, _, code := runPalm(t, "prompt", "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

// --- Completion ---

func TestE2E_CompletionZsh(t *testing.T) {
	out, _, code := runPalm(t, "completion", "zsh")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if len(out) == 0 {
		t.Error("expected zsh completion output, got empty")
	}
}

// --- Matrix ---

func TestE2E_Matrix(t *testing.T) {
	_, _, code := runPalm(t, "matrix")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

// --- Pirate ---

func TestE2E_PirateStatus(t *testing.T) {
	_, _, code := runPalm(t, "pirate", "status")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

// --- Shield ---

func TestE2E_Shield(t *testing.T) {
	_, _, code := runPalm(t, "shield")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

// --- Actlog ---

func TestE2E_Actlog(t *testing.T) {
	_, _, code := runPalm(t, "log")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

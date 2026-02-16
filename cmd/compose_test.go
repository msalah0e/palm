package cmd

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestLoadComposeFile_Valid(t *testing.T) {
	dir := t.TempDir()
	content := `name = "test-workflow"
description = "A test workflow"

[[steps]]
name = "step1"
run = "echo hello"

[[steps]]
name = "step2"
run = "echo world"
depends_on = ["step1"]
`
	path := filepath.Join(dir, ".palm-compose.toml")
	os.WriteFile(path, []byte(content), 0644)

	cf, err := loadComposeFile(path)
	if err != nil {
		t.Fatalf("loadComposeFile failed: %v", err)
	}
	if cf.Name != "test-workflow" {
		t.Errorf("expected name 'test-workflow', got %q", cf.Name)
	}
	if cf.Description != "A test workflow" {
		t.Errorf("expected description 'A test workflow', got %q", cf.Description)
	}
	if len(cf.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(cf.Steps))
	}
}

func TestLoadComposeFile_MissingName(t *testing.T) {
	dir := t.TempDir()
	content := `[[steps]]
run = "echo hello"
`
	path := filepath.Join(dir, "workflow.toml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := loadComposeFile(path)
	if err == nil {
		t.Error("expected error for step missing name")
	}
}

func TestLoadComposeFile_NoRunOrTool(t *testing.T) {
	dir := t.TempDir()
	content := `[[steps]]
name = "empty-step"
`
	path := filepath.Join(dir, "workflow.toml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := loadComposeFile(path)
	if err == nil {
		t.Error("expected error for step with no run or tool")
	}
}

func TestLoadComposeFile_DuplicateNames(t *testing.T) {
	dir := t.TempDir()
	content := `[[steps]]
name = "same"
run = "echo 1"

[[steps]]
name = "same"
run = "echo 2"
`
	path := filepath.Join(dir, "workflow.toml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := loadComposeFile(path)
	if err == nil {
		t.Error("expected error for duplicate step names")
	}
}

func TestLoadComposeFile_UnknownDependency(t *testing.T) {
	dir := t.TempDir()
	content := `[[steps]]
name = "step1"
run = "echo hello"
depends_on = ["nonexistent"]
`
	path := filepath.Join(dir, "workflow.toml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := loadComposeFile(path)
	if err == nil {
		t.Error("expected error for unknown dependency")
	}
}

func TestLoadComposeFile_NotFound(t *testing.T) {
	_, err := loadComposeFile("/nonexistent/path/workflow.toml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestResolveExecutionOrder_Linear(t *testing.T) {
	wf := &ComposeFile{
		Steps: []ComposeStep{
			{Name: "a", Run: "echo a"},
			{Name: "b", Run: "echo b", DependsOn: []string{"a"}},
			{Name: "c", Run: "echo c", DependsOn: []string{"b"}},
		},
	}

	levels := resolveExecutionOrder(wf)

	if len(levels) != 3 {
		t.Fatalf("expected 3 levels for linear chain, got %d", len(levels))
	}
	if len(levels[0]) != 1 || levels[0][0].Name != "a" {
		t.Errorf("level 0 should be [a], got %v", stepNames(levels[0]))
	}
	if len(levels[1]) != 1 || levels[1][0].Name != "b" {
		t.Errorf("level 1 should be [b], got %v", stepNames(levels[1]))
	}
	if len(levels[2]) != 1 || levels[2][0].Name != "c" {
		t.Errorf("level 2 should be [c], got %v", stepNames(levels[2]))
	}
}

func TestResolveExecutionOrder_Parallel(t *testing.T) {
	wf := &ComposeFile{
		Steps: []ComposeStep{
			{Name: "a", Run: "echo a"},
			{Name: "b", Run: "echo b"},
			{Name: "c", Run: "echo c"},
		},
	}

	levels := resolveExecutionOrder(wf)

	if len(levels) != 1 {
		t.Fatalf("expected 1 level for independent steps, got %d", len(levels))
	}
	if len(levels[0]) != 3 {
		t.Errorf("expected 3 steps in parallel, got %d", len(levels[0]))
	}
}

func TestResolveExecutionOrder_Diamond(t *testing.T) {
	// A → B, A → C, B+C → D
	wf := &ComposeFile{
		Steps: []ComposeStep{
			{Name: "a", Run: "echo a"},
			{Name: "b", Run: "echo b", DependsOn: []string{"a"}},
			{Name: "c", Run: "echo c", DependsOn: []string{"a"}},
			{Name: "d", Run: "echo d", DependsOn: []string{"b", "c"}},
		},
	}

	levels := resolveExecutionOrder(wf)

	if len(levels) != 3 {
		t.Fatalf("expected 3 levels for diamond, got %d", len(levels))
	}
	if len(levels[0]) != 1 {
		t.Errorf("level 0 should have 1 step, got %d", len(levels[0]))
	}
	if len(levels[1]) != 2 {
		t.Errorf("level 1 should have 2 parallel steps, got %d", len(levels[1]))
	}
	if len(levels[2]) != 1 {
		t.Errorf("level 2 should have 1 step, got %d", len(levels[2]))
	}
}

func TestResolveExecutionOrder_Empty(t *testing.T) {
	wf := &ComposeFile{Steps: []ComposeStep{}}
	levels := resolveExecutionOrder(wf)
	if len(levels) != 0 {
		t.Errorf("expected 0 levels for empty workflow, got %d", len(levels))
	}
}

func TestResolveInput_Step(t *testing.T) {
	outputs := map[string]string{
		"step1": "hello from step1",
	}
	var mu sync.Mutex

	result := resolveInput("step:step1", outputs, &mu)
	if result != "hello from step1" {
		t.Errorf("expected 'hello from step1', got %q", result)
	}
}

func TestResolveInput_File(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "input.txt")
	os.WriteFile(path, []byte("file content"), 0644)

	outputs := make(map[string]string)
	var mu sync.Mutex

	result := resolveInput("file:"+path, outputs, &mu)
	if result != "file content" {
		t.Errorf("expected 'file content', got %q", result)
	}
}

func TestResolveInput_Literal(t *testing.T) {
	outputs := make(map[string]string)
	var mu sync.Mutex

	result := resolveInput("just some text", outputs, &mu)
	if result != "just some text" {
		t.Errorf("expected 'just some text', got %q", result)
	}
}

func TestResolveInput_MultipleSteps(t *testing.T) {
	outputs := map[string]string{
		"s1": "output1",
		"s2": "output2",
	}
	var mu sync.Mutex

	result := resolveInput("step:s1,step:s2", outputs, &mu)
	if result != "output1\n\noutput2" {
		t.Errorf("expected combined outputs, got %q", result)
	}
}

func TestResolveInput_MissingStep(t *testing.T) {
	outputs := make(map[string]string)
	var mu sync.Mutex

	result := resolveInput("step:nonexistent", outputs, &mu)
	if result != "" {
		t.Errorf("expected empty for missing step, got %q", result)
	}
}

func TestExecuteComposeStep_ShellCommand(t *testing.T) {
	step := ComposeStep{
		Name: "test",
		Run:  "echo hello world",
	}

	result := executeComposeStep(step, os.Environ(), "", false)

	if result.Error != "" {
		t.Errorf("expected no error, got %q", result.Error)
	}
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
	if result.Output != "hello world\n" {
		t.Errorf("expected 'hello world\\n', got %q", result.Output)
	}
	if result.Step != "test" {
		t.Errorf("expected step name 'test', got %q", result.Step)
	}
	if result.Duration <= 0 {
		t.Error("expected positive duration")
	}
}

func TestExecuteComposeStep_FailingCommand(t *testing.T) {
	step := ComposeStep{
		Name: "fail",
		Run:  "false",
	}

	result := executeComposeStep(step, os.Environ(), "", false)

	if result.Error == "" {
		t.Error("expected error for failing command")
	}
	if result.ExitCode != 1 {
		t.Errorf("expected exit code 1, got %d", result.ExitCode)
	}
}

func TestExecuteComposeStep_WithStdin(t *testing.T) {
	step := ComposeStep{
		Name: "stdin-test",
		Run:  "cat",
	}

	result := executeComposeStep(step, os.Environ(), "piped input", false)

	if result.Error != "" {
		t.Errorf("unexpected error: %q", result.Error)
	}
	if result.Output != "piped input" {
		t.Errorf("expected 'piped input', got %q", result.Output)
	}
}

func TestExecuteComposeStep_Timeout(t *testing.T) {
	step := ComposeStep{
		Name:    "slow",
		Run:     "sleep 10",
		Timeout: 1,
	}

	result := executeComposeStep(step, os.Environ(), "", false)

	if result.Error != "timeout" {
		t.Errorf("expected timeout error, got %q", result.Error)
	}
	if result.ExitCode != -1 {
		t.Errorf("expected exit code -1 for timeout, got %d", result.ExitCode)
	}
}

func TestComposeInit(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	composeInit()

	path := filepath.Join(dir, ".palm-compose.toml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("compose init should create .palm-compose.toml")
	}

	// Verify it's valid TOML by loading it
	cf, err := loadComposeFile(path)
	if err != nil {
		t.Fatalf("generated compose file should be valid: %v", err)
	}
	if cf.Name == "" {
		t.Error("generated compose file should have a name")
	}
	if len(cf.Steps) == 0 {
		t.Error("generated compose file should have steps")
	}
}

func stepNames(steps []ComposeStep) []string {
	names := make([]string, len(steps))
	for i, s := range steps {
		names[i] = s.Name
	}
	return names
}

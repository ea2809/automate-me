package app

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ea2809/automate-me/internal/core"
)

type fakeUI struct{}

func (f fakeUI) ClearScreen() {}

func (f fakeUI) SelectTask(tasks []core.TaskRecord, state SelectionState) (core.TaskRecord, SelectionState, error) {
	return core.TaskRecord{}, state, ErrUserCanceled
}

func (f fakeUI) PromptInputs(inputs []core.InputSpec, defaults map[string]any) (map[string]any, error) {
	if defaults != nil {
		return defaults, nil
	}
	return map[string]any{}, nil
}

func (f fakeUI) RenderRunning(taskID, pluginTitle string) {}

func (f fakeUI) RenderLoading(message string) {}

func (f fakeUI) WaitForEnter() error { return nil }

func TestRunTaskByIDNotFound(t *testing.T) {
	base := t.TempDir()
	repo := createRepoWithLocalConfig(t, base)
	chdirTo(t, repo)
	if err := RunTaskByID(fakeUI{}, "missing:task"); err == nil {
		t.Fatal("expected error for missing task")
	}
}

func TestRunTaskByIDExecutes(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping shell script test on windows")
	}
	base := t.TempDir()
	repo, specDir := createRepoWithLocalSpecsDir(t, base)
	outputFile := filepath.Join(base, "out.txt")
	script := filepath.Join(base, "run.sh")
	scriptBody := "#!/bin/sh\n" +
		"echo \"ok\" > \"$OUTPUT_FILE\"\n"
	if err := os.WriteFile(script, []byte(scriptBody), 0o755); err != nil {
		t.Fatal(err)
	}
	os.Setenv("OUTPUT_FILE", outputFile)
	defer os.Unsetenv("OUTPUT_FILE")

	spec := `{
  "schemaVersion": 1,
  "plugin": {"id": "p", "title": "P", "exec": "` + script + `"},
  "tasks": [{"name": "t", "title": "t"}]
}`
	writeSpecFile(t, specDir, "p.json", spec)
	chdirTo(t, repo)

	if err := RunTaskByID(fakeUI{}, "p:t"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(outputFile); err != nil {
		t.Fatalf("expected output file: %v", err)
	}
}

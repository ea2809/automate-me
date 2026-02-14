package app

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ea2809/automate-me/internal/core"
)

type cancelUI struct{}

func (c cancelUI) ClearScreen() {}

func (c cancelUI) SelectTask(tasks []core.TaskRecord, state SelectionState) (core.TaskRecord, SelectionState, error) {
	return core.TaskRecord{}, state, ErrUserCanceled
}

func (c cancelUI) PromptInputs(inputs []core.InputSpec, defaults map[string]any) (map[string]any, error) {
	return map[string]any{}, nil
}

func (c cancelUI) RenderRunning(taskID, pluginTitle string) {}

func (c cancelUI) RenderLoading(message string) {}

func (c cancelUI) WaitForEnter() error { return nil }

type sequenceUI struct {
	argsSequence []map[string]any
	index        int
}

func (s *sequenceUI) ClearScreen() {}

func (s *sequenceUI) SelectTask(tasks []core.TaskRecord, state SelectionState) (core.TaskRecord, SelectionState, error) {
	if s.index >= len(s.argsSequence) {
		return core.TaskRecord{}, state, ErrUserCanceled
	}
	return tasks[0], state, nil
}

func (s *sequenceUI) PromptInputs(inputs []core.InputSpec, defaults map[string]any) (map[string]any, error) {
	args := s.argsSequence[s.index]
	if defaults != nil {
		for key, value := range defaults {
			if _, ok := args[key]; !ok {
				args[key] = value
			}
		}
	}
	return args, nil
}

func (s *sequenceUI) RenderRunning(taskID, pluginTitle string) {}

func (s *sequenceUI) RenderLoading(message string) {}

func (s *sequenceUI) WaitForEnter() error {
	s.index++
	return nil
}

func TestRunInteractiveRespectsDefaults(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping shell script test on windows")
	}
	base := t.TempDir()
	repo, specDir := createRepoWithLocalSpecsDir(t, base)
	outputFile := filepath.Join(base, "out.txt")
	script := filepath.Join(base, "run.sh")
	scriptBody := "#!/bin/sh\n" +
		"cat > \"$OUTPUT_FILE\"\n"
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

	ui := &sequenceUI{
		argsSequence: []map[string]any{
			{"value": "first"},
			{"value": "second"},
		},
	}
	if err := RunInteractive(ui); err != ErrUserCanceled {
		t.Fatalf("expected cancel, got %v", err)
	}
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "" {
		t.Fatal("expected output to be written")
	}
	if string(data) == "first\n" {
		t.Fatal("expected second run to overwrite output")
	}
}

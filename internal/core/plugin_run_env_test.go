package core

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunPluginTaskSetsEnv(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping shell script test on windows")
	}
	base := t.TempDir()
	outputFile := filepath.Join(base, "env.txt")
	script := filepath.Join(base, "env.sh")
	content := "#!/bin/sh\n" +
		"echo \"$AUTOMATE_ME_REPO_ROOT|$AUTOMATE_ME_CWD|$AUTOMATE_ME_TASK_ID|$AUTOMATE_ME_PLUGIN_ID|$AUTOMATE_ME_TASK_NAME|$AUTOMATE_ME_SCOPE\" > \"$OUTPUT_FILE\"\n"
	if err := os.WriteFile(script, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	os.Setenv("OUTPUT_FILE", outputFile)
	defer os.Unsetenv("OUTPUT_FILE")

	repoRoot := filepath.Join(base, "repo")
	cwd := filepath.Join(repoRoot, "sub")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatal(err)
	}

	task := TaskRecord{
		PluginID:   "p",
		Task:       TaskSpec{Name: "t"},
		PluginPath: script,
		Scope:      ScopeGlobal,
		DirectExec: true,
	}

	if err := RunPluginTask(task, repoRoot, cwd, map[string]any{}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(strings.TrimSpace(string(data)), "|")
	if len(parts) != 6 {
		t.Fatalf("expected 6 env parts, got %d", len(parts))
	}
	if parts[0] != repoRoot || parts[1] != cwd || parts[2] != "p:t" || parts[3] != "p" || parts[4] != "t" || parts[5] != string(ScopeGlobal) {
		t.Fatalf("unexpected env values: %v", parts)
	}
}

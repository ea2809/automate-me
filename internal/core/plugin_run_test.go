package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRunPluginTask(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping shell script test on windows")
	}
	base := t.TempDir()
	outputFile := filepath.Join(base, "out.json")
	outputEnvFile := filepath.Join(base, "env.txt")
	script := filepath.Join(base, "plugin.sh")
	content := "#!/bin/sh\n" +
		"if [ \"$1\" != \"run\" ]; then exit 2; fi\n" +
		"if [ \"$2\" != \"test\" ]; then exit 3; fi\n" +
		"cat > \"$OUTPUT_FILE\"\n" +
		"echo \"$AUTOMATE_ME_TASK_ID\" > \"$OUTPUT_ENV_FILE\"\n"
	if err := os.WriteFile(script, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	os.Setenv("OUTPUT_FILE", outputFile)
	os.Setenv("OUTPUT_ENV_FILE", outputEnvFile)
	defer os.Unsetenv("OUTPUT_FILE")
	defer os.Unsetenv("OUTPUT_ENV_FILE")

	task := TaskRecord{
		PluginID:   "plug",
		Task:       TaskSpec{Name: "test"},
		PluginPath: script,
		Scope:      ScopeLocal,
	}
	args := map[string]any{"pattern": "foo"}
	repoRoot := filepath.Join(base, "repo")
	cwd := filepath.Join(repoRoot, "sub")
	if err := RunPluginTask(task, repoRoot, cwd, args); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatal(err)
	}
	ctx := payload["ctx"].(map[string]any)
	if ctx["repoRoot"] != repoRoot {
		t.Fatalf("expected repoRoot %s, got %v", repoRoot, ctx["repoRoot"])
	}

	envData, err := os.ReadFile(outputEnvFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(envData) != "plug:test\n" {
		t.Fatalf("expected task id plug:test, got %s", string(envData))
	}
}

package core

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunPluginTaskDirectExec(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping shell script test on windows")
	}
	base := t.TempDir()
	outputArgs := filepath.Join(base, "args.txt")
	outputPwd := filepath.Join(base, "pwd.txt")
	script := filepath.Join(base, "direct.sh")
	content := "#!/bin/sh\n" +
		"echo \"$#\" > \"$OUTPUT_ARGS\"\n" +
		"pwd > \"$OUTPUT_PWD\"\n"
	if err := os.WriteFile(script, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	os.Setenv("OUTPUT_ARGS", outputArgs)
	os.Setenv("OUTPUT_PWD", outputPwd)
	defer os.Unsetenv("OUTPUT_ARGS")
	defer os.Unsetenv("OUTPUT_PWD")

	repoRoot := filepath.Join(base, "repo")
	if err := os.MkdirAll(repoRoot, 0o755); err != nil {
		t.Fatal(err)
	}

	task := TaskRecord{
		PluginID:   "p",
		Task:       TaskSpec{Name: "t"},
		PluginPath: script,
		Scope:      ScopeLocal,
		DirectExec: true,
	}

	if err := RunPluginTask(task, repoRoot, repoRoot, map[string]any{}); err != nil {
		t.Fatal(err)
	}

	argsData, err := os.ReadFile(outputArgs)
	if err != nil {
		t.Fatal(err)
	}
	if string(argsData) != "0\n" {
		t.Fatalf("expected no args, got %s", string(argsData))
	}

	pwdData, err := os.ReadFile(outputPwd)
	if err != nil {
		t.Fatal(err)
	}
	gotPwd := strings.TrimSpace(string(pwdData))
	resolvedRepo, err := filepath.EvalSymlinks(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	resolvedPwd, err := filepath.EvalSymlinks(gotPwd)
	if err != nil {
		t.Fatal(err)
	}
	if resolvedPwd != resolvedRepo {
		t.Fatalf("expected cwd %s, got %s", resolvedRepo, resolvedPwd)
	}
}

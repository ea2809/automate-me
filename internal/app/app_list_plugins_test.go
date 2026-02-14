package app

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestListTasksOutputs(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping shell script test on windows")
	}
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	specDir := filepath.Join(repo, ".automate-me", "specs")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	spec := `{
  "schemaVersion": 1,
  "plugin": {"id": "p", "title": "P", "exec": "/bin/echo"},
  "tasks": [{"name": "t", "title": "Title", "description": "Desc"}]
}`
	if err := os.WriteFile(filepath.Join(specDir, "p.json"), []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := ListTasksWithWriter(&buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	if !strings.Contains(output, "p:t\tTitle\tDesc") {
		t.Fatalf("unexpected output: %s", output)
	}
}

func TestListPluginsOutputs(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping shell script test on windows")
	}
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	specDir := filepath.Join(repo, ".automate-me", "specs")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	spec := `{
  "schemaVersion": 1,
  "plugin": {"id": "p", "title": "P", "exec": "/bin/echo"},
  "tasks": []
}`
	if err := os.WriteFile(filepath.Join(specDir, "p.json"), []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := ListPluginsWithWriter(&buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	if !strings.Contains(output, "p\tlocal") && !strings.Contains(output, "p\tglobal") {
		t.Fatalf("unexpected output: %s", output)
	}
}

func TestListPluginsEmpty(t *testing.T) {
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	if err := os.MkdirAll(filepath.Join(repo, ".automate-me"), 0o755); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := ListPluginsWithWriter(&buf); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(buf.String()) != "no plugins found" {
		t.Fatalf("unexpected output: %s", buf.String())
	}
}

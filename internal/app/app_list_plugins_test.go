package app

import (
	"bytes"
	"runtime"
	"strings"
	"testing"
)

func TestListTasksOutputs(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping shell script test on windows")
	}
	base := t.TempDir()
	repo, specDir := createRepoWithLocalSpecsDir(t, base)
	spec := `{
  "schemaVersion": 1,
  "plugin": {"id": "p", "title": "P", "exec": "/bin/echo"},
  "tasks": [{"name": "t", "title": "Title", "description": "Desc"}]
}`
	writeSpecFile(t, specDir, "p.json", spec)
	chdirTo(t, repo)

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
	repo, specDir := createRepoWithLocalSpecsDir(t, base)
	spec := `{
  "schemaVersion": 1,
  "plugin": {"id": "p", "title": "P", "exec": "/bin/echo"},
  "tasks": []
}`
	writeSpecFile(t, specDir, "p.json", spec)
	chdirTo(t, repo)

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
	repo := createRepoWithLocalConfig(t, base)
	chdirTo(t, repo)

	var buf bytes.Buffer
	if err := ListPluginsWithWriter(&buf); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(buf.String()) != "no plugins found" {
		t.Fatalf("unexpected output: %s", buf.String())
	}
}

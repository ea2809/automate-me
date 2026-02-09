package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImportSpecFileRequiresExec(t *testing.T) {
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(base, "spec.json")
	data := `{
  "schemaVersion": 1,
  "plugin": {"id": "p"},
  "tasks": []
}`
	if err := os.WriteFile(src, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportSpecFile(src, repo, ScopeLocal); err == nil {
		t.Fatal("expected error for missing plugin.exec")
	}
}

func TestImportSpecFileLocalNoRepo(t *testing.T) {
	base := t.TempDir()
	src := filepath.Join(base, "spec.json")
	data := `{
  "schemaVersion": 1,
  "plugin": {"id": "p", "exec": "/bin/echo"},
  "tasks": []
}`
	if err := os.WriteFile(src, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportSpecFile(src, "", ScopeLocal); err == nil {
		t.Fatal("expected error for local import without repo root")
	}
}

func TestLoadSpecsExecModeProtocol(t *testing.T) {
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	localDir := filepath.Join(repo, ".automate-me", "specs")
	if err := os.MkdirAll(localDir, 0o755); err != nil {
		t.Fatal(err)
	}
	spec := `{
  "schemaVersion": 1,
  "plugin": {"id": "p", "exec": "/bin/echo", "execMode": "protocol"},
  "tasks": [{"name": "t", "title": "t"}]
}`
	path := filepath.Join(localDir, "p.json")
	if err := os.WriteFile(path, []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}

	specs, err := loadSpecs(repo)
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 1 {
		t.Fatalf("expected 1 spec, got %d", len(specs))
	}
	if specs[0].DirectExec {
		t.Fatal("expected DirectExec false for execMode protocol")
	}
}

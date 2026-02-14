package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImportSpecFile(t *testing.T) {
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}

	spec := `{
  "schemaVersion": 1,
  "plugin": {"id": "repo", "title": "Repo", "exec": "/bin/echo"},
  "tasks": [{"name": "test", "title": "Run tests"}]
}`
	src := filepath.Join(base, "spec.json")
	if err := os.WriteFile(src, []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}

	dest, err := ImportSpecFile(src, repo, ScopeLocal)
	if err != nil {
		t.Fatal(err)
	}
	if dest == "" {
		t.Fatal("expected destination path")
	}
	if _, err := os.Stat(dest); err != nil {
		t.Fatalf("expected spec written: %v", err)
	}
	paths := newPathConfig(repo)
	localSpecsDir, err := paths.localSpecs()
	if err != nil {
		t.Fatal(err)
	}
	expectedDest := filepath.Join(localSpecsDir, "repo.json")
	if dest != expectedDest {
		t.Fatalf("expected destination %s, got %s", expectedDest, dest)
	}
}

package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSpecsSkipsInvalid(t *testing.T) {
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	paths := newPathConfig(repo)
	specDir, err := paths.localSpecs()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}

	badJSON := filepath.Join(specDir, "bad.json")
	if err := os.WriteFile(badJSON, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	missingExec := filepath.Join(specDir, "noexec.json")
	if err := os.WriteFile(missingExec, []byte(`{"schemaVersion":1,"plugin":{"id":"p"},"tasks":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	records, err := loadSpecs(repo)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 0 {
		t.Fatalf("expected 0 specs, got %d", len(records))
	}
}

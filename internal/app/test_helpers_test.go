package app

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	testLocalConfigDirName = ".automate-me"
	testSpecsDirName       = "specs"
)

func createRepoWithLocalConfig(t *testing.T, base string) string {
	t.Helper()
	repo := filepath.Join(base, "repo")
	if err := os.MkdirAll(filepath.Join(repo, testLocalConfigDirName), 0o755); err != nil {
		t.Fatal(err)
	}
	return repo
}

func createRepoWithLocalSpecsDir(t *testing.T, base string) (string, string) {
	t.Helper()
	repo := filepath.Join(base, "repo")
	specDir := filepath.Join(repo, testLocalConfigDirName, testSpecsDirName)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	return repo, specDir
}

func writeSpecFile(t *testing.T, specDir, filename, spec string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(specDir, filename), []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}
}

func chdirTo(t *testing.T, dir string) {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
}

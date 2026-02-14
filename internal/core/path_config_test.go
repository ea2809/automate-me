package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathConfigLocalAndGlobal(t *testing.T) {
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	paths := newPathConfig(repo)

	localRoot, err := paths.localRoot()
	if err != nil {
		t.Fatal(err)
	}
	if localRoot != filepath.Join(repo, localConfigDirName) {
		t.Fatalf("unexpected local root: %s", localRoot)
	}

	localBin, err := paths.localBin()
	if err != nil {
		t.Fatal(err)
	}
	if localBin != filepath.Join(repo, localConfigDirName, binDirName) {
		t.Fatalf("unexpected local bin: %s", localBin)
	}

	localSpecs, err := paths.localSpecs()
	if err != nil {
		t.Fatal(err)
	}
	if localSpecs != filepath.Join(repo, localConfigDirName, specsDirName) {
		t.Fatalf("unexpected local specs: %s", localSpecs)
	}

	configDir := filepath.Join(base, "config")
	os.Setenv("XDG_CONFIG_HOME", configDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	globalRoot, err := paths.globalRoot()
	if err != nil {
		t.Fatal(err)
	}
	if globalRoot != filepath.Join(configDir, globalConfigDirName) {
		t.Fatalf("unexpected global root: %s", globalRoot)
	}

	globalBin, err := paths.globalBin()
	if err != nil {
		t.Fatal(err)
	}
	if globalBin != filepath.Join(configDir, globalConfigDirName, binDirName) {
		t.Fatalf("unexpected global bin: %s", globalBin)
	}

	globalSpecs, err := paths.globalSpecs()
	if err != nil {
		t.Fatal(err)
	}
	if globalSpecs != filepath.Join(configDir, globalConfigDirName, specsDirName) {
		t.Fatalf("unexpected global specs: %s", globalSpecs)
	}
}

func TestPathConfigLocalPathRequiresRepoRoot(t *testing.T) {
	paths := newPathConfig("")
	if _, err := paths.localRoot(); err == nil {
		t.Fatal("expected error for local root without repo")
	}
	if _, err := paths.localBin(); err == nil {
		t.Fatal("expected error for local bin without repo")
	}
	if _, err := paths.localSpecs(); err == nil {
		t.Fatal("expected error for local specs without repo")
	}
}

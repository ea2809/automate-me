package core

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestFindRepoRoot(t *testing.T) {
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	sub := filepath.Join(repo, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(repo, ".automate-me"), 0o755); err != nil {
		t.Fatal(err)
	}

	root, ok, err := FindRepoRoot(sub)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || root != repo {
		t.Fatalf("expected repo root %s, got %s", repo, root)
	}
}

func TestDiscoverPluginCandidates(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping exec bit test on windows")
	}
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	localBin := filepath.Join(repo, ".automate-me", "bin")
	globalConfig := filepath.Join(base, "config")
	globalBin := filepath.Join(globalConfig, "automate-me", "bin")
	if err := os.MkdirAll(localBin, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(globalBin, 0o755); err != nil {
		t.Fatal(err)
	}

	localPlugin := filepath.Join(localBin, "plugin-local")
	globalPlugin := filepath.Join(globalBin, "plugin-global")
	if err := os.WriteFile(localPlugin, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(globalPlugin, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	os.Setenv("XDG_CONFIG_HOME", globalConfig)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	candidates, err := discoverPluginCandidates(repo)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(candidates))
	}
}
